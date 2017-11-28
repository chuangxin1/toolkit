package toolkit

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/julienschmidt/httprouter"
)

// HTTPServerOptions default http transport options
func HTTPServerOptions(logger log.Logger) []httptransport.ServerOption {
	return []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(HTTPEncodeError),
		httptransport.ServerBefore(PopulateRequestContext),
	}
}

// StartServer new server and start
func StartServer(
	addr string,
	router *httprouter.Router,
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
