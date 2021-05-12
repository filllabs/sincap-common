// Package json provides utility methods for centralising json rendering in order to make changes later easier for performance reasons.
package json

import (
	"bytes"
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

// ToMap decodes json byte stream to a map with int support
func ToMap(in *[]byte) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(*in))
	decoder.UseNumber()
	err := decoder.Decode(&parsed)
	if err != nil {
		return parsed, err
	}
	for key, val := range parsed {
		n, ok := val.(json.Number)
		if !ok {
			continue
		}
		if i, err := n.Int64(); err == nil {
			parsed[key] = i
			continue
		}
		if f, err := n.Float64(); err == nil {
			parsed[key] = f
			continue
		}
	}
	return parsed, nil
}
