package server

import (
	"net/http"

	"gitlab.com/sincap/sincap-common/logging"
)

func GracefulStop(wg *logging.LoggedWaitGroup) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wg.Addl("Route", 1)
			next.ServeHTTP(w, r)
			wg.Donel("Route")
		})
	}
}
