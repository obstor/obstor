/*
 * MinIO Cloud Storage, (C) 2017-2020 MinIO, Inc.
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

package s3

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	obstorgo "github.com/obstor/obstor-go/v7"
	"github.com/obstor/obstor-go/v7/pkg/credentials"
	"github.com/obstor/obstor-go/v7/pkg/encrypt"
	"github.com/obstor/obstor-go/v7/pkg/s3utils"
	"github.com/obstor/obstor-go/v7/pkg/tags"
	obstor "github.com/obstor/obstor/cmd"
	xhttp "github.com/obstor/obstor/cmd/http"
	"github.com/obstor/obstor/cmd/logger"
	"github.com/obstor/obstor/pkg/auth"
	"github.com/obstor/obstor/pkg/bucket/policy"
	"github.com/obstor/obstor/pkg/madmin"
	"github.com/urfave/cli"
)

func init() {
	const s3BackendTemplate = `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} {{if .VisibleFlags}}[FLAGS]{{end}} [ENDPOINT]
{{if .VisibleFlags}}
FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}
ENDPOINT:
  s3 server endpoint. Default ENDPOINT is https://s3.amazonaws.com

EXAMPLES:
  1. Start obstor backend server for AWS S3 backend
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_USER{{.AssignmentOperator}}accesskey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_PASSWORD{{.AssignmentOperator}}secretkey
     {{.Prompt}} {{.HelpName}}

  2. Start obstor backend server for AWS S3 backend with edge caching enabled
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_USER{{.AssignmentOperator}}accesskey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_ROOT_PASSWORD{{.AssignmentOperator}}secretkey
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_DRIVES{{.AssignmentOperator}}"/mnt/drive1,/mnt/drive2,/mnt/drive3,/mnt/drive4"
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_EXCLUDE{{.AssignmentOperator}}"bucket1/*,*.png"
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_QUOTA{{.AssignmentOperator}}90
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_AFTER{{.AssignmentOperator}}3
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_WATERMARK_LOW{{.AssignmentOperator}}75
     {{.Prompt}} {{.EnvVarSetCommand}} OBSTOR_CACHE_WATERMARK_HIGH{{.AssignmentOperator}}85
     {{.Prompt}} {{.HelpName}}
`

	_ = obstor.RegisterBackendCommand(cli.Command{
		Name:               obstor.S3Backend,
		Usage:              "Amazon Simple Storage Service (S3)",
		Action:             s3BackendMain,
		CustomHelpTemplate: s3BackendTemplate,
		HideHelp:           true,
	})
}

// Handler for 'obstor backend s3' command line.
func s3BackendMain(ctx *cli.Context) {
	args := ctx.Args()
	if !ctx.Args().Present() {
		args = cli.Args{"https://s3.amazonaws.com"}
	}

	serverAddr := ctx.GlobalString("address")
	if serverAddr == "" || serverAddr == ":"+obstor.GlobalObstorDefaultPort {
		serverAddr = ctx.String("address")
	}
	// Validate backend arguments.
	logger.FatalIf(obstor.ValidateBackendArguments(serverAddr, args.First()), "Invalid argument")

	// Start the backend..
	obstor.StartBackend(ctx, &S3{args.First()})
}

// S3 implements Backend.
type S3 struct {
	host string
}

// Name implements Backend interface.
func (g *S3) Name() string {
	return obstor.S3Backend
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// randString generates random names and prepends them with a known prefix.
func randString(n int, src rand.Source, prefix string) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return prefix + string(b[0:30-len(prefix)])
}

// Chains all credential types, in the following order:
//   - AWS env vars (i.e. AWS_ACCESS_KEY_ID)
//   - AWS creds file (i.e. AWS_SHARED_CREDENTIALS_FILE or ~/.aws/credentials)
//   - Static credentials provided by user (i.e. OBSTOR_ROOT_USER/OBSTOR_ACCESS_KEY)
var defaultProviders = []credentials.Provider{
	&credentials.EnvAWS{},
	&credentials.FileAWSCredentials{},
	&credentials.EnvObstor{},
}

// Chains all credential types, in the following order:
//   - AWS env vars (i.e. AWS_ACCESS_KEY_ID)
//   - AWS creds file (i.e. AWS_SHARED_CREDENTIALS_FILE or ~/.aws/credentials)
//   - IAM profile based credentials. (performs an HTTP
//     call to a pre-defined endpoint, only valid inside
//     configured ec2 instances)
//   - Static credentials provided by user (i.e. OBSTOR_ROOT_USER/OBSTOR_ACCESS_KEY)
var defaultAWSCredProviders = []credentials.Provider{
	&credentials.EnvAWS{},
	&credentials.FileAWSCredentials{},
	&credentials.IAM{
		Client: &http.Client{
			Transport: obstor.NewBackendHTTPTransport(),
		},
	},
	&credentials.EnvObstor{},
}

// newS3 - Initializes a new client by auto probing S3 server signature.
func newS3(urlStr string, tripper http.RoundTripper) (*obstorgo.Core, error) {
	if urlStr == "" {
		urlStr = "https://s3.amazonaws.com"
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// Override default params if the host is provided
	endpoint, secure, err := obstor.ParseBackendEndpoint(urlStr)
	if err != nil {
		return nil, err
	}

	var creds *credentials.Credentials
	if s3utils.IsAmazonEndpoint(*u) {
		// If we see an Amazon S3 endpoint, then we use more ways to fetch backend credentials.
		// Specifically IAM style rotating credentials are only supported with AWS S3 endpoint.
		creds = credentials.NewChainCredentials(defaultAWSCredProviders)

	} else {
		creds = credentials.NewChainCredentials(defaultProviders)
	}

	options := &obstorgo.Options{
		Creds:        creds,
		Secure:       secure,
		Region:       s3utils.GetRegionFromURL(*u),
		BucketLookup: obstorgo.BucketLookupAuto,
		Transport:    tripper,
	}

	clnt, err := obstorgo.New(endpoint, options)
	if err != nil {
		return nil, err
	}

	return &obstorgo.Core{Client: clnt}, nil
}

// NewBackendLayer returns s3 ObjectLayer.
func (g *S3) NewBackendLayer(creds auth.Credentials) (obstor.ObjectLayer, error) {
	metrics := obstor.NewMetrics()

	t := &obstor.MetricsTransport{
		Transport: obstor.NewBackendHTTPTransport(),
		Metrics:   metrics,
	}

	// Creds are ignored here, since S3 backend implements chaining
	// all credentials.
	clnt, err := newS3(g.host, t)
	if err != nil {
		return nil, err
	}

	probeBucketName := randString(60, rand.NewSource(time.Now().UnixNano()), "probe-bucket-sign-")

	// Check if the provided keys are valid.
	if _, err = clnt.BucketExists(context.Background(), probeBucketName); err != nil {
		if obstorgo.ToErrorResponse(err).Code != "AccessDenied" {
			return nil, err
		}
	}

	s := s3Objects{
		Client:  clnt,
		Metrics: metrics,
		HTTPClient: &http.Client{
			Transport: t,
		},
	}

	// Enables single encryption of KMS is configured.
	if obstor.GlobalKMS != nil {
		encS := s3EncObjects{s}

		// Start stale enc multipart uploads cleanup routine.
		go encS.cleanupStaleEncMultipartUploads(obstor.GlobalContext,
			obstor.GlobalStaleUploadsCleanupInterval, obstor.GlobalStaleUploadsExpiry)

		return &encS, nil
	}
	return &s, nil
}

// S3 backend is production-ready
func (g *S3) Production() bool {
	return true
}

// s3Objects implements backend for Obstor and S3 compatible object storage servers.
type s3Objects struct {
	obstor.BackendUnsupported
	Client     *obstorgo.Core
	HTTPClient *http.Client
	Metrics    *obstor.BackendMetrics
}

// GetMetrics returns this backend's metrics
func (l *s3Objects) GetMetrics(ctx context.Context) (*obstor.BackendMetrics, error) {
	return l.Metrics, nil
}

// Shutdown saves any backend metadata to disk
// if necessary and reload upon next restart.
func (l *s3Objects) Shutdown(ctx context.Context) error {
	return nil
}

// StorageInfo is not relevant to S3 backend.
func (l *s3Objects) StorageInfo(ctx context.Context) (si obstor.StorageInfo, _ []error) {
	si.Backend.Type = madmin.Gateway
	host := l.Client.EndpointURL().Host
	if l.Client.EndpointURL().Port() == "" {
		host = l.Client.EndpointURL().Host + ":" + l.Client.EndpointURL().Scheme
	}
	si.Backend.GatewayOnline = obstor.IsBackendOnline(ctx, host)
	return si, nil
}

// MakeBucket creates a new container on S3 backend.
func (l *s3Objects) MakeBucketWithLocation(ctx context.Context, bucket string, opts obstor.BucketOptions) error {
	if opts.LockEnabled || opts.VersioningEnabled {
		return obstor.NotImplemented{}
	}

	// Verify if bucket name is valid.
	// We are using a separate helper function here to validate bucket
	// names instead of IsValidBucketName() because there is a possibility
	// that certains users might have buckets which are non-DNS compliant
	// in us-east-1 and we might severely restrict them by not allowing
	// access to these buckets.
	// Ref - http://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html
	if s3utils.CheckValidBucketName(bucket) != nil {
		return obstor.BucketNameInvalid{Bucket: bucket}
	}
	err := l.Client.MakeBucket(ctx, bucket, obstorgo.MakeBucketOptions{Region: opts.Location})
	if err != nil {
		return obstor.ErrorRespToObjectError(err, bucket)
	}
	return err
}

// GetBucketInfo gets bucket metadata..
func (l *s3Objects) GetBucketInfo(ctx context.Context, bucket string) (bi obstor.BucketInfo, e error) {
	buckets, err := l.Client.ListBuckets(ctx)
	if err != nil {
		// Listbuckets may be disallowed, proceed to check if
		// bucket indeed exists, if yes return success.
		var ok bool
		if ok, err = l.Client.BucketExists(ctx, bucket); err != nil {
			return bi, obstor.ErrorRespToObjectError(err, bucket)
		}
		if !ok {
			return bi, obstor.BucketNotFound{Bucket: bucket}
		}
		return obstor.BucketInfo{
			Name:    bi.Name,
			Created: time.Now().UTC(),
		}, nil
	}

	for _, bi := range buckets {
		if bi.Name != bucket {
			continue
		}

		return obstor.BucketInfo{
			Name:    bi.Name,
			Created: bi.CreationDate,
		}, nil
	}

	return bi, obstor.BucketNotFound{Bucket: bucket}
}

// ListBuckets lists all S3 buckets
func (l *s3Objects) ListBuckets(ctx context.Context) ([]obstor.BucketInfo, error) {
	buckets, err := l.Client.ListBuckets(ctx)
	if err != nil {
		return nil, obstor.ErrorRespToObjectError(err)
	}

	b := make([]obstor.BucketInfo, len(buckets))
	for i, bi := range buckets {
		b[i] = obstor.BucketInfo{
			Name:    bi.Name,
			Created: bi.CreationDate,
		}
	}

	return b, err
}

// DeleteBucket deletes a bucket on S3
func (l *s3Objects) DeleteBucket(ctx context.Context, bucket string, forceDelete bool) error {
	err := l.Client.RemoveBucket(ctx, bucket)
	if err != nil {
		return obstor.ErrorRespToObjectError(err, bucket)
	}
	return nil
}

// ListObjects lists all blobs in S3 bucket filtered by prefix
func (l *s3Objects) ListObjects(ctx context.Context, bucket string, prefix string, marker string, delimiter string, maxKeys int) (loi obstor.ListObjectsInfo, e error) {
	result, err := l.Client.ListObjects(bucket, prefix, marker, delimiter, maxKeys)
	if err != nil {
		return loi, obstor.ErrorRespToObjectError(err, bucket)
	}

	return obstor.FromObstorClientListBucketResult(bucket, result), nil
}

// ListObjectsV2 lists all blobs in S3 bucket filtered by prefix
func (l *s3Objects) ListObjectsV2(ctx context.Context, bucket, prefix, continuationToken, delimiter string, maxKeys int, fetchOwner bool, startAfter string) (loi obstor.ListObjectsV2Info, e error) {
	result, err := l.Client.ListObjectsV2(bucket, prefix, startAfter, continuationToken, delimiter, maxKeys)
	if err != nil {
		return loi, obstor.ErrorRespToObjectError(err, bucket)
	}

	return obstor.FromObstorClientListBucketV2Result(bucket, result), nil
}

// GetObjectNInfo - returns object info and locked object ReadCloser
func (l *s3Objects) GetObjectNInfo(ctx context.Context, bucket, object string, rs *obstor.HTTPRangeSpec, h http.Header, lockType obstor.LockType, opts obstor.ObjectOptions) (gr *obstor.GetObjectReader, err error) {
	var objInfo obstor.ObjectInfo
	objInfo, err = l.GetObjectInfo(ctx, bucket, object, opts)
	if err != nil {
		return nil, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	fn, off, length, err := obstor.NewGetObjectReader(rs, objInfo, opts)
	if err != nil {
		return nil, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	pr, pw := io.Pipe()
	go func() {
		err := l.getObject(ctx, bucket, object, off, length, pw, objInfo.ETag, opts)
		pw.CloseWithError(err)
	}()

	// Setup cleanup function to cause the above go-routine to
	// exit in case of partial read
	pipeCloser := func() { _ = pr.Close() }
	return fn(pr, h, opts.CheckPrecondFn, pipeCloser)
}

// GetObject reads an object from S3. Supports additional
// parameters like offset and length which are synonymous with
// HTTP Range requests.
//
// startOffset indicates the starting read location of the object.
// length indicates the total length of the object.
func (l *s3Objects) getObject(ctx context.Context, bucket string, key string, startOffset int64, length int64, writer io.Writer, etag string, o obstor.ObjectOptions) error {
	if length < 0 && length != -1 {
		return obstor.ErrorRespToObjectError(obstor.InvalidRange{}, bucket, key)
	}

	opts := obstorgo.GetObjectOptions{}
	opts.ServerSideEncryption = o.ServerSideEncryption

	if startOffset >= 0 && length >= 0 {
		if err := opts.SetRange(startOffset, startOffset+length-1); err != nil {
			return obstor.ErrorRespToObjectError(err, bucket, key)
		}
	}

	if etag != "" {
		_ = opts.SetMatchETag(etag)
	}

	object, _, _, err := l.Client.GetObject(ctx, bucket, key, opts)
	if err != nil {
		return obstor.ErrorRespToObjectError(err, bucket, key)
	}
	defer object.Close()
	if _, err := io.Copy(writer, object); err != nil {
		return obstor.ErrorRespToObjectError(err, bucket, key)
	}
	return nil
}

// GetObjectInfo reads object info and replies back ObjectInfo
func (l *s3Objects) GetObjectInfo(ctx context.Context, bucket string, object string, opts obstor.ObjectOptions) (objInfo obstor.ObjectInfo, err error) {
	oi, err := l.Client.StatObject(ctx, bucket, object, obstorgo.StatObjectOptions{
		ServerSideEncryption: opts.ServerSideEncryption,
	})
	if err != nil {
		return obstor.ObjectInfo{}, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	return obstor.FromObstorClientObjectInfo(bucket, oi), nil
}

// PutObject creates a new object with the incoming data,
func (l *s3Objects) PutObject(ctx context.Context, bucket string, object string, r *obstor.PutObjReader, opts obstor.ObjectOptions) (objInfo obstor.ObjectInfo, err error) {
	data := r.Reader
	var tagMap map[string]string
	if tagstr, ok := opts.UserDefined[xhttp.AmzObjectTagging]; ok && tagstr != "" {
		tagObj, err := tags.ParseObjectTags(tagstr)
		if err != nil {
			return objInfo, obstor.ErrorRespToObjectError(err, bucket, object)
		}
		tagMap = tagObj.ToMap()
		delete(opts.UserDefined, xhttp.AmzObjectTagging)
	}
	PutOpts := obstorgo.PutObjectOptions{
		UserMetadata:         opts.UserDefined,
		ServerSideEncryption: opts.ServerSideEncryption,
		UserTags:             tagMap,
	}
	ui, err := l.Client.PutObject(ctx, bucket, object, data, data.Size(), data.MD5Base64String(), data.SHA256HexString(), PutOpts)
	if err != nil {
		return objInfo, obstor.ErrorRespToObjectError(err, bucket, object)
	}
	// On success, populate the key & metadata so they are present in the notification
	oi := obstorgo.ObjectInfo{
		ETag:     ui.ETag,
		Size:     ui.Size,
		Key:      object,
		Metadata: obstor.ToObstorClientObjectInfoMetadata(opts.UserDefined),
	}

	return obstor.FromObstorClientObjectInfo(bucket, oi), nil
}

// CopyObject copies an object from source bucket to a destination bucket.
func (l *s3Objects) CopyObject(ctx context.Context, srcBucket string, srcObject string, dstBucket string, dstObject string, srcInfo obstor.ObjectInfo, srcOpts, dstOpts obstor.ObjectOptions) (objInfo obstor.ObjectInfo, err error) {
	if srcOpts.CheckPrecondFn != nil && srcOpts.CheckPrecondFn(srcInfo) {
		return obstor.ObjectInfo{}, obstor.PreConditionFailed{}
	}
	// Set this header such that following CopyObject() always sets the right metadata on the destination.
	// metadata input is already a trickled down value from interpreting x-amz-metadata-directive at
	// handler layer. So what we have right now is supposed to be applied on the destination object anyways.
	// So preserve it by adding "REPLACE" directive to save all the metadata set by CopyObject API.
	srcInfo.UserDefined["x-amz-metadata-directive"] = "REPLACE"
	srcInfo.UserDefined["x-amz-copy-source-if-match"] = srcInfo.ETag
	header := make(http.Header)
	if srcOpts.ServerSideEncryption != nil {
		encrypt.SSECopy(srcOpts.ServerSideEncryption).Marshal(header)
	}

	if dstOpts.ServerSideEncryption != nil {
		dstOpts.ServerSideEncryption.Marshal(header)
	}

	for k, v := range header {
		srcInfo.UserDefined[k] = v[0]
	}

	if _, err = l.Client.CopyObject(ctx, srcBucket, srcObject, dstBucket, dstObject, srcInfo.UserDefined, obstorgo.CopySrcOptions{}, obstorgo.PutObjectOptions{}); err != nil {
		return objInfo, obstor.ErrorRespToObjectError(err, srcBucket, srcObject)
	}
	return l.GetObjectInfo(ctx, dstBucket, dstObject, dstOpts)
}

// DeleteObject deletes a blob in bucket
func (l *s3Objects) DeleteObject(ctx context.Context, bucket string, object string, opts obstor.ObjectOptions) (obstor.ObjectInfo, error) {
	err := l.Client.RemoveObject(ctx, bucket, object, obstorgo.RemoveObjectOptions{})
	if err != nil {
		return obstor.ObjectInfo{}, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	return obstor.ObjectInfo{
		Bucket: bucket,
		Name:   object,
	}, nil
}

func (l *s3Objects) DeleteObjects(ctx context.Context, bucket string, objects []obstor.ObjectToDelete, opts obstor.ObjectOptions) ([]obstor.DeletedObject, []error) {
	errs := make([]error, len(objects))
	dobjects := make([]obstor.DeletedObject, len(objects))
	for idx, object := range objects {
		_, errs[idx] = l.DeleteObject(ctx, bucket, object.ObjectName, opts)
		if errs[idx] == nil {
			dobjects[idx] = obstor.DeletedObject{
				ObjectName: object.ObjectName,
			}
		}
	}
	return dobjects, errs
}

// ListMultipartUploads lists all multipart uploads.
func (l *s3Objects) ListMultipartUploads(ctx context.Context, bucket string, prefix string, keyMarker string, uploadIDMarker string, delimiter string, maxUploads int) (lmi obstor.ListMultipartsInfo, e error) {
	result, err := l.Client.ListMultipartUploads(ctx, bucket, prefix, keyMarker, uploadIDMarker, delimiter, maxUploads)
	if err != nil {
		return lmi, err
	}

	return obstor.FromObstorClientListMultipartsInfo(result), nil
}

// NewMultipartUpload upload object in multiple parts
func (l *s3Objects) NewMultipartUpload(ctx context.Context, bucket string, object string, o obstor.ObjectOptions) (uploadID string, err error) {
	var tagMap map[string]string
	if tagStr, ok := o.UserDefined[xhttp.AmzObjectTagging]; ok {
		tagObj, err := tags.Parse(tagStr, true)
		if err != nil {
			return uploadID, obstor.ErrorRespToObjectError(err, bucket, object)
		}
		tagMap = tagObj.ToMap()
		delete(o.UserDefined, xhttp.AmzObjectTagging)
	}
	// Create PutObject options
	opts := obstorgo.PutObjectOptions{
		UserMetadata:         o.UserDefined,
		ServerSideEncryption: o.ServerSideEncryption,
		UserTags:             tagMap,
	}
	uploadID, err = l.Client.NewMultipartUpload(ctx, bucket, object, opts)
	if err != nil {
		return uploadID, obstor.ErrorRespToObjectError(err, bucket, object)
	}
	return uploadID, nil
}

// PutObjectPart puts a part of object in bucket
func (l *s3Objects) PutObjectPart(ctx context.Context, bucket string, object string, uploadID string, partID int, r *obstor.PutObjReader, opts obstor.ObjectOptions) (pi obstor.PartInfo, e error) {
	data := r.Reader
	info, err := l.Client.PutObjectPart(ctx, bucket, object, uploadID, partID, data, data.Size(), obstorgo.PutObjectPartOptions{
		Md5Base64: data.MD5Base64String(),
		Sha256Hex: data.SHA256HexString(),
		SSE:       opts.ServerSideEncryption,
	})
	if err != nil {
		return pi, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	return obstor.FromObstorClientObjectPart(info), nil
}

// CopyObjectPart creates a part in a multipart upload by copying
// existing object or a part of it.
func (l *s3Objects) CopyObjectPart(ctx context.Context, srcBucket, srcObject, destBucket, destObject, uploadID string,
	partID int, startOffset, length int64, srcInfo obstor.ObjectInfo, srcOpts, dstOpts obstor.ObjectOptions) (p obstor.PartInfo, err error) {
	if srcOpts.CheckPrecondFn != nil && srcOpts.CheckPrecondFn(srcInfo) {
		return obstor.PartInfo{}, obstor.PreConditionFailed{}
	}
	srcInfo.UserDefined = map[string]string{
		"x-amz-copy-source-if-match": srcInfo.ETag,
	}
	header := make(http.Header)
	if srcOpts.ServerSideEncryption != nil {
		encrypt.SSECopy(srcOpts.ServerSideEncryption).Marshal(header)
	}

	if dstOpts.ServerSideEncryption != nil {
		dstOpts.ServerSideEncryption.Marshal(header)
	}
	for k, v := range header {
		srcInfo.UserDefined[k] = v[0]
	}

	completePart, err := l.Client.CopyObjectPart(ctx, srcBucket, srcObject, destBucket, destObject,
		uploadID, partID, startOffset, length, srcInfo.UserDefined)
	if err != nil {
		return p, obstor.ErrorRespToObjectError(err, srcBucket, srcObject)
	}
	p.PartNumber = completePart.PartNumber
	p.ETag = completePart.ETag
	return p, nil
}

// GetMultipartInfo returns multipart info of the uploadId of the object
func (l *s3Objects) GetMultipartInfo(ctx context.Context, bucket, object, uploadID string, opts obstor.ObjectOptions) (result obstor.MultipartInfo, err error) {
	result.Bucket = bucket
	result.Object = object
	result.UploadID = uploadID
	return result, nil
}

// ListObjectParts returns all object parts for specified object in specified bucket
func (l *s3Objects) ListObjectParts(ctx context.Context, bucket string, object string, uploadID string, partNumberMarker int, maxParts int, opts obstor.ObjectOptions) (lpi obstor.ListPartsInfo, e error) {
	result, err := l.Client.ListObjectParts(ctx, bucket, object, uploadID, partNumberMarker, maxParts)
	if err != nil {
		return lpi, err
	}
	lpi = obstor.FromObstorClientListPartsInfo(result)
	if lpi.IsTruncated && maxParts > len(lpi.Parts) {
		partNumberMarker = lpi.NextPartNumberMarker
		for {
			result, err = l.Client.ListObjectParts(ctx, bucket, object, uploadID, partNumberMarker, maxParts)
			if err != nil {
				return lpi, err
			}

			nlpi := obstor.FromObstorClientListPartsInfo(result)

			partNumberMarker = nlpi.NextPartNumberMarker

			lpi.Parts = append(lpi.Parts, nlpi.Parts...)
			if !nlpi.IsTruncated {
				break
			}
		}
	}
	return lpi, nil
}

// AbortMultipartUpload aborts a ongoing multipart upload
func (l *s3Objects) AbortMultipartUpload(ctx context.Context, bucket string, object string, uploadID string, opts obstor.ObjectOptions) error {
	err := l.Client.AbortMultipartUpload(ctx, bucket, object, uploadID)
	return obstor.ErrorRespToObjectError(err, bucket, object)
}

// CompleteMultipartUpload completes ongoing multipart upload and finalizes object
func (l *s3Objects) CompleteMultipartUpload(ctx context.Context, bucket string, object string, uploadID string, uploadedParts []obstor.CompletePart, opts obstor.ObjectOptions) (oi obstor.ObjectInfo, e error) {
	uploadInfo, err := l.Client.CompleteMultipartUpload(ctx, bucket, object, uploadID, obstor.ToObstorClientCompleteParts(uploadedParts), obstorgo.PutObjectOptions{})
	if err != nil {
		return oi, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	return obstor.ObjectInfo{Bucket: bucket, Name: object, ETag: strings.Trim(uploadInfo.ETag, "\"")}, nil
}

// SetBucketPolicy sets policy on bucket
func (l *s3Objects) SetBucketPolicy(ctx context.Context, bucket string, bucketPolicy *policy.Policy) error {
	data, err := json.Marshal(bucketPolicy)
	if err != nil {
		// This should not happen.
		logger.LogIf(ctx, err)
		return obstor.ErrorRespToObjectError(err, bucket)
	}

	if err := l.Client.SetBucketPolicy(ctx, bucket, string(data)); err != nil {
		return obstor.ErrorRespToObjectError(err, bucket)
	}

	return nil
}

// GetBucketPolicy will get policy on bucket
func (l *s3Objects) GetBucketPolicy(ctx context.Context, bucket string) (*policy.Policy, error) {
	data, err := l.Client.GetBucketPolicy(ctx, bucket)
	if err != nil {
		return nil, obstor.ErrorRespToObjectError(err, bucket)
	}

	bucketPolicy, err := policy.ParseConfig(strings.NewReader(data), bucket)
	return bucketPolicy, obstor.ErrorRespToObjectError(err, bucket)
}

// DeleteBucketPolicy deletes all policies on bucket
func (l *s3Objects) DeleteBucketPolicy(ctx context.Context, bucket string) error {
	if err := l.Client.SetBucketPolicy(ctx, bucket, ""); err != nil {
		return obstor.ErrorRespToObjectError(err, bucket, "")
	}
	return nil
}

// GetObjectTags gets the tags set on the object
func (l *s3Objects) GetObjectTags(ctx context.Context, bucket string, object string, opts obstor.ObjectOptions) (*tags.Tags, error) {
	var err error
	if _, err = l.GetObjectInfo(ctx, bucket, object, opts); err != nil {
		return nil, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	t, err := l.Client.GetObjectTagging(ctx, bucket, object, obstorgo.GetObjectTaggingOptions{})
	if err != nil {
		return nil, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	return t, nil
}

// PutObjectTags attaches the tags to the object
func (l *s3Objects) PutObjectTags(ctx context.Context, bucket, object string, tagStr string, opts obstor.ObjectOptions) (obstor.ObjectInfo, error) {
	tagObj, err := tags.Parse(tagStr, true)
	if err != nil {
		return obstor.ObjectInfo{}, obstor.ErrorRespToObjectError(err, bucket, object)
	}
	if err = l.Client.PutObjectTagging(ctx, bucket, object, tagObj, obstorgo.PutObjectTaggingOptions{VersionID: opts.VersionID}); err != nil {
		return obstor.ObjectInfo{}, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	objInfo, err := l.GetObjectInfo(ctx, bucket, object, opts)
	if err != nil {
		return obstor.ObjectInfo{}, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	return objInfo, nil
}

// DeleteObjectTags removes the tags attached to the object
func (l *s3Objects) DeleteObjectTags(ctx context.Context, bucket, object string, opts obstor.ObjectOptions) (obstor.ObjectInfo, error) {
	if err := l.Client.RemoveObjectTagging(ctx, bucket, object, obstorgo.RemoveObjectTaggingOptions{}); err != nil {
		return obstor.ObjectInfo{}, obstor.ErrorRespToObjectError(err, bucket, object)
	}
	objInfo, err := l.GetObjectInfo(ctx, bucket, object, opts)
	if err != nil {
		return obstor.ObjectInfo{}, obstor.ErrorRespToObjectError(err, bucket, object)
	}

	return objInfo, nil
}

// IsCompressionSupported returns whether compression is applicable for this layer.
func (l *s3Objects) IsCompressionSupported() bool {
	return false
}

// IsEncryptionSupported returns whether server side encryption is implemented for this layer.
func (l *s3Objects) IsEncryptionSupported() bool {
	return obstor.GlobalKMS != nil || obstor.GlobalBackendSSE.IsSet()
}

func (l *s3Objects) IsTaggingSupported() bool {
	return true
}
