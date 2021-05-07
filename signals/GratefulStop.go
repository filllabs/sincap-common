// Package signals provides commonly needed signal scenario implementations
package signals

import (
	"os"
	"os/signal"
	"syscall"
)

// GracefulStop is the single and the only channel for listening cancel signals.
// Every program shoud implement their defers as a go routine which listens this channel.
/*
func main (){
	go graceful(fs)
}
func graceful(fs *flag.FlagSet) {
	sig := <-signal.GracefulStop
	logging.Logger.Info("App is stopping gracefuly", zap.Any("signal", sig))
	os.Exit(0)
}*/
// //TODO: make is developer friendly
var GracefulStop = make(chan os.Signal)

// RegisterGracefulStop calls signal.Notify for all necessary signals
func RegisterGracefulStop() {
	signal.Notify(GracefulStop, syscall.SIGTERM)
	signal.Notify(GracefulStop, syscall.SIGINT)
}
