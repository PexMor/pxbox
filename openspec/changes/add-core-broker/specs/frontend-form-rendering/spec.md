## ADDED Requirements

### Requirement: Form Rendering from JSON Schema

The frontend SHALL render HTML forms dynamically from JSON Schema, enabling users to provide structured data input.

#### Scenario: Render form from schema

- **WHEN** a responder opens an inquiry with a JSON Schema
- **THEN** the frontend renders an HTML form with appropriate input fields based on the schema

#### Scenario: Render form from JSON example

- **WHEN** a responder opens an inquiry with a JSON example object
- **THEN** the frontend infers field types and renders a form

### Requirement: Schema-Based Field Types

The frontend SHALL map JSON Schema types to appropriate HTML input types and components.

#### Scenario: Render string field

- **WHEN** a schema defines a field with `type: "string"`
- **THEN** the frontend renders a text input field

#### Scenario: Render number field

- **WHEN** a schema defines a field with `type: "number"`
- **THEN** the frontend renders a number input field

#### Scenario: Render boolean field

- **WHEN** a schema defines a field with `type: "boolean"`
- **THEN** the frontend renders a checkbox or toggle

#### Scenario: Render array field

- **WHEN** a schema defines a field with `type: "array"`
- **THEN** the frontend renders a list input with add/remove controls

#### Scenario: Render object field

- **WHEN** a schema defines a field with `type: "object"`
- **THEN** the frontend renders nested form fields

### Requirement: UI Hints Support

The frontend SHALL render UI hints (help text, placeholders, examples) provided in the request's `uiHints` field.

#### Scenario: Display help text

- **WHEN** a field has `uiHints` with `ui:help` text
- **THEN** the frontend displays the help text near the field

#### Scenario: Display placeholder

- **WHEN** a field has `uiHints` with `ui:placeholder` text
- **THEN** the frontend sets the input placeholder attribute

#### Scenario: Display example

- **WHEN** a field has `uiHints` with `ui:example` value
- **THEN** the frontend displays the example as guidance

### Requirement: Prefill Data

The frontend SHALL populate form fields with prefill data when provided in the request.

#### Scenario: Populate prefill values

- **WHEN** a request includes `prefill` data
- **THEN** the frontend pre-populates matching form fields with those values

#### Scenario: Prefill with partial data

- **WHEN** a request includes `prefill` with only some fields
- **THEN** only those fields are pre-populated, others remain empty

### Requirement: Form Validation

The frontend SHALL validate form input against the JSON Schema before submission.

#### Scenario: Validate required fields

- **WHEN** a schema defines `required` fields
- **AND** a user attempts to submit without filling required fields
- **THEN** the frontend displays validation errors

#### Scenario: Validate field types

- **WHEN** a user enters invalid data (e.g., text in number field)
- **THEN** the frontend displays type validation errors

#### Scenario: Validate patterns

- **WHEN** a schema defines a `pattern` for a string field
- **AND** a user enters text that doesn't match the pattern
- **THEN** the frontend displays pattern validation error

#### Scenario: Live validation

- **WHEN** a user types in a form field
- **THEN** the frontend validates the field in real-time and shows errors immediately

### Requirement: File Upload Integration

The frontend SHALL support file uploads as part of form responses, using presigned URLs.

#### Scenario: Upload file via presigned URL

- **WHEN** a user selects a file to upload
- **THEN** the frontend requests a presigned URL, uploads the file, and includes the file reference in the response

#### Scenario: Validate file against policy

- **WHEN** a request includes `filesPolicy` with size or MIME type restrictions
- **AND** a user selects a file that violates the policy
- **THEN** the frontend rejects the file and displays an error

#### Scenario: Multiple file uploads

- **WHEN** a request allows multiple files
- **THEN** the frontend supports uploading multiple files and includes all references in the response

### Requirement: Response Submission

The frontend SHALL submit form responses via WebSocket (primary) or REST (fallback).

#### Scenario: Submit via WebSocket

- **WHEN** a user submits a completed form
- **AND** WebSocket connection is available
- **THEN** the frontend sends the response via WebSocket `postResponse` command

#### Scenario: Submit via REST fallback

- **WHEN** a user submits a completed form
- **AND** WebSocket connection is unavailable
- **THEN** the frontend falls back to `POST /v1/requests/:id/response` REST endpoint

#### Scenario: Handle submission errors

- **WHEN** a response submission fails (validation error, network error)
- **THEN** the frontend displays an error message and allows the user to retry

### Requirement: Form Library Integration

The frontend SHALL use @rjsf/core (React JSON Schema Form) via preact/compat for form rendering.

#### Scenario: Integrate @rjsf/core

- **WHEN** the frontend renders a form
- **THEN** it uses @rjsf/core components configured for Preact compatibility

#### Scenario: Customize form widgets

- **WHEN** UI hints specify custom widgets
- **THEN** the frontend uses appropriate @rjsf widgets or custom components

### Requirement: Loading and Error States

The frontend SHALL display appropriate loading and error states during form operations.

#### Scenario: Loading state while fetching request

- **WHEN** a user opens an inquiry
- **THEN** the frontend displays a loading indicator while fetching request details

#### Scenario: Error state for invalid request

- **WHEN** a request cannot be loaded (not found, unauthorized)
- **THEN** the frontend displays an error message

#### Scenario: Loading state during submission

- **WHEN** a user submits a form
- **THEN** the frontend displays a loading indicator and disables the submit button
