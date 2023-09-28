package server

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/filllabs/sincap-common/db"
	"github.com/filllabs/sincap-common/logging"
	"github.com/filllabs/sincap-common/server/graceful"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

var server *http.Server

// Start starts a http server with the given router.
// As an extra it sets start time as a new seed for the random
func Start(config Config, r chi.Router, close func()) {
	rand.Seed(time.Now().UTC().UnixNano())
	logging.Logger.Info("Server is starting", zap.String("domain", config.Domain), zap.Int64("port", config.Port))
	graceful.Register()
	server = &http.Server{Addr: config.GetHost(), Handler: r}
	// Start the server async
	go func() {
		graceful.WG.Addl("Server", 1)

		if err := http.ListenAndServe(config.GetHost(), r); err != nil {
			logging.Logger.Panic("Server Error", zap.Error(err))
			os.Exit(1)
		}
	}()
	// and wait for the Shutdown signals
	StopGracefully(close)
}

func stop() {
	server.Shutdown(context.Background())
	graceful.WG.Donel("Server")
	graceful.WG.Wait()
	db.CloseAll()
	logging.Logger.Named("Server").Info("stopped gracefully")
	logging.Logger.Sync()
	os.Exit(0)
}

// StopGracefully listens for os signals stops the server gracefully.
// Double Command/Control + C forces to close immediately
func StopGracefully(cb func()) {
	// if signal comes than start stopping process
	sig := <-graceful.SignalChan
	logging.Logger.Named("Server").Info("stopping gracefuly", zap.Any("signal", sig))
	// Call callbacks to let application finalize what is necessary
	cb()
	// start graceful stop procedure it is async because of force stop feature.
	go stop()
	// wait for the second signal if comes force stop and exit with an error
	sig = <-graceful.SignalChan
	logging.Logger.Named("Server").Info("stopped by force", zap.Any("signal", sig))
	logging.Logger.Sync()
	os.Exit(130)
}
