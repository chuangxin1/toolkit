package toolkit

import (
	"net/http"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// ServiceHander hander
type ServiceHander struct {
	Method string
	Router string
	Hander http.Handler
}

// EndpointHander endpoint hander
type EndpointHander struct {
	HasAuth  bool
	Method   string
	Router   string
	Dec      httptransport.DecodeRequestFunc
	Enc      EncodeResponseFunc
	Endpoint endpoint.Endpoint
}
