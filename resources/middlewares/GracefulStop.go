package middlewares

import (
	"net/http"
	"sync"
)

// WaitGrouped adds wait group add and done calls for each handler. Useful especially for graceful stops.
func WaitGrouped(wg *sync.WaitGroup) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wg.Add(1)
			next.ServeHTTP(w, r)
			wg.Done()
		})
	}
}
