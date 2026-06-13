package s3

import (
	"net/http"

	"github.com/gorilla/mux"
	obstor "github.com/obstor/obstor/cmd"
)

// RegisterTestAPI registers specific S3 API handler functions on test routers.
func RegisterTestAPI(apiRouter *mux.Router, bucketRouter *mux.Router, apiFunctions ...string) {
	api := Handlers{ObjectAPIHandlers: obstor.ObjectAPIHandlers{
		ObjectAPI: obstor.NewObjectLayerFn,
		CacheAPI:  obstor.NewCachedObjectLayerFn,
	}}

	// Register ListBuckets handler.
	apiRouter.Methods(http.MethodGet).HandlerFunc(api.ListBucketsHandler)

	for _, apiFunction := range apiFunctions {
		switch apiFunction {
		case "PostPolicy":
			bucketRouter.Methods(http.MethodPost).HeadersRegexp("Content-Type", "multipart/form-data*").HandlerFunc(api.PostPolicyBucketHandler)
		case "HeadObject":
			bucketRouter.Methods("Head").Path("/{object:.+}").HandlerFunc(api.HeadObjectHandler)
		case "GetObject":
			bucketRouter.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(api.GetObjectHandler)
		case "PutObject":
			bucketRouter.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(api.PutObjectHandler)
		case "DeleteObject":
			bucketRouter.Methods(http.MethodDelete).Path("/{object:.+}").HandlerFunc(api.DeleteObjectHandler)
		case "CopyObject":
			bucketRouter.Methods(http.MethodPut).Path("/{object:.+}").HeadersRegexp("X-Amz-Copy-Source", ".*?(\\/|%2F).*?").HandlerFunc(api.CopyObjectHandler)
		case "PutBucketPolicy":
			bucketRouter.Methods(http.MethodPut).HandlerFunc(api.PutBucketPolicyHandler).Queries("policy", "")
		case "DeleteBucketPolicy":
			bucketRouter.Methods(http.MethodDelete).HandlerFunc(api.DeleteBucketPolicyHandler).Queries("policy", "")
		case "GetBucketPolicy":
			bucketRouter.Methods(http.MethodGet).HandlerFunc(api.GetBucketPolicyHandler).Queries("policy", "")
		case "GetBucketLifecycle":
			bucketRouter.Methods(http.MethodGet).HandlerFunc(api.GetBucketLifecycleHandler).Queries("lifecycle", "")
		case "PutBucketLifecycle":
			bucketRouter.Methods(http.MethodPut).HandlerFunc(api.PutBucketLifecycleHandler).Queries("lifecycle", "")
		case "DeleteBucketLifecycle":
			bucketRouter.Methods(http.MethodDelete).HandlerFunc(api.DeleteBucketLifecycleHandler).Queries("lifecycle", "")
		case "GetBucketLocation":
			bucketRouter.Methods(http.MethodGet).HandlerFunc(api.GetBucketLocationHandler).Queries("location", "")
		case "HeadBucket":
			bucketRouter.Methods(http.MethodHead).HandlerFunc(api.HeadBucketHandler)
		case "DeleteMultipleObjects":
			bucketRouter.Methods(http.MethodPost).HandlerFunc(api.DeleteMultipleObjectsHandler).Queries("delete", "")
		case "NewMultipart":
			bucketRouter.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(api.NewMultipartUploadHandler).Queries("uploads", "")
		case "CopyObjectPart":
			bucketRouter.Methods(http.MethodPut).Path("/{object:.+}").HeadersRegexp("X-Amz-Copy-Source", ".*?(\\/|%2F).*?").HandlerFunc(api.CopyObjectPartHandler).Queries("partNumber", "{partNumber:[0-9]+}", "uploadId", "{uploadId:.*}")
		case "PutObjectPart":
			bucketRouter.Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(api.PutObjectPartHandler).Queries("partNumber", "{partNumber:[0-9]+}", "uploadId", "{uploadId:.*}")
		case "ListObjectParts":
			bucketRouter.Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(api.ListObjectPartsHandler).Queries("uploadId", "{uploadId:.*}")
		case "ListMultipartUploads":
			bucketRouter.Methods(http.MethodGet).HandlerFunc(api.ListMultipartUploadsHandler).Queries("uploads", "")
		case "CompleteMultipart":
			bucketRouter.Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(api.CompleteMultipartUploadHandler).Queries("uploadId", "{uploadId:.*}")
		case "AbortMultipart":
			bucketRouter.Methods(http.MethodDelete).Path("/{object:.+}").HandlerFunc(api.AbortMultipartUploadHandler).Queries("uploadId", "{uploadId:.*}")
		case "GetBucketNotification":
			bucketRouter.Methods(http.MethodGet).HandlerFunc(api.GetBucketNotificationHandler).Queries("notification", "")
		case "PutBucketNotification":
			bucketRouter.Methods(http.MethodPut).HandlerFunc(api.PutBucketNotificationHandler).Queries("notification", "")
		case "ListenNotification":
			bucketRouter.Methods(http.MethodGet).HandlerFunc(api.ListenNotificationHandler).Queries("events", "{events:.*}")
		case "RemoveBucket":
			bucketRouter.Methods(http.MethodDelete).HandlerFunc(api.DeleteBucketHandler)
		}
	}
}
