## ADDED Requirements

### Requirement: Inquiry Listing

The system SHALL allow users to list their pending inquiries with filtering, sorting, and pagination.

#### Scenario: List pending inquiries

- **WHEN** a user requests `GET /v1/inquiries?status=PENDING&entityId=...`
- **THEN** the system returns paginated list of pending inquiries

#### Scenario: Filter by status

- **WHEN** a user filters inquiries by status `CLAIMED`
- **THEN** only inquiries with that status are returned

#### Scenario: Sort by deadline

- **WHEN** a user sorts inquiries by `deadline_at`
- **THEN** inquiries are returned ordered by deadline (soonest first)

### Requirement: Inquiry Read Status

The system SHALL track whether inquiries have been read by the user.

#### Scenario: Mark inquiry as read

- **WHEN** a user marks an inquiry as read
- **THEN** the `read_at` timestamp is set

#### Scenario: Filter unread inquiries

- **WHEN** a user requests unread inquiries
- **THEN** only inquiries with `read_at IS NULL` are returned

### Requirement: Inquiry Snooze

The system SHALL allow users to snooze inquiries, scheduling reminders for later.

#### Scenario: Snooze inquiry

- **WHEN** a user snoozes an inquiry with `remindAt` timestamp
- **THEN** a reminder is created and the user is notified at that time

#### Scenario: Prevent snooze past deadline

- **WHEN** a user attempts to snooze past the inquiry's deadline
- **THEN** the system rejects the snooze and suggests nearest valid time

### Requirement: Inquiry Cancellation

The system SHALL allow users to cancel their pending inquiries.

#### Scenario: Cancel inquiry

- **WHEN** a user cancels an inquiry
- **THEN** the inquiry status changes to `CANCELLED` and linked flow resumes if applicable

#### Scenario: Cancel claimed inquiry

- **WHEN** a non-claimer attempts to cancel a `CLAIMED` inquiry
- **THEN** the system prompts for confirmation or requires owner override

### Requirement: Inquiry Soft Delete

The system SHALL support soft-deleting inquiries from the user's inbox without removing them from the system.

#### Scenario: Delete inquiry from inbox

- **WHEN** a user deletes an inquiry
- **THEN** the `deleted_at` timestamp is set and the inquiry is hidden from inbox

#### Scenario: Restore deleted inquiry

- **WHEN** a user requests inquiries with `includeDeleted=true`
- **THEN** deleted inquiries are included in results

### Requirement: Inquiry Grouping

The system SHALL support grouping inquiries by attention status (needs attention, due soon, all pending).

#### Scenario: Group by attention status

- **WHEN** a user views inbox grouped by attention
- **THEN** inquiries are organized into: "Needs attention", "Due soon", "All pending"

### Requirement: Inquiry Details

The system SHALL provide full inquiry details including schema, UI hints, prefill, deadlines, and linked flow.

#### Scenario: Get inquiry details

- **WHEN** a user requests inquiry details
- **THEN** the system returns complete inquiry information including all metadata
