package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/sincap/sincap-common/validator"
)

func Validator(key string) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		r := ctx.Locals(key)
		if r == nil {
			return ctx.Status(400).JSON(map[string]string{"error": "invalid request"})
		}
		if err := validator.Validate.Struct(r); err != nil {
			return ctx.Status(422).JSON(map[string]string{"error": err.Error()})
		}
		return ctx.Next()
	}
}
