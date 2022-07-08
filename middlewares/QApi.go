package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/sincap/sincap-common/middlewares/qapi"
)

// QApi parses the query params for the query
func QApi(ctx *fiber.Ctx) error {
	query := qapi.Query{}
	params := make(map[string]string, 7)
	params["_q"] = ctx.Query("_q", "")
	params["_fields"] = ctx.Query("_fields", "")
	params["_preloads"] = ctx.Query("_preloads", "")
	params["_offset"] = ctx.Query("_offset", "")
	params["_limit"] = ctx.Query("_limit", "")
	params["_sort"] = ctx.Query("_sort", "")
	params["_filter"] = ctx.Query("_filter", "")

	if err := query.Parse(params); err != nil {
		// no query found no problem.
	}
	ctx.Locals("qapi", &query)
	return ctx.Next()
}
