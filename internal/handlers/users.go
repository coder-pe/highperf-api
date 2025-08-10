// internal/handlers/users.go
package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"highperf-api/internal/encoding/jsonx"
)

type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func Healthz(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok")) // pequeño, está bien copiar
}

// GET /users/:id
func GetUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Deadline corto por ruta caliente
	ctx, cancel := context.WithTimeout(r.Context(), 80*time.Millisecond)
	defer cancel()

	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	user, err := fetchUser(ctx, id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "timeout", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// JSON con pool de buffers para minimizar allocs
	buf := jsonx.GetBuffer()
	defer jsonx.PutBuffer(buf)

	if err := jsonx.MarshalToBuffer(user, buf); err != nil {
		http.Error(w, "encode", http.StatusInternalServerError)
		return
	}

	// Evita copiar: escribe el []byte del buffer directo
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = w.Write(buf.Bytes())
}

// POST /users
func CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Streaming del body sin leer todo a memoria
	dec := jsonx.NewDecoder(r.Body) // json.Decoder envuelto con optimizaciones
	var u User
	if err := dec.Decode(&u); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	// ... validar y persistir
	w.WriteHeader(http.StatusCreated)
}

// Static files: usa sendfile bajo el capó (zero-copy kernel)
func ServeStatic(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.ServeFile(w, r, "./public"+ps.ByName("path"))
}

// Simulación de fetch (coloca tu repo/cache)
func fetchUser(ctx context.Context, id int64) (User, error) {
	select {
	case <-ctx.Done():
		return User{}, ctx.Err()
	case <-time.After(2 * time.Millisecond):
		return User{ID: id, Name: "Ada"}, nil
	}
}
