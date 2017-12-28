package toolkit

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/chuangxin1/httprouter"
	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	"github.com/sony/gobreaker"

	httptransport "github.com/go-kit/kit/transport/http"
)

// RouteVars returns the route variables for the current request, if any.
func RouteVars(ctx context.Context) map[string]string {
	return httprouter.ContextVars(ctx)
}

// ContextVars returns the
func ContextVars(ctx context.Context, key interface{}) interface{} {
	return ctx.Value(key)
}

// CopyURL copy url
func CopyURL(base *url.URL, path string) *url.URL {
	next := *base
	next.Path = path
	return &next
}

func routePath(ctx context.Context, req *http.Request) {
	path := req.URL.Path
	prefix, _ := ctx.Value(ContextKeyGateWayPrefix).(string)
	routePath := httprouter.ContextRoutePath(ctx)
	if prefix != "" && routePath != "" {
		if strings.Contains(routePath, prefix) {
			absolutePath := routePath[len(prefix):]
			if absolutePath != path {
				req.URL.Path = absolutePath
			}
		}
	}
}

// ClientEncodeGetRequest client get encode request
func ClientEncodeGetRequest(
	ctx context.Context,
	req *http.Request, request interface{}) error {
	values := URLValuesStruct(request)

	auth, ok := ctx.Value(ContextKeyRequestAuthorization).(string)
	if ok {
		req.Header.Set(HTTPHeaderAuthorization, auth)
	}
	routePath(ctx, req)
	token, _ := ctx.Value(ContextKeyAccessToken).(string)
	if token != "" {
		values.Set(VarUserAuthorization, token)
	}

	req.URL.RawQuery = values.Encode()
	return nil
}

// ClientEncodeJSONRequest is an EncodeRequestFunc that serializes the request
// as a JSON object to the Request body. Many JSON-over-HTTP services can use
// it as a sensible default. If the request implements Headerer, the provided
// headers will be applied to the request.
func ClientEncodeJSONRequest(
	ctx context.Context,
	req *http.Request,
	request interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	routePath(ctx, req)
	if headerer, ok := request.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			req.Header.Set(k, headerer.Headers().Get(k))
		}
	}
	if auth, ok := ctx.Value(ContextKeyRequestAuthorization).(string); ok {
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
	method, router string,
	dec DecodeResponseFunc) endpoint.Endpoint {
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
		httptransport.DecodeResponseFunc(dec),
		options...,
	).Endpoint()

	e = circuitbreaker.Gobreaker(
		gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
	//e = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), qps))(e)

	return e
}

// ClientLoadBalancer load balance
func ClientLoadBalancer(endpoints sd.FixedEndpointer) endpoint.Endpoint {
	balancer := lb.NewRoundRobin(endpoints)
	return lb.Retry(retryMax, retryTimeout, balancer)
}
