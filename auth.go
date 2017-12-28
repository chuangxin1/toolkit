package toolkit

import (
	"encoding/json"
	"errors"

	jwt "github.com/dgrijalva/jwt-go"
)

// AccessToken access token
type AccessToken struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Expires int64  `json:"expires_in"`
}

// CacheAccessToken cache access token
type CacheAccessToken struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Status  int    `json:"status"`
	Expires int64  `json:"expires_in"`
	Message string `json:"message"`
}

var (
	accessTokenKey []byte
)

// cacheKey cache token key
func cacheKey(id string) string {
	return `user:user:` + id
}

// AccessTokenStorageCache storage CacheAccessToken to redis
func AccessTokenStorageCache(id string, token CacheAccessToken) error {
	bytes, err := json.Marshal(token)
	if err != nil {
		return err
	}
	redis := NewRedisCache()
	return redis.Set(cacheKey(id), string(bytes), 0)
}

// AccessTokenGetCache get CacheAccessToken from redis
func AccessTokenGetCache(id string) (CacheAccessToken, error) {
	redis := NewRedisCache()
	data, err := redis.Get(cacheKey(id))
	var token CacheAccessToken
	if err != nil {
		return token, err
	}
	err = json.Unmarshal([]byte(data), &token)
	return token, err
}

// SetAccessTokenKey set jwt key
func SetAccessTokenKey(key string) {
	accessTokenKey = []byte(key)
}

// NewAccessToken new token
func NewAccessToken(tok AccessToken) (string, error) {
	claims := jwt.MapClaims{
		"id":         tok.ID,
		"name":       tok.Name,
		"expires_in": tok.Expires}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(accessTokenKey)
}

// ParseAccessToken parse token
func ParseAccessToken(accessToken string) (AccessToken, error) {
	var (
		tok AccessToken
		e   error
	)

	token, err := jwt.Parse(
		accessToken,
		func(token *jwt.Token) (interface{}, error) {
			return accessTokenKey, nil
		})

	if token.Valid {
		claims := token.Claims.(jwt.MapClaims)
		if id, ok := claims["id"]; ok {
			tok.ID = id.(string)
		}
		if name, ok := claims["name"]; ok {
			tok.Name = name.(string)
		}
		if expires, ok := claims["expires_in"]; ok {
			tok.Expires = int64(expires.(float64))
		}
		e = nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			//fmt.Println("That's not even a token")
			e = errors.New(`错误的认证信息`)
		} else if ve.Errors&
			(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			//fmt.Println("Timing is everything")
			e = errors.New(`认证信息已过期`)
		} else {
			e = errors.New(`无效的认证信息`)
			//fmt.Println("Couldn't handle this token:", err)
		}
	}
	return tok, e
}
