# Project Context

## Purpose

PxBox is a universal, schema-driven data-entry broker that enables any application, CLI, or server to request structured data input from users or automated agents. The system acts as a proxy between requestors (who need data) and responders (who provide data), supporting long-running flows that can suspend and resume across application restarts.

**Core Value Proposition**: Decouple data collection from user interaction, enabling asynchronous, durable workflows where human input can be requested, paused, and resumed seamlessly.

## Tech Stack

### Backend

- **Language**: Go 1.21+
- **HTTP Router**: chi
- **WebSocket**: gorilla/websocket
- **Database**: PostgreSQL (via pgx + sqlc for type-safe queries)
- **Pub/Sub**: Redis (go-redis)
- **Background Jobs**: asynq
- **JSON Schema**: github.com/santhosh-tekuri/jsonschema/v5
- **Auth**: golang-jwt/jwt/v5
- **IDs**: ulid (time-sortable)
- **Migrations**: pressly/goose
- **Logging**: uber-go/zap
- **Object Storage**: minio-go or aws-sdk-go-v2 (S3-compatible)

### Frontend

- **Build Tool**: Vite
- **Framework**: Preact (with preact/compat for React ecosystem compatibility)
- **Package Manager**: Yarn v2 Berry (PnP)
- **Form Rendering**: @rjsf/core (via preact/compat)
- **State Management**: Zustand + Immer
- **Styling**: Modern CSS (consider CSS Modules or Tailwind)

### Testing

- **Test Client**: Python (demonstrates REST, WebSocket, and resumable flows)

## Project Conventions

### Code Style

- **Go**: Follow standard `gofmt` formatting, use `golangci-lint` for linting
- **TypeScript/JavaScript**: Use Prettier with standard config, ESLint for Preact
- **Naming**: Use clear, descriptive names; prefer verb-noun for functions (`createRequest`, `getInquiry`)
- **File Organization**: Group by feature/capability, not by type

### Architecture Patterns

- **API-First**: All functionality exposed via well-defined API contracts (REST + WebSocket)
- **Transport Agnostic**: Core business logic independent of transport layer
- **Event-Driven**: Use Redis pub/sub for real-time notifications and event fanout
- **Durable State**: All flows and requests persist to PostgreSQL; support suspend/resume
- **Schema-Driven**: JSON Schema as the single source of truth for data validation and UI generation

### Testing Strategy

- **Unit Tests**: Test business logic independently of transport
- **Integration Tests**: Test API endpoints (REST + WebSocket) against test database
- **E2E Tests**: Python test app demonstrates real-world usage patterns
- **Test Data**: Use factories/fixtures for consistent test data

### Git Workflow

- Use feature branches; merge via pull requests
- Follow conventional commits for changelog generation
- Tag releases with semantic versioning

## Domain Context

### Core Concepts

**Entity**: A routable target (user, group, role, bot) identified by `entity_id` and optional aliases (email, handle, URN).

**Request (Inquiry)**: A data-entry request containing:

- Schema (JSON Schema, JSON example, or `$ref` URL)
- UI hints (help text, field descriptions, examples)
- Prefill data (default values)
- Target entity
- Deadlines and attention timestamps
- File upload policies

**Response**: The data provided by a responder, validated against the request's schema, optionally including file references.

**Flow**: A durable, suspendable workflow that can emit inquiries, wait for responses, and resume execution. Flows survive application restarts via checkpointed state (`cursor`).

**Transport**: Two transport options:

- **WebSocket (Primary)**: Bidirectional, real-time, supports commands and events
- **REST (Polling)**: Stateless HTTP endpoints for compatibility and simple clients

### Request Lifecycle

`PENDING → CLAIMED → ANSWERED | CANCELLED | EXPIRED`

### Flow Lifecycle

`RUNNING → SUSPENDED/WAITING_INPUT → RUNNING → COMPLETED | CANCELLED | FAILED`

## Important Constraints

- **API-First**: All features must be accessible via API before UI implementation
- **WebSocket Primary**: WebSocket is the preferred transport; REST provides polling fallback
- **Durability**: Flows and requests must survive application restarts
- **Schema Validation**: All responses must validate against provided JSON Schema
- **File Storage**: Support local filesystem initially; S3/GCS support can be added later
- **Timezone Handling**: Store all timestamps in UTC; display in user's timezone (Europe/Prague default)

## External Dependencies

- **PostgreSQL**: Primary data store for entities, requests, responses, flows
- **Redis**: Pub/sub for real-time events, job queue backend
- **Object Storage** (optional): S3/GCS/MinIO for file uploads (can use local filesystem initially)
- **JSON Schema Validator**: github.com/santhosh-tekuri/jsonschema/v5
