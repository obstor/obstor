/*
 * MinIO Cloud Storage, (C) 2017-2019 MinIO, Inc.
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

package cmd

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/obstor/obstor/cmd/config"
	xhttp "github.com/obstor/obstor/cmd/http"
	"github.com/obstor/obstor/cmd/logger"
	"github.com/obstor/obstor/pkg/env"
	"github.com/obstor/obstor/pkg/hash"
	xnet "github.com/obstor/obstor/pkg/net"

	obstor "github.com/obstor/obstor-go/v7"
)

var (
	// CleanMetadataKeys provides cleanMetadataKeys function alias.
	CleanMetadataKeys = cleanMetadataKeys

	// ListObjects function alias.
	ListObjects = listObjects

	// FilterListEntries function alias.
	FilterListEntries = filterListEntries
)

// FromObstorClientMetadata converts obstor metadata to map[string]string
func FromObstorClientMetadata(metadata map[string][]string) map[string]string {
	mm := make(map[string]string, len(metadata))
	for k, v := range metadata {
		mm[http.CanonicalHeaderKey(k)] = v[0]
	}
	return mm
}

// FromObstorClientObjectPart converts obstor ObjectPart to PartInfo
func FromObstorClientObjectPart(op obstor.ObjectPart) PartInfo {
	return PartInfo{
		Size:         op.Size,
		ETag:         CanonicalizeETag(op.ETag),
		LastModified: op.LastModified,
		PartNumber:   op.PartNumber,
	}
}

// FromObstorClientListPartsInfo converts obstor ListObjectPartsResult to ListPartsInfo
func FromObstorClientListPartsInfo(lopr obstor.ListObjectPartsResult) ListPartsInfo {
	// Convert obstor ObjectPart to PartInfo
	fromObstorClientObjectParts := func(parts []obstor.ObjectPart) []PartInfo {
		toParts := make([]PartInfo, len(parts))
		for i, part := range parts {
			toParts[i] = FromObstorClientObjectPart(part)
		}
		return toParts
	}

	return ListPartsInfo{
		UploadID:             lopr.UploadID,
		Bucket:               lopr.Bucket,
		Object:               lopr.Key,
		StorageClass:         "",
		PartNumberMarker:     lopr.PartNumberMarker,
		NextPartNumberMarker: lopr.NextPartNumberMarker,
		MaxParts:             lopr.MaxParts,
		IsTruncated:          lopr.IsTruncated,
		Parts:                fromObstorClientObjectParts(lopr.ObjectParts),
	}
}

// FromObstorClientListMultipartsInfo converts obstor ListMultipartUploadsResult to ListMultipartsInfo
func FromObstorClientListMultipartsInfo(lmur obstor.ListMultipartUploadsResult) ListMultipartsInfo {
	uploads := make([]MultipartInfo, len(lmur.Uploads))

	for i, um := range lmur.Uploads {
		uploads[i] = MultipartInfo{
			Object:    um.Key,
			UploadID:  um.UploadID,
			Initiated: um.Initiated,
		}
	}

	commonPrefixes := make([]string, len(lmur.CommonPrefixes))
	for i, cp := range lmur.CommonPrefixes {
		commonPrefixes[i] = cp.Prefix
	}

	return ListMultipartsInfo{
		KeyMarker:          lmur.KeyMarker,
		UploadIDMarker:     lmur.UploadIDMarker,
		NextKeyMarker:      lmur.NextKeyMarker,
		NextUploadIDMarker: lmur.NextUploadIDMarker,
		MaxUploads:         int(lmur.MaxUploads),
		IsTruncated:        lmur.IsTruncated,
		Uploads:            uploads,
		Prefix:             lmur.Prefix,
		Delimiter:          lmur.Delimiter,
		CommonPrefixes:     commonPrefixes,
		EncodingType:       lmur.EncodingType,
	}

}

// FromObstorClientObjectInfo converts obstor ObjectInfo to backend ObjectInfo
func FromObstorClientObjectInfo(bucket string, oi obstor.ObjectInfo) ObjectInfo {
	userDefined := FromObstorClientMetadata(oi.Metadata)
	userDefined[xhttp.ContentType] = oi.ContentType

	return ObjectInfo{
		Bucket:          bucket,
		Name:            oi.Key,
		ModTime:         oi.LastModified,
		Size:            oi.Size,
		ETag:            CanonicalizeETag(oi.ETag),
		UserDefined:     userDefined,
		ContentType:     oi.ContentType,
		ContentEncoding: oi.Metadata.Get(xhttp.ContentEncoding),
		StorageClass:    oi.StorageClass,
		Expires:         oi.Expires,
	}
}

// FromObstorClientListBucketV2Result converts obstor ListBucketResult to ListObjectsInfo
func FromObstorClientListBucketV2Result(bucket string, result obstor.ListBucketV2Result) ListObjectsV2Info {
	objects := make([]ObjectInfo, len(result.Contents))

	for i, oi := range result.Contents {
		objects[i] = FromObstorClientObjectInfo(bucket, oi)
	}

	prefixes := make([]string, len(result.CommonPrefixes))
	for i, p := range result.CommonPrefixes {
		prefixes[i] = p.Prefix
	}

	return ListObjectsV2Info{
		IsTruncated: result.IsTruncated,
		Prefixes:    prefixes,
		Objects:     objects,

		ContinuationToken:     result.ContinuationToken,
		NextContinuationToken: result.NextContinuationToken,
	}
}

// FromObstorClientListBucketResult converts obstor ListBucketResult to ListObjectsInfo
func FromObstorClientListBucketResult(bucket string, result obstor.ListBucketResult) ListObjectsInfo {
	objects := make([]ObjectInfo, len(result.Contents))

	for i, oi := range result.Contents {
		objects[i] = FromObstorClientObjectInfo(bucket, oi)
	}

	prefixes := make([]string, len(result.CommonPrefixes))
	for i, p := range result.CommonPrefixes {
		prefixes[i] = p.Prefix
	}

	return ListObjectsInfo{
		IsTruncated: result.IsTruncated,
		NextMarker:  result.NextMarker,
		Prefixes:    prefixes,
		Objects:     objects,
	}
}

// FromObstorClientListBucketResultToV2Info converts obstor ListBucketResult to ListObjectsV2Info
func FromObstorClientListBucketResultToV2Info(bucket string, result obstor.ListBucketResult) ListObjectsV2Info {
	objects := make([]ObjectInfo, len(result.Contents))

	for i, oi := range result.Contents {
		objects[i] = FromObstorClientObjectInfo(bucket, oi)
	}

	prefixes := make([]string, len(result.CommonPrefixes))
	for i, p := range result.CommonPrefixes {
		prefixes[i] = p.Prefix
	}

	return ListObjectsV2Info{
		IsTruncated:           result.IsTruncated,
		Prefixes:              prefixes,
		Objects:               objects,
		ContinuationToken:     result.Marker,
		NextContinuationToken: result.NextMarker,
	}
}

// ToObstorClientObjectInfoMetadata convertes metadata to map[string][]string
func ToObstorClientObjectInfoMetadata(metadata map[string]string) map[string][]string {
	mm := make(map[string][]string, len(metadata))
	for k, v := range metadata {
		mm[http.CanonicalHeaderKey(k)] = []string{v}
	}
	return mm
}

// ToObstorClientMetadata converts metadata to map[string]string
func ToObstorClientMetadata(metadata map[string]string) map[string]string {
	mm := make(map[string]string, len(metadata))
	for k, v := range metadata {
		mm[http.CanonicalHeaderKey(k)] = v
	}
	return mm
}

// ToObstorClientCompletePart converts CompletePart to obstor CompletePart
func ToObstorClientCompletePart(part CompletePart) obstor.CompletePart {
	return obstor.CompletePart{
		ETag:       part.ETag,
		PartNumber: part.PartNumber,
	}
}

// ToObstorClientCompleteParts converts []CompletePart to obstor []CompletePart
func ToObstorClientCompleteParts(parts []CompletePart) []obstor.CompletePart {
	mparts := make([]obstor.CompletePart, len(parts))
	for i, part := range parts {
		mparts[i] = ToObstorClientCompletePart(part)
	}
	return mparts
}

// IsBackendOnline - verifies if the backend is reachable
// by performing a GET request on the URL. returns 'true'
// if backend is reachable.
func IsBackendOnline(ctx context.Context, host string) bool {
	var d net.Dialer

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return false
	}

	_ = conn.Close()
	return true
}

// ErrorRespToObjectError converts Obstor errors to obstor object layer errors.
func ErrorRespToObjectError(err error, params ...string) error {
	if err == nil {
		return nil
	}

	bucket := ""
	object := ""
	if len(params) >= 1 {
		bucket = params[0]
	}
	if len(params) == 2 {
		object = params[1]
	}

	if xnet.IsNetworkOrHostDown(err, false) {
		return BackendDown{}
	}

	obstorErr, ok := err.(obstor.ErrorResponse)
	if !ok {
		// We don't interpret non Obstor errors. As obstor errors will
		// have StatusCode to help to convert to object errors.
		return err
	}

	switch obstorErr.Code {
	case "BucketAlreadyOwnedByYou":
		err = BucketAlreadyOwnedByYou{}
	case "BucketNotEmpty":
		err = BucketNotEmpty{}
	case "NoSuchBucketPolicy":
		err = BucketPolicyNotFound{}
	case "NoSuchLifecycleConfiguration":
		err = BucketLifecycleNotFound{}
	case "InvalidBucketName":
		err = BucketNameInvalid{Bucket: bucket}
	case "InvalidPart":
		err = InvalidPart{}
	case "NoSuchBucket":
		err = BucketNotFound{Bucket: bucket}
	case "NoSuchKey":
		if object != "" {
			err = ObjectNotFound{Bucket: bucket, Object: object}
		} else {
			err = BucketNotFound{Bucket: bucket}
		}
	case "XObstorInvalidObjectName":
		err = ObjectNameInvalid{}
	case "AccessDenied":
		err = PrefixAccessDenied{
			Bucket: bucket,
			Object: object,
		}
	case "XAmzContentSHA256Mismatch":
		err = hash.SHA256Mismatch{}
	case "NoSuchUpload":
		err = InvalidUploadID{}
	case "EntityTooSmall":
		err = PartTooSmall{}
	}

	return err
}

// ComputeCompleteMultipartMD5 calculates MD5 ETag for complete multipart responses
func ComputeCompleteMultipartMD5(parts []CompletePart) string {
	return getCompleteMultipartMD5(parts)
}

// Parse backend sse env variable
func parseBackendSSE(s string) (backendSSE, error) {
	l := strings.Split(s, ";")
	var gwSlice backendSSE
	for _, val := range l {
		v := strings.ToUpper(val)
		switch v {
		case "":
			continue
		case backendSSES3:
			fallthrough
		case backendSSEC:
			gwSlice = append(gwSlice, v)
			continue
		default:
			return nil, config.ErrInvalidGWSSEValue(nil).Msg("backend SSE cannot be (%s) ", v)
		}
	}
	return gwSlice, nil
}

// Handle backend env vars
func backendHandleEnvVars() {
	// Handle common env vars.
	handleCommonEnvVars()

	if !globalActiveCred.IsValid() {
		logger.Fatal(config.ErrInvalidCredentials(nil),
			"Unable to validate credentials inherited from the shell environment")
	}

	gwsseVal := env.Get("OBSTOR_BACKEND_SSE", "")
	if gwsseVal != "" {
		var err error
		GlobalBackendSSE, err = parseBackendSSE(gwsseVal)
		if err != nil {
			logger.Fatal(err, "Unable to parse OBSTOR_BACKEND_SSE value (`%s`)", gwsseVal)
		}
	}
}

// shouldMeterRequest checks whether incoming request should be added to prometheus backend metrics
func shouldMeterRequest(req *http.Request) bool {
	return !guessIsBrowserReq(req) && !guessIsHealthCheckReq(req) && !guessIsMetricsReq(req)
}

// MetricsTransport is a custom wrapper around Transport to track metrics
type MetricsTransport struct {
	Transport *http.Transport
	Metrics   *BackendMetrics
}

// RoundTrip implements the RoundTrip method for MetricsTransport
func (m MetricsTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	metered := shouldMeterRequest(r)
	if metered && (r.Method == http.MethodPost || r.Method == http.MethodPut) {
		m.Metrics.IncRequests(r.Method)
		if r.ContentLength > 0 {
			m.Metrics.IncBytesSent(uint64(r.ContentLength))
		}
	}
	// Make the request to the server.
	resp, err := m.Transport.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if metered && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
		m.Metrics.IncRequests(r.Method)
		if resp.ContentLength > 0 {
			m.Metrics.IncBytesReceived(uint64(resp.ContentLength))
		}
	}
	return resp, nil
}
