package toolkit

import (
	"errors"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

// AccessToken access token
type AccessToken struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Expires int    `json:"expires_in"`
}

var (
	accessTokenKey []byte
)

// SetAccessTokenKey set jwt key
func SetAccessTokenKey(key []byte) {
	accessTokenKey = key
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
			tok.Expires = expires.(int)
		}
		e = nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			//fmt.Println("That's not even a token")
			e = errors.New(`That's not even a token`)
		} else if ve.Errors&
			(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			//fmt.Println("Timing is everything")
			e = errors.New(`expired`)
		} else {
			e = errors.New(`Couldn't handle this token`)
			//fmt.Println("Couldn't handle this token:", err)
		}
	}
	return tok, e
}

// Strips 'Bearer ' prefix from bearer token string
func stripBearerPrefixFromTokenString(tok string) (string, error) {
	// Should be a bearer token
	if len(tok) > 6 && strings.ToUpper(tok[0:7]) == "BEARER " {
		return tok[7:], nil
	}
	return tok, nil
}
