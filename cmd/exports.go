/*
 * MinIO Cloud Storage, (C) 2016-2020 MinIO, Inc.
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
	"fmt"
	"net/http"

	"github.com/obstor/obstor/pkg/auth"
	iampolicy "github.com/obstor/obstor/pkg/iam/policy"
	obstorpkg "github.com/obstor/obstor/pkg/obstor"
)

// Exported aliases for sub-packages (cmd/protocols/s3, cmd/protocols/sftp).

type ObjectAPIHandlers = objectAPIHandlers

var (
	NewObjectLayerFn       = newObjectLayerFn
	NewCachedObjectLayerFn = newCachedObjectLayerFn
	CollectAPIStats        = collectAPIStats
	HTTPTraceAll           = httpTraceAll
	HTTPTraceHdrs          = httpTraceHdrs
	MaxClients             = maxClients
	ErrorResponseHandler   = http.HandlerFunc(errorResponseHandler)
	IsErrBucketNotFound    = isErrBucketNotFound
	IsErrObjectNotFound    = isErrObjectNotFound
	IsErrVersionNotFound   = isErrVersionNotFound
	PathJoin               = pathJoin
	MustGetUUID            = mustGetUUID
	IsStringEqual          = isStringEqual
)

const ObstorReservedBucket = obstorpkg.ReservedBucket
const ReadLock = readLock

func MethodNotAllowedHandler(api string) func(w http.ResponseWriter, r *http.Request) {
	return methodNotAllowedHandler(api)
}

func GetActiveCred() auth.Credentials { return globalActiveCred }
func GetIAMSys() *IAMSys              { return globalIAMSys }
func GetCertsDir() *ConfigDir         { return globalCertsDir }
func GetDomainNames() []string        { return globalDomainNames }

// Enforce per bucket IAM policy
func CheckSFTPAccess(accessKey, bucket, objectName string, action iampolicy.Action) error {
	if bucket != "" && !IsBucketSFTPEnabled(bucket) {
		return fmt.Errorf("SFTP access is disabled for bucket %q", bucket)
	}
	if accessKey == "" {
		return fmt.Errorf("SFTP requires authenticated access")
	}
	// Root user / active cred is always allowed.
	if accessKey == globalActiveCred.AccessKey {
		return nil
	}
	if !globalIAMSys.IsAllowed(iampolicy.Args{
		AccountName: accessKey,
		Action:      action,
		BucketName:  bucket,
		ObjectName:  objectName,
		IsOwner:     false,
	}) {
		return fmt.Errorf("access denied: %s on %s/%s", action, bucket, objectName)
	}
	return nil
}

// IAM action constants exposed for use by the SFTP driver.
var (
	SFTPActionGetObject    iampolicy.Action = iampolicy.GetObjectAction
	SFTPActionPutObject    iampolicy.Action = iampolicy.PutObjectAction
	SFTPActionDeleteObject iampolicy.Action = iampolicy.DeleteObjectAction
	SFTPActionListBucket   iampolicy.Action = iampolicy.ListBucketAction
	SFTPActionGetBucket    iampolicy.Action = iampolicy.GetBucketLocationAction
	SFTPActionCreateBucket iampolicy.Action = iampolicy.CreateBucketAction
	SFTPActionDeleteBucket iampolicy.Action = iampolicy.DeleteBucketAction
	SFTPActionListBuckets  iampolicy.Action = iampolicy.ListAllMyBucketsAction
)
