package temp

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/filllabs/sincap-common/logging"
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
	return o, nil
}

// Read reads the given named file from the temp directory.
func Read(fileName string) (string, *[]byte, error) {
	content, err := ioutil.ReadFile(Folder + "/" + fileName)
	if err != nil {
		return "", nil, err
	}
	path := strings.Split(fileName, "-")
	return path[0], &content, nil
}

// NewReader returns reader for the given named file from the temp directory.
func NewReader(fileName string) (string, io.Reader, error) {
	file, err := os.Open(path.Join(Folder, fileName))
	if err != nil {
		return "", nil, err
	}
	path := strings.Split(fileName, "-")
	return path[0], bufio.NewReader(file), nil
}

// Delete deletes the given named file from the temp directory.
func Delete(fileName string) error {
	err := os.Remove(path.Join(Folder, fileName))
	return err
}
