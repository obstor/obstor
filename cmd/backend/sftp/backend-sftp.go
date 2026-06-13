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
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/obstor/obstor-go/v7/pkg/s3utils"
	obstor "github.com/obstor/obstor/cmd"
	"github.com/obstor/obstor/cmd/logger"
	"github.com/obstor/obstor/pkg/auth"
	"github.com/obstor/obstor/pkg/env"
	"github.com/obstor/obstor/pkg/madmin"
	xsftp "github.com/pkg/sftp"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	sftpSeparator = obstor.SlashSeparator
)

func sftpBackendHostKeyCallback() (ssh.HostKeyCallback, error) {
	if khPath := env.Get("OBSTOR_BACKEND_SFTP_KNOWN_HOSTS", ""); khPath != "" {
		cb, err := knownhosts.New(khPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load OBSTOR_BACKEND_SFTP_KNOWN_HOSTS %s: %w", khPath, err)
		}
		return cb, nil
	}

	if hostKey := env.Get("OBSTOR_BACKEND_SFTP_HOST_KEY", ""); hostKey != "" {
		pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostKey))
		if err != nil {
			return nil, fmt.Errorf("unable to parse OBSTOR_BACKEND_SFTP_HOST_KEY: %w", err)
		}
		return ssh.FixedHostKey(pub), nil
	}

	if skip := env.Get("OBSTOR_BACKEND_SFTP_INSECURE_SKIP_HOST_KEY", ""); skip == "on" || skip == "true" {
		logger.Info("WARNING: SFTP backend host key verification is disabled via OBSTOR_BACKEND_SFTP_INSECURE_SKIP_HOST_KEY; the backend connection is vulnerable to man-in-the-middle attacks")
		return ssh.InsecureIgnoreHostKey(), nil
	}

	return nil, fmt.Errorf("SFTP backend host key verification is required: set OBSTOR_BACKEND_SFTP_KNOWN_HOSTS or OBSTOR_BACKEND_SFTP_HOST_KEY, or set OBSTOR_BACKEND_SFTP_INSECURE_SKIP_HOST_KEY=on to disable it (not recommended)")
}

func init() {
	const sftpBackendTemplate = `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} {{if .VisibleFlags}}[FLAGS]{{end}} SFTP-ENDPOINT
{{if .VisibleFlags}}
FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}
SFTP-ENDPOINT:
  SFTP server endpoint e.g. sftp-server:22/data

ENVIRONMENT VARIABLES:
  OBSTOR_BACKEND_SFTP_USER:     SSH username
  OBSTOR_BACKEND_SFTP_PASSWORD: SSH password (if not using key auth)
  OBSTOR_BACKEND_SFTP_KEY:      Path to SSH private key file

EXAMPLES:
  1. Start obstor backend server for SFTP backend with password auth
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_USER{{.AssignmentOperator}}accesskey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_PASSWORD{{.AssignmentOperator}}secretkey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_BACKEND_SFTP_USER{{.AssignmentOperator}}sftpuser
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_BACKEND_SFTP_PASSWORD{{.AssignmentOperator}}sftppassword
     {{.Prompt}} {{.HelpName}} sftp-server:22/data

  2. Start obstor backend server for SFTP backend with key auth
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_USER{{.AssignmentOperator}}accesskey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_PASSWORD{{.AssignmentOperator}}secretkey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_BACKEND_SFTP_USER{{.AssignmentOperator}}sftpuser
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_BACKEND_SFTP_KEY{{.AssignmentOperator}}/path/to/id_rsa
     {{.Prompt}} {{.HelpName}} sftp-server:22/data
`

	_ = obstor.RegisterBackendCommand(cli.Command{
		Name:               obstor.SFTPBackend,
		Usage:              "SSH File Transfer Protocol (SFTP)",
		Action:             sftpBackendMain,
		CustomHelpTemplate: sftpBackendTemplate,
		HideHelp:           true,
	})
}

// Handler for 'obstor backend sftp' command line.
func sftpBackendMain(ctx *cli.Context) {
	if ctx.Args().First() == "help" || !ctx.Args().Present() {
		cli.ShowCommandHelpAndExit(ctx, obstor.SFTPBackend, 1)
	}

	obstor.StartBackend(ctx, &SFTP{endpoint: ctx.Args().First()})
}

// SFTP implements Backend.
type SFTP struct {
	endpoint string
}

// Name implements Backend interface.
func (g *SFTP) Name() string {
	return obstor.SFTPBackend
}

// SFTP backend is production-ready
func (g *SFTP) Production() bool {
	return true
}

// Parses an endpoint like "host:port/path"
func parseSFTPEndpoint(endpoint string) (host, port, basePath string) {
	basePath = "/"

	// Split off the path
	slashIdx := strings.Index(endpoint, "/")
	hostPort := endpoint
	if slashIdx >= 0 {
		hostPort = endpoint[:slashIdx]
		basePath = endpoint[slashIdx:]
	}

	// Split host and port
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		// Default to 22
		host = hostPort
		port = "22"
	}

	return host, port, basePath
}

// NewBackendLayer returns sftp backend layer.
func (g *SFTP) NewBackendLayer(creds auth.Credentials) (obstor.ObjectLayer, error) {
	host, port, basePath := parseSFTPEndpoint(g.endpoint)

	sshUser := env.Get("OBSTOR_BACKEND_SFTP_USER", "")
	sshPassword := env.Get("OBSTOR_BACKEND_SFTP_PASSWORD", "")
	sshKeyPath := env.Get("OBSTOR_BACKEND_SFTP_KEY", "")

	if sshUser == "" {
		return nil, fmt.Errorf("OBSTOR_BACKEND_SFTP_USER must be set")
	}

	hostKeyCallback, err := sftpBackendHostKeyCallback()
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            sshUser,
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
	}

	if sshKeyPath != "" {
		keyBytes, err := os.ReadFile(sshKeyPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read SSH private key %s: %w", sshKeyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("unable to parse SSH private key %s: %w", sshKeyPath, err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if sshPassword != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(sshPassword)}
	} else {
		return nil, fmt.Errorf("either OBSTOR_BACKEND_SFTP_PASSWORD or OBSTOR_BACKEND_SFTP_KEY must be set")
	}

	addr := net.JoinHostPort(host, port)
	sshConn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to SFTP server %s: %w", addr, err)
	}

	client, err := xsftp.NewClient(sshConn)
	if err != nil {
		_ = sshConn.Close()
		return nil, fmt.Errorf("unable to initialize SFTP client: %w", err)
	}

	// Create tmp directory for temporary uploads.
	tmpPath := obstor.PathJoin(basePath, sftpSeparator, obstorMetaTmpBucket)
	if err = client.MkdirAll(tmpPath); err != nil {
		_ = client.Close()
		_ = sshConn.Close()
		return nil, fmt.Errorf("unable to create tmp directory %s: %w", tmpPath, err)
	}

	return &sftpObjects{
		clnt:     client,
		sshConn:  sshConn,
		subPath:  basePath,
		listPool: obstor.NewTreeWalkPool(time.Minute * 30),
	}, nil
}

// sftpObjects implements backend for SFTP-compatible storage servers.
type sftpObjects struct {
	obstor.BackendUnsupported
	clnt     *xsftp.Client
	sshConn  *ssh.Client
	subPath  string
	listPool *obstor.TreeWalkPool
}

func sftpToObjectErr(ctx context.Context, err error, params ...string) error {
	if err == nil {
		return nil
	}
	bucket := ""
	object := ""
	uploadID := ""
	switch len(params) {
	case 3:
		uploadID = params[2]
		fallthrough
	case 2:
		object = params[1]
		fallthrough
	case 1:
		bucket = params[0]
	}

	switch {
	case os.IsNotExist(err):
		if uploadID != "" {
			return obstor.InvalidUploadID{UploadID: uploadID}
		}
		if object != "" {
			return obstor.ObjectNotFound{Bucket: bucket, Object: object}
		}
		return obstor.BucketNotFound{Bucket: bucket}
	case os.IsExist(err):
		if object != "" {
			return obstor.PrefixAccessDenied{Bucket: bucket, Object: object}
		}
		return obstor.BucketAlreadyOwnedByYou{Bucket: bucket}
	case os.IsPermission(err):
		return obstor.PrefixAccessDenied{Bucket: bucket, Object: object}
	default:
		logger.LogIf(ctx, err)
		return err
	}
}

// sftpIsValidBucketName verifies whether a bucket name is valid.
func sftpIsValidBucketName(bucket string) bool {
	return s3utils.CheckValidBucketNameStrict(bucket) == nil
}

func (n *sftpObjects) sftpPathJoin(args ...string) string {
	return obstor.PathJoin(append([]string{n.subPath, sftpSeparator}, args...)...)
}

func (n *sftpObjects) Shutdown(ctx context.Context) error {
	_ = n.clnt.Close()
	return n.sshConn.Close()
}

func (n *sftpObjects) LocalStorageInfo(ctx context.Context) (si obstor.StorageInfo, errs []error) {
	return n.StorageInfo(ctx)
}

func (n *sftpObjects) StorageInfo(ctx context.Context) (si obstor.StorageInfo, errs []error) {
	// SFTP has no StatFs equivalent. Probe connectivity instead.
	_, err := n.clnt.Stat(n.sftpPathJoin())
	si.Backend.Type = madmin.Gateway
	si.Backend.GatewayOnline = err == nil
	if err != nil {
		return si, []error{err}
	}
	return si, nil
}

func (n *sftpObjects) MakeBucketWithLocation(ctx context.Context, bucket string, opts obstor.BucketOptions) error {
	if opts.LockEnabled || opts.VersioningEnabled {
		return obstor.NotImplemented{}
	}
	if !sftpIsValidBucketName(bucket) {
		return obstor.BucketNameInvalid{Bucket: bucket}
	}
	return sftpToObjectErr(ctx, n.clnt.Mkdir(n.sftpPathJoin(bucket)), bucket)
}

func (n *sftpObjects) GetBucketInfo(ctx context.Context, bucket string) (bi obstor.BucketInfo, err error) {
	fi, err := n.clnt.Stat(n.sftpPathJoin(bucket))
	if err != nil {
		return bi, sftpToObjectErr(ctx, err, bucket)
	}
	return obstor.BucketInfo{
		Name:    bucket,
		Created: fi.ModTime(),
	}, nil
}

func (n *sftpObjects) ListBuckets(ctx context.Context) (buckets []obstor.BucketInfo, err error) {
	entries, err := n.clnt.ReadDir(n.sftpPathJoin())
	if err != nil {
		logger.LogIf(ctx, err)
		return nil, sftpToObjectErr(ctx, err)
	}

	for _, entry := range entries {
		// Ignore all reserved bucket names and invalid bucket names.
		if isReservedOrInvalidBucket(entry.Name(), false) {
			continue
		}
		if !entry.IsDir() {
			continue
		}
		buckets = append(buckets, obstor.BucketInfo{
			Name:    entry.Name(),
			Created: entry.ModTime(),
		})
	}

	sort.Sort(byBucketName(buckets))
	return buckets, nil
}

func (n *sftpObjects) DeleteBucket(ctx context.Context, bucket string, forceDelete bool) error {
	if !sftpIsValidBucketName(bucket) {
		return obstor.BucketNameInvalid{Bucket: bucket}
	}
	bucketPath := n.sftpPathJoin(bucket)
	if forceDelete {
		// Walk and remove all contents recursively.
		walker := n.clnt.Walk(bucketPath)
		var files []string
		var dirs []string
		for walker.Step() {
			if walker.Err() != nil {
				continue
			}
			if walker.Stat().IsDir() {
				dirs = append(dirs, walker.Path())
			} else {
				files = append(files, walker.Path())
			}
		}
		// Remove files first, then directories in reverse order.
		for _, f := range files {
			_ = n.clnt.Remove(f)
		}
		for i := len(dirs) - 1; i >= 0; i-- {
			_ = n.clnt.RemoveDirectory(dirs[i])
		}
		return nil
	}
	return sftpToObjectErr(ctx, n.clnt.RemoveDirectory(bucketPath), bucket)
}

func (n *sftpObjects) isLeafDir(bucket, leafPath string) bool {
	return n.isObjectDir(context.Background(), bucket, leafPath)
}

func (n *sftpObjects) isLeaf(bucket, leafPath string) bool {
	return !strings.HasSuffix(leafPath, sftpSeparator)
}

func (n *sftpObjects) listDirFactory() obstor.ListDirFunc {
	listDir := func(bucket, prefixDir, prefixEntry string) (emptyDir bool, entries []string, delayIsLeaf bool) {
		fis, err := n.clnt.ReadDir(n.sftpPathJoin(bucket, prefixDir))
		if err != nil {
			if os.IsNotExist(err) {
				err = nil
			}
			logger.LogIf(obstor.GlobalContext, err)
			return
		}
		if len(fis) == 0 {
			return true, nil, false
		}
		for _, fi := range fis {
			if fi.IsDir() {
				entries = append(entries, fi.Name()+sftpSeparator)
			} else {
				entries = append(entries, fi.Name())
			}
		}
		entries, delayIsLeaf = obstor.FilterListEntries(bucket, prefixDir, entries, prefixEntry, n.isLeaf)
		return false, entries, delayIsLeaf
	}
	return listDir
}

// ListObjects lists all blobs in SFTP bucket filtered by prefix.
func (n *sftpObjects) ListObjects(ctx context.Context, bucket, prefix, marker, delimiter string, maxKeys int) (loi obstor.ListObjectsInfo, err error) {
	fileInfos := make(map[string]os.FileInfo)
	targetPath := n.sftpPathJoin(bucket, prefix)

	var targetFileInfo os.FileInfo
	if targetFileInfo, err = n.populateDirectoryListing(targetPath, fileInfos); err != nil {
		return loi, sftpToObjectErr(ctx, err, bucket)
	}

	// If the user is trying to list a single file, return it directly.
	if !targetFileInfo.IsDir() {
		return obstor.ListObjectsInfo{
			IsTruncated: false,
			NextMarker:  "",
			Objects: []obstor.ObjectInfo{
				fileInfoToObjectInfo(bucket, prefix, targetFileInfo),
			},
			Prefixes: []string{},
		}, nil
	}

	getObjectInfo := func(ctx context.Context, bucket, entry string) (obstor.ObjectInfo, error) {
		filePath := path.Clean(n.sftpPathJoin(bucket, entry))
		fi, ok := fileInfos[filePath]

		if !ok {
			parentPath := path.Dir(filePath)
			if _, err := n.populateDirectoryListing(parentPath, fileInfos); err != nil {
				return obstor.ObjectInfo{}, sftpToObjectErr(ctx, err, bucket)
			}
			fi, ok = fileInfos[filePath]
			if !ok {
				err = fmt.Errorf("could not get FileInfo for path '%s'", filePath)
				return obstor.ObjectInfo{}, sftpToObjectErr(ctx, err, bucket, entry)
			}
		}

		objectInfo := fileInfoToObjectInfo(bucket, entry, fi)
		delete(fileInfos, filePath)
		return objectInfo, nil
	}

	return obstor.ListObjects(ctx, n, bucket, prefix, marker, delimiter, maxKeys, n.listPool, n.listDirFactory(), n.isLeaf, n.isLeafDir, getObjectInfo, getObjectInfo)
}

func fileInfoToObjectInfo(bucket string, entry string, fi os.FileInfo) obstor.ObjectInfo {
	return obstor.ObjectInfo{
		Bucket:  bucket,
		Name:    entry,
		ModTime: fi.ModTime(),
		Size:    fi.Size(),
		IsDir:   fi.IsDir(),
	}
}

func (n *sftpObjects) populateDirectoryListing(filePath string, fileInfos map[string]os.FileInfo) (os.FileInfo, error) {
	dirStat, err := n.clnt.Stat(filePath)
	if err != nil {
		return nil, err
	}

	key := path.Clean(filePath)
	if !dirStat.IsDir() {
		return dirStat, nil
	}

	fileInfos[key] = dirStat
	infos, err := n.clnt.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range infos {
		entryPath := obstor.PathJoin(filePath, fileInfo.Name())
		fileInfos[entryPath] = fileInfo
	}

	return dirStat, nil
}

// ListObjectsV2 lists all blobs in SFTP bucket filtered by prefix.
func (n *sftpObjects) ListObjectsV2(ctx context.Context, bucket, prefix, continuationToken, delimiter string, maxKeys int,
	fetchOwner bool, startAfter string) (loi obstor.ListObjectsV2Info, err error) {
	marker := continuationToken
	if marker == "" {
		marker = startAfter
	}
	resultV1, err := n.ListObjects(ctx, bucket, prefix, marker, delimiter, maxKeys)
	if err != nil {
		return loi, err
	}
	return obstor.ListObjectsV2Info{
		Objects:               resultV1.Objects,
		Prefixes:              resultV1.Prefixes,
		ContinuationToken:     continuationToken,
		NextContinuationToken: resultV1.NextMarker,
		IsTruncated:           resultV1.IsTruncated,
	}, nil
}

// deleteObject deletes a file path and recursively cleans up empty parent directories.
func (n *sftpObjects) deleteObject(basePath, deletePath string) error {
	if basePath == deletePath {
		return nil
	}

	if err := n.clnt.Remove(deletePath); err != nil {
		// Ignore errors if the directory is not empty.
		if isNotEmpty(err) {
			return nil
		}
		return err
	}

	deletePath = strings.TrimSuffix(deletePath, sftpSeparator)
	deletePath = path.Dir(deletePath)

	// Delete parent directory. Errors for parent directories shouldn't trickle down.
	_ = n.deleteObject(basePath, deletePath)
	return nil
}

// isNotEmpty checks if an SFTP error indicates a non-empty directory.
func isNotEmpty(err error) bool {
	if err == nil {
		return false
	}
	// SFTP servers typically return "Failure" (SSH_FX_FAILURE) for non-empty dirs.
	// Check the error string as a fallback.
	return strings.Contains(err.Error(), "not empty") ||
		strings.Contains(err.Error(), "directory is not empty")
}

func (n *sftpObjects) DeleteObject(ctx context.Context, bucket, object string, opts obstor.ObjectOptions) (obstor.ObjectInfo, error) {
	err := sftpToObjectErr(ctx, n.deleteObject(n.sftpPathJoin(bucket), n.sftpPathJoin(bucket, object)), bucket, object)
	return obstor.ObjectInfo{
		Bucket: bucket,
		Name:   object,
	}, err
}

func (n *sftpObjects) DeleteObjects(ctx context.Context, bucket string, objects []obstor.ObjectToDelete, opts obstor.ObjectOptions) ([]obstor.DeletedObject, []error) {
	errs := make([]error, len(objects))
	dobjects := make([]obstor.DeletedObject, len(objects))
	for idx, object := range objects {
		_, errs[idx] = n.DeleteObject(ctx, bucket, object.ObjectName, opts)
		if errs[idx] == nil {
			dobjects[idx] = obstor.DeletedObject{
				ObjectName: object.ObjectName,
			}
		}
	}
	return dobjects, errs
}

func (n *sftpObjects) GetObjectNInfo(ctx context.Context, bucket, object string, rs *obstor.HTTPRangeSpec, h http.Header, lockType obstor.LockType, opts obstor.ObjectOptions) (gr *obstor.GetObjectReader, err error) {
	objInfo, err := n.GetObjectInfo(ctx, bucket, object, opts)
	if err != nil {
		return nil, err
	}

	var startOffset, length int64
	startOffset, length, err = rs.GetOffsetLength(objInfo.Size)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		nerr := n.getObject(ctx, bucket, object, startOffset, length, pw, objInfo.ETag, opts)
		pw.CloseWithError(nerr)
	}()

	pipeCloser := func() { _ = pr.Close() }
	return obstor.NewGetObjectReaderFromReader(pr, objInfo, opts, pipeCloser)
}

func (n *sftpObjects) CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string, srcInfo obstor.ObjectInfo, srcOpts, dstOpts obstor.ObjectOptions) (obstor.ObjectInfo, error) {
	cpSrcDstSame := obstor.IsStringEqual(n.sftpPathJoin(srcBucket, srcObject), n.sftpPathJoin(dstBucket, dstObject))
	if cpSrcDstSame {
		return n.GetObjectInfo(ctx, srcBucket, srcObject, obstor.ObjectOptions{})
	}

	return n.PutObject(ctx, dstBucket, dstObject, srcInfo.PutObjReader, obstor.ObjectOptions{
		ServerSideEncryption: dstOpts.ServerSideEncryption,
		UserDefined:          srcInfo.UserDefined,
	})
}

func (n *sftpObjects) getObject(ctx context.Context, bucket, key string, startOffset, length int64, writer io.Writer, etag string, opts obstor.ObjectOptions) error {
	if _, err := n.clnt.Stat(n.sftpPathJoin(bucket)); err != nil {
		return sftpToObjectErr(ctx, err, bucket)
	}

	rd, err := n.clnt.Open(n.sftpPathJoin(bucket, key))
	if err != nil {
		return sftpToObjectErr(ctx, err, bucket, key)
	}
	defer func() { _ = rd.Close() }()

	if startOffset > 0 {
		if _, err = rd.Seek(startOffset, io.SeekStart); err != nil {
			return sftpToObjectErr(ctx, err, bucket, key)
		}
	}

	if _, err = io.CopyN(writer, rd, length); err != nil {
		if err == io.ErrClosedPipe {
			err = nil
		}
	}
	return sftpToObjectErr(ctx, err, bucket, key)
}

func (n *sftpObjects) isObjectDir(ctx context.Context, bucket, object string) bool {
	fis, err := n.clnt.ReadDir(n.sftpPathJoin(bucket, object))
	if err != nil {
		return false
	}
	return len(fis) == 0
}

// GetObjectInfo reads object info and replies back ObjectInfo.
func (n *sftpObjects) GetObjectInfo(ctx context.Context, bucket, object string, opts obstor.ObjectOptions) (objInfo obstor.ObjectInfo, err error) {
	_, err = n.clnt.Stat(n.sftpPathJoin(bucket))
	if err != nil {
		return objInfo, sftpToObjectErr(ctx, err, bucket)
	}
	if strings.HasSuffix(object, sftpSeparator) && !n.isObjectDir(ctx, bucket, object) {
		return objInfo, sftpToObjectErr(ctx, os.ErrNotExist, bucket, object)
	}

	fi, err := n.clnt.Stat(n.sftpPathJoin(bucket, object))
	if err != nil {
		return objInfo, sftpToObjectErr(ctx, err, bucket, object)
	}
	return obstor.ObjectInfo{
		Bucket:  bucket,
		Name:    object,
		ModTime: fi.ModTime(),
		Size:    fi.Size(),
		IsDir:   fi.IsDir(),
	}, nil
}

func (n *sftpObjects) PutObject(ctx context.Context, bucket string, object string, r *obstor.PutObjReader, opts obstor.ObjectOptions) (objInfo obstor.ObjectInfo, err error) {
	_, err = n.clnt.Stat(n.sftpPathJoin(bucket))
	if err != nil {
		return objInfo, sftpToObjectErr(ctx, err, bucket)
	}

	name := n.sftpPathJoin(bucket, object)

	// If it's a directory create a prefix.
	if strings.HasSuffix(object, sftpSeparator) && r.Size() == 0 {
		if err = n.clnt.MkdirAll(name); err != nil {
			_ = n.deleteObject(n.sftpPathJoin(bucket), name)
			return objInfo, sftpToObjectErr(ctx, err, bucket, object)
		}
	} else {
		tmpname := n.sftpPathJoin(obstorMetaTmpBucket, obstor.MustGetUUID())
		w, err := n.clnt.Create(tmpname)
		if err != nil {
			return objInfo, sftpToObjectErr(ctx, err, bucket, object)
		}
		defer func() { _ = n.deleteObject(n.sftpPathJoin(obstorMetaTmpBucket), tmpname) }()
		if _, err = io.Copy(w, r); err != nil {
			_ = w.Close()
			return objInfo, sftpToObjectErr(ctx, err, bucket, object)
		}
		_ = w.Close()

		dir := path.Dir(name)
		if dir != "" {
			if err = n.clnt.MkdirAll(dir); err != nil {
				_ = n.deleteObject(n.sftpPathJoin(bucket), dir)
				return objInfo, sftpToObjectErr(ctx, err, bucket, object)
			}
		}

		// SFTP Rename overwrites by default (POSIX semantics).
		if err = n.clnt.Rename(tmpname, name); err != nil {
			return objInfo, sftpToObjectErr(ctx, err, bucket, object)
		}
	}

	fi, err := n.clnt.Stat(name)
	if err != nil {
		return objInfo, sftpToObjectErr(ctx, err, bucket, object)
	}
	return obstor.ObjectInfo{
		Bucket:  bucket,
		Name:    object,
		ETag:    r.MD5CurrentHexString(),
		ModTime: fi.ModTime(),
		Size:    fi.Size(),
		IsDir:   fi.IsDir(),
	}, nil
}
