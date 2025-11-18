## ADDED Requirements

### Requirement: Python Test Client Application

The system SHALL include a Python test client application that demonstrates REST, WebSocket, and resumable flow usage patterns.

#### Scenario: Demonstrate REST API usage

- **WHEN** the Python client runs
- **THEN** it demonstrates creating requests, polling status, and submitting responses via REST endpoints

#### Scenario: Demonstrate WebSocket API usage

- **WHEN** the Python client runs
- **THEN** it demonstrates WebSocket connection, command sending, and event handling

#### Scenario: Demonstrate resumable flows

- **WHEN** the Python client starts a flow
- **AND** the application is interrupted
- **AND** the application is restarted
- **THEN** the client resumes the flow from its last checkpoint

### Requirement: REST Client Implementation

The Python test client SHALL implement REST API client functionality.

#### Scenario: Create request via REST

- **WHEN** the client calls `create_request()` with schema and entity
- **THEN** it sends `POST /v1/requests` and returns request ID

#### Scenario: Poll request status

- **WHEN** the client calls `poll_request_status(request_id)`
- **THEN** it repeatedly calls `GET /v1/requests/:id` until status changes

#### Scenario: Submit response via REST

- **WHEN** the client calls `submit_response(request_id, payload)`
- **THEN** it sends `POST /v1/requests/:id/response` with validated payload

### Requirement: WebSocket Client Implementation

The Python test client SHALL implement WebSocket client functionality.

#### Scenario: Connect to WebSocket

- **WHEN** the client calls `connect_websocket(token)`
- **THEN** it establishes a WebSocket connection to `/v1/ws` with JWT authentication

#### Scenario: Subscribe to channels

- **WHEN** the client calls `subscribe(channel)`
- **THEN** it sends a subscribe command and receives events on that channel

#### Scenario: Send commands via WebSocket

- **WHEN** the client calls `send_command(op, data)`
- **THEN** it sends a command message and receives acknowledgment or event response

#### Scenario: Handle events

- **WHEN** the client receives an event message
- **THEN** it processes the event and acknowledges it

### Requirement: Resumable Flow Demonstration

The Python test client SHALL demonstrate creating, suspending, and resuming flows.

#### Scenario: Create flow

- **WHEN** the client calls `create_flow(kind, initial_cursor)`
- **THEN** it creates a flow and stores the flow ID

#### Scenario: Suspend flow waiting for input

- **WHEN** a flow emits an inquiry
- **THEN** the client saves the flow state (flow ID, cursor) to disk

#### Scenario: Resume flow after restart

- **WHEN** the client restarts
- **AND** it loads saved flow state from disk
- **THEN** it calls `resume_flow(flow_id, event, data)` to continue execution

#### Scenario: Handle flow completion

- **WHEN** a flow completes
- **THEN** the client cleans up saved state and displays completion message

### Requirement: Example Scenarios

The Python test client SHALL include example scenarios demonstrating common use cases.

#### Scenario: Shipping address collection

- **WHEN** the client runs the shipping address example
- **THEN** it creates a request for shipping address, waits for response, and displays the result

#### Scenario: User profile collection

- **WHEN** the client runs the user profile example
- **THEN** it demonstrates a multi-step flow collecting user profile information

#### Scenario: File upload demonstration

- **WHEN** the client runs the file upload example
- **THEN** it demonstrates requesting file uploads and handling file references

### Requirement: State Persistence

The Python test client SHALL persist flow state to survive application restarts.

#### Scenario: Save flow state

- **WHEN** a flow suspends
- **THEN** the client saves flow ID, cursor, and metadata to a local file

#### Scenario: Load flow state

- **WHEN** the client starts
- **THEN** it checks for saved state files and offers to resume flows

#### Scenario: Clear completed flows

- **WHEN** a flow completes or is cancelled
- **THEN** the client removes the saved state file

### Requirement: Error Handling

The Python test client SHALL handle errors gracefully and provide useful feedback.

#### Scenario: Handle network errors

- **WHEN** a network request fails
- **THEN** the client displays an error message and allows retry

#### Scenario: Handle validation errors

- **WHEN** a response fails schema validation
- **THEN** the client displays validation error details

#### Scenario: Handle authentication errors

- **WHEN** authentication fails
- **THEN** the client prompts for valid credentials

### Requirement: Configuration

The Python test client SHALL support configuration via environment variables or config file.

#### Scenario: Configure broker URL

- **WHEN** the client starts
- **THEN** it reads broker URL from `BROKER_URL` environment variable or config file

#### Scenario: Configure authentication

- **WHEN** the client starts
- **THEN** it reads JWT token from `BROKER_TOKEN` environment variable or config file
