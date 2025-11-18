<!-- OPENSPEC:START -->

# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:

- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:

- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

---

# PxBox Architecture & Development Guide

This document provides architectural overviews, design decisions, and technical guidance for both human developers and AI agents working with the PxBox codebase.

## Architecture Overview

PxBox is a universal, schema-driven data-entry broker that decouples data collection from user interaction. The system acts as a proxy between requestors (applications/agents) and responders (users/entities), supporting asynchronous, durable workflows.

### Core Components

- **Backend Service** (`cmd/pxbox-api/`): Go application providing REST and WebSocket APIs
- **Web UI** (`frontend/`): Preact web application (pxbox-wui) for responders
- **Database**: PostgreSQL for persistent storage (entities, requests, responses, flows)
- **Message Bus**: Redis for pub/sub and event streaming
- **Job Queue**: Redis-backed asynq for background tasks

### Key Design Decisions

#### API-First Architecture
All functionality is exposed via well-defined API contracts before UI implementation. This enables:
- Independent client development
- Testing via API before UI completion
- Multiple client types (CLI, web, mobile)

#### Dual Transport Pattern
- **WebSocket (Primary)**: Bidirectional, real-time communication for commands and events
- **REST (Fallback)**: Stateless HTTP endpoints for compatibility and simple clients

Both transports access the same service layer, ensuring consistency.

#### Durable Flows
Flows persist checkpointed state (`cursor`) to PostgreSQL, enabling suspend/resume across application restarts. Flow recovery runs on application startup.

#### Schema-Driven Validation
JSON Schema serves as the single source of truth for:
- Data validation
- UI form generation
- API contracts

See [Schema References](docs/schema-refs.md) for `$ref` resolution details.

## Technology Stack

### Backend
- **Go 1.21+**: Primary language
- **chi**: HTTP router
- **gorilla/websocket**: WebSocket server
- **pgx + sqlc**: Type-safe PostgreSQL queries
- **go-redis**: Redis client
- **asynq**: Background job processing
- **jsonschema/v5**: JSON Schema validation
- **golang-jwt/jwt/v5**: JWT authentication
- **ulid**: Time-sortable IDs
- **pressly/goose**: Database migrations (optional)
- **uber-go/zap**: Structured logging

### Frontend
- **Vite**: Build tool and dev server
- **Preact**: Lightweight React alternative
- **preact/compat**: React ecosystem compatibility
- **@rjsf/core**: JSON Schema Form renderer
- **Zustand**: State management
- **Yarn v2 Berry**: Package manager (PnP)

### Infrastructure
- **PostgreSQL 14+**: Primary data store
- **Redis 7+**: Pub/sub and job queue
- **Docker Compose**: Local development environment

## Project Structure

```
pxbox/
├── cmd/pxbox-api/       # Main application entry point
│   ├── main.go          # Server startup and initialization
│   ├── migrate.go       # Custom migration runner
│   └── goose.go         # Goose migration runner (optional)
├── internal/
│   ├── api/             # HTTP/WebSocket handlers
│   │   ├── routes.go    # Route registration
│   │   ├── requests.go  # Request endpoints
│   │   ├── websocket.go # WebSocket handler
│   │   └── ...
│   ├── db/              # Database layer
│   │   ├── pool.go      # Connection pool
│   │   └── queries.go   # Type-safe queries (sqlc)
│   ├── service/         # Business logic
│   │   ├── request.go   # Request management
│   │   ├── flow.go      # Flow management
│   │   ├── flow_runner.go # Flow execution
│   │   └── ...
│   ├── ws/              # WebSocket hub and connections
│   ├── pubsub/          # Redis pub/sub and streams
│   ├── jobs/            # Background job handlers
│   ├── schema/          # JSON Schema validation
│   └── storage/         # File storage abstraction
├── migrations/          # Database migrations
├── frontend/            # Vite + Preact web UI (pxbox-wui)
├── test/               # Integration and E2E tests
└── docs/               # Detailed documentation
```

## Key Patterns

### Service Layer Pattern
Business logic lives in `internal/service/`, independent of transport:
- `RequestService`: Request lifecycle management
- `FlowService`: Flow execution and recovery
- `EntityService`: Entity resolution

### Event-Driven Architecture
- Redis pub/sub for real-time event broadcasting
- Redis Streams for event replay and resume
- WebSocket for client notifications
- Background jobs for scheduled tasks (deadlines, reminders)

### Transport Abstraction
Both REST and WebSocket handlers call shared services:
```go
// REST handler
func (d Dependencies) createRequest(w http.ResponseWriter, r *http.Request) {
    req, err := d.RequestService.CreateRequest(ctx, input)
    // ...
}

// WebSocket command handler
func (h *CommandHandler) handleCreateRequest(...) {
    req, err := h.requestSvc.CreateRequest(ctx, input)
    // ...
}
```

## Development Workflow

### OpenSpec Integration
This project uses OpenSpec for spec-driven development. See `openspec/AGENTS.md` for:
- Creating change proposals
- Implementing approved changes
- Archiving completed changes

**Quick Reference:**
- `openspec list` - View active changes
- `openspec show <id>` - View change details
- `openspec validate <id> --strict` - Validate changes

### Testing Strategy
- **Unit Tests**: `internal/*/*_test.go` - Test business logic independently
- **Integration Tests**: `test/*_test.go` - Test API endpoints with test database
- **E2E Tests**: Python client demonstrates real-world usage

See [Testing Guide](docs/testing.md) for details.

### Database Migrations
Two migration systems supported:
1. **Custom Runner**: `go run ./cmd/pxbox-api migrate`
2. **Goose**: `go run ./cmd/pxbox-api goose-migrate`

Migrations are SQL files in `migrations/` directory.

## API Documentation

- **[REST API](docs/api.md)**: Complete REST endpoint documentation
- **[WebSocket Protocol](docs/websocket.md)**: WebSocket message formats and examples
- **[Flow Checkpoints](docs/flow-checkpoint.md)**: Flow checkpoint format and resume patterns

## Common Tasks

### Adding a New API Endpoint
1. Add route in `internal/api/routes.go`
2. Implement handler in appropriate file (`requests.go`, `flows.go`, etc.)
3. Add service method if business logic needed
4. Write integration test in `test/`
5. Update API documentation in `docs/api.md`

### Adding a WebSocket Command
1. Add command handler in `internal/ws/commands.go`
2. Register in `CommandHandler.HandleCommand`
3. Add service method if needed
4. Write integration test
5. Update WebSocket documentation in `docs/websocket.md`

### Adding a Background Job
1. Define job type in `internal/jobs/jobs.go`
2. Implement handler
3. Schedule job in service layer (e.g., `RequestService.CreateRequest`)
4. Write test in `test/jobs_test.go`

## Configuration

Key environment variables:
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_ADDR`: Redis address (default: `localhost:6379`)
- `ADDR`: HTTP server address (default: `:8080`)
- `JWT_SECRET`: Secret key for JWT authentication
- `STORAGE_BASE_DIR`: Local file storage directory
- `STORAGE_BASE_URL`: Base URL for file access

## Security Considerations

- JWT authentication for both REST and WebSocket
- File policy validation (size, MIME type, extensions)
- `$ref` URL allowlist for JSON Schema (prevents SSRF)
- Input validation via JSON Schema
- SQL injection prevention via sqlc type-safe queries

## Performance Considerations

- Connection pooling for PostgreSQL
- LRU cache for compiled JSON Schemas
- Redis pub/sub for efficient event broadcasting
- Background jobs for heavy operations (webhooks, notifications)

## Troubleshooting

### Flow Recovery Issues
- Check `flows` table for `RUNNING` or `SUSPENDED` status
- Verify `cursor` field contains valid JSON
- Check application logs for recovery errors

### WebSocket Connection Issues
- Verify JWT token validity
- Check Redis connection for pub/sub
- Review WebSocket hub logs

### File Upload Issues
- Verify `STORAGE_BASE_DIR` exists and is writable
- Check file policy validation errors
- Review presigned URL generation

## References

- [Project README](README.md) - Quick start and overview
- [OpenSpec Guide](openspec/AGENTS.md) - Spec-driven development workflow
- [API Documentation](docs/api.md) - REST API reference
- [WebSocket Protocol](docs/websocket.md) - WebSocket protocol details
- [Testing Guide](docs/testing.md) - Testing instructions
