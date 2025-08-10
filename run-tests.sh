#!/bin/bash

# Script para ejecutar todas las pruebas
set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Ejecutando pruebas unitarias...${NC}"

# Ejecutar pruebas con coverage
echo "Running tests with coverage..."
go test -v -race -coverprofile=coverage.out ./...

# Mostrar coverage
echo -e "\n${BLUE}📊 Coverage report:${NC}"
go tool cover -func=coverage.out

# Generar HTML coverage report
echo -e "\n${BLUE}📄 Generando reporte HTML...${NC}"
go tool cover -html=coverage.out -o coverage.html
echo "Coverage report saved to coverage.html"

# Ejecutar benchmarks
echo -e "\n${BLUE}🏃‍♂️ Ejecutando benchmarks...${NC}"
go test -bench=. -benchmem ./...

# Ejecutar vet para análisis estático
echo -e "\n${BLUE}🔍 Análisis estático con go vet...${NC}"
go vet ./...

# Verificar formato
echo -e "\n${BLUE}📝 Verificando formato de código...${NC}"
if ! gofmt -l . | grep -q .; then
    echo -e "${GREEN}✓ Código correctamente formateado${NC}"
else
    echo -e "${RED}✗ Código no está formateado correctamente:${NC}"
    gofmt -l .
    echo "Ejecuta: gofmt -w ."
    exit 1
fi

echo -e "\n${GREEN}✅ Todas las pruebas completadas exitosamente${NC}"