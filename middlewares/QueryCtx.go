package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/sincap/sincap-common/resources/query"
)

// Context parses the query params for the query
func Context(contextKey string) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		query := query.Query{}
		params := map[string]string{}
		params["_q"] = ctx.Query("_q", "")
		params["_fields"] = ctx.Query("_fields", "")
		params["_preloads"] = ctx.Query("_preloads", "")
		params["_offset"] = ctx.Query("_offset", "")
		params["_limit"] = ctx.Query("_limit", "")
		params["_sort"] = ctx.Query("_sort", "")
		params["_filter"] = ctx.Query("_filter", "")

		if err := query.Parse(params); err == nil {
			ctx.Locals("queryapi", &query)
		}
		return ctx.Next()
	}
}
