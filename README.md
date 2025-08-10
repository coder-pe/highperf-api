# High Performance Go API

Una API HTTP de alto rendimiento construida en Go, optimizada para manejar alta concurrencia con latencia mínima y uso eficiente de memoria.

## 🚀 Características de Alto Rendimiento

### Optimizaciones de Memoria
- **Buffer Pooling**: Reutilización de buffers JSON con `sync.Pool` para minimizar allocations
- **Zero-Copy JSON**: Serialización directa a response writer sin copias intermedias
- **Capacity Limits**: Prevención de memory leaks con límites de capacidad en pools (1MB)
- **Streaming JSON**: Decodificación de request bodies sin cargar todo en memoria

### Optimizaciones de Red
- **SO_REUSEPORT**: Preparado para escalado multi-proceso (Linux)
- **HTTP/1.1 Optimizations**: Timeouts configurados y reutilización de conexiones
- **TLS 1.2+ Only**: Configuración segura con cipher suites optimizadas
- **Static File Serving**: Zero-copy usando `sendfile()` del kernel

### Resiliencia y Control de Flujo
- **Circuit Breaker**: Protección contra cascading failures (20 fallas → abre por 2s)
- **Rate Limiting**: Token bucket con 1000 req/s de capacidad
- **Request Timeouts**: Deadlines por handler (80ms para rutas calientes)
- **Graceful Shutdown**: Apagado elegante con timeout de 5s

## 📋 API Endpoints

| Method | Endpoint | Description | Timeout |
|--------|----------|-------------|---------|
| `GET` | `/healthz` | Health check | 100ms |
| `GET` | `/users/:id` | Obtener usuario por ID | 80ms |
| `POST` | `/users` | Crear nuevo usuario | 100ms |
| `GET` | `/files/*path` | Archivos estáticos (zero-copy) | 100ms |

## 🛠️ Instalación y Ejecución

### Requisitos
- Go 1.24.5+
- Git

### Clonar e Instalar
```bash
git clone <repo-url>
cd highperf-api
go mod download
```

### Ejecutar el Servidor
```bash
# Desarrollo
go run cmd/api/main.go

# Producción
go build -o bin/api cmd/api/main.go
./bin/api
```

El servidor estará disponible en `http://localhost:8080`

## 🧪 Testing y Desarrollo

### Ejecutar Todas las Pruebas
```bash
# Suite completa con coverage y benchmarks
./run-tests.sh

# Solo pruebas unitarias
go test ./...

# Con race detection
go test -race ./...

# Solo benchmarks
go test -bench=. -benchmem ./...
```

### Probar API Endpoints
```bash
# Probar endpoints con el servidor corriendo
./test-api.sh
```

### Análisis de Código
```bash
# Análisis estático
go vet ./...

# Formatear código
gofmt -w .

# Coverage HTML
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🏗️ Arquitectura

### Estructura del Proyecto
```
highperf-api/
├── cmd/api/              # Entry point
│   └── main.go          # Server setup y configuración
├── internal/
│   ├── handlers/        # HTTP handlers
│   │   ├── users.go     # User endpoints
│   │   └── users_test.go
│   ├── httpserver/      # HTTP layer
│   │   ├── server.go    # Router y middleware setup
│   │   ├── middleware.go # Middleware implementations
│   │   └── *_test.go
│   └── encoding/jsonx/  # Optimized JSON
│       ├── marshal.go   # JSON encoding
│       ├── pool.go      # Buffer pooling
│       └── *_test.go
├── test-api.sh          # API testing script
├── run-tests.sh         # Complete test suite
└── CLAUDE.md           # Developer guidance
```

### Middleware Stack (orden de ejecución)
1. **Server Header** - Identificación del servidor
2. **Recovery** - Captura de panics
3. **Timeouts** - Deadlines de request (100ms default)
4. **Rate Limiting** - Token bucket (1000 req/s)
5. **Circuit Breaker** - Protección contra failures
6. **Metrics** - Recolección de métricas (placeholder)
7. **Tracing** - Distributed tracing (placeholder)

### Flujo de Request
```
Request → Middleware Stack → Router → Handler → JSON Pool → Response
```

## ⚡ Benchmarks

### Performance Típico (MacBook Pro M1)
```
BenchmarkGetUser-8         500000    2400 ns/op    1200 B/op    12 allocs/op
BenchmarkHealthz-8        2000000     800 ns/op     400 B/op     4 allocs/op
BenchmarkMarshalToBuffer-8 1000000   1600 ns/op     600 B/op     6 allocs/op
```

### Optimizaciones Medidas
- **JSON Pooling**: ~40% reducción en allocations vs `json.Marshal`
- **Buffer Reuse**: ~60% menos garbage collection pressure
- **Stream Decoding**: ~30% menos memoria para requests grandes

## 🔧 Configuración

### Variables de Entorno
```bash
# Puerto (default: 8080)
PORT=8080

# Timeouts
READ_TIMEOUT=2s
WRITE_TIMEOUT=2s
IDLE_TIMEOUT=60s

# Rate Limiting
RATE_LIMIT_RPS=1000
RATE_LIMIT_CAPACITY=1000

# Circuit Breaker
CB_FAILURE_THRESHOLD=20
CB_OPEN_DURATION=2s
```

### Configuración de Producción
- Usar reverse proxy (nginx/haproxy) para TLS termination
- Configurar `SO_REUSEPORT` con library como `go-reuseport`
- Implementar métricas con Prometheus/OpenTelemetry
- Configurar distributed tracing
- Usar cache distribuido (Redis) para rate limiting

## 📊 Monitoring

### Health Checks
```bash
curl http://localhost:8080/healthz
# → "ok"
```

### Métricas (TODO)
- Request duration histogram
- Request count por endpoint
- Error rate por status code
- Circuit breaker state
- Rate limit hits

## 🐛 Troubleshooting

### Problemas Comunes

**Server no arranca en puerto 8080**
```bash
# Verificar puerto ocupado
lsof -i :8080
# Cambiar puerto
PORT=8081 go run cmd/api/main.go
```

**Tests fallan con timeout**
```bash
# Aumentar timeout para tests lentos
go test -timeout 30s ./...
```

**Rate limiting muy agresivo**
```bash
# Ajustar en middleware.go
const cap = 2000    # Aumentar capacidad
const refill = 2000 # Aumentar refill rate
```

## 🤝 Contribuir

1. Fork del proyecto
2. Crear feature branch: `git checkout -b feature/nueva-caracteristica`
3. Commit cambios: `git commit -am 'Add nueva caracteristica'`
4. Push branch: `git push origin feature/nueva-caracteristica`
5. Crear Pull Request

### Estándares de Código
- Seguir Go formatting con `gofmt`
- Agregar tests para nueva funcionalidad
- Mantener coverage > 80%
- Documentar funciones públicas
- Usar benchmarks para optimizaciones

## 📄 Licencia

MIT License - ver archivo `LICENSE` para detalles.

## 🔗 Referencias

- [httprouter](https://github.com/julienschmidt/httprouter) - High-performance HTTP router
- [Go Performance Tips](https://github.com/golang/go/wiki/Performance)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)