#!/bin/bash

# API Template Generator
# Converts this project into a reusable template for new REST APIs

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 High Performance Go API Template Generator${NC}"
echo

# Check if new project name is provided
if [ -z "$1" ]; then
    echo -e "${RED}Usage: $0 <new-project-name> [output-directory]${NC}"
    echo "Example: $0 my-awesome-api ../projects/"
    exit 1
fi

NEW_PROJECT_NAME="$1"
OUTPUT_DIR="${2:-../}"
TARGET_DIR="$OUTPUT_DIR/$NEW_PROJECT_NAME"

# Validate project name
if [[ ! "$NEW_PROJECT_NAME" =~ ^[a-zA-Z][a-zA-Z0-9_-]*$ ]]; then
    echo -e "${RED}❌ Invalid project name. Use only letters, numbers, hyphens, and underscores.${NC}"
    exit 1
fi

# Check if target directory already exists
if [ -d "$TARGET_DIR" ]; then
    echo -e "${RED}❌ Directory $TARGET_DIR already exists!${NC}"
    exit 1
fi

echo -e "${YELLOW}📦 Creating new project: $NEW_PROJECT_NAME${NC}"
echo -e "${YELLOW}📁 Target directory: $TARGET_DIR${NC}"
echo

# Create target directory
mkdir -p "$TARGET_DIR"

# Copy all files except excluded ones
echo -e "${BLUE}📋 Copying project files...${NC}"
rsync -av \
    --exclude='.git' \
    --exclude='bin' \
    --exclude='*.log' \
    --exclude='coverage.out' \
    --exclude='coverage.html' \
    --exclude='node_modules' \
    --exclude='.DS_Store' \
    --exclude='template-generator.sh' \
    ./ "$TARGET_DIR/"

# Replace module name in go.mod
echo -e "${BLUE}🔧 Updating module name...${NC}"
sed -i.bak "s/module highperf-api/module $NEW_PROJECT_NAME/g" "$TARGET_DIR/go.mod"
rm "$TARGET_DIR/go.mod.bak"

# Replace import paths in all Go files
echo -e "${BLUE}🔧 Updating import paths...${NC}"
find "$TARGET_DIR" -name "*.go" -type f -exec sed -i.bak "s|highperf-api/|$NEW_PROJECT_NAME/|g" {} \;
find "$TARGET_DIR" -name "*.go.bak" -delete

# Update docker-compose.yml
echo -e "${BLUE}🔧 Updating Docker configuration...${NC}"
sed -i.bak "s/api_db/${NEW_PROJECT_NAME}_db/g" "$TARGET_DIR/docker-compose.yml"
rm "$TARGET_DIR/docker-compose.yml.bak"

# Update Makefile
echo -e "${BLUE}🔧 Updating Makefile...${NC}"
sed -i.bak "s/APP_NAME := highperf-api/APP_NAME := $NEW_PROJECT_NAME/g" "$TARGET_DIR/Makefile"
rm "$TARGET_DIR/Makefile.bak"

# Create project-specific .env.example
echo -e "${BLUE}🔧 Creating environment template...${NC}"
cat > "$TARGET_DIR/.env.example" << EOF
# Server Configuration
PORT=8080
SERVER_HOST=0.0.0.0
READ_TIMEOUT=5s
WRITE_TIMEOUT=10s
IDLE_TIMEOUT=60s
GRACEFUL_TIMEOUT=15s

# Database Configuration
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=${NEW_PROJECT_NAME}_db
DB_USER=postgres
DB_PASSWORD=your-db-password
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Authentication
JWT_SECRET=your-super-secret-jwt-key-change-in-production
TOKEN_EXPIRY=24h
REFRESH_EXPIRY=168h

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_ADD_SOURCE=true

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9090
METRICS_PATH=/metrics
EOF

# Create migration directory structure
echo -e "${BLUE}🔧 Creating migration structure...${NC}"
mkdir -p "$TARGET_DIR/migrations"
cat > "$TARGET_DIR/migrations/001_create_users_table.up.sql" << EOF
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
EOF

cat > "$TARGET_DIR/migrations/001_create_users_table.down.sql" << EOF
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
EOF

# Create monitoring configuration
echo -e "${BLUE}🔧 Setting up monitoring...${NC}"
mkdir -p "$TARGET_DIR/monitoring"

cat > "$TARGET_DIR/monitoring/prometheus.yml" << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: '${NEW_PROJECT_NAME}'
    static_configs:
      - targets: ['api:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
EOF

# Create updated README
echo -e "${BLUE}📝 Generating project README...${NC}"
cat > "$TARGET_DIR/README.md" << EOF
# $NEW_PROJECT_NAME

A high-performance REST API built with Go, featuring advanced performance optimizations and production-ready patterns.

## 🚀 Quick Start

\`\`\`bash
# Clone and setup
git clone <your-repo-url>
cd $NEW_PROJECT_NAME
cp .env.example .env
# Edit .env with your configuration

# Run with Docker Compose
make docker-compose-up

# Or run locally
make deps
make run
\`\`\`

## 📋 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| \`GET\` | \`/healthz\` | Health check |
| \`GET\` | \`/users/:id\` | Get user by ID |
| \`POST\` | \`/users\` | Create user |
| \`PUT\` | \`/users/:id\` | Update user |
| \`DELETE\` | \`/users/:id\` | Delete user |
| \`GET\` | \`/users\` | List users (paginated) |

## 🛠️ Development

\`\`\`bash
# Setup development environment
make dev-setup

# Run tests
make test

# Run with coverage
make test-coverage

# Format and lint
make check

# Build
make build
\`\`\`

## 📦 Deployment

\`\`\`bash
# Build Docker image
make docker-build

# Deploy with docker-compose
make docker-compose-up

# View logs
make docker-compose-logs
\`\`\`

## 🏗️ Architecture

This API is built with:

- **High Performance**: Buffer pooling, zero-copy JSON, connection pooling
- **Reliability**: Circuit breaker, rate limiting, graceful shutdown
- **Security**: JWT authentication, password hashing, request validation
- **Observability**: Structured logging, metrics, health checks
- **Database**: PostgreSQL with repository pattern
- **Caching**: Redis for session storage and caching

## 📄 License

MIT License
EOF

# Clean up and initialize git
echo -e "${BLUE}🔧 Finalizing project...${NC}"
cd "$TARGET_DIR"

# Initialize git repository
git init
git add .
git commit -m "Initial commit: Generated from high-performance API template

Features:
- High-performance HTTP server with optimizations
- JWT authentication and password hashing
- Database layer with repository pattern
- Request validation and error handling
- Structured logging and configuration management
- Docker and docker-compose setup
- Comprehensive testing suite
- Production-ready middleware stack"

echo
echo -e "${GREEN}✅ Project '$NEW_PROJECT_NAME' created successfully!${NC}"
echo
echo -e "${YELLOW}📍 Location: $TARGET_DIR${NC}"
echo
echo -e "${BLUE}🎯 Next steps:${NC}"
echo "1. cd $TARGET_DIR"
echo "2. cp .env.example .env"
echo "3. Edit .env with your configuration"
echo "4. make docker-compose-up"
echo "5. Visit http://localhost:8080/healthz"
echo
echo -e "${BLUE}📚 Available commands:${NC}"
echo "• make help              - Show all available commands"
echo "• make dev-setup         - Setup development environment"
echo "• make test             - Run tests"
echo "• make docker-compose-up - Start all services"
echo
echo -e "${GREEN}🎉 Happy coding!${NC}"
