package toolkit

import (
	"net/http"
	"runtime"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter new http router with default route
func NewRouter() *httprouter.Router {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(notFoundHandler)
	router.PanicHandler = panicHandler

	router.GET("/", defaultHandler)
	router.GET("/health", defaultHandler)
	router.Handler("GET", "/metrics", promhttp.Handler())

	return router
}

func defaultHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ver := map[string]string{"version": "0.2.0", "comments": "chuangxin1.com API"}
	HTTPWriteJSON(w, RowReplyData(ver))
}

func panicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	e := err.(runtime.Error)

	HTTPWriteJSON(w, ErrReplyData(ErrException, e.Error()))
}

func notFoundHandler(w http.ResponseWriter, req *http.Request) {
	HTTPWriteJSON(w, ErrReplyData(ErrNotFound, `NotFound`))
}
