package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	Healthz(w, req, nil)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if body != "ok" {
		t.Errorf("Expected body 'ok', got %q", body)
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{"valid user ID", "123", http.StatusOK},
		{"invalid user ID", "invalid", http.StatusBadRequest},
		{"negative user ID", "-1", http.StatusOK}, // parseint accepts negative numbers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			params := httprouter.Params{
				{Key: "id", Value: tt.userID},
			}

			GetUser(w, req, params)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == http.StatusOK {
				contentType := resp.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type 'application/json', got %q", contentType)
				}

				body := w.Body.String()
				if !strings.Contains(body, `"id":`) {
					t.Errorf("Expected JSON with id field, got %q", body)
				}
			}
		})
	}
}

func TestGetUserTimeout(t *testing.T) {
	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(1 * time.Millisecond)

	req := httptest.NewRequest("GET", "/users/123", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	params := httprouter.Params{
		{Key: "id", Value: "123"},
	}

	GetUser(w, req, params)

	resp := w.Result()
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("Expected status %d, got %d", http.StatusGatewayTimeout, resp.StatusCode)
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{"valid JSON", `{"name":"John Doe"}`, http.StatusCreated},
		{"invalid JSON", `{"name":}`, http.StatusBadRequest},
		{"empty body", ``, http.StatusBadRequest},
		{"malformed JSON", `not json`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/users", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			CreateUser(w, req, nil)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestServeStatic(t *testing.T) {
	// Create a temporary file for testing
	req := httptest.NewRequest("GET", "/files/test.txt", nil)
	w := httptest.NewRecorder()

	params := httprouter.Params{
		{Key: "path", Value: "/nonexistent.txt"},
	}

	ServeStatic(w, req, params)

	resp := w.Result()
	// Should return 404 for nonexistent file
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d for nonexistent file, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestFetchUser(t *testing.T) {
	ctx := context.Background()

	user, err := fetchUser(ctx, 123)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if user.ID != 123 {
		t.Errorf("Expected user ID 123, got %d", user.ID)
	}

	if user.Name != "Ada" {
		t.Errorf("Expected user name 'Ada', got %q", user.Name)
	}
}

func TestFetchUserTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(1 * time.Millisecond)

	_, err := fetchUser(ctx, 123)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}

func BenchmarkGetUser(b *testing.B) {
	req := httptest.NewRequest("GET", "/users/123", nil)
	params := httprouter.Params{
		{Key: "id", Value: "123"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		GetUser(w, req, params)
	}
}

func BenchmarkHealthz(b *testing.B) {
	req := httptest.NewRequest("GET", "/healthz", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		Healthz(w, req, nil)
	}
}
