## ADDED Requirements

### Requirement: Deadline Notifications

The system SHALL notify users when inquiries approach or reach their deadlines.

#### Scenario: Notify before deadline

- **WHEN** an inquiry's `deadline_at` approaches (e.g., 1 hour before)
- **THEN** the system emits `request.deadline_near` event on the entity channel

#### Scenario: Notify on deadline expiry

- **WHEN** an inquiry's `deadline_at` is reached
- **THEN** the system marks the inquiry as `EXPIRED` and emits `request.expired` event

### Requirement: Attention Notifications

The system SHALL notify users when inquiries require attention based on `attention_at` timestamp.

#### Scenario: Notify at attention time

- **WHEN** an inquiry's `attention_at` timestamp is reached
- **THEN** the system emits `request.attention` event on the entity channel

### Requirement: Auto-Cancel with Grace Period

The system SHALL support auto-cancelling inquiries after deadline plus optional grace period.

#### Scenario: Auto-cancel after grace period

- **WHEN** an inquiry's deadline plus `autocancel_grace` period expires
- **THEN** the inquiry status changes to `CANCELLED` and linked flow resumes on timeout branch

#### Scenario: Cancel without grace period

- **WHEN** an inquiry has no `autocancel_grace` set
- **THEN** the inquiry expires at deadline but does not auto-cancel

### Requirement: Notification Delivery

The system SHALL deliver notifications via WebSocket events (primary) and optionally via email/webhook.

#### Scenario: WebSocket notification

- **WHEN** a notification event occurs
- **THEN** the event is broadcast to all subscribed clients on the relevant channel

#### Scenario: Email notification (optional)

- **WHEN** an entity has email configured
- **AND** a high-priority notification occurs
- **THEN** an email notification is sent (if email provider configured)

### Requirement: Reminder Scheduling

The system SHALL schedule reminders for snoozed inquiries using background jobs.

#### Scenario: Schedule reminder

- **WHEN** a user snoozes an inquiry with `remindAt` timestamp
- **THEN** a background job is scheduled to emit notification at that time

#### Scenario: Cancel reminder

- **WHEN** an inquiry is answered or cancelled before reminder time
- **THEN** the scheduled reminder job is cancelled

### Requirement: Notification Events

The system SHALL emit the following notification events:

- `request.created` - New inquiry created
- `request.attention` - Inquiry needs attention
- `request.deadline_near` - Deadline approaching
- `request.expired` - Deadline reached
- `request.cancelled` - Inquiry cancelled
- `request.answered` - Response submitted
- `flow.updated` - Flow status changed

#### Scenario: Emit request created event

- **WHEN** a new inquiry is created
- **THEN** `request.created` event is emitted to `entity:<entityId>` channel

#### Scenario: Emit flow updated event

- **WHEN** a flow status changes
- **THEN** `flow.updated` event is emitted to `entity:<ownerEntity>` channel
