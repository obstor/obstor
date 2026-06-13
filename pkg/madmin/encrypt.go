/*
 * MinIO Cloud Storage, (C) 2018-2021 MinIO, Inc.
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
 *
 */

package madmin

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"github.com/obstor/obstor/pkg/argon2"
	"github.com/obstor/obstor/pkg/fips"
	sio "github.com/obstor/sio"
	"golang.org/x/crypto/pbkdf2"
)

// EncryptData encrypts the data with an unique key
// derived from password using the Argon2id PBKDF.
//
// The returned ciphertext data consists of:
//
//	salt | AEAD ID | nonce | encrypted data
//	 32      1         8      ~ len(data)
func EncryptData(password string, data []byte) ([]byte, error) {
	salt := mustRandom(32)

	var (
		id  byte
		key []byte
	)
	if fips.Enabled() {
		key = pbkdf2.Key([]byte(password), salt, pbkdf2Cost, 32, sha256.New)
		id = pbkdf2AESGCM
	} else {
		key = argon2.IDKey([]byte(password), salt, argon2idTime, argon2idMemory, argon2idThreads, 32)
		id = argon2idAESGCM
	}

	nonce := mustRandom(8)
	var nonce12 [12]byte
	copy(nonce12[:], nonce)

	var ciphertext bytes.Buffer
	ciphertext.Write(salt)
	ciphertext.WriteByte(id)
	ciphertext.Write(nonce)

	w, err := sio.EncryptWriter(&ciphertext, sio.Config{
		Key:          key,
		CipherSuites: []byte{sio.AES_GCM},
		Nonce:        &nonce12,
	})
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(data); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return ciphertext.Bytes(), nil
}

// ErrMaliciousData indicates that the stream cannot be
// decrypted by provided credentials.
var ErrMaliciousData = errors.New("madmin: data is not authentic")

// DecryptData decrypts the data with the key derived
// from the salt (part of data) and the password using
// the PBKDF used in EncryptData. DecryptData returns
// the decrypted plaintext on success.
//
// The data must be a valid ciphertext produced by
// EncryptData. Otherwise, the decryption will fail.
func DecryptData(password string, data io.Reader) ([]byte, error) {
	var (
		salt  [32]byte
		id    [1]byte
		nonce [8]byte // This depends on the AEAD but both used ciphers have the same nonce length.
	)

	if _, err := io.ReadFull(data, salt[:]); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(data, id[:]); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(data, nonce[:]); err != nil {
		return nil, err
	}

	var key []byte
	var cipher byte
	switch id[0] {
	case argon2idAESGCM:
		key = argon2.IDKey([]byte(password), salt[:], argon2idTime, argon2idMemory, argon2idThreads, 32)
		cipher = sio.AES_GCM
	case argon2idChaCHa20Poly1305:
		key = argon2.IDKey([]byte(password), salt[:], argon2idTime, argon2idMemory, argon2idThreads, 32)
		cipher = sio.CHACHA20_POLY1305
	case pbkdf2AESGCM:
		key = pbkdf2.Key([]byte(password), salt[:], pbkdf2Cost, 32, sha256.New)
		cipher = sio.AES_GCM
	default:
		return nil, errors.New("madmin: invalid encryption algorithm ID")
	}

	var nonce12 [12]byte
	copy(nonce12[:], nonce[:])
	decReader, err := sio.DecryptReader(data, sio.Config{
		Key:          key,
		CipherSuites: []byte{cipher},
		Nonce:        &nonce12,
	})
	if err != nil {
		return nil, err
	}
	return io.ReadAll(decReader)
}

func mustRandom(n int) []byte {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}
	return b
}

const (
	argon2idAESGCM           = 0x00
	argon2idChaCHa20Poly1305 = 0x01
	pbkdf2AESGCM             = 0x02
)

const (
	argon2idTime    = 1
	argon2idMemory  = 64 * 1024
	argon2idThreads = 4
	pbkdf2Cost      = 8192
)
