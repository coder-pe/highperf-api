#!/bin/bash

# Copyright (C) 2025 Miguel Mamani <miguel.coder.per@gmail.com>
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published
# by the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program. If not, see <https://www.gnu.org/licenses/>.

# Test script para la API
set -e

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🚀 Probando API en $BASE_URL"

# Test 1: Health check
echo -n "Testing /healthz... "
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/healthz")
if [ "$response" == "200" ]; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAILED (HTTP $response)${NC}"
    exit 1
fi

# Test 2: Get user
echo -n "Testing GET /users/123... "
response=$(curl -s -w "%{http_code}" "$BASE_URL/users/123")
http_code=$(echo "$response" | tail -c 4)
body=$(echo "$response" | head -c -4)

if [ "$http_code" == "200" ]; then
    echo -e "${GREEN}✓ OK${NC}"
    echo "  Response: $body"
else
    echo -e "${RED}✗ FAILED (HTTP $http_code)${NC}"
    echo "  Response: $body"
fi

# Test 3: Get user with invalid ID
echo -n "Testing GET /users/invalid... "
response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/users/invalid")
if [ "$response" == "400" ]; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAILED (Expected 400, got $response)${NC}"
fi

# Test 4: Create user
echo -n "Testing POST /users... "
response=$(curl -s -w "%{http_code}" -X POST "$BASE_URL/users" \
    -H "Content-Type: application/json" \
    -d '{"name":"Test User"}')
http_code=$(echo "$response" | tail -c 4)

if [ "$http_code" == "201" ]; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAILED (HTTP $http_code)${NC}"
fi

# Test 5: Create user with invalid JSON
echo -n "Testing POST /users with invalid JSON... "
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/users" \
    -H "Content-Type: application/json" \
    -d '{"invalid json}')
if [ "$response" == "400" ]; then
    echo -e "${GREEN}✓ OK${NC}"
else
    echo -e "${RED}✗ FAILED (Expected 400, got $response)${NC}"
fi

# Test 6: Rate limiting (envía muchas requests)
echo -n "Testing rate limiting... "
success_count=0
for i in {1..10}; do
    response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/healthz")
    if [ "$response" == "200" ]; then
        ((success_count++))
    fi
done

if [ $success_count -ge 8 ]; then
    echo -e "${GREEN}✓ OK (Rate limiting working)${NC}"
else
    echo -e "${RED}✗ FAILED (Too many requests blocked)${NC}"
fi

echo -e "\n${GREEN}✅ API tests completed${NC}"