package toolkit

import (
	"context"
	"encoding/base64"
	"errors"
	"time"
	//"strings"

	"github.com/go-kit/kit/endpoint"
)

type jwtKey string

const (
	// JWTToken jwt access token
	JWTToken jwtKey = `jwt_access_token`
)

func checkAuth(accessToken AccessToken) (CacheAccessToken, error) {
	var (
		id, uid []byte
		err     error
		token   CacheAccessToken
	)
	id, err = base64.StdEncoding.DecodeString(accessToken.ID)
	if err != nil {
		return token, err
	}
	aes := NewAesCrypto()
	uid, err = aes.Decrypt(id)
	if err != nil {
		return token, err
	}

	token, err = AccessTokenGetCache(string(uid))
	t := time.Now()
	if token.Expires <= t.Unix() {
		err = errors.New(`认证信息已过期`)
	}

	return token, err
}

// AuthMiddleware auth
func AuthMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(
			ctx context.Context,
			request interface{}) (interface{}, error) {
			if token, ok := ctx.Value(ContextKeyAccessToken).(string); ok {
				var (
					tok    AccessToken
					ctoken CacheAccessToken
					err    error
				)
				if token == "" {
					return NewReplyData(ErrUnAuthorized), nil
				}
				if tok, err = ParseAccessToken(token); err != nil {
					return ErrReplyData(ErrUnAuthorized, err.Error()), nil
				}
				if ctoken, err = checkAuth(tok); err != nil {
					return ErrReplyData(ErrUnAuthorized, err.Error()), nil
				}
				ctx = context.WithValue(ctx, JWTToken, ctoken)
				return next(ctx, request)
			}
			return NewReplyData(ErrUnAuthorized), nil
		}
	}
}
