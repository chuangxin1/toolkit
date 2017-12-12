package toolkit

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// AuthMiddleware auth
func AuthMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(
			ctx context.Context,
			request interface{}) (interface{}, error) {
			auth, _ := ctx.Value(ContextKeyRequestAuthorization).(string)
			token, _ := ctx.Value(ContextKeyAccessToken).(string)

			//fmt.Println(HTTPHeaderAuthorization, auth)
			//fmt.Println(VarUserAuthorization, token)
			if auth == "" && token == "" {
				return NewReplyData(ErrUnAuthorized), nil
			}
			return next(ctx, request)
		}
	}
}
