package temp

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.com/sincap/sincap-common/logging"
	"go.uber.org/zap"
)

// Folder holds the temp directory for the app
var Folder string

func init() {
	var err error
	appname := filepath.Base(os.Args[0])
	Folder, err = ioutil.TempDir("", appname)
	logging.Logger.Debug("Creating temp folder", zap.String("path", Folder))
	if err != nil {
		logging.Logger.Panic("Can't Create temp folder", zap.Error(err))
		return
	}
}

// Write writes the given multipart file to the temp directory.
func Write(content *[]byte, prefix string) (*os.File, error) {
	o, err := ioutil.TempFile(Folder, prefix+"-")
	if err != nil {
		return nil, err
	}

	if _, err := o.Write(*content); err != nil {
		return nil, err
	}
	if err := o.Close(); err != nil {
		return nil, err
	}
	logging.Logger.Debug("Temp File written", zap.String("name", o.Name()))
	return o, nil
}
