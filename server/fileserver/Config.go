package fileserver

// Config holds file serving configuration
type Config struct {
	Folder string `json:"folder" yaml:"folder"`
	Path   string `json:"path" yaml:"path"`
}
