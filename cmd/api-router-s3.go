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
	"net"
	"net/http"

	xhttp "github.com/obstor/obstor/cmd/http"
	"github.com/obstor/obstor/pkg/wildcard"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Return the request host, stripping port if present.
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
			HandlerFunc(collectAPIStats(r.api, httpTraceAll(NotImplementedHandler))).
			Queries(r.queries...)
		if r.path != "" {
			t.Path(r.path)
		}
	}
}

// Register S3 API routes directly (test fallback).
func registerAPIRouterDirect(router *mux.Router) {
	api := objectAPIHandlers{
		ObjectAPI: newObjectLayerFn,
		CacheAPI:  newCachedObjectLayerFn,
	}

	apiRouter := router.PathPrefix(SlashSeparator).Subrouter()

	var routers []*mux.Router
	for _, domainName := range globalDomainNames {
		if IsKubernetes() {
			routers = append(routers, apiRouter.MatcherFunc(func(r *http.Request, match *mux.RouteMatch) bool {
				host, _, err := net.SplitHostPort(getHost(r))
				if err != nil {
					host = r.Host
				}
				// Skip obstor.<domain> so path-style works on the base k8s service domain.
				return host != obstorReservedBucket+"."+domainName
			}).Host("{bucket:.+}."+domainName).Subrouter())
		} else {
			routers = append(routers, apiRouter.Host("{bucket:.+}."+domainName).Subrouter())
		}
	}
	routers = append(routers, apiRouter.PathPrefix("/{bucket}").Subrouter())

	for _, router := range routers {
		rejectUnsupportedAPIs(router)
		// Object operations
		router.Methods(http.MethodHead).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("headobject", maxClients(httpTraceAll(api.HeadObjectHandler))))
		router.Methods(http.MethodPut).Path("/{object:.+}").
			HeadersRegexp(xhttp.AmzCopySource, ".*?(\\/|%2F).*?").
			HandlerFunc(collectAPIStats("copyobjectpart", maxClients(httpTraceAll(api.CopyObjectPartHandler)))).
			Queries("partNumber", "{partNumber:[0-9]+}", "uploadId", "{uploadId:.*}")
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("putobjectpart", maxClients(httpTraceHdrs(api.PutObjectPartHandler)))).Queries("partNumber", "{partNumber:[0-9]+}", "uploadId", "{uploadId:.*}")
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("listobjectparts", maxClients(httpTraceAll(api.ListObjectPartsHandler)))).Queries("uploadId", "{uploadId:.*}")
		router.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("completemutipartupload", maxClients(httpTraceAll(api.CompleteMultipartUploadHandler)))).Queries("uploadId", "{uploadId:.*}")
		router.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("newmultipartupload", maxClients(httpTraceAll(api.NewMultipartUploadHandler)))).Queries("uploads", "")
		router.Methods(http.MethodDelete).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("abortmultipartupload", maxClients(httpTraceAll(api.AbortMultipartUploadHandler)))).Queries("uploadId", "{uploadId:.*}")
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("getobjectacl", maxClients(httpTraceHdrs(api.GetObjectACLHandler)))).Queries("acl", "")
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("putobjectacl", maxClients(httpTraceHdrs(api.PutObjectACLHandler)))).Queries("acl", "")
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("getobjecttagging", maxClients(httpTraceHdrs(api.GetObjectTaggingHandler)))).Queries("tagging", "")
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("putobjecttagging", maxClients(httpTraceHdrs(api.PutObjectTaggingHandler)))).Queries("tagging", "")
		router.Methods(http.MethodDelete).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("deleteobjecttagging", maxClients(httpTraceHdrs(api.DeleteObjectTaggingHandler)))).Queries("tagging", "")
		router.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("selectobjectcontent", maxClients(httpTraceHdrs(api.SelectObjectContentHandler)))).Queries("select", "").Queries("select-type", "2")
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("getobjectretention", maxClients(httpTraceAll(api.GetObjectRetentionHandler)))).Queries("retention", "")
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("getobjectlegalhold", maxClients(httpTraceAll(api.GetObjectLegalHoldHandler)))).Queries("legal-hold", "")
		router.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("getobject", maxClients(httpTraceHdrs(api.GetObjectHandler))))
		router.Methods(http.MethodPut).Path("/{object:.+}").HeadersRegexp(xhttp.AmzCopySource, ".*?(\\/|%2F).*?").HandlerFunc(
			collectAPIStats("copyobject", maxClients(httpTraceAll(api.CopyObjectHandler))))
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("putobjectretention", maxClients(httpTraceAll(api.PutObjectRetentionHandler)))).Queries("retention", "")
		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("putobjectlegalhold", maxClients(httpTraceAll(api.PutObjectLegalHoldHandler)))).Queries("legal-hold", "")

		router.Methods(http.MethodPut).Path("/{object:.+}").HeadersRegexp(xhttp.AmzSnowballExtract, "true").HandlerFunc(
			collectAPIStats("putobject", maxClients(httpTraceHdrs(api.PutObjectExtractHandler))))

		router.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("putobject", maxClients(httpTraceHdrs(api.PutObjectHandler))))

		router.Methods(http.MethodDelete).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("deleteobject", maxClients(httpTraceAll(api.DeleteObjectHandler))))

		router.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(
			collectAPIStats("restoreobject", maxClients(httpTraceAll(api.PostRestoreObjectHandler)))).Queries("restore", "")

		/// Bucket operations
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketlocation", maxClients(httpTraceAll(api.GetBucketLocationHandler)))).Queries("location", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketpolicy", maxClients(httpTraceAll(api.GetBucketPolicyHandler)))).Queries("policy", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketlifecycle", maxClients(httpTraceAll(api.GetBucketLifecycleHandler)))).Queries("lifecycle", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketencryption", maxClients(httpTraceAll(api.GetBucketEncryptionHandler)))).Queries("encryption", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketobjectlockconfiguration", maxClients(httpTraceAll(api.GetBucketObjectLockConfigHandler)))).Queries("object-lock", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketreplicationconfiguration", maxClients(httpTraceAll(api.GetBucketReplicationConfigHandler)))).Queries("replication", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketversioning", maxClients(httpTraceAll(api.GetBucketVersioningHandler)))).Queries("versioning", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketnotification", maxClients(httpTraceAll(api.GetBucketNotificationHandler)))).Queries("notification", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("listennotification", maxClients(httpTraceAll(api.ListenNotificationHandler)))).Queries("events", "{events:.*}")

		// Dummy Bucket Calls
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketacl", maxClients(httpTraceAll(api.GetBucketACLHandler)))).Queries("acl", "")
		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucketacl", maxClients(httpTraceAll(api.PutBucketACLHandler)))).Queries("acl", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketcors", maxClients(httpTraceAll(api.GetBucketCorsHandler)))).Queries("cors", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketwebsite", maxClients(httpTraceAll(api.GetBucketWebsiteHandler)))).Queries("website", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketaccelerate", maxClients(httpTraceAll(api.GetBucketAccelerateHandler)))).Queries("accelerate", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketrequestpayment", maxClients(httpTraceAll(api.GetBucketRequestPaymentHandler)))).Queries("requestPayment", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketlogging", maxClients(httpTraceAll(api.GetBucketLoggingHandler)))).Queries("logging", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbuckettagging", maxClients(httpTraceAll(api.GetBucketTaggingHandler)))).Queries("tagging", "")
		router.Methods(http.MethodDelete).HandlerFunc(
			collectAPIStats("deletebucketwebsite", maxClients(httpTraceAll(api.DeleteBucketWebsiteHandler)))).Queries("website", "")
		router.Methods(http.MethodDelete).HandlerFunc(
			collectAPIStats("deletebuckettagging", maxClients(httpTraceAll(api.DeleteBucketTaggingHandler)))).Queries("tagging", "")

		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("listmultipartuploads", maxClients(httpTraceAll(api.ListMultipartUploadsHandler)))).Queries("uploads", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("listobjectsv2M", maxClients(httpTraceAll(api.ListObjectsV2MHandler)))).Queries("list-type", "2", "metadata", "true")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("listobjectsv2", maxClients(httpTraceAll(api.ListObjectsV2Handler)))).Queries("list-type", "2")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("listobjectversions", maxClients(httpTraceAll(api.ListObjectVersionsHandler)))).Queries("versions", "")
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getpolicystatus", maxClients(httpTraceAll(api.GetBucketPolicyStatusHandler)))).Queries("policyStatus", "")
		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucketlifecycle", maxClients(httpTraceAll(api.PutBucketLifecycleHandler)))).Queries("lifecycle", "")
		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucketreplicationconfiguration", maxClients(httpTraceAll(api.PutBucketReplicationConfigHandler)))).Queries("replication", "")
		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucketencryption", maxClients(httpTraceAll(api.PutBucketEncryptionHandler)))).Queries("encryption", "")

		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucketpolicy", maxClients(httpTraceAll(api.PutBucketPolicyHandler)))).Queries("policy", "")

		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucketobjectlockconfig", maxClients(httpTraceAll(api.PutBucketObjectLockConfigHandler)))).Queries("object-lock", "")
		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbuckettagging", maxClients(httpTraceAll(api.PutBucketTaggingHandler)))).Queries("tagging", "")
		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucketversioning", maxClients(httpTraceAll(api.PutBucketVersioningHandler)))).Queries("versioning", "")
		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucketnotification", maxClients(httpTraceAll(api.PutBucketNotificationHandler)))).Queries("notification", "")
		router.Methods(http.MethodPut).HandlerFunc(
			collectAPIStats("putbucket", maxClients(httpTraceAll(api.PutBucketHandler))))
		router.Methods(http.MethodHead).HandlerFunc(
			collectAPIStats("headbucket", maxClients(httpTraceAll(api.HeadBucketHandler))))
		router.Methods(http.MethodPost).HeadersRegexp(xhttp.ContentType, "multipart/form-data*").HandlerFunc(
			collectAPIStats("postpolicybucket", maxClients(httpTraceHdrs(api.PostPolicyBucketHandler))))
		router.Methods(http.MethodPost).HandlerFunc(
			collectAPIStats("deletemultipleobjects", maxClients(httpTraceAll(api.DeleteMultipleObjectsHandler)))).Queries("delete", "")
		router.Methods(http.MethodDelete).HandlerFunc(
			collectAPIStats("deletebucketpolicy", maxClients(httpTraceAll(api.DeleteBucketPolicyHandler)))).Queries("policy", "")
		router.Methods(http.MethodDelete).HandlerFunc(
			collectAPIStats("deletebucketreplicationconfiguration", maxClients(httpTraceAll(api.DeleteBucketReplicationConfigHandler)))).Queries("replication", "")
		router.Methods(http.MethodDelete).HandlerFunc(
			collectAPIStats("deletebucketlifecycle", maxClients(httpTraceAll(api.DeleteBucketLifecycleHandler)))).Queries("lifecycle", "")
		router.Methods(http.MethodDelete).HandlerFunc(
			collectAPIStats("deletebucketencryption", maxClients(httpTraceAll(api.DeleteBucketEncryptionHandler)))).Queries("encryption", "")
		router.Methods(http.MethodDelete).HandlerFunc(
			collectAPIStats("deletebucket", maxClients(httpTraceAll(api.DeleteBucketHandler))))
		// Obstor extension API for replication.
		//
		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("getbucketreplicationmetrics", maxClients(httpTraceAll(api.GetBucketReplicationMetricsHandler)))).Queries("replication-metrics", "")

		router.Methods(http.MethodGet).HandlerFunc(
			collectAPIStats("listobjectsv1", maxClients(httpTraceAll(api.ListObjectsV1Handler))))

	}

	/// Root operation

	// ListenNotification
	apiRouter.Methods(http.MethodGet).Path(SlashSeparator).HandlerFunc(
		collectAPIStats("listennotification", maxClients(httpTraceAll(api.ListenNotificationHandler)))).Queries("events", "{events:.*}")

	// ListBuckets
	apiRouter.Methods(http.MethodGet).Path(SlashSeparator).HandlerFunc(
		collectAPIStats("listbuckets", maxClients(httpTraceAll(api.ListBucketsHandler))))

	// S3 browser adds '//' for ListBuckets; handle it rather than 400.
	apiRouter.Methods(http.MethodGet).Path(SlashSeparator + SlashSeparator).HandlerFunc(
		collectAPIStats("listbuckets", maxClients(httpTraceAll(api.ListBucketsHandler))))

	// Explicit 405 for non-GET on root; gorilla/mux would 400 otherwise.
	apiRouter.Methods(http.MethodPut, http.MethodPost, http.MethodDelete).Path(SlashSeparator).HandlerFunc(
		collectAPIStats("methodnotallowed", httpTraceAll(methodNotAllowedHandler("S3"))))

	// If none of the routes match add default error handler routes
	apiRouter.NotFoundHandler = collectAPIStats("notfound", httpTraceAll(errorResponseHandler))
	apiRouter.MethodNotAllowedHandler = collectAPIStats("methodnotallowed", httpTraceAll(methodNotAllowedHandler("S3")))

}

// Configure CORS for S3 API responses.
func corsHandlerDirect(handler http.Handler) http.Handler {
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
			for _, allowedOrigin := range GlobalGetCorsAllowOrigins() {
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
