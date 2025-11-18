## ADDED Requirements

### Requirement: File Upload Support

The system SHALL support file uploads as part of data-entry responses, with configurable policies per request.

#### Scenario: Request with file policy

- **WHEN** a requestor creates a request with `filesPolicy: {maxTotalMB: 50, mime: ["image/*"]}`
- **THEN** the policy is stored and enforced when files are uploaded

#### Scenario: Upload file via presigned URL

- **WHEN** a responder requests a presigned upload URL
- **THEN** the system generates a presigned PUT URL with expiration

### Requirement: Storage Abstraction

The system SHALL support multiple storage backends through a common interface: local filesystem (initial) and S3/GCS (extensible).

#### Scenario: Use local filesystem storage

- **WHEN** storage backend is configured as "local"
- **THEN** files are stored in a configured directory on the server filesystem

#### Scenario: Use S3-compatible storage

- **WHEN** storage backend is configured as "s3"
- **THEN** files are stored in the configured S3 bucket using presigned URLs

### Requirement: Presigned URLs

The system SHALL generate presigned URLs for both upload (PUT) and download (GET) operations.

#### Scenario: Generate presigned PUT URL

- **WHEN** a responder requests `POST /v1/files/sign?name=photo.png&contentType=image/png`
- **THEN** the system returns `{putUrl: "...", getUrl: "...", expiresAt: "..."}`

#### Scenario: Presigned URL expiration

- **WHEN** a presigned URL expires
- **THEN** the URL is no longer valid and returns 403 Forbidden

### Requirement: File Metadata

The system SHALL store file metadata including: name, URL, size, MIME type, checksum (SHA256), and storage location.

#### Scenario: Store file reference

- **WHEN** a responder includes file references in a response
- **THEN** the response stores file metadata: `[{name, url, size, mime, sha256}]`

### Requirement: File Validation

The system SHALL validate uploaded files against the request's `filesPolicy` (size limits, MIME types).

#### Scenario: Validate file size

- **WHEN** a file exceeds `maxTotalMB` limit
- **THEN** the upload is rejected with an error

#### Scenario: Validate MIME type

- **WHEN** a file's MIME type is not in the allowed list
- **THEN** the upload is rejected with an error

### Requirement: File Cleanup

The system SHALL support cleanup of orphaned files (files referenced in cancelled/expired requests).

#### Scenario: Cleanup orphaned files

- **WHEN** a request is cancelled or expired
- **THEN** associated files can be marked for cleanup (manual or scheduled)
