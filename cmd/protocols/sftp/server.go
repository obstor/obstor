/*
 * PGG Obstor, (C) 2021-2026 PGG, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sftp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	obstor "github.com/obstor/obstor/cmd"
	"github.com/obstor/obstor/cmd/logger"
	xsftp "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func init() {
	obstor.GlobalSFTPStartFn = Start
}

const (
	sftpMaxConnections = 256
	sftpIdleTimeout    = 10 * time.Minute
)

var sftpActiveConns int64

// Launch the SFTP server.
func Start() {
	cfg := obstor.GlobalSFTPConfig

	sshConfig := &ssh.ServerConfig{
		PasswordCallback: sftpPasswordCallback,
		MaxAuthTries:     3,
		ServerVersion:    "SSH-2.0-Obstor",
		BannerCallback:   nil,
	}

	sshConfig.KeyExchanges = []string{
		"curve25519-sha256",
		"curve25519-sha256@libssh.org",
		"ecdh-sha2-nistp256",
		"ecdh-sha2-nistp384",
	}
	sshConfig.Ciphers = []string{
		"chacha20-poly1305@openssh.com",
		"aes256-gcm@openssh.com",
		"aes128-gcm@openssh.com",
	}
	sshConfig.MACs = []string{
		"hmac-sha2-256-etm@openssh.com",
		"hmac-sha2-256",
	}

	hostKey, err := loadOrGenerateHostKey(cfg.HostKeyPath)
	if err != nil {
		logger.Fatal(err, "Unable to load/generate SFTP host key")
	}
	sshConfig.AddHostKey(hostKey)

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		logger.Fatal(err, "Unable to start SFTP server on %s", cfg.Address)
	}

	logger.Info("SFTP server listening on %s", cfg.Address)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				logger.LogIf(obstor.GlobalContext, fmt.Errorf("sftp accept error: %w", err))
				continue
			}

			if atomic.LoadInt64(&sftpActiveConns) >= sftpMaxConnections {
				_ = conn.Close()
				continue
			}
			atomic.AddInt64(&sftpActiveConns, 1)
			go func() {
				defer atomic.AddInt64(&sftpActiveConns, -1)
				handleSFTPConnection(conn, sshConfig)
			}()
		}
	}()
}

func sftpPasswordCallback(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	accessKey := c.User()
	secretKey := string(pass)

	activeCred := obstor.GetActiveCred()
	keyMatch := subtle.ConstantTimeCompare([]byte(accessKey), []byte(activeCred.AccessKey)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(secretKey), []byte(activeCred.SecretKey)) == 1
	if keyMatch && passMatch {
		return &ssh.Permissions{
			Extensions: map[string]string{
				"accessKey": accessKey,
			},
		}, nil
	}

	if obstor.GetIAMSys() != nil {
		cred, ok := obstor.GetIAMSys().GetUser(accessKey)
		if ok && cred.IsValid() &&
			subtle.ConstantTimeCompare([]byte(secretKey), []byte(cred.SecretKey)) == 1 {
			return &ssh.Permissions{
				Extensions: map[string]string{
					"accessKey": accessKey,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid credentials")
}

func handleSFTPConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		logger.LogIf(obstor.GlobalContext, fmt.Errorf("sftp ssh handshake error from %s: %w", conn.RemoteAddr(), err))
		return
	}
	defer func() { _ = sshConn.Close() }()

	_ = conn.SetDeadline(time.Time{})

	logger.Info("SFTP connection from %s (user: %s)", conn.RemoteAddr(), sshConn.User())

	done := make(chan struct{})
	defer close(done)
	go func() {
		timer := time.NewTimer(sftpIdleTimeout)
		defer timer.Stop()
		select {
		case <-timer.C:
			logger.Info("SFTP idle timeout for user %s from %s", sshConn.User(), conn.RemoteAddr())
			_ = sshConn.Close()
		case <-done:
		}
	}()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			logger.LogIf(obstor.GlobalContext, fmt.Errorf("sftp channel accept error: %w", err))
			return
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := req.Type == "subsystem" && len(req.Payload) > 4 && string(req.Payload[4:]) == "sftp"
				_ = req.Reply(ok, nil)
			}
		}(requests)

		driver := newSFTPDriver(sshConn.User())
		server := xsftp.NewRequestServer(channel, xsftp.Handlers{
			FileGet:  driver,
			FilePut:  driver,
			FileCmd:  driver,
			FileList: driver,
		})

		if err := server.Serve(); err != nil {
			if err != io.EOF {
				logger.LogIf(obstor.GlobalContext, fmt.Errorf("sftp server error for user %s: %w", sshConn.User(), err))
			}
		}
		_ = server.Close()
	}
}

func loadOrGenerateHostKey(hostKeyPath string) (ssh.Signer, error) {
	if hostKeyPath != "" {
		keyBytes, err := os.ReadFile(hostKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read host key %s: %w", hostKeyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse host key %s: %w", hostKeyPath, err)
		}
		return signer, nil
	}

	// Auto-generate and persist a host key.
	autoKeyPath := filepath.Join(obstor.GetCertsDir().Get(), "sftp_host_key")
	if keyBytes, err := os.ReadFile(autoKeyPath); err == nil {
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err == nil {
			return signer, nil
		}
	}

	logger.Info("Generating SFTP host key at %s", autoKeyPath)
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate host key: %w", err)
	}

	keyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	if err := os.WriteFile(autoKeyPath, keyBytes, 0600); err != nil {
		return nil, fmt.Errorf("failed to save host key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated host key: %w", err)
	}
	return signer, nil
}
