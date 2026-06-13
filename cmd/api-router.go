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
	"net/http"

	"github.com/gorilla/mux"
	xhttp "github.com/obstor/obstor/cmd/http"
)

func newHTTPServerFn() *xhttp.Server {
	globalObjLayerMutex.RLock()
	defer globalObjLayerMutex.RUnlock()
	return globalHTTPServer
}

func setHTTPServer(h *xhttp.Server) {
	globalObjLayerMutex.Lock()
	globalHTTPServer = h
	globalObjLayerMutex.Unlock()
}

func newObjectLayerFn() ObjectLayer {
	globalObjLayerMutex.RLock()
	defer globalObjLayerMutex.RUnlock()
	return globalObjectAPI
}

func newCachedObjectLayerFn() CacheObjectLayer {
	globalObjLayerMutex.RLock()
	defer globalObjLayerMutex.RUnlock()
	return globalCacheObjectAPI
}

func setCacheObjectLayer(c CacheObjectLayer) {
	globalObjLayerMutex.Lock()
	globalCacheObjectAPI = c
	globalObjLayerMutex.Unlock()
}

func setObjectLayer(o ObjectLayer) {
	globalObjLayerMutex.Lock()
	globalObjectAPI = o
	globalObjLayerMutex.Unlock()
}

// objectAPIHandler implements and provides http handlers for S3 API.
type objectAPIHandlers struct {
	ObjectAPI func() ObjectLayer
	CacheAPI  func() CacheObjectLayer
}

// NotImplementedHandler responds with ErrNotImplemented.
func NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(r.Context(), w, errorCodes.ToAPIErr(ErrNotImplemented), r.URL, guessIsBrowserReq(r))
}

// Register S3 API routes via protocol package
func registerAPIRouter(router *mux.Router) {
	if GlobalS3RegisterAPIRouterFn != nil {
		GlobalS3RegisterAPIRouterFn(router)
		return
	}
	// Fallback for test binaries without the s3 package.
	registerAPIRouterDirect(router)
}

// Delegate to the S3 package or falls back to direct.
func corsHandler(handler http.Handler) http.Handler {
	if GlobalS3CorsHandlerFn != nil {
		return GlobalS3CorsHandlerFn(handler)
	}
	return corsHandlerDirect(handler)
}
