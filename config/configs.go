package config

import (
	"strconv"

	"gitlab.com/sincap/sincap-common/dbconn"
	"go.uber.org/zap"
)

// Config is the configuration of the application
type Config struct {
	Server      Server            `json:"server"`
	FrontendURL string            `json:"frontendURL"`
	BackendURL  string            `json:"backendURL"`
	FileServer  []FileServer      `json:"fileServer"`
	DB          []dbconn.DBConfig `json:"db"`
	Auth        Auth              `json:"auth"`
	Log         zap.Config        `json:"log"`
	Metrics     Metrics           `json:"metrics"`
	Mail        Mail              `json:"mail"`
	Recaptcha   Recaptcha         `json:"recaptcha"`
}

// FileServer holds file serving configuration
type FileServer struct {
	Folder string `json:"folder"`
	Path   string `json:"path"`
}

// Server holds server configuration
type Server struct {
	Domain          string `json:"domain"`
	Port            int64  `json:"port"`
	Cors            bool   `json:"cors"`
	SecurityHeaders bool   `json:"securityHeaders"`
}

// GetHost returns the combination of the domain, port and more
func (s *Server) GetHost() string {
	return s.Domain + ":" + strconv.FormatInt(s.Port, 10)
}

// GetURL returns the host with the protocol
func (s *Server) GetURL() string {
	return "http://" + s.GetHost()
}

// Auth holds authentication configuration
type Auth struct {
	Algo     string `json:"algo"`
	Secret   string `json:"secret"`
	SignKey  string `json:"signKey"`
	Timeout  int64  `json:"timeout"`
	HTTPOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

// Metrics holds metric reporting configuration
type Metrics struct {
	Interval int `json:"interval"`
}

// Mail holds the mail server configuration.
type Mail struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Recaptcha holds conf. of Google Recaptcha.
type Recaptcha struct {
	Key string `json:"key"`
}
