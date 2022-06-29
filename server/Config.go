package server

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// Config holds server configuration
type Config struct {
	Domain          string          `json:"domain" yaml:"domain"`
	Port            int64           `json:"port" yaml:"port"`
	FrontendURL     string          `json:"frontendURL" yaml:"frontendURL"`
	BackendURL      string          `json:"backendURL" yaml:"backendURL"`
	SecurityHeaders bool            `json:"securityHeaders" yaml:"securityHeaders"`
	Etag            *etag.Config    `json:"etag,omitempty" yaml:"etag,omitempty"`
	Cors            *cors.Config    `json:"cors,omitempty" yaml:"cors,omitempty"`
	Limiter         *limiter.Config `json:"limiter,omitempty" yaml:"limiter,omitempty"`
	fiber.Config
}

// GetHost returns the combination of the domain, port and more
func (s *Config) GetHost() string {
	return s.Domain + ":" + strconv.FormatInt(s.Port, 10)
}
