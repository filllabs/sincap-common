// Package claims holds EncryptedClaims, DecryptedClaims and util functions.
// These are necessary for auth and token (jwt)
package claims

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
)

// FromContext decodes the given jwt from the context and returns a decrypted version of the claims
func FromContext(context context.Context, secret string) (*DecryptedClaims, error) {
	token, eclaims, err := readEncrypted(context)
	if err != nil {
		return nil, fmt.Errorf("token error read token. %v", err)
	}
	if token == nil || !token.Valid {
		return nil, errors.New("token error token not valid")
	}
	return eclaims.Decrypt(secret)
}

func readEncrypted(ctx context.Context) (*jwt.Token, *EncryptedClaims, error) {
	token, _ := ctx.Value(jwtauth.TokenCtxKey).(*jwt.Token)

	var claims jwt.MapClaims
	if token != nil {
		if tokenClaims, ok := token.Claims.(jwt.MapClaims); ok {
			claims = tokenClaims
		} else {
			return nil, nil, fmt.Errorf("jwtauth: unknown type of Claims: %+v", token.Claims)
		}
	} else {
		claims = jwt.MapClaims{}
	}
	eclaims := EncryptedClaims{}
	eclaims.Fill(claims)

	err, _ := ctx.Value(jwtauth.ErrorCtxKey).(error)

	return token, &eclaims, err
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return strings.Split(IPAddress, ":")[0]
}
