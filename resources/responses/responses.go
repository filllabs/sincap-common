// Package responses contains shorthand functions for everyday responses
package responses

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/filllabs/sincap-common/json"
	"github.com/go-chi/render"
)

// Err renders the given error as json with the given status
func Err(w http.ResponseWriter, r *http.Request, err error, status int) {
	render.Status(r, status)
	json.Render(w, r, err.Error())
}

// Data renders the given data as a json with X-Total-Count header
func Data(w http.ResponseWriter, r *http.Request, data interface{}, count ...int) {
	if len(count) > 0 {
		w.Header().Set("X-Total-Count", strconv.Itoa(count[0]))
	}
	json.Render(w, r, data)
}

// Status204 returns 204 no content
func Status204(w http.ResponseWriter, r *http.Request) {
	render.NoContent(w, r)
}

// Status401 renders the 401 error as json with the status Unauthorized
func Status401(w http.ResponseWriter, r *http.Request) {
	Err(w, r, errors.New(http.StatusText(http.StatusUnauthorized)), http.StatusUnauthorized)
}

// Status403 renders the 403 error as json with the status Forbidden
func Status403(w http.ResponseWriter, r *http.Request) {
	Err(w, r, errors.New(http.StatusText(http.StatusForbidden)), http.StatusForbidden)
}

// Status404 renders the 404 error as json with the status Not Found
func Status404(w http.ResponseWriter, r *http.Request) {
	Err(w, r, errors.New(http.StatusText(http.StatusNotFound)), http.StatusNotFound)
}

// Status500 renders the 500 error as json with the status Internal Server Error
func Status500(w http.ResponseWriter, r *http.Request, err error) {
	Err(w, r, err, http.StatusInternalServerError)
}
