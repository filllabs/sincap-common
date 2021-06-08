package temp

import (
	"io/ioutil"
	"os"
	"time"

	"go.uber.org/zap"
)

// Cleaner starts a go routine for cleanin temp files created by upload.
// It cleans only files older than maxAge given
func Cleaner(l *zap.Logger, maxAge time.Duration) {
	logger := l.Named("TempFolder Cleaner")
	logger.Info("Starting")

	files, err := ioutil.ReadDir(Folder)
	if err != nil {
		logger.Error("Can't read temp dir", zap.Error(err))
	}
	for _, f := range files {
		if time.Since(f.ModTime()) >= maxAge {
			logger.Debug("Deleting temp file", zap.String("name", f.Name()), zap.String("time", f.ModTime().String()))
			err := os.Remove(Folder + "/" + f.Name())
			if err != nil {
				logger.Warn("Can't delete temp file", zap.String("name", f.Name()), zap.Error(err))
			}
		}
	}
}
