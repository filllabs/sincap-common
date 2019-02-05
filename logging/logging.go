package logging

import (
	"log"

	"go.uber.org/zap"
)

// Logger is the root logger of the application.
var Logger *zap.Logger

// Sugar is the root sugar (belongs to root logger) of the application.
var Sugar *zap.SugaredLogger

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Can't initialize logger. Please check configuration. ->" + err.Error())
	}
	Logger = logger
	sugar := Logger.Sugar()
	Sugar = sugar
	Logger.Info("Logger initialized")
}

//Configure helps you to configure the global logger according to your projects configuration
func Configure(conf zap.Config) {
	logger, err := conf.Build()
	if err != nil {
		log.Fatal("Can't configure logger. Will use production congfig from zap. Please check configuration." + err.Error())
	}
	Logger = logger
	sugar := Logger.Sugar()
	Sugar = sugar
	Logger.Info("Logger Configured", zap.Any("config", conf))
}
