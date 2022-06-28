package middlewares

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders is a middleware for adding security headers to the response
// Cache-Control: no-cache, no-store, max-age=0, must-revalidate
// Pragma: no-cache
// Expires: 0
// X-Content-Type-Options: nosniff
// Strict-Transport-Security: max-age=31536000 ; includeSubDomains
// X-Frame-Options: DENY
// X-XSS-Protection: 1; mode=block
func SecurityHeaders(ctx *fiber.Ctx) error {

	ctx.Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	ctx.Set("Pragma", "no-cache")
	ctx.Set("Expires", "0")
	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.Set("Strict-Transport-Security", "max-age=31536000 ; includeSubDomains")
	ctx.Set("X-Frame-Options", "DENY")
	ctx.Set("X-XSS-Protection", "1")

	return nil
}
