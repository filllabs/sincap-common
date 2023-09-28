// Package ws provides necessary Connect, Disconnect functions and Writers
package ws

import (
	"sync"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/filllabs/sincap-common/auth/claims"
	"github.com/go-chi/jwtauth"
	"go.uber.org/zap"
	melody "gopkg.in/olahol/melody.v1"
)

// NewEncryptedHandleConnect returns default implementation of connect handler for melody websockets. Decodes jwt with the given secret.
func NewEncryptedHandleConnect(users *sync.Map, logger *zap.Logger, beforeResponse func(*melody.Session), secret string) func(*melody.Session) {
	return func(s *melody.Session) {
		if claims, err := claims.FromContext(s.Request.Context(), secret); err != nil {
			logger.Warn("Can not read token from request context")
		} else {
			if claims.Username == "" {
				logger.Warn("Can not read Username from token")
			} else {
				if claims.UserID == 0 {
					logger.Warn("Can not read UserId from token")
				} else {
					s.Set("Username", claims.Username)
					s.Set("UserID", claims.UserID)
					users.Store(s, claims.Username)
					logger.Info("Web Socket Connected", zap.String("Username", claims.Username))
					beforeResponse(s)
				}
			}
		}
	}
}

// NewHandleConnect returns default implementation of connect handler for melody websockets.HandleConnect
func NewHandleConnect(users *sync.Map, logger *zap.Logger, beforeResponse func(*melody.Session)) func(*melody.Session) {
	return func(s *melody.Session) {
		if token, _, err := jwtauth.FromContext(s.Request.Context()); err != nil {
			logger.Warn("Can not read token from request context")
		} else {
			mapClaims := token.Claims.(jwt.MapClaims)
			if username := mapClaims["Username"]; username == nil {
				logger.Warn("Can not read Username from token")
			} else {
				if userID := mapClaims["UserID"]; userID == nil {
					logger.Warn("Can not read UserId from token")
				} else {
					s.Set("Username", username)
					s.Set("UserID", uint(userID.(float64)))
					users.Store(s, username.(string))
					logger.Info("Web Socket Connected", zap.String("Username", username.(string)))
					beforeResponse(s)
				}
			}
		}
	}
}

// NewHandleDisconnect returns default implementation of disconnect handler for melody websockets.HandleConnect
func NewHandleDisconnect(users *sync.Map, logger *zap.Logger, beforeResponse func(*melody.Session)) func(*melody.Session) {
	return func(s *melody.Session) {
		if user, ok := users.Load(s); ok {
			logger.Info("Web Socket Disconnected", zap.String("Username", user.(string)))
			users.Delete(s)
			beforeResponse(s)
		}
	}
}
