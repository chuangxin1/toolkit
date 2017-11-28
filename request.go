package toolkit

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	"github.com/sony/gobreaker"

	httptransport "github.com/go-kit/kit/transport/http"
)

// CopyURL copy url
func CopyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

// ClientEncodeGetRequest client get encode request
func ClientEncodeGetRequest(ctx context.Context, req *http.Request, request interface{}) error {
	values := URLValuesStruct(request)

	auth, ok := ctx.Value(ContextKeyRequestAuthorization).(string)
	if ok {
		req.Header.Set(HTTPHeaderAuthorization, auth)
	}

	token, _ := ctx.Value(ContextKeyAccessToken).(string)
	if token != "" {
		values.Set(VarUserAuthorization, token)
	}

	req.URL.RawQuery = values.Encode()
	return nil
}

// ClientEncodeJSONRequest is an EncodeRequestFunc that serializes the request as a
// JSON object to the Request body. Many JSON-over-HTTP services can use it as
// a sensible default. If the request implements Headerer, the provided headers
// will be applied to the request.
func ClientEncodeJSONRequest(ctx context.Context, req *http.Request, request interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	if headerer, ok := request.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			req.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	auth, ok := ctx.Value(ContextKeyRequestAuthorization).(string)
	if ok {
		req.Header.Set("Authorization", auth)
	}
	values := url.Values{}
	token, _ := ctx.Value(ContextKeyAccessToken).(string)
	if token != "" {
		values.Set(VarUserAuthorization, token)
	}
	req.URL.RawQuery = values.Encode()

	var b bytes.Buffer
	req.Body = ioutil.NopCloser(&b)
	return json.NewEncoder(&b).Encode(request)
}

// ClientRequestEndpoint client request Endpoint
func ClientRequestEndpoint(
	ctx context.Context,
	u *url.URL,
	method, router string) endpoint.Endpoint {
	var e endpoint.Endpoint
	options := []httptransport.ClientOption{}
	var enc httptransport.EncodeRequestFunc

	switch method {
	case "POST":
		enc = ClientEncodeJSONRequest
	default:
		enc = ClientEncodeGetRequest
	}
	e = httptransport.NewClient(
		method,
		CopyURL(u, router),
		enc,
		HTTPDecodeResponse,
		options...,
	).Endpoint()

	e = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
	//e = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(e)

	return e
}

// ClientLoadBalancer load balance
func ClientLoadBalancer(endpoints sd.FixedEndpointer) endpoint.Endpoint {
	balancer := lb.NewRoundRobin(endpoints)
	return lb.Retry(retryMax, retryTimeout, balancer)
}