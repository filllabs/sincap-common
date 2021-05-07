package config

import "strconv"

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
