# PxBox - Universal Data Entry Proxy

A universal, schema-driven data-entry broker that enables dynamic form generation from JSON Schema and handles data collection workflows. PxBox acts as a proxy between requestors (agents, applications) and responders (users, entities), managing the entire lifecycle of data-entry requests.

## Features

- **Schema-Driven Forms**: Generate HTML forms dynamically from JSON Schema
- **Dual Transport**: REST API (polling) and WebSocket (real-time bidirectional)
- **Durable Flows**: Suspendable workflows that survive application restarts
- **File Uploads**: Support for local filesystem and S3/GCS (via presigned URLs)
- **Event-Driven**: Real-time notifications via Redis pub/sub and WebSocket
- **Background Jobs**: Automated deadline notifications, expiry, and reminders

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis 7+
- Node.js 18+ (for frontend)
- Docker & Docker Compose (optional, for local development)

### Using Docker Compose (Recommended)

1. Clone the repository:

```bash
git clone <repository-url>
cd pxbox
```

2. Start services:

```bash
docker-compose up -d
```

This starts:

- PostgreSQL on port 5432
- Redis on port 6379
- pxbox-api on port 8080

3. Run migrations:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/pxbox?sslmode=disable"
go run ./cmd/pxbox-api migrate
```

4. Start the API server:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/pxbox?sslmode=disable"
export REDIS_ADDR="localhost:6379"
go run ./cmd/pxbox-api serve
```

### Manual Setup

1. **Database Setup**:

```bash
createdb pxbox
export DATABASE_URL="postgres://user:password@localhost:5432/pxbox?sslmode=disable"
go run ./cmd/pxbox-api migrate
```

2. **Redis Setup**:

```bash
redis-server
export REDIS_ADDR="localhost:6379"
```

3. **Start API Server**:

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/pxbox?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export ADDR=":8080"
go run ./cmd/pxbox-api serve
```

4. **Web UI Setup**:

```bash
cd frontend
yarn install
yarn dev
```

## Configuration

Key environment variables:

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_ADDR`: Redis address (default: `localhost:6379`)
- `ADDR`: HTTP server address (default: `:8080`)
- `JWT_SECRET`: Secret key for JWT authentication

See [Architecture Guide](AGENTS.md) for complete configuration options.

## Documentation

- **[Quick Start Guide](docs/quickstart.md)**: Get started quickly with Docker Compose, Python examples, and web UI
- **[Architecture Guide](AGENTS.md)**: Technical overview, design decisions, and development patterns
- **[REST API](docs/api.md)**: Complete REST endpoint documentation
- **[WebSocket Protocol](docs/websocket.md)**: WebSocket message formats and examples
- **[Flow Checkpoints](docs/flow-checkpoint.md)**: Flow checkpoint format and resume patterns
- **[Schema References](docs/schema-refs.md)**: JSON Schema `$ref` resolution with allowlist
- **[Testing Guide](docs/testing.md)**: Testing instructions and best practices

**Quick API Reference:**

- Base URL: `http://localhost:8080/v1`
- WebSocket: `ws://localhost:8080/v1/ws`
- Key endpoints: `POST /requests`, `GET /requests/:id`, `POST /requests/:id/response`, `POST /files/sign`, `GET /inquiries`, `POST /flows`, `POST /flows/:id/resume`

## Development

See [Architecture Guide](AGENTS.md) for:

- Project structure
- Development workflow
- Common tasks
- Testing strategy
- Database migrations

## Deployment

### Requirements

- **PostgreSQL**: 14+ with JSONB support
- **Redis**: 7+ for pub/sub and job queue
- **Go**: 1.21+ for backend
- **Node.js**: 18+ for frontend build

### Production Considerations

- Set `JWT_SECRET` to a secure random value
- Use connection pooling for PostgreSQL
- Configure Redis persistence if needed
- Set up reverse proxy (nginx/traefik) for HTTPS
- Enable CORS appropriately for frontend
- Configure file storage (S3/GCS) for production
- Set up monitoring and logging

See [Architecture Guide](AGENTS.md) for detailed deployment guidance.

## License

[Add your license here]
