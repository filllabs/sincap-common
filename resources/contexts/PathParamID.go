package contexts

import (
	"context"
	"net/http"
	"reflect"

	"github.com/go-chi/chi"
	"gitlab.com/sincap/sincap-common/db"
	"gitlab.com/sincap/sincap-common/logging"
	"gitlab.com/sincap/sincap-common/resources/responses"
	"gitlab.com/sincap/sincap-common/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PathParamID is a ready to use context for reading "id" path param.
// Reads the parameter and receives from the database to put in to the context with the given key
func PathParamID(key ContextKey, i interface{}) func(next http.Handler) http.Handler {
	t := reflect.TypeOf(i)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := chi.URLParam(r, "id")
			id, err := types.ParseUint(idParam)
			if err != nil {
				responses.Status404(w, r)
				return
			}
			record := reflect.New(t).Interface()
			if err := read(db.DB(), record, id); err != nil {
				responses.Status404(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), key, record)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PathParamIDUnscoped is a ready to use context for reading "id" path param with Unscoped support.
// Reads the parameter and receives from the database to put in to the context with the given key
func PathParamIDUnscoped(key ContextKey, i interface{}) func(next http.Handler) http.Handler {
	t := reflect.TypeOf(i)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := chi.URLParam(r, "id")
			id, err := types.ParseUint(idParam)
			if err != nil {
				responses.Status404(w, r)
				return
			}
			record := reflect.New(t).Interface()
			if err := read(db.DB().Unscoped(), record, id); err != nil {
				responses.Status404(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), key, record)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func read(DB *gorm.DB, record interface{}, id uint, preloads ...string) error {
	result := DB.First(record, id)
	if result.Error != nil {
		logging.Logger.Error("Read error", zap.Any("Model", reflect.TypeOf(record)), zap.Error(result.Error), zap.Uint("id", id))
	}
	return result.Error
}
