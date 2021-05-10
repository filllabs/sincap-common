package server

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
)

// Start starts a http server with the given router.
// As an extra it sets start time as a new seed for the random
func Start(config Config, r chi.Router) {
	rand.Seed(time.Now().UTC().UnixNano())
	logging.Logger.Info("Server is starting", zap.String("domain", config.Domain), zap.Int64("port", config.Port))
	err := http.ListenAndServe(config.GetHost(), r)
	logging.Logger.Panic("Server Error", zap.Error(err))
}
