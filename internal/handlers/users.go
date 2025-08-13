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

// internal/handlers/users.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"highperf-api/internal/auth"
	"highperf-api/internal/encoding/jsonx"
	"highperf-api/internal/models"
	"highperf-api/internal/repository"
	"highperf-api/internal/validator"
)

// UserHandler encapsula las dependencias para los manejadores de usuario.
type UserHandler struct {
	repo   repository.UserRepository
	hasher *auth.PasswordHasher
}

// NewUserHandler crea una nueva instancia de UserHandler.
func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{
		repo:   repo,
		hasher: auth.NewPasswordHasher(), // Instanciamos el hasher
	}
}

func Healthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// GetUser es el manejador para GET /users/:id
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx, cancel := context.WithTimeout(r.Context(), 80*time.Millisecond)
	defer cancel()

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "timeout", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	userResponse := user.ToResponse()
	buf := jsonx.GetBuffer()
	defer jsonx.PutBuffer(buf)

	if err := jsonx.MarshalToBuffer(userResponse, buf); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = w.Write(buf.Bytes())
}

// CreateUser es el manejador para POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	dec := jsonx.NewDecoder(r.Body)
	var req models.UserCreateRequest
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := validator.Validate(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	passwordHash, err := h.hasher.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user := req.ToModel(passwordHash)

	createdUser, err := h.repo.Create(r.Context(), user)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Escribimos la respuesta del usuario creado directamente al writer
	enc := json.NewEncoder(w)
	enc.Encode(createdUser.ToResponse())
}

// ServeStatic sirve archivos estáticos.
func ServeStatic(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.ServeFile(w, r, "./public"+ps.ByName("path"))
}
