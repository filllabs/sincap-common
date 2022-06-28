// Package middlewares provides a set of usefull middlewares
package middlewares

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gitlab.com/sincap/sincap-common/logging"
	"gitlab.com/sincap/sincap-common/net"
	"gitlab.com/sincap/sincap-common/resources/middlewares"
	"gitlab.com/sincap/sincap-common/server"
	"go.uber.org/zap"
)

// AddDefaultMiddlewares adds all predefined middlewares to the router.
// You may add RequestMetrics manually by adding 	AddRequestMetrics(r)
func AddDefaultMiddlewares(r fiber.Router, config server.Config) {
	r.Use(requestid.New())
	r.Use(logger.New())
	r.Use(recover.New())

	logger := logging.Logger.Named("Server")

	if config.Cors == nil {
		config.Cors = &cors.ConfigDefault
	}
	logger.Info("Adding CORS", zap.Any("config", config.Cors))
	cors := cors.New(*config.Cors)
	r.Use(cors)

	if config.SecurityHeaders {
		logging.Logger.Named("Server").Info("Adding SecurityHeaders")
		r.Use(middlewares.SecurityHeaders)
	}

	if config.Etag != nil {
		logging.Logger.Named("Server").Info("Adding ETag")
		r.Use(etag.New(*config.Etag))
	}

	if config.Limiter != nil {
		logging.Logger.Named("Server").Info("Adding Limiter")
		config.Limiter.KeyGenerator = func(c *fiber.Ctx) string {
			return net.ReadUserIP(c)
		}
		config.Limiter.LimitReached = func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).
				SendString(fmt.Sprintf("Too many requests. Limit is %d reqs per %d mins.", config.Limiter.Max, config.Limiter.Expiration))
		}
		r.Use(limiter.New(*config.Limiter))
	}
}
