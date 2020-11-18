// Package resources contains the utilities for recources.
package resources

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"gitlab.com/sincap/sincap-common/json"
	"gitlab.com/sincap/sincap-common/random"
)

// Resource is the the definition of a single service resource with minimum set of predefined information.
type Resource struct {
	PathParamCxtKey ContextKey
	BodyCtxKey      ContextKey
}

// NewResource returns a new instance of resource
func NewResource() Resource {
	r := Resource{
		PathParamCxtKey: NewContextKey(),
		BodyCtxKey:      NewContextKey(),
	}
	return r
}

// Handler is the shorthand type for a fucntion which takes http.Handler and returns http.Handler
type Handler func(next http.Handler) http.Handler

// ContextKey is a special key type for using resource context key
type ContextKey string

// NewContextKey is a special random method for using resource context key 32 byte length
func NewContextKey() ContextKey {
	key := random.GetString()
	return ContextKey(key)
}

// ResponseErr renders the given error as json with the given status
func ResponseErr(w http.ResponseWriter, r *http.Request, err error, status int) {
	render.Status(r, status)
	json.Render(w, r, err.Error())
}

// ResponseData renders the given data as a json with X-Total-Count header
func ResponseData(w http.ResponseWriter, r *http.Request, data interface{}, count ...int) {
	if len(count) > 0 {
		w.Header().Set("X-Total-Count", strconv.Itoa(count[0]))
	}
	json.Render(w, r, data)
}

// Response200 returns
func Response204(w http.ResponseWriter, r *http.Request) {
	render.NoContent(w, r)
}

// Response401 renders the 401 error as json with the status Unauthorized
func Response401(w http.ResponseWriter, r *http.Request) {
	ResponseErr(w, r, errors.New(http.StatusText(http.StatusUnauthorized)), http.StatusUnauthorized)
}

// Response403 renders the 403 error as json with the status Forbidden
func Response403(w http.ResponseWriter, r *http.Request) {
	ResponseErr(w, r, errors.New(http.StatusText(http.StatusForbidden)), http.StatusForbidden)
}

// Response404 renders the 404 error as json with the status Not Found
func Response404(w http.ResponseWriter, r *http.Request) {
	ResponseErr(w, r, errors.New(http.StatusText(http.StatusNotFound)), http.StatusNotFound)
}

// Response500 renders the 500 error as json with the status Internal Server Error
func Response500(w http.ResponseWriter, r *http.Request, err error) {
	ResponseErr(w, r, err, http.StatusInternalServerError)
}
