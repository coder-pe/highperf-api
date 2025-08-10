package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWithServerHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := withServerHeader(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	serverHeader := w.Header().Get("Server")
	if serverHeader != "go-highperf" {
		t.Errorf("Expected Server header 'go-highperf', got %q", serverHeader)
	}
}

func TestWithRecover(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := withRecover(panicHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Should not panic
	middleware.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d after panic, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "internal") {
		t.Errorf("Expected error message containing 'internal', got %q", body)
	}
}

func TestWithTimeouts(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			// Context was cancelled/timed out
			w.WriteHeader(http.StatusRequestTimeout)
		case <-time.After(200 * time.Millisecond):
			// Handler completed normally
			w.WriteHeader(http.StatusOK)
		}
	})

	middleware := withTimeouts(slowHandler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	middleware.ServeHTTP(w, req)
	duration := time.Since(start)

	// Should timeout around 100ms
	if duration > 150*time.Millisecond {
		t.Errorf("Expected timeout around 100ms, took %v", duration)
	}

	// The request context is created by the middleware, check the response status
	if w.Code != http.StatusRequestTimeout {
		t.Errorf("Expected timeout status %d, got %d", http.StatusRequestTimeout, w.Code)
	}
}

func TestWithRateLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := withRateLimit(handler)

	// Test that initial requests succeed
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed with status %d", i, w.Code)
		}
	}

	// Test concurrent access
	var wg sync.WaitGroup
	successes := 0
	failures := 0
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			mu.Lock()
			if w.Code == http.StatusOK {
				successes++
			} else {
				failures++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Should have some successful requests and possibly some rate limited
	if successes == 0 {
		t.Error("Expected some successful requests")
	}
}

func TestWithBreaker(t *testing.T) {
	failCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failCount++
		if failCount <= 20 { // Fail exactly the threshold
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	middleware := withBreaker(handler)

	// Send enough failing requests to trip the circuit breaker
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		expectedCode := http.StatusInternalServerError
		if i >= 20 { // After threshold, should be circuit broken
			expectedCode = http.StatusServiceUnavailable
		}

		if w.Code != expectedCode {
			t.Errorf("Expected %d on request %d, got %d", expectedCode, i, w.Code)
		}
	}

	// Next request should be circuit broken
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected circuit breaker to return 503, got %d", w.Code)
	}

	// After the timeout, circuit should reset
	time.Sleep(3 * time.Second) // Wait longer than openFor (2s)

	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	// Should allow the request through (handler now returns 200)
	if w.Code != http.StatusOK {
		t.Errorf("Expected circuit breaker to reset and return 200, got %d", w.Code)
	}
}

func TestRespRecorder(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := &respRecorder{ResponseWriter: w, code: 200}

	// Test default code
	if recorder.code != 200 {
		t.Errorf("Expected default code 200, got %d", recorder.code)
	}

	// Test WriteHeader
	recorder.WriteHeader(404)
	if recorder.code != 404 {
		t.Errorf("Expected code 404 after WriteHeader, got %d", recorder.code)
	}

	resp := w.Result()
	if resp.StatusCode != 404 {
		t.Errorf("Expected underlying ResponseWriter to have status 404, got %d", resp.StatusCode)
	}
}

func TestWithMetrics(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := withMetrics(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Should pass through without issues
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestWithTracing(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := withTracing(handler)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Should pass through without issues
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func BenchmarkWithServerHeader(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := withServerHeader(handler)
	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
	}
}

func BenchmarkWithRateLimit(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := withRateLimit(handler)
	req := httptest.NewRequest("GET", "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)
	}
}
