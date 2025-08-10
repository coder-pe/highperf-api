// internal/httpserver/middleware.go
package httpserver

import (
	"context"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func withServerHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "go-highperf")
		next.ServeHTTP(w, r)
	})
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				http.Error(w, "internal", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func withTimeouts(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dl := 100 * time.Millisecond
		ctx, cancel := context.WithTimeout(r.Context(), dl)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Token bucket simple (per-process). Para producción usa un lib distribuido (Redis/leaky).
func withRateLimit(next http.Handler) http.Handler {
	const cap = 1000
	const refill = 1000
	const per = time.Second

	var mu sync.Mutex
	tokens := cap
	last := time.Now()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		mu.Lock()
		// refill
		elapsed := now.Sub(last)
		if elapsed > 0 {
			n := int(float64(refill) * elapsed.Seconds() / per.Seconds())
			if n > 0 {
				tokens = min(cap, tokens+n)
				last = now
			}
		}
		if tokens <= 0 {
			mu.Unlock()
			http.Error(w, "busy", http.StatusTooManyRequests)
			return
		}
		tokens--
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Placeholder middleware for metrics collection
func withMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement Prometheus/OpenTelemetry metrics
		next.ServeHTTP(w, r)
	})
}

// Placeholder middleware for distributed tracing
func withTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement OpenTelemetry tracing
		next.ServeHTTP(w, r)
	})
}

// Circuit breaker muy compacto (para demo). En prod: sony/gobreaker.
func withBreaker(next http.Handler) http.Handler {
	const failureThreshold = 20
	const openFor = 2 * time.Second

	var mu sync.Mutex
	failures := 0
	openUntil := time.Time{}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		if time.Now().Before(openUntil) {
			mu.Unlock()
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		mu.Unlock()

		// Jitter para evitar thundering herd
		jitter := time.Duration(rand.Intn(2)) * time.Millisecond
		time.Sleep(jitter)

		rr := &respRecorder{ResponseWriter: w, code: 200}
		next.ServeHTTP(rr, r)

		mu.Lock()
		if rr.code >= 500 {
			failures++
			if failures >= failureThreshold {
				openUntil = time.Now().Add(openFor)
				failures = 0
			}
		} else {
			// éxito resetea lentamente
			failures = int(math.Max(0, float64(failures-2)))
		}
		mu.Unlock()
	})
}

type respRecorder struct {
	http.ResponseWriter
	code int
}

func (r *respRecorder) WriteHeader(status int) {
	r.code = status
	r.ResponseWriter.WriteHeader(status)
}
