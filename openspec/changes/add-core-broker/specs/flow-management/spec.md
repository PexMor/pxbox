## ADDED Requirements

### Requirement: Flow Creation

The system SHALL support creating durable flows that can suspend and resume execution.

#### Scenario: Create flow

- **WHEN** a requestor creates a flow with kind and initial cursor
- **THEN** the flow is stored with status `RUNNING` and assigned a flow ID

### Requirement: Flow State Persistence

The system SHALL persist flow state (cursor) to PostgreSQL, enabling resume after application restart.

#### Scenario: Persist flow cursor

- **WHEN** a flow updates its state
- **THEN** the cursor is saved to the database

#### Scenario: Resume after restart

- **WHEN** the application restarts
- **AND** a flow has status `RUNNING` or `WAITING_INPUT`
- **THEN** the flow can be resumed from its last checkpoint

### Requirement: Flow Suspension

The system SHALL support suspending flows when waiting for input, with checkpoint state.

#### Scenario: Suspend flow waiting for input

- **WHEN** a flow emits an inquiry and waits for response
- **THEN** the flow status changes to `WAITING_INPUT` and cursor is saved

#### Scenario: Link inquiry to flow

- **WHEN** a flow suspends waiting for input
- **THEN** the created inquiry is linked to the flow via `flow_id`

### Requirement: Flow Resumption

The system SHALL support resuming flows from suspension points when events occur (response answered, timeout, cancellation).

#### Scenario: Resume on response

- **WHEN** an inquiry linked to a flow is answered
- **THEN** the flow resumes execution with the response data in its cursor

#### Scenario: Resume on timeout

- **WHEN** an inquiry's deadline expires
- **THEN** the flow resumes on the timeout branch with timeout event

#### Scenario: Resume on cancellation

- **WHEN** an inquiry is cancelled
- **THEN** the flow resumes on the cancellation branch

### Requirement: Flow Lifecycle

The system SHALL manage flow status transitions: `RUNNING → WAITING_INPUT → RUNNING → COMPLETED | CANCELLED | FAILED`.

#### Scenario: Complete flow

- **WHEN** a flow finishes all steps
- **THEN** the flow status changes to `COMPLETED`

#### Scenario: Cancel flow

- **WHEN** a flow is cancelled (by user or system)
- **THEN** the flow status changes to `CANCELLED` and all open inquiries are cancelled

#### Scenario: Fail flow

- **WHEN** a flow encounters an unhandled error
- **THEN** the flow status changes to `FAILED` and error is logged

### Requirement: Flow Checkpointing

The system SHALL store flow checkpoints (cursor) as JSONB, allowing flexible state per flow type.

#### Scenario: Update checkpoint

- **WHEN** a flow progresses to a new step
- **THEN** the cursor is updated with new state

#### Scenario: Resume from checkpoint

- **WHEN** a flow resumes
- **THEN** execution continues from the state stored in cursor

### Requirement: Flow Event Handling

The system SHALL support resuming flows with different event types (response answered, timeout, cancellation).

#### Scenario: Resume with event data

- **WHEN** `ResumeFlow(flowId, event="request.answered", data={...})` is called
- **THEN** the flow receives the event data in its cursor's `lastEvent` field
