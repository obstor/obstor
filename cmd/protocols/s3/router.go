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

package s3

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	obstor "github.com/obstor/obstor/cmd"
	xhttp "github.com/obstor/obstor/cmd/http"
	"github.com/obstor/obstor/pkg/wildcard"
	"github.com/rs/cors"
)

func init() {
	obstor.GlobalS3RegisterAPIRouterFn = RegisterAPIRouter
	obstor.GlobalS3CorsHandlerFn = CorsHandler
	obstor.GlobalS3RegisterTestAPIFn = RegisterTestAPI
}

// getHost tries its best to return the request host.
// According to section 14.23 of RFC 2616 the Host header
// can include the port number if the default value of 80 is not used.
func getHost(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.Host
	}
	return r.Host
}

type rejectedAPI struct {
	api     string
	methods []string
	queries []string
	path    string
}

var rejectedAPIs = []rejectedAPI{
	{
		api:     "inventory",
		methods: []string{http.MethodGet, http.MethodPut, http.MethodDelete},
		queries: []string{"inventory", ""},
	},
	{
		api:     "cors",
		methods: []string{http.MethodPut, http.MethodDelete},
		queries: []string{"cors", ""},
	},
	{
		api:     "metrics",
		methods: []string{http.MethodGet, http.MethodPut, http.MethodDelete},
		queries: []string{"metrics", ""},
	},
	{
		api:     "website",
		methods: []string{http.MethodPut},
		queries: []string{"website", ""},
	},
	{
		api:     "logging",
		methods: []string{http.MethodPut, http.MethodDelete},
		queries: []string{"logging", ""},
	},
	{
		api:     "accelerate",
		methods: []string{http.MethodPut, http.MethodDelete},
		queries: []string{"accelerate", ""},
	},
	{
		api:     "requestPayment",
		methods: []string{http.MethodPut, http.MethodDelete},
		queries: []string{"requestPayment", ""},
	},
	{
		api:     "torrent",
		methods: []string{http.MethodPut, http.MethodDelete, http.MethodGet},
		queries: []string{"torrent", ""},
		path:    "/{object:.+}",
	},
	{
		api:     "acl",
		methods: []string{http.MethodDelete},
		queries: []string{"acl", ""},
		path:    "/{object:.+}",
	},
	{
		api:     "acl",
		methods: []string{http.MethodDelete, http.MethodPut, http.MethodHead},
		queries: []string{"acl", ""},
	},
	{
		api:     "publicAccessBlock",
		methods: []string{http.MethodDelete, http.MethodPut, http.MethodGet},
		queries: []string{"publicAccessBlock", ""},
	},
	{
		api:     "ownershipControls",
		methods: []string{http.MethodDelete, http.MethodPut, http.MethodGet},
		queries: []string{"ownershipControls", ""},
	},
	{
		api:     "intelligent-tiering",
		methods: []string{http.MethodDelete, http.MethodPut, http.MethodGet},
		queries: []string{"intelligent-tiering", ""},
	},
	{
		api:     "analytics",
		methods: []string{http.MethodDelete, http.MethodPut, http.MethodGet},
		queries: []string{"analytics", ""},
	},
}

func rejectUnsupportedAPIs(router *mux.Router) {
	for _, r := range rejectedAPIs {
		t := router.Methods(r.methods...).
			HandlerFunc(obstor.CollectAPIStats(r.api, obstor.HTTPTraceAll(obstor.NotImplementedHandler))).
			Queries(r.queries...)
		if r.path != "" {
			t.Path(r.path)
		}
	}
}

// RegisterAPIRouter registers S3 compatible APIs.
func RegisterAPIRouter(router *mux.Router) {
	// Initialize API.
	api := Handlers{
		ObjectAPIHandlers: obstor.ObjectAPIHandlers{
			ObjectAPI: obstor.NewObjectLayerFn,
			CacheAPI:  obstor.NewCachedObjectLayerFn,
		},
	}

	// API Router
	apiRouter := router.PathPrefix(obstor.SlashSeparator).Subrouter()

	var routers []*mux.Router
	for _, domainName := range obstor.GetDomainNames() {
		if obstor.IsKubernetes() {
			routers = append(routers, apiRouter.MatcherFunc(func(r *http.Request, match *mux.RouteMatch) bool {
				host, _, err := net.SplitHostPort(getHost(r))
				if err != nil {
					host = r.Host
				}
				// Make sure to skip matching obstor.<domain>` this is
				// specifically meant for operator/k8s deployment
				// The reason we need to skip this is for a special
				// usecase where we need to make sure that
				// obstor.<namespace>.svc.<cluster_domain> is ignored
				// by the bucketDNS style to ensure that path style
				// is available and honored at this domain.
				//
				// All other `<bucket>.<namespace>.svc.<cluster_domain>`
				// makes sure that buckets are routed through this matcher
				// to match for `<bucket>`
				return host != obstor.ObstorReservedBucket+"."+domainName
			}).Host("{bucket:.+}."+domainName).Subrouter())
		} else {
			routers = append(routers, apiRouter.Host("{bucket:.+}."+domainName).Subrouter())
		}
	}
	routers = append(routers, apiRouter.PathPrefix("/{bucket}").Subrouter())

	for _, router := range routers {
		rejectUnsupportedAPIs(router)
		// Object operations
		// HeadObject
		router.Methods(http.MethodHead).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("headobject", obstor.MaxClients(obstor.HTTPTraceAll(api.HeadObjectHandler))))
		// CopyObjectPart
		router.Methods(http.MethodPut).Path("/{object:.+}").
			HeadersRegexp(xhttp.AmzCopySource, ".*?(\\/|%2F).*?").
			HandlerFunc(obstor.CollectAPIStats("copyobjectpart", obstor.MaxClients(obstor.HTTPTraceAll(api.CopyObjectPartHandler)))).
			Queries("partNumber", "{partNumber:[0-9]+}", "uploadId", "{uploadId:.*}")
		// PutObjectPart
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("putobjectpart", obstor.MaxClients(obstor.HTTPTraceHdrs(api.PutObjectPartHandler)))).Queries("partNumber", "{partNumber:[0-9]+}", "uploadId", "{uploadId:.*}")
		// ListObjectParts
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("listobjectparts", obstor.MaxClients(obstor.HTTPTraceAll(api.ListObjectPartsHandler)))).Queries("uploadId", "{uploadId:.*}")
		// CompleteMultipartUpload
		router.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("completemutipartupload", obstor.MaxClients(obstor.HTTPTraceAll(api.CompleteMultipartUploadHandler)))).Queries("uploadId", "{uploadId:.*}")
		// NewMultipartUpload
		router.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("newmultipartupload", obstor.MaxClients(obstor.HTTPTraceAll(api.NewMultipartUploadHandler)))).Queries("uploads", "")
		// AbortMultipartUpload
		router.Methods(http.MethodDelete).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("abortmultipartupload", obstor.MaxClients(obstor.HTTPTraceAll(api.AbortMultipartUploadHandler)))).Queries("uploadId", "{uploadId:.*}")
		// GetObjectACL - this is a dummy call.
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("getobjectacl", obstor.MaxClients(obstor.HTTPTraceHdrs(api.GetObjectACLHandler)))).Queries("acl", "")
		// PutObjectACL - this is a dummy call.
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("putobjectacl", obstor.MaxClients(obstor.HTTPTraceHdrs(api.PutObjectACLHandler)))).Queries("acl", "")
		// GetObjectTagging
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("getobjecttagging", obstor.MaxClients(obstor.HTTPTraceHdrs(api.GetObjectTaggingHandler)))).Queries("tagging", "")
		// PutObjectTagging
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("putobjecttagging", obstor.MaxClients(obstor.HTTPTraceHdrs(api.PutObjectTaggingHandler)))).Queries("tagging", "")
		// DeleteObjectTagging
		router.Methods(http.MethodDelete).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("deleteobjecttagging", obstor.MaxClients(obstor.HTTPTraceHdrs(api.DeleteObjectTaggingHandler)))).Queries("tagging", "")
		// SelectObjectContent
		router.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("selectobjectcontent", obstor.MaxClients(obstor.HTTPTraceHdrs(api.SelectObjectContentHandler)))).Queries("select", "").Queries("select-type", "2")
		// GetObjectRetention
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("getobjectretention", obstor.MaxClients(obstor.HTTPTraceAll(api.GetObjectRetentionHandler)))).Queries("retention", "")
		// GetObjectLegalHold
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("getobjectlegalhold", obstor.MaxClients(obstor.HTTPTraceAll(api.GetObjectLegalHoldHandler)))).Queries("legal-hold", "")
		// GetObject
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("getobject", obstor.MaxClients(obstor.HTTPTraceHdrs(api.GetObjectHandler))))
		// CopyObject
		router.Methods(http.MethodPut).Path("/{object:.+}").HeadersRegexp(xhttp.AmzCopySource, ".*?(\\/|%2F).*?").HandlerFunc(
			obstor.CollectAPIStats("copyobject", obstor.MaxClients(obstor.HTTPTraceAll(api.CopyObjectHandler))))
		// PutObjectRetention
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("putobjectretention", obstor.MaxClients(obstor.HTTPTraceAll(api.PutObjectRetentionHandler)))).Queries("retention", "")
		// PutObjectLegalHold
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("putobjectlegalhold", obstor.MaxClients(obstor.HTTPTraceAll(api.PutObjectLegalHoldHandler)))).Queries("legal-hold", "")

		// PutObject with auto-extract support for zip
		router.Methods(http.MethodPut).Path("/{object:.+}").HeadersRegexp(xhttp.AmzSnowballExtract, "true").HandlerFunc(
			obstor.CollectAPIStats("putobject", obstor.MaxClients(obstor.HTTPTraceHdrs(api.PutObjectExtractHandler))))

		// PutObject
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("putobject", obstor.MaxClients(obstor.HTTPTraceHdrs(api.PutObjectHandler))))

		// DeleteObject
		router.Methods(http.MethodDelete).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("deleteobject", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteObjectHandler))))

		// PostRestoreObject
		router.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(
			obstor.CollectAPIStats("restoreobject", obstor.MaxClients(obstor.HTTPTraceAll(api.PostRestoreObjectHandler)))).Queries("restore", "")

		/// Bucket operations
		// GetBucketLocation
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketlocation", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketLocationHandler)))).Queries("location", "")
		// GetBucketPolicy
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketpolicy", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketPolicyHandler)))).Queries("policy", "")
		// GetBucketLifecycle
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketlifecycle", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketLifecycleHandler)))).Queries("lifecycle", "")
		// GetBucketEncryption
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketencryption", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketEncryptionHandler)))).Queries("encryption", "")
		// GetBucketObjectLockConfig
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketobjectlockconfiguration", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketObjectLockConfigHandler)))).Queries("object-lock", "")
		// GetBucketReplicationConfig
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketreplicationconfiguration", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketReplicationConfigHandler)))).Queries("replication", "")
		// GetBucketVersioning
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketversioning", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketVersioningHandler)))).Queries("versioning", "")
		// GetBucketNotification
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketnotification", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketNotificationHandler)))).Queries("notification", "")
		// ListenNotification
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("listennotification", obstor.MaxClients(obstor.HTTPTraceAll(api.ListenNotificationHandler)))).Queries("events", "{events:.*}")

		// Dummy Bucket Calls
		// GetBucketACL -- this is a dummy call.
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketacl", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketACLHandler)))).Queries("acl", "")
		// PutBucketACL -- this is a dummy call.
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucketacl", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketACLHandler)))).Queries("acl", "")
		// GetBucketCors - this is a dummy call.
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketcors", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketCorsHandler)))).Queries("cors", "")
		// GetBucketWebsiteHandler - this is a dummy call.
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketwebsite", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketWebsiteHandler)))).Queries("website", "")
		// GetBucketAccelerateHandler - this is a dummy call.
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketaccelerate", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketAccelerateHandler)))).Queries("accelerate", "")
		// GetBucketRequestPaymentHandler - this is a dummy call.
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketrequestpayment", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketRequestPaymentHandler)))).Queries("requestPayment", "")
		// GetBucketLoggingHandler - this is a dummy call.
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketlogging", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketLoggingHandler)))).Queries("logging", "")
		// GetBucketTaggingHandler
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbuckettagging", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketTaggingHandler)))).Queries("tagging", "")
		//DeleteBucketWebsiteHandler
		router.Methods(http.MethodDelete).HandlerFunc(
			obstor.CollectAPIStats("deletebucketwebsite", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteBucketWebsiteHandler)))).Queries("website", "")
		// DeleteBucketTaggingHandler
		router.Methods(http.MethodDelete).HandlerFunc(
			obstor.CollectAPIStats("deletebuckettagging", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteBucketTaggingHandler)))).Queries("tagging", "")

		// ListMultipartUploads
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("listmultipartuploads", obstor.MaxClients(obstor.HTTPTraceAll(api.ListMultipartUploadsHandler)))).Queries("uploads", "")
		// ListObjectsV2M
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("listobjectsv2M", obstor.MaxClients(obstor.HTTPTraceAll(api.ListObjectsV2MHandler)))).Queries("list-type", "2", "metadata", "true")
		// ListObjectsV2
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("listobjectsv2", obstor.MaxClients(obstor.HTTPTraceAll(api.ListObjectsV2Handler)))).Queries("list-type", "2")
		// ListObjectVersions
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("listobjectversions", obstor.MaxClients(obstor.HTTPTraceAll(api.ListObjectVersionsHandler)))).Queries("versions", "")
		// GetBucketPolicyStatus
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getpolicystatus", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketPolicyStatusHandler)))).Queries("policyStatus", "")
		// PutBucketLifecycle
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucketlifecycle", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketLifecycleHandler)))).Queries("lifecycle", "")
		// PutBucketReplicationConfig
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucketreplicationconfiguration", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketReplicationConfigHandler)))).Queries("replication", "")
		// PutBucketEncryption
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucketencryption", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketEncryptionHandler)))).Queries("encryption", "")

		// PutBucketPolicy
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucketpolicy", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketPolicyHandler)))).Queries("policy", "")

		// PutBucketObjectLockConfig
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucketobjectlockconfig", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketObjectLockConfigHandler)))).Queries("object-lock", "")
		// PutBucketTaggingHandler
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbuckettagging", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketTaggingHandler)))).Queries("tagging", "")
		// PutBucketVersioning
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucketversioning", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketVersioningHandler)))).Queries("versioning", "")
		// PutBucketNotification
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucketnotification", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketNotificationHandler)))).Queries("notification", "")
		// PutBucket
		router.Methods(http.MethodPut).HandlerFunc(
			obstor.CollectAPIStats("putbucket", obstor.MaxClients(obstor.HTTPTraceAll(api.PutBucketHandler))))
		// HeadBucket
		router.Methods(http.MethodHead).HandlerFunc(
			obstor.CollectAPIStats("headbucket", obstor.MaxClients(obstor.HTTPTraceAll(api.HeadBucketHandler))))
		// PostPolicy
		router.Methods(http.MethodPost).HeadersRegexp(xhttp.ContentType, "multipart/form-data*").HandlerFunc(
			obstor.CollectAPIStats("postpolicybucket", obstor.MaxClients(obstor.HTTPTraceHdrs(api.PostPolicyBucketHandler))))
		// DeleteMultipleObjects
		router.Methods(http.MethodPost).HandlerFunc(
			obstor.CollectAPIStats("deletemultipleobjects", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteMultipleObjectsHandler)))).Queries("delete", "")
		// DeleteBucketPolicy
		router.Methods(http.MethodDelete).HandlerFunc(
			obstor.CollectAPIStats("deletebucketpolicy", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteBucketPolicyHandler)))).Queries("policy", "")
		// DeleteBucketReplication
		router.Methods(http.MethodDelete).HandlerFunc(
			obstor.CollectAPIStats("deletebucketreplicationconfiguration", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteBucketReplicationConfigHandler)))).Queries("replication", "")
		// DeleteBucketLifecycle
		router.Methods(http.MethodDelete).HandlerFunc(
			obstor.CollectAPIStats("deletebucketlifecycle", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteBucketLifecycleHandler)))).Queries("lifecycle", "")
		// DeleteBucketEncryption
		router.Methods(http.MethodDelete).HandlerFunc(
			obstor.CollectAPIStats("deletebucketencryption", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteBucketEncryptionHandler)))).Queries("encryption", "")
		// DeleteBucket
		router.Methods(http.MethodDelete).HandlerFunc(
			obstor.CollectAPIStats("deletebucket", obstor.MaxClients(obstor.HTTPTraceAll(api.DeleteBucketHandler))))
		// Obstor extension API for replication.
		//
		// GetBucketReplicationMetrics
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("getbucketreplicationmetrics", obstor.MaxClients(obstor.HTTPTraceAll(api.GetBucketReplicationMetricsHandler)))).Queries("replication-metrics", "")

		// S3 ListObjectsV1 (Legacy)
		router.Methods(http.MethodGet).HandlerFunc(
			obstor.CollectAPIStats("listobjectsv1", obstor.MaxClients(obstor.HTTPTraceAll(api.ListObjectsV1Handler))))

	}

	/// Root operation

	// ListenNotification
	apiRouter.Methods(http.MethodGet).Path(obstor.SlashSeparator).HandlerFunc(
		obstor.CollectAPIStats("listennotification", obstor.MaxClients(obstor.HTTPTraceAll(api.ListenNotificationHandler)))).Queries("events", "{events:.*}")

	// ListBuckets
	apiRouter.Methods(http.MethodGet).Path(obstor.SlashSeparator).HandlerFunc(
		obstor.CollectAPIStats("listbuckets", obstor.MaxClients(obstor.HTTPTraceAll(api.ListBucketsHandler))))

	// S3 browser with signature v4 adds '//' for ListBuckets request, so rather
	// than failing with UnknownAPIRequest we simply handle it for now.
	apiRouter.Methods(http.MethodGet).Path(obstor.SlashSeparator + obstor.SlashSeparator).HandlerFunc(
		obstor.CollectAPIStats("listbuckets", obstor.MaxClients(obstor.HTTPTraceAll(api.ListBucketsHandler))))

	// Reject non-GET methods on root path with 405 MethodNotAllowed.
	apiRouter.Methods(http.MethodPut, http.MethodPost, http.MethodDelete).Path(obstor.SlashSeparator).HandlerFunc(
		obstor.CollectAPIStats("methodnotallowed", obstor.HTTPTraceAll(obstor.MethodNotAllowedHandler("S3"))))

	// If none of the routes match add default error handler routes
	apiRouter.NotFoundHandler = obstor.CollectAPIStats("notfound", obstor.HTTPTraceAll(obstor.ErrorResponseHandler))
	apiRouter.MethodNotAllowedHandler = obstor.CollectAPIStats("methodnotallowed", obstor.HTTPTraceAll(obstor.MethodNotAllowedHandler("S3")))

}

// CorsHandler handler for CORS (Cross Origin Resource Sharing)
func CorsHandler(handler http.Handler) http.Handler {
	commonS3Headers := []string{
		xhttp.Date,
		xhttp.ETag,
		xhttp.ServerInfo,
		xhttp.Connection,
		xhttp.AcceptRanges,
		xhttp.ContentRange,
		xhttp.ContentEncoding,
		xhttp.ContentLength,
		xhttp.ContentType,
		xhttp.ContentDisposition,
		xhttp.LastModified,
		xhttp.ContentLanguage,
		xhttp.CacheControl,
		xhttp.RetryAfter,
		xhttp.AmzBucketRegion,
		xhttp.Expires,
		"X-Amz*",
		"x-amz*",
		"*",
	}

	return cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			for _, allowedOrigin := range obstor.GlobalGetCorsAllowOrigins() {
				if wildcard.MatchSimple(allowedOrigin, origin) {
					return true
				}
			}
			return false
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPut,
			http.MethodHead,
			http.MethodPost,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodPatch,
		},
		AllowedHeaders:   commonS3Headers,
		ExposedHeaders:   commonS3Headers,
		AllowCredentials: true,
	}).Handler(handler)
}
