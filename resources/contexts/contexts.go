package contexts

import (
	"context"
	"net/http"
	"reflect"

	"github.com/go-chi/chi"
	"gitlab.com/sincap/sincap-common/dbutil"
	"gitlab.com/sincap/sincap-common/resources"
	"gitlab.com/sincap/sincap-common/types"
)

// PathParamIDCtx is a ready to use context for reading "id" path param and receiving related
func PathParamIDCtx(ctxKey resources.ContextKey, in interface{}) resources.Handler {
	t := reflect.TypeOf(in)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := chi.URLParam(r, "id")
			id, err := types.ParseUint(idParam)
			if err != nil {
				resources.Response404(w, r)
				return
			}
			record := reflect.New(t).Interface()
			if err := dbutil.Read(dbutil.DB(), record, id); err != nil {
				resources.Response404(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey, record)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
