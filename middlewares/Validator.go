package middlewares

import (
	"fmt"
	"reflect"

	"github.com/filllabs/sincap-common/validator"
	"github.com/gofiber/fiber/v2"
)

func Validator(key string) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		r := ctx.Locals(key)
		if r == nil {
			return ctx.Status(400).JSON(map[string]string{"error": "invalid request"})
		}
		if err := validator.Validate.Struct(r); err != nil {
			// TODO: beautify error messages
			return ctx.Status(422).JSON(map[string]string{"error": err.Error()})
		}
		return ctx.Next()
	}
}
func ValidatorMap(key string, t any) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		r, ok := ctx.Locals(key).(*map[string]any)
		if !ok {
			return ctx.Status(400).JSON(map[string]string{"error": "invalid request"})
		}

		if r == nil {
			return ctx.Status(400).JSON(map[string]string{"error": "invalid request"})
		}
		rules := mapRulesFromFields(t, *r)
		if results := validator.Validate.ValidateMap(*r, rules); len(results) > 0 {
			// TODO: beautify error messages

			return ctx.Status(422).JSON(map[string]string{"error": fmt.Sprintf("%v", results)})
		}
		return ctx.Next()
	}
}

func mapRulesFromFields(t any, rec map[string]any) map[string]any {
	//TODO: add caching
	fields := reflect.VisibleFields(reflect.TypeOf(t))
	rules := make(map[string]any, len(fields))
	for _, field := range fields {
		if _, ok := rec[field.Name]; !ok {
			continue
		}
		tag := field.Tag.Get("validate")
		if len(tag) > 0 {
			rules[field.Name] = tag
		}
	}
	return rules
}
