package server

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	gometrics "github.com/rcrowley/go-metrics"
	"gitlab.com/sincap/sincap-common/logging"
	"gitlab.com/sincap/sincap-common/resources/middlewares"
)

// AddDefaultMiddlewares adds all predefined middlewares to the router.
// You may add RequestMetrics manually by adding 	AddRequestMetrics(r)
func AddDefaultMiddlewares(r chi.Router, config Config) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	allowedURL := "*"
	if config.Cors {
		logging.Logger.Named("Server").Info("Adding CORS")
		allowedURL = config.FrontendURL
	}
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{allowedURL},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		ExposedHeaders:   []string{"Link", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)
	if config.SecurityHeaders {
		logging.Logger.Named("Server").Info("Adding SecurityHeaders")
		r.Use(middlewares.SecurityHeaders)
	}
}

// AddRequestMetrics adds a middleware for measuring request metrict for all paths
// Dont't forget to add some code to write the registry. Like,
// go gometrics.WriteJSON(gometrics.DefaultRegistry, time.Second*time.Duration(interval), writer)
func AddRequestMetrics(r *chi.Mux) gometrics.Registry {
	mdlwr, registry := RequestMetrics()
	r.Use(mdlwr)
	return registry
}
