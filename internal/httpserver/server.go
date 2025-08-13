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

// internal/httpserver/server.go
package httpserver

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"highperf-api/internal/handlers"
	"highperf-api/internal/repository"
)

func NewRouter() http.Handler {
	r := httprouter.New()

	// =========================================================================
	// Inicialización de Dependencias (Simulado)
	var userRepo repository.UserRepository
	userRepo = repository.NewUserRepository(nil, nil) // CUIDADO: Esto es temporal
	userHandler := handlers.NewUserHandler(userRepo)
	// =========================================================================

	// =========================================================================
	// Inicialización de Middlewares
	rateLimiter := NewRateLimiter(1000, 1000, time.Second)
	circuitBreaker := NewCircuitBreaker(20, 2*time.Second)
	// =========================================================================

	// Middlewares (orden importa)
	var h http.Handler = r
	h = withServerHeader(h)
	h = withRecover(h)
	h = withTimeouts(h)
	h = rateLimiter.Middleware(h)
	h = circuitBreaker.Middleware(h)
	h = withMetrics(h)
	h = withTracing(h)

	// Rutas
	r.GET("/healthz", handlers.Healthz)
	r.GET("/users/:id", userHandler.GetUser)
	r.POST("/users", userHandler.CreateUser)
	r.GET("/files/*path", handlers.ServeStatic) // zero-copy

	return h
}
