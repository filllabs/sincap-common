// Package json provides utility methods for centralising json rendering in order to make changes later easier for performance reasons.
package json

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
)

// Decode helps to deserialize json streams
func Decode(r io.Reader, v interface{}) error {
	err := render.DecodeJSON(r, v)
	return err
}

// Render helps to serialize json to streams and close context status.
func Render(w http.ResponseWriter, r *http.Request, v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if status, ok := r.Context().Value(render.StatusCtxKey).(int); ok {
		w.WriteHeader(status)
	}
	w.Write(buf)
}
