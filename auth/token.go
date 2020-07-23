package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
)

// DecodeFromContext decodes the given jwt from the context and returns a decrypted version of the claims
func DecodeFromContext(context context.Context, secret string) (*DecryptedClaims, error) {
	token, eclaims, err := fromContext(context)
	if err != nil {
		return nil, fmt.Errorf("Token error read token. %v", err)
	}
	if token == nil || !token.Valid {
		return nil, errors.New("Token error token not valid")
	}
	return eclaims.Decrypt(secret)
}

func fromContext(ctx context.Context) (*jwt.Token, *EncryptedClaims, error) {
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

// ValidateClaimsWithRequest validates the request information with the info inside of the claims
func ValidateClaimsWithRequest(r http.Request, claims DecryptedClaims) error {
	if ip := ReadUserIP(&r); ip != claims.UserIP {
		return fmt.Errorf("token error, ip info is not equal. User ip : %v  Claims ip : %v", ip, claims.UserIP)
	}
	if r.Header.Get("User-Agent") != claims.UserAgent {
		return fmt.Errorf("token error, user agents info is not equal. User user-agent : %v  Claims user-agent : %v", r.Header.Get("User-Agent"), claims.UserAgent)
	}
	return nil
}

// ReadUserIP gets real IP of the user
func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return strings.Split(IPAddress, ":")[0]
}
