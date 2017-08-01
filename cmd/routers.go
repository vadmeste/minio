/*
 * Minio Cloud Storage, (C) 2015, 2016 Minio, Inc.
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
	"io"
	stdlog "log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"

	router "github.com/gorilla/mux"
)

func newObjectLayerFn() (layer ObjectLayer) {
	globalObjLayerMutex.RLock()
	layer = globalObjectAPI
	globalObjLayerMutex.RUnlock()
	return
}

// Composed function registering routers for only distributed XL setup.
func registerDistXLRouters(mux *router.Router, endpoints EndpointList) error {
	// Register storage rpc router only if its a distributed setup.
	err := registerStorageRPCRouters(mux, endpoints)
	if err != nil {
		return err
	}

	// Register distributed namespace lock.
	err = registerDistNSLockRouter(mux, endpoints)
	if err != nil {
		return err
	}

	// Register S3 peer communication router.
	err = registerS3PeerRPCRouter(mux)
	if err != nil {
		return err
	}

	// Register RPC router for web related calls.
	return registerBrowserPeerRPCRouter(mux)
}

type loggingHandler struct {
	writer  io.Writer
	handler http.Handler
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	timeBanner := "<<" + time.Now().String() + ">>\n\n"
	h.writer.Write([]byte(timeBanner))

	logMaxLen := 1024 * 2

	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
	}

	if len(requestDump) > logMaxLen {
		requestDump = append(requestDump[:logMaxLen], []byte("...")...)
	}
	h.writer.Write(requestDump)
	h.writer.Write([]byte("\n\n"))

	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, r)

	respDump, err := httputil.DumpResponse(rec.Result(), true)
	if err != nil {
		stdlog.Fatal(err)
	}

	if len(respDump) > logMaxLen {
		respDump = append(respDump[:logMaxLen], []byte("...")...)
	}
	h.writer.Write(respDump)
	h.writer.Write([]byte("\n\n"))

	// we copy the captured response headers to our new response
	for k, v := range rec.Header() {
		w.Header()[k] = v
	}

	// grab the captured response body
	w.WriteHeader(rec.Code)
	w.Write(rec.Body.Bytes())
}

// Enables logging handler..
func setLoggingHandler(h http.Handler) http.Handler {
	logDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		stdlog.Fatal(err)
	}
	logFile, err := os.Create(filepath.Join(logDir, os.Args[0]+"-ws-log-"+time.Now().Format("2006-01-02T15:04:05")))
	if err != nil {
		stdlog.Fatal(err)
	}
	return loggingHandler{writer: logFile, handler: h}
}

// configureServer handler returns final handler for the http server.
func configureServerHandler(endpoints EndpointList) (http.Handler, error) {
	// Initialize router. `SkipClean(true)` stops gorilla/mux from
	// normalizing URL path minio/minio#3256
	mux := router.NewRouter().SkipClean(true)

	// Initialize distributed NS lock.
	if globalIsDistXL {
		registerDistXLRouters(mux, endpoints)
	}

	// Add Admin RPC router
	err := registerAdminRPCRouter(mux)
	if err != nil {
		return nil, err
	}

	// Register web router when its enabled.
	if globalIsBrowserEnabled {
		if err := registerWebRouter(mux); err != nil {
			return nil, err
		}
	}

	// Add Admin router.
	registerAdminRouter(mux)

	// Add API router.
	registerAPIRouter(mux)

	// List of some generic handlers which are applied for all incoming requests.
	var handlerFns = []HandlerFunc{
		// Validate all the incoming paths.
		setPathValidityHandler,
		// Network statistics
		setHTTPStatsHandler,
		// Limits all requests size to a maximum fixed limit
		setRequestSizeLimitHandler,
		// Adds 'crossdomain.xml' policy handler to serve legacy flash clients.
		setCrossDomainPolicy,
		// Redirect some pre-defined browser request paths to a static location prefix.
		setBrowserRedirectHandler,
		// Validates if incoming request is for restricted buckets.
		setPrivateBucketHandler,
		// Adds cache control for all browser requests.
		setBrowserCacheControlHandler,
		// Validates all incoming requests to have a valid date header.
		setTimeValidityHandler,
		// CORS setting for all browser API requests.
		setCorsHandler,
		// Validates all incoming URL resources, for invalid/unsupported
		// resources client receives a HTTP error.
		setIgnoreResourcesHandler,
		// Auth handler verifies incoming authorization headers and
		// routes them accordingly. Client receives a HTTP error for
		// invalid/unsupported signatures.
		setAuthHandler,
		// Add logging handler.
		setLoggingHandler,
		// Add new handlers here.
	}

	// Register rest of the handlers.
	return registerHandlers(mux, handlerFns...), nil
}
