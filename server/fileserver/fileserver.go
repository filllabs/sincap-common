package fileserver

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/filllabs/sincap-common/logging"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Add adds multiple file servers from the given configurations.
func Add(r chi.Router, configs ...Config) {
	workDir, _ := os.Getwd()
	for _, config := range configs {
		filesDir := filepath.Join(workDir, config.Folder)
		fileServer(r, config.Path, http.Dir(filesDir))
	}
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.Dir) {
	if strings.ContainsAny(path, "{}*") {
		logging.Logger.Panic("FileServer does not permit URL parameters.")
	}
	fs := http.StripPrefix(path, http.FileServer(root))

	logging.Logger.Info("FileServer is mounting", zap.String("path", path), zap.Any("root", root))
	r.Mount(path, fs)
}
