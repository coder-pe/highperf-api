# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Building and Running
- `go run cmd/api/main.go` - Start the HTTP server on port 8080
- `go build -o bin/api cmd/api/main.go` - Build the binary
- `go mod tidy` - Clean up module dependencies

### Testing and Linting
- `./run-tests.sh` - Run comprehensive test suite with coverage and benchmarks
- `./test-api.sh` - Test API endpoints (requires running server)
- `go test ./...` - Run all tests
- `go test -bench=. -benchmem ./...` - Run benchmarks
- `go test -race -coverprofile=coverage.out ./...` - Run tests with race detection and coverage
- `go vet ./...` - Run static analysis
- `gofmt -w .` - Format all Go files

## Architecture

This is a high-performance Go HTTP API built with performance optimization in mind. Key architectural patterns:

### Core Structure
- **Entry Point**: `cmd/api/main.go` - Contains server setup with optimized timeouts, TLS config, and graceful shutdown
- **HTTP Layer**: `internal/httpserver/` - Router setup and middleware chain
- **Handlers**: `internal/handlers/users.go` - Request handlers with context timeouts and optimized JSON processing
- **Custom JSON**: `internal/encoding/jsonx/` - Optimized JSON encoding with buffer pooling to minimize allocations

### Performance Optimizations
- **Buffer Pooling**: Uses `sync.Pool` for JSON buffer reuse (jsonx package)
- **Zero-Copy JSON**: Direct buffer-to-response writing without intermediate allocations
- **Context Timeouts**: Per-handler deadline management (80ms for hot paths like GetUser)
- **Middleware Chain**: Ordered middleware stack including rate limiting, circuit breaker, recovery, and metrics
- **SO_REUSEPORT**: Prepared for multi-process scaling (reusePortListen function)

### Middleware Stack (applied in order)
1. Server header identification
2. Panic recovery
3. Request timeouts (100ms default)
4. Rate limiting (token bucket: 1000 req/sec capacity)
5. Circuit breaker (opens after 20 failures for 2s)
6. Metrics collection (placeholder)
7. Tracing (placeholder)

### Dependencies
- `github.com/julienschmidt/httprouter` - High-performance HTTP router
- Go 1.24.5+ required

### Key Files
- `internal/httpserver/server.go:10` - Router configuration and middleware setup
- `internal/httpserver/middleware.go:42` - Rate limiting implementation
- `internal/httpserver/middleware.go:78` - Circuit breaker implementation  
- `internal/encoding/jsonx/pool.go:9` - Buffer pool with 1MB cap limit
- `cmd/api/main.go:35` - SO_REUSEPORT listener setup

## Development Notes

- The codebase prioritizes performance over abstraction
- Middleware functions are designed to fail fast with appropriate HTTP status codes
- JSON encoding avoids `json.Marshal` in favor of streaming to pooled buffers
- Context deadlines are used extensively to prevent cascading failures
- Static file serving uses `http.ServeFile` for zero-copy kernel operations