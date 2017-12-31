package toolkit

import (
	"context"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
)

var (
	// retry max
	retryMax = 3
	// retry timeout
	retryTimeout = 500 * time.Millisecond
)

func factory(
	ctx context.Context,
	method, router string,
	dec DecodeResponseFunc) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}
		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}

		return ClientRequestEndpoint(ctx, tgt, method, router, dec), nil, nil
	}
}

// FactoryLoadBalancer factory load balance
func FactoryLoadBalancer(
	ctx context.Context,
	instancer *consulsd.Instancer,
	method, router string,
	dec DecodeResponseFunc,
	logger log.Logger) endpoint.Endpoint {

	endpointer := sd.NewEndpointer(
		instancer,
		factory(ctx, method, router, dec),
		logger)
	balancer := lb.NewRoundRobin(endpointer)
	return lb.Retry(retryMax, retryTimeout, balancer)
}
