// Package claims holds EncryptedClaims, DecryptedClaims and util functions.
// These are necessary for auth and token (jwt)
package claims

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt/v4"
)

// FromContext decodes the given jwt from the context and returns a decrypted version of the claims
func FromContext(context context.Context, secret string) (*DecryptedClaims, error) {
	token, _ := context.Value(jwtauth.TokenCtxKey).(*jwt.Token)
	eclaims, err := readEncrypted(token)
	if err != nil {
		return nil, fmt.Errorf("token error read token. %v", err)
	}
	if token == nil || !token.Valid {
		return nil, errors.New("token error token not valid")
	}
	return eclaims.Decrypt(secret)
}

// FromContext decodes the given jwt from the context and returns a decrypted version of the claims
func FromToken(token *jwt.Token, secret string) (*DecryptedClaims, error) {
	eclaims, err := readEncrypted(token)
	if err != nil {
		return nil, fmt.Errorf("token error read token. %v", err)
	}
	if token == nil || !token.Valid {
		return nil, errors.New("token error token not valid")
	}
	return eclaims.Decrypt(secret)
}

func readEncrypted(token *jwt.Token) (*EncryptedClaims, error) {
	var claims jwt.MapClaims
	if token != nil {
		if tokenClaims, ok := token.Claims.(jwt.MapClaims); ok {
			claims = tokenClaims
		} else {
			return nil, fmt.Errorf("jwtauth: unknown type of Claims: %+v", token.Claims)
		}
	} else {
		claims = jwt.MapClaims{}
	}
	eclaims := EncryptedClaims{}
	eclaims.Fill(claims)

	return &eclaims, nil
}
