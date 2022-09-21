package middlewares

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

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
func BodyParserMap(key string, forbiddenFields ...string) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		t := make(map[string]any, 1)
		err := ctx.BodyParser(&t)
		if err != nil {
			return err
		}
		for _, key := range forbiddenFields {
			if _, ok := t[key]; ok {
				return ctx.Status(422).JSON(map[string]string{"error": fmt.Sprintf("You cannot update %s Field", key)})
			}
		}
		ctx.Locals(key, &t)
		return ctx.Next()
	}
}
