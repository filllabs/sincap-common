package server

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"gitlab.com/sincap/sincap-common/logging"

	"go.uber.org/zap"

	"github.com/go-chi/chi"
	"github.com/rcrowley/go-metrics"
	gometrics "github.com/rcrowley/go-metrics"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	return n, err
}
func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("not a Hijacker")
}

// RequestMetrics logs response code and average service times metrics
func RequestMetrics() (func(next http.Handler) http.Handler, gometrics.Registry) {
	registry := metrics.NewRegistry()
	return func(next http.Handler) http.Handler {
		logging.Logger.Info("Metrics registered.", zap.String("source", "metrics.middleware"))

		fn := func(w http.ResponseWriter, r *http.Request) {
			rctx := chi.RouteContext(r.Context())
			nc := chi.NewRouteContext()
			rctx.Routes.Match(nc, r.Method, r.URL.Path)
			routePattern := nc.RoutePattern()

			if ws := r.Header.Get("Upgrade"); ws == "websocket" {
				next.ServeHTTP(w, r)
				return
			}

			// Save service histograms
			his := gometrics.GetOrRegisterHistogram("api."+routePattern+"."+r.Method, registry, gometrics.NewUniformSample(1028))
			t1 := time.Now()
			sw := statusWriter{ResponseWriter: w}

			next.ServeHTTP(&sw, r)

			// Save counts by Response Status
			his.Update(time.Since(t1).Nanoseconds())
			responseStatus := strconv.Itoa(sw.status)
			g := gometrics.GetOrRegisterCounter("responseStatus."+responseStatus, registry)
			g.Inc(1)

		}
		return http.HandlerFunc(fn)
	}, registry
}
