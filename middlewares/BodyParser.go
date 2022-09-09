package middlewares

import "github.com/gofiber/fiber/v2"

func BodyParser[T any](key string) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		t := new(T)
		err := ctx.BodyParser(t)
		if err != nil {
			return err
		}
		ctx.Locals(key, t)
		return ctx.Next()
	}
}
