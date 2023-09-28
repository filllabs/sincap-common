package images

import (
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/filllabs/sincap-common/logging"
	"go.uber.org/zap"
)

// IsJPGPNG checks the multipart file is jpg or png
func IsJPGPNG(file multipart.File, fileName string) (bool, string) {
	var extension = strings.ToLower(filepath.Ext(fileName)[1:])

	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		logging.Logger.Error("error on", zap.Error(err))
		return false, ""
	}

	// set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		logging.Logger.Error("error on", zap.Error(err))
		return false, ""
	}

	mimeType := http.DetectContentType(fileHeader)

	return mimeType == "image/jpg" || mimeType == "image/jpeg" || mimeType == "image/png", strings.ToLower(extension)
}
