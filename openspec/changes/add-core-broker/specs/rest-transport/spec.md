## ADDED Requirements

### Requirement: REST API Endpoints

The system SHALL provide REST endpoints for all core operations, enabling polling-based access for clients that cannot use WebSocket.

#### Scenario: Create request via POST

- **WHEN** a requestor sends `POST /v1/requests` with request data
- **THEN** the system creates the request and returns `{requestId, status: "PENDING"}`

#### Scenario: Get request via GET

- **WHEN** a requestor sends `GET /v1/requests/:id`
- **THEN** the system returns the request details including current status

### Requirement: Request Management Endpoints

The system SHALL provide REST endpoints for:

- `POST /v1/requests` - Create request
- `GET /v1/requests/:id` - Get request details
- `POST /v1/requests/:id/cancel` - Cancel request
- `GET /v1/entities/:id/queue` - List pending inquiries for entity
- `POST /v1/requests/:id/claim` - Claim a request
- `POST /v1/requests/:id/response` - Submit response

#### Scenario: List entity queue

- **WHEN** a responder sends `GET /v1/entities/:id/queue?status=PENDING`
- **THEN** the system returns paginated list of pending requests for that entity

#### Scenario: Claim request via REST

- **WHEN** a responder sends `POST /v1/requests/:id/claim`
- **THEN** the request status changes to `CLAIMED` if available

### Requirement: Flow Management Endpoints

The system SHALL provide REST endpoints for:

- `POST /v1/flows` - Create flow
- `GET /v1/flows/:id` - Get flow details
- `POST /v1/flows/:id/resume` - Resume suspended flow
- `POST /v1/flows/:id/cancel` - Cancel flow

#### Scenario: Create flow via REST

- **WHEN** a requestor sends `POST /v1/flows` with flow definition
- **THEN** the system creates the flow and returns flow ID

#### Scenario: Resume flow via REST

- **WHEN** a client sends `POST /v1/flows/:id/resume` with event data
- **THEN** the system resumes the flow from its checkpoint

### Requirement: Inquiry Management Endpoints

The system SHALL provide REST endpoints for:

- `GET /v1/inquiries` - List inquiries with filters
- `POST /v1/inquiries/:id/markRead` - Mark inquiry as read
- `POST /v1/inquiries/:id/snooze` - Snooze inquiry
- `POST /v1/inquiries/:id/cancel` - Cancel inquiry
- `DELETE /v1/inquiries/:id` - Soft delete inquiry

#### Scenario: List inquiries with filters

- **WHEN** a user sends `GET /v1/inquiries?status=PENDING&entityId=...`
- **THEN** the system returns filtered, paginated list of inquiries

#### Scenario: Snooze inquiry

- **WHEN** a user sends `POST /v1/inquiries/:id/snooze` with `remindAt` timestamp
- **THEN** the system schedules a reminder and updates the inquiry

### Requirement: Authentication

All REST endpoints SHALL require JWT authentication via `Authorization: Bearer <token>` header.

#### Scenario: Authenticated request

- **WHEN** a client sends a request with valid JWT token
- **THEN** the request is processed

#### Scenario: Unauthenticated request

- **WHEN** a client sends a request without valid token
- **THEN** the system returns 401 Unauthorized

### Requirement: Polling Pattern

REST endpoints SHALL support polling for status updates, with appropriate caching headers.

#### Scenario: Poll request status

- **WHEN** a requestor polls `GET /v1/requests/:id` repeatedly
- **THEN** the response includes current status and `Last-Modified` header for caching

### Requirement: Error Responses

REST endpoints SHALL return consistent error responses with appropriate HTTP status codes and error details.

#### Scenario: Validation error

- **WHEN** a request fails schema validation
- **THEN** the system returns 422 Unprocessable Entity with validation error details

#### Scenario: Not found

- **WHEN** a request references a non-existent resource
- **THEN** the system returns 404 Not Found
