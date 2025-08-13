/*
 * Copyright (C) 2025 Miguel Mamani <miguel.coder.per@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

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
				http.Error(w, "internal server error", http.StatusInternalServerError)
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

// --- Rate Limiter con estado encapsulado ---

type RateLimiter struct {
	mu     sync.Mutex
	tokens int
	last   time.Time
	cap    int
	refill int
	per    time.Duration
}

func NewRateLimiter(capacity, refillRate int, per time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens: capacity,
		last:   time.Now(),
		cap:    capacity,
		refill: refillRate,
		per:    per,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		rl.mu.Lock()

		elapsed := now.Sub(rl.last)
		if elapsed > 0 {
			n := int(float64(rl.refill) * elapsed.Seconds() / rl.per.Seconds())
			if n > 0 {
				rl.tokens = min(rl.cap, rl.tokens+n)
				rl.last = now
			}
		}

		if rl.tokens <= 0 {
			rl.mu.Unlock()
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		rl.tokens--
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- Circuit Breaker con estado encapsulado ---

type CircuitBreaker struct {
	mu               sync.Mutex
	failures         int
	openUntil        time.Time
	failureThreshold int
	openFor          time.Duration
}

func NewCircuitBreaker(failureThreshold int, openFor time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		openFor:          openFor,
	}
}

func (cb *CircuitBreaker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cb.mu.Lock()
		if time.Now().Before(cb.openUntil) {
			cb.mu.Unlock()
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		cb.mu.Unlock()

		jitter := time.Duration(rand.Intn(2)) * time.Millisecond
		time.Sleep(jitter)

		rr := &respRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rr, r)

		cb.mu.Lock()
		if rr.code >= http.StatusInternalServerError {
			cb.failures++
			if cb.failures >= cb.failureThreshold {
				cb.openUntil = time.Now().Add(cb.openFor)
				cb.failures = 0
			}
		} else {
			cb.failures = int(math.Max(0, float64(cb.failures-2)))
		}
		cb.mu.Unlock()
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

// --- Middlewares de Observabilidad (Implementación básica) ---

func withMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Metrics: received request for %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func withTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Tracing: starting trace for %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
