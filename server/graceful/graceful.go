package graceful

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/filllabs/sincap-common/logging"
)

// SignalChan is the single and the only channel for listening cancel signals.
/*
func main (){
	go graceful(fs)
}
func graceful(fs *flag.FlagSet) {
	sig := <-graceful.SignalChan
	logging.Logger.Info("App is stopping gracefuly", zap.Any("signal", sig))
	os.Exit(0)
}*/
var SignalChan = make(chan os.Signal)

// WG is a wait group to help stopping server gracefuly
var WG = logging.LoggedWaitGroup{Name: "QuitWG"}

// Register calls signal.Notify for all necessary signals
func Register() {
	signal.Notify(SignalChan, syscall.SIGTERM)
	signal.Notify(SignalChan, syscall.SIGINT)
}
