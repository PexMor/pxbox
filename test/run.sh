#!/bin/bash
set -e

echo "Cleaning up any existing test containers..."
docker-compose -f docker-compose.test.yaml down 2>/dev/null || true

# Stop any containers using our test ports
docker ps --filter "publish=6380" --format "{{.ID}}" | xargs -r docker stop 2>/dev/null || true
docker ps --filter "publish=5433" --format "{{.ID}}" | xargs -r docker stop 2>/dev/null || true
docker ps --filter "publish=8082" --format "{{.ID}}" | xargs -r docker stop 2>/dev/null || true

echo "Starting test services..."
docker-compose -f docker-compose.test.yaml up -d

echo "Waiting for services to be ready..."
sleep 5

echo "Running migrations..."
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/pxbox_test?sslmode=disable"
export DATABASE_URL="$TEST_DATABASE_URL"
go run ./cmd/pxbox-api migrate 2>&1 | grep -v "already exists" || true

echo "Running tests..."
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/pxbox_test?sslmode=disable"
export TEST_REDIS_ADDR="localhost:6380"
export TEST_API_URL="http://localhost:8082"
go test -v ./test/... -timeout 60s

echo "Cleaning up..."
docker-compose -f docker-compose.test.yaml down
