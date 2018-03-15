package toolkit

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
)

// EncodeResponseFunc encodes the passed response object to the HTTP response
// writer. It's designed to be used in HTTP servers, for server-side
// endpoints. One straightforward EncodeResponseFunc could be something that
// JSON encodes the object directly to the response body.
type EncodeResponseFunc func(context.Context,
	http.ResponseWriter, interface{}) error

// DecodeResponseFunc extracts a user-domain response object from an HTTP
// response object. It's designed to be used in HTTP clients, for client-side
// endpoints. One straightforward DecodeResponseFunc could be something that
// JSON decodes from the response body to the concrete response type.
type DecodeResponseFunc func(
	context.Context, *http.Response) (response interface{}, err error)

// HTTPTansportServerOptions default http transport options
func HTTPTansportServerOptions(
	logger log.Logger) []httptransport.ServerOption {
	return []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(HTTPEncodeError),
		httptransport.ServerBefore(PopulateRequestContext),
	}
}

// NewHTTPTansportServer new server hander
func NewHTTPTansportServer(
	hasAuth bool,
	e endpoint.Endpoint,
	dec httptransport.DecodeRequestFunc,
	enc EncodeResponseFunc,
	logger log.Logger) *httptransport.Server {
	options := HTTPTansportServerOptions(logger)
	if hasAuth {
		e = AuthMiddleware()(e)
	}
	return httptransport.NewServer(
		e,
		dec,
		httptransport.EncodeResponseFunc(enc),
		options...,
	)
}

// StartServer new server and start
func StartServer(
	addr string,
	router http.Handler,
	readTimeout, writeTimeout time.Duration,
	maxHeaderBytes int,
	logger log.Logger) {
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("Protocol", "HTTP", "addr", addr)
		errc <- server.ListenAndServe()
		logger.Log("Exit server", "Quit")
	}()

	// Run!
	logger.Log("Exit", <-errc)
}
