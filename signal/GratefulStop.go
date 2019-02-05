package signal

import (
	"os"
	"os/signal"
	"syscall"
)

// GracefulStop is the single and the only chanel for listening cancel signals.
// Every command shoud implement their defers as a go routine which listes this channel
//TODO: make is developer friendly
var GracefulStop = make(chan os.Signal)

func Register() {
	signal.Notify(GracefulStop, syscall.SIGTERM)
	signal.Notify(GracefulStop, syscall.SIGINT)
}
