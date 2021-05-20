package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

type route struct {
	Path   string
	Method string
}

// GetAllRoutes walks all paths and returns them as a slice
func GetAllRoutes(r *chi.Mux) []route {
	routes := []route{}
	walkFunc := func(method string, path string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		cleanRoute := strings.Replace(path, "/*", "", -1)
		if strings.HasSuffix(cleanRoute, "/") {
			cleanRoute = cleanRoute[0 : len(cleanRoute)-1]
		}
		routes = append(routes, route{Path: cleanRoute, Method: method})
		return nil
	}
	chi.Walk(r, walkFunc)
	return routes
}
