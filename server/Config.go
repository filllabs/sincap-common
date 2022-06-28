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
	Domain          string          `json:"domain"`
	Port            int64           `json:"port"`
	FrontendURL     string          `json:"frontendURL"`
	BackendURL      string          `json:"backendURL"`
	SecurityHeaders bool            `json:"securityHeaders"`
	Etag            *etag.Config    `json:"etag,omitempty"`
	Cors            *cors.Config    `json:"cors,omitempty"`
	Limiter         *limiter.Config `json:"limiter,omitempty"`
	fiber.Config
}

// GetHost returns the combination of the domain, port and more
func (s *Config) GetHost() string {
	return s.Domain + ":" + strconv.FormatInt(s.Port, 10)
}
