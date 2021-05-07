package config

// FileServer holds file serving configuration
type FileServer struct {
	Folder string `json:"folder"`
	Path   string `json:"path"`
}
