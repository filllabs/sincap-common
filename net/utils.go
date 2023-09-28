package net

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ReadUserIP(ctx *fiber.Ctx) string {
	IPAddress := ctx.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = ctx.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = ctx.IP()
	}
	return strings.Split(IPAddress, ":")[0]
}
