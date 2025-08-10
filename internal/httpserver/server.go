// internal/httpserver/server.go
package httpserver

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"highperf-api/internal/handlers"
)

func NewRouter() http.Handler {
	r := httprouter.New()

	// Middlewares (orden importa)
	var h http.Handler = r
	h = withServerHeader(h)
	h = withRecover(h)   // no panics cruzando el límite
	h = withTimeouts(h)  // ctx deadlines por ruta
	h = withRateLimit(h) // token bucket
	h = withBreaker(h)   // circuit breaker
	h = withMetrics(h)   // prom/otel
	h = withTracing(h)   // otel

	// Rutas
	r.GET("/healthz", handlers.Healthz)
	r.GET("/users/:id", handlers.GetUser)
	r.POST("/users", handlers.CreateUser)
	r.GET("/files/*path", handlers.ServeStatic) // zero-copy

	return h
}
