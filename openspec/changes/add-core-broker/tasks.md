## 1. Database Schema and Migrations

- [x] 1.1 Create PostgreSQL migration for `entities` table
- [x] 1.2 Create PostgreSQL migration for `requests` table with all fields
- [x] 1.3 Create PostgreSQL migration for `responses` table
- [x] 1.4 Create PostgreSQL migration for `flows` table with cursor field
- [x] 1.5 Create PostgreSQL migration for `reminders` table
- [x] 1.6 Add indexes for performance (entity_id, status, deadline_at, attention_at, flow_id)
- [x] 1.7 Set up goose migration tooling

## 2. Go Backend Foundation

- [x] 2.1 Initialize Go module and project structure (`cmd/pxbox-api`, `internal/`)
- [x] 2.2 Set up database connection (pgx pool)
- [x] 2.3 Set up Redis connection (go-redis)
- [x] 2.4 Set up logging (zap)
- [x] 2.5 Set up configuration management (env vars, viper optional)
- [x] 2.6 Create sqlc query files for entities, requests, responses, flows
- [x] 2.7 Generate type-safe database code with sqlc (manual implementation)

## 3. Core Domain Models

- [x] 3.1 Define `Entity` model and repository
- [x] 3.2 Define `Request` model with status enum
- [x] 3.3 Define `Response` model
- [x] 3.4 Define `Flow` model with status enum and cursor type
- [x] 3.5 Implement entity resolution (by ID or handle)

## 4. JSON Schema Validation

- [x] 4.1 Integrate jsonschema/v5 compiler
- [x] 4.2 Implement schema compilation and caching (LRU cache)
- [x] 4.3 Implement schema validation for responses
- [x] 4.4 Handle schema inference from JSON examples
- [x] 4.5 Implement `$ref` resolution with allowlist (optional)

## 5. Request Management Service

- [x] 5.1 Implement `CreateRequest` service method
- [x] 5.2 Implement `GetRequest` service method
- [x] 5.3 Implement `ClaimRequest` service method
- [x] 5.4 Implement `PostResponse` service method with validation
- [x] 5.5 Implement `CancelRequest` service method
- [x] 5.6 Implement request status transitions

## 6. Flow Management Service

- [x] 6.1 Implement `CreateFlow` service method
- [x] 6.2 Implement `GetFlow` service method
- [x] 6.3 Implement `ResumeFlow` service method
- [x] 6.4 Implement `CancelFlow` service method
- [x] 6.5 Implement flow checkpoint persistence
- [x] 6.6 Implement flow runner interface and basic implementation
- [x] 6.7 Implement flow recovery on application start

## 7. File Storage Service

- [x] 7.1 Define `Storage` interface (PresignPut, PresignGet, Delete)
- [x] 7.2 Implement local filesystem storage backend
- [x] 7.3 Implement file policy validation (size, MIME type)
- [x] 7.4 Implement presigned URL generation for local storage
- [x] 7.5 Add file metadata storage (name, URL, size, checksum)

## 8. Redis Pub/Sub and Events

- [x] 8.1 Implement Redis pub/sub wrapper
- [x] 8.2 Implement event publishing (request.created, request.answered, etc.)
- [x] 8.3 Implement Redis Streams for event replay (or in-memory ring for dev)
- [x] 8.4 Implement sequence number tracking per channel
- [x] 8.5 Implement event acknowledgment tracking

## 9. WebSocket Transport

- [x] 9.1 Set up gorilla/websocket server
- [x] 9.2 Implement WebSocket connection handler with JWT auth
- [x] 9.3 Implement message envelope parsing (JSON)
- [x] 9.4 Implement channel subscription management
- [x] 9.5 Implement command routing (createRequest, postResponse, etc.)
- [x] 9.6 Implement event broadcasting to subscribed channels
- [x] 9.7 Implement sequence numbers and acknowledgment
- [x] 9.8 Implement resume from last position
- [x] 9.9 Implement connection lifecycle (connect, disconnect, reconnect)
- [x] 9.10 Add heartbeat/ping-pong for connection health

## 10. REST Transport

- [x] 10.1 Set up chi router
- [x] 10.2 Implement JWT authentication middleware
- [x] 10.3 Implement `POST /v1/requests` endpoint
- [x] 10.4 Implement `GET /v1/requests/:id` endpoint
- [x] 10.5 Implement `POST /v1/requests/:id/cancel` endpoint
- [x] 10.6 Implement `GET /v1/entities/:id/queue` endpoint with pagination
- [x] 10.7 Implement `POST /v1/requests/:id/claim` endpoint
- [x] 10.8 Implement `POST /v1/requests/:id/response` endpoint
- [x] 10.9 Implement `POST /v1/flows` endpoint
- [x] 10.10 Implement `GET /v1/flows/:id` endpoint
- [x] 10.11 Implement `POST /v1/flows/:id/resume` endpoint
- [x] 10.12 Implement `POST /v1/flows/:id/cancel` endpoint
- [x] 10.13 Implement `GET /v1/inquiries` endpoint with filters
- [x] 10.14 Implement `POST /v1/inquiries/:id/markRead` endpoint
- [x] 10.15 Implement `POST /v1/inquiries/:id/snooze` endpoint
- [x] 10.16 Implement `POST /v1/inquiries/:id/cancel` endpoint
- [x] 10.17 Implement `DELETE /v1/inquiries/:id` endpoint
- [x] 10.18 Implement `POST /v1/files/sign` endpoint
- [x] 10.19 Add consistent error response formatting
- [x] 10.20 Add request/response logging middleware

## 11. Background Jobs (asynq)

- [x] 11.1 Set up asynq server and client
- [x] 11.2 Implement deadline notification job (1h before deadline)
- [x] 11.3 Implement expiry job (mark EXPIRED at deadline)
- [x] 11.4 Implement auto-cancel job (after grace period)
- [x] 11.5 Implement attention notification job
- [x] 11.6 Implement reminder job for snoozed inquiries
- [x] 11.7 Schedule jobs when requests/flows are created

## 12. Frontend Setup (Vite + Preact)

- [x] 12.1 Initialize Vite project with Preact template
- [x] 12.2 Configure Yarn v2 Berry (PnP)
- [x] 12.3 Set up preact/compat for React ecosystem compatibility
- [x] 12.4 Install @rjsf/core and dependencies
- [x] 12.5 Set up Zustand for state management
- [x] 12.6 Configure build and dev scripts

## 13. Frontend WebSocket Client

- [x] 13.1 Implement WebSocket client wrapper class
- [x] 13.2 Implement connection management (connect, disconnect, reconnect)
- [x] 13.3 Implement channel subscription
- [x] 13.4 Implement command sending (createRequest, postResponse, etc.)
- [x] 13.5 Implement event handling and acknowledgment
- [x] 13.6 Implement resume from last position
- [x] 13.7 Create `useBrokerWS` Preact hook

## 14. Frontend Inbox UI

- [x] 14.1 Create inbox page component
- [x] 14.2 Implement inquiry list with filtering (status, entity)
- [x] 14.3 Implement inquiry grouping (needs attention, due soon, all)
- [x] 14.4 Implement inquiry card component with deadline display
- [x] 14.5 Implement mark as read functionality
- [x] 14.6 Implement snooze functionality with date picker
- [x] 14.7 Implement cancel inquiry functionality
- [x] 14.8 Implement delete inquiry functionality
- [x] 14.9 Add real-time updates via WebSocket events

## 15. Frontend Form Rendering

- [x] 15.1 Create request form page component
- [x] 15.2 Integrate @rjsf/core form renderer
- [x] 15.3 Implement schema loading from request
- [x] 15.4 Implement UI hints rendering (help text, examples)
- [x] 15.5 Implement prefill data population
- [x] 15.6 Implement form validation
- [x] 15.7 Implement file upload UI (presigned URL flow)
- [x] 15.8 Implement response submission (WebSocket + REST fallback)
- [x] 15.9 Add loading and error states

## 16. Python Test Client

- [x] 16.1 Create Python project structure
- [x] 16.2 Implement REST client for request creation
- [x] 16.3 Implement REST client for polling request status
- [x] 16.4 Implement WebSocket client (websockets library)
- [x] 16.5 Implement WebSocket command/event handling
- [x] 16.6 Implement resumable flow demonstration:
  - [x] 16.6.1 Create flow that requests input
  - [x] 16.6.2 Suspend flow and save state
  - [x] 16.6.3 Restart application
  - [x] 16.6.4 Resume flow from checkpoint
- [x] 16.7 Add example scenarios (shipping address, user profile, etc.)

## 17. Testing

- [x] 17.1 Write unit tests for request management service
- [x] 17.2 Write unit tests for flow management service
- [x] 17.3 Write unit tests for schema validation
- [x] 17.4 Write integration tests for REST endpoints
- [x] 17.5 Write integration tests for WebSocket commands
- [x] 17.6 Write E2E test with Python client
- [x] 17.7 Test flow suspend/resume across restarts
- [x] 17.8 Test deadline and notification jobs (requires background jobs implementation)

## 18. Documentation

- [x] 18.1 Document REST API endpoints (OpenAPI/Swagger)
- [x] 18.2 Document WebSocket protocol and message formats
- [x] 18.3 Document flow checkpoint format
- [x] 18.4 Create README with setup instructions
- [x] 18.5 Document deployment requirements (PostgreSQL, Redis)
- [x] 18.6 Add code comments for complex logic

## 19. DevOps and Deployment

- [x] 19.1 Create Dockerfile for Go backend
- [x] 19.2 Create docker-compose.yml (PostgreSQL, Redis, broker)
- [x] 19.3 Set up development environment setup script
- [x] 19.4 Configure environment variables
- [x] 19.5 Add health check endpoints (`/healthz`, `/readyz`)
