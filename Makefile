.PHONY: help test test-unit test-integration build run serve docker-build docker-up docker-down docker-test migrate

# Default target
.DEFAULT_GOAL := help

# Show help message
help:
	@echo "PxBox Makefile Commands:"
	@echo ""
	@echo "  make build          - Build the pxbox-api binary"
	@echo "  make run            - Run pxbox-api server (requires postgres & redis)"
	@echo "  make serve          - Alias for 'make run'"
	@echo "  make migrate        - Run database migrations"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests (unit + integration)"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build Docker images"
	@echo "  make docker-up     - Start Docker services"
	@echo "  make docker-down   - Stop Docker services"
	@echo "  make docker-test   - Run tests in Docker"
	@echo "  make dev           - Start all services with docker-compose"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean         - Clean build artifacts and storage"

# Build the application
build:
	go build -o bin/pxbox-api ./cmd/pxbox-api

# Run the pxbox-api server (requires postgres & redis to be running)
run: build
	@echo "Starting pxbox-api server..."
	@echo "Make sure postgres and redis are running (use 'make docker-up' to start them)"
	@mkdir -p storage
	@DATABASE_URL="postgres://postgres:postgres@localhost:5432/pxbox?sslmode=disable" \
	REDIS_ADDR="localhost:6379" \
	ADDR=":8082" \
	STORAGE_BASE_DIR="./storage" \
	STORAGE_BASE_URL="http://localhost:8082" \
	JWT_SECRET="default-secret-key-change-in-production" \
	./bin/pxbox-api serve

# Alias for run
serve: run

# Run unit tests
test-unit:
	go test -v -short ./internal/...

# Run integration tests (requires docker-compose)
test-integration:
	./test/run.sh

# Run all tests
test: test-unit test-integration

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-test:
	docker-compose up -d postgres redis
	sleep 5
	docker-compose run --rm pxbox-api migrate
	docker-compose run --rm pxbox-api go test -v ./test/...
	docker-compose down

# Run migrations
migrate:
	go run ./cmd/pxbox-api migrate

# Development
dev:
	docker-compose up

# Clean
clean:
	rm -rf bin/
	rm -rf storage/
	docker-compose down -v

