## ADDED Requirements

### Requirement: Entity Management

The system SHALL support entities (users, groups, roles, bots) that can receive data-entry requests. Each entity SHALL have a unique identifier (`entity_id`) and optional aliases (email, handle, URN) for routing.

#### Scenario: Create entity with handle

- **WHEN** a requestor creates an entity with handle "alice@example.com"
- **THEN** the entity is stored with a UUID `entity_id` and the handle is available for routing

#### Scenario: Resolve entity by handle

- **WHEN** a request references entity handle "alice@example.com"
- **THEN** the system resolves it to the corresponding `entity_id`

### Requirement: Request Creation

The system SHALL allow requestors to create data-entry requests specifying:

- Target entity (by ID or handle)
- Schema (JSON Schema, JSON example, or `$ref` URL)
- UI hints (help text, field descriptions, examples)
- Prefill data (default values)
- Deadlines and attention timestamps
- File upload policies

#### Scenario: Create request with JSON Schema

- **WHEN** a requestor creates a request with a JSON Schema and target entity
- **THEN** the request is stored with status `PENDING` and assigned a unique request ID (ULID)

#### Scenario: Create request with JSON example

- **WHEN** a requestor creates a request with a JSON example object
- **THEN** the system infers a minimal schema and stores the request

#### Scenario: Create request with schema reference

- **WHEN** a requestor creates a request with `{"$ref": "https://example.com/schema.json"}`
- **THEN** the system fetches and validates the schema (if allowlisted) and stores it

### Requirement: Request Lifecycle

The system SHALL manage request status transitions: `PENDING → CLAIMED → ANSWERED | CANCELLED | EXPIRED`.

#### Scenario: Claim request

- **WHEN** a responder claims a `PENDING` request
- **THEN** the request status changes to `CLAIMED` and is locked to that responder

#### Scenario: Answer request

- **WHEN** a responder submits a response for a `CLAIMED` request
- **AND** the response validates against the request's schema
- **THEN** the request status changes to `ANSWERED` and the response is stored

#### Scenario: Cancel request

- **WHEN** a requestor or responder cancels a request
- **THEN** the request status changes to `CANCELLED`

#### Scenario: Expire request

- **WHEN** a request's `deadline_at` is reached
- **THEN** the request status changes to `EXPIRED`

### Requirement: Schema Validation

The system SHALL validate all responses against the request's JSON Schema before accepting them.

#### Scenario: Valid response

- **WHEN** a responder submits a response that matches the schema
- **THEN** the response is accepted and stored

#### Scenario: Invalid response

- **WHEN** a responder submits a response that violates the schema
- **THEN** the system rejects the response with validation errors

### Requirement: Response Storage

The system SHALL store responses with:

- Request ID reference
- Responder entity ID
- Validated JSON payload
- Optional file references
- Timestamp

#### Scenario: Store response with payload

- **WHEN** a responder submits a valid response
- **THEN** the response is stored with all metadata and linked to the request

#### Scenario: Store response with file references

- **WHEN** a responder submits a response with file references
- **THEN** the response includes file metadata (name, URL, size, checksum)
