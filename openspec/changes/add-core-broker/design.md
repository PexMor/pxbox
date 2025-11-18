## Context

PxBox is a greenfield project implementing a universal data-entry broker. The system must support both simple one-off data requests and complex, long-running workflows that can suspend and resume. The architecture prioritizes API-first design with dual transport options (WebSocket primary, REST fallback) to accommodate diverse client needs.

## Goals / Non-Goals

### Goals

- **API-First**: All functionality accessible via well-defined API contracts before UI implementation
- **Transport Flexibility**: Support both WebSocket (real-time, bidirectional) and REST (polling) transports
- **Durability**: Flows and requests persist across application restarts
- **Schema-Driven**: JSON Schema as single source of truth for validation and UI generation
- **Extensibility**: File storage abstraction allowing local filesystem initially, S3/GCS later

### Non-Goals

- **Template Management**: Template versioning and feedback system (can be added later)
- **Advanced Auth**: OIDC/OAuth2 implementation details (assume JWT tokens provided)
- **Multi-tenancy**: Focus on single-instance deployment initially
- **Webhook Delivery**: External webhook callbacks (can use asynq jobs later)

## Decisions

### Decision: WebSocket as Primary Transport

**Rationale**: WebSocket provides bidirectional, real-time communication essential for:

- Instant notifications to responders
- Real-time status updates to requestors
- Efficient command/event pattern
- Firewall-friendly single-port communication

**Alternatives Considered**:

- HTTP long-polling: More complex, higher latency, multiple connections
- Server-Sent Events: One-way only, requires separate HTTP requests for commands

### Decision: Dual Transport Pattern

**Rationale**: REST provides compatibility for simple clients (CLI tools, curl scripts) and environments where WebSocket is restricted. Both transports access the same business logic layer.

**Implementation**: Transport-agnostic service layer; REST and WebSocket handlers call shared services.

### Decision: Flow Checkpointing via Cursor

**Rationale**: Store flow state as JSONB `cursor` field in PostgreSQL. This allows:

- Simple resume logic (read cursor, continue from checkpoint)
- Flexible state structure per flow type
- No need for complex state machine libraries initially

**Alternatives Considered**:

- Dedicated state machine library: Overkill for initial implementation
- Event sourcing: More complex, can be added later if needed

### Decision: Redis Streams for Event Replay

**Rationale**: Use Redis Streams (or in-memory ring buffer for dev) to enable:

- Per-channel sequence numbers for ordering
- Resume from last acknowledged position
- At-least-once delivery guarantees

**Alternatives Considered**:

- Database event log: Higher latency, more complex queries
- Message queue (RabbitMQ/Kafka): Overkill for initial scale

### Decision: Local Filesystem First for File Storage

**Rationale**: Start simple with local filesystem; abstract storage interface allows S3/GCS swap later without API changes.

**Storage Interface**: Define `Storage` interface with `PresignPut`, `PresignGet`, `Delete` methods.

## Risks / Trade-offs

### Risk: WebSocket Connection Management

**Mitigation**: Implement connection pooling, heartbeat/ping-pong, automatic reconnection with resume support. Monitor connection counts and implement backpressure.

### Risk: Flow State Corruption

**Mitigation**: Validate cursor structure on resume, provide clear error messages. Consider cursor versioning if schema changes needed.

### Risk: File Storage Scalability

**Mitigation**: Abstract storage interface from day one. Local filesystem works for single-instance deployments; S3/GCS swap is straightforward when needed.

### Trade-off: REST Polling vs WebSocket Efficiency

**Acceptance**: REST polling is less efficient but provides compatibility. Document WebSocket as preferred transport; REST for simple clients.

## Migration Plan

N/A - Greenfield project.

## Open Questions

- Should we support multiple responders claiming the same request? (Initial: single claimer)
- How should we handle schema `$ref` resolution? (Initial: allowlist domains, cache by hash)
- What's the maximum payload size for WebSocket messages? (Initial: 1MB, chunk larger payloads)
