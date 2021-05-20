package server

import "strconv"

// Config holds server configuration
type Config struct {
	Domain          string `json:"domain"`
	Port            int64  `json:"port"`
	Cors            bool   `json:"cors"`
	SecurityHeaders bool   `json:"securityHeaders"`
	FrontendURL     string `json:"frontendURL"`
	BackendURL      string `json:"backendURL"`
}

// GetHost returns the combination of the domain, port and more
func (s *Config) GetHost() string {
	return s.Domain + ":" + strconv.FormatInt(s.Port, 10)
}
