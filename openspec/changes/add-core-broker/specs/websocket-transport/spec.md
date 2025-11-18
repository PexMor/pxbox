## ADDED Requirements

### Requirement: WebSocket Connection

The system SHALL provide a WebSocket endpoint (`/v1/ws`) that accepts connections with JWT authentication via subprotocol or Authorization header.

#### Scenario: Connect with JWT

- **WHEN** a client connects to `/v1/ws` with valid JWT token
- **THEN** the connection is established and authenticated

#### Scenario: Reject invalid token

- **WHEN** a client connects with invalid or missing JWT
- **THEN** the connection is rejected with 401 Unauthorized

### Requirement: Message Envelope

The system SHALL use a JSON envelope format for all WebSocket messages with fields: `id`, `type`, `channel`, `op`, `data`, `seq`, `since`, `meta`.

#### Scenario: Send command message

- **WHEN** a client sends a command message
- **THEN** the message includes `type: "cmd"`, `op` (operation name), `channel`, and `data` payload

#### Scenario: Receive event message

- **WHEN** the server sends an event to a client
- **THEN** the message includes `type: "event"`, `channel`, `seq` (sequence number), and `data` payload

### Requirement: Channel Subscription

The system SHALL support subscribing to logical channels: `entity:<entityId>`, `request:<requestId>`, `requestor:<clientId>`, `template:<templateId>`.

#### Scenario: Subscribe to entity inbox

- **WHEN** a responder subscribes to `entity:<entityId>`
- **THEN** the responder receives all events for that entity's inbox

#### Scenario: Subscribe to request updates

- **WHEN** a client subscribes to `request:<requestId>`
- **THEN** the client receives all events for that specific request

### Requirement: Sequence Numbers and Acknowledgment

The system SHALL assign sequence numbers per channel and support client acknowledgment for at-least-once delivery.

#### Scenario: Receive sequenced event

- **WHEN** the server sends an event on a channel
- **THEN** the event includes a monotonically increasing `seq` number

#### Scenario: Acknowledge event

- **WHEN** a client sends `{type: "ack", channel, seq: N}`
- **THEN** the server records the acknowledgment and can resume from `seq: N+1` if reconnected

### Requirement: Resume from Last Position

The system SHALL support resuming a channel subscription from the last acknowledged sequence number.

#### Scenario: Resume subscription

- **WHEN** a client sends `{type: "resume", channel, since: N}`
- **THEN** the server replays all events from `seq: N+1` to the current position

### Requirement: WebSocket Commands

The system SHALL support the following commands via WebSocket:

- `createRequest`: Create a new data-entry request
- `getRequest`: Retrieve request details
- `claimRequest`: Claim a pending request
- `postResponse`: Submit a response
- `cancelRequest`: Cancel a request
- `createFlow`: Create a new flow
- `resumeFlow`: Resume a suspended flow
- `cancelFlow`: Cancel a flow
- `listInquiries`: List pending inquiries for an entity

#### Scenario: Create request via WebSocket

- **WHEN** a requestor sends `{type: "cmd", op: "createRequest", data: {...}}`
- **THEN** the system creates the request and responds with `{type: "event", data: {event: "request.created", requestId: "..."}}`

#### Scenario: Post response via WebSocket

- **WHEN** a responder sends `{type: "cmd", op: "postResponse", data: {requestId, payload}}`
- **THEN** the system validates, stores the response, and emits `request.answered` events

### Requirement: Event Broadcasting

The system SHALL broadcast events to all subscribed clients on relevant channels.

#### Scenario: Broadcast request created

- **WHEN** a request is created for entity `E`
- **THEN** all clients subscribed to `entity:<E>` receive `request.created` event

#### Scenario: Broadcast response answered

- **WHEN** a response is submitted
- **THEN** events are sent to `request:<requestId>` and `requestor:<clientId>` channels

### Requirement: Connection Management

The system SHALL handle connection lifecycle: connect, disconnect, reconnect with resume support.

#### Scenario: Handle disconnect

- **WHEN** a WebSocket connection is closed
- **THEN** the server cleans up subscriptions and allows reconnection

#### Scenario: Reconnect with resume

- **WHEN** a client reconnects and sends resume commands
- **THEN** the server replays missed events from the last acknowledged position
