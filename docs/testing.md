# Testing Guide

This document describes how to run tests for PxBox using docker-compose.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for local testing)

## Running Tests

### Unit Tests (No Docker Required)

```bash
# Run all unit tests
make test-unit

# Or directly with go test
go test -v -short ./internal/...
```

### Integration Tests (Requires Docker Compose)

Integration tests use docker-compose to spin up PostgreSQL and Redis containers.

```bash
# Run integration tests (automatically starts/stops services)
make test-integration

# Or manually:
./test/run.sh
```

The test script will:

1. Start PostgreSQL and Redis containers
2. Wait for services to be ready
3. Run database migrations
4. Execute integration tests
5. Clean up containers

### Using Docker Compose Directly

```bash
# Start test services
docker-compose -f test/docker-compose.test.yaml up -d

# Set environment variables
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/pxbox_test?sslmode=disable"
export TEST_REDIS_ADDR="localhost:6380"

# Run tests
go test -v ./test/...

# Stop services
docker-compose -f test/docker-compose.test.yaml down
```

## Test Structure

- `internal/service/*_test.go` - Unit tests for services
- `internal/schema/*_test.go` - Unit tests for schema validation
- `test/integration_test.go` - Integration tests for API endpoints
- `test/helpers.go` - Test helper functions

## Writing Tests

### Unit Tests

Unit tests should be placed alongside the code they test:

```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMyFunction(t *testing.T) {
    // Test implementation
    assert.Equal(t, expected, actual)
}
```

### Integration Tests

Integration tests should be in the `test/` directory and use the test helpers:

```go
package test

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestMyIntegration(t *testing.T) {
    server, _, cleanup := setupTestServer(t)
    defer cleanup()

    // Test implementation
}
```

## Continuous Integration

For CI/CD pipelines, use:

```bash
# Build and test in Docker
docker-compose build
docker-compose run --rm pxbox-api go test -v ./...
```

## Troubleshooting

### Tests fail with "connection refused"

- Ensure Docker containers are running: `docker-compose -f test/docker-compose.test.yaml ps`
- Check service health: `docker-compose -f test/docker-compose.test.yaml logs`

### Migration errors

- Migrations are run automatically by the test script
- If migrations fail, check database connection string
- Ensure PostgreSQL container is healthy before running tests

### Port conflicts

- Test services use ports 5433 (PostgreSQL) and 6380 (Redis)
- Ensure these ports are available or modify `test/docker-compose.test.yaml`
