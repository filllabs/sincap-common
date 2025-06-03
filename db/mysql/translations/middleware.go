package translations

import (
	"context"
	"errors"

	"github.com/filllabs/sincap-common/db"
	"github.com/gofiber/fiber/v2"
)

type contextKey string

const LANG_KEY contextKey = "language"
const fiberCtxKey contextKey = "fiberCtx"

func TranslationMiddleware(c *fiber.Ctx) error {
	lang := c.Get("Accept-Language")

	langCodes, err := ListCodes(db.DB())
	if err != nil {
		lang = DEFAULT_LANG_CODE
	} else {
		found := false
		for _, langCode := range langCodes {
			if lang == langCode {
				found = true
				break
			}
		}
		if !found {
			lang = DEFAULT_LANG_CODE
		}
	}

	c.Locals("lang", lang)
	ctx := context.WithValue(c.Context(), fiberCtxKey, c)
	ctx = context.WithValue(ctx, LANG_KEY, lang)
	c.SetUserContext(ctx)

	return c.Next()
}

func GetFiberCtx(ctx context.Context) (*fiber.Ctx, error) {
	fiberCtx, ok := ctx.Value(fiberCtxKey).(*fiber.Ctx)
	if !ok {
		return nil, errors.New("fiber context not found")
	}
	return fiberCtx, nil
}

// GetLanguage now directly checks the context first before falling back to Fiber context
func GetLanguage(ctx context.Context) string {
	if lang, ok := ctx.Value(LANG_KEY).(string); ok && lang != "" {
		return lang
	}
	fiberCtx, err := GetFiberCtx(ctx)
	if err != nil {
		return DEFAULT_LANG_CODE
	}
	lang, ok := fiberCtx.Locals("lang").(string)
	if !ok || lang == "" {
		return DEFAULT_LANG_CODE
	}
	return lang
}
