package validator

import (
	"context"
	"net/http"
	"reflect"

	"gitlab.com/sincap/sincap-common/json"
	"gitlab.com/sincap/sincap-common/resources"
	"gitlab.com/sincap/sincap-common/resources/contexts"
	"gitlab.com/sincap/sincap-common/resources/responses"
)

// Context middleware is used to parse interfaces and validate them.
func Context(contextKey contexts.ContextKey, in interface{}) resources.Handler {
	t := reflect.TypeOf(in)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			record := reflect.New(t).Interface()
			if err := json.Decode(r.Body, record); err != nil {
				responses.Err(w, r, err, http.StatusBadRequest)
				return
			}
			if err := Validate.Struct(record); err != nil {
				responses.Err(w, r, err, http.StatusUnprocessableEntity)
				return
			}
			ctx := context.WithValue(r.Context(), contextKey, record)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
