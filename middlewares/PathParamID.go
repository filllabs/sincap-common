package middlewares

import (
	"reflect"

	"github.com/filllabs/sincap-common/logging"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func PathParamID(contextKey string, in interface{}, paramKey string, db *gorm.DB) func(ctx *fiber.Ctx) error {
	t := reflect.TypeOf(in)
	return func(ctx *fiber.Ctx) error {
		// parse handles the mime type
		id, err := ctx.ParamsInt(paramKey, 0)
		if err != nil || id == 0 {
			return fiber.NewError(fiber.StatusNotFound)
		}

		record := reflect.New(t).Interface()
		if err := read(db.Unscoped(), record, id); err != nil {
			return fiber.NewError(fiber.StatusNotFound)
		}
		ctx.Locals(contextKey, record)
		return ctx.Next()
	}
}

func read(db *gorm.DB, record interface{}, id int, preloads ...string) error {
	result := db.First(record, id)
	if result.Error != nil {
		logging.Logger.Error("Read error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Int("id", id))
	}
	return result.Error
}
