# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Core broker system with request/response lifecycle
- Dual transport support: REST API and WebSocket (primary)
- Durable flows with suspend/resume capability
- User inbox with CRUD operations for inquiries
- File upload support with presigned URLs (local filesystem)
- JSON Schema validation with `$ref` resolution and allowlist
- Background jobs for deadline notifications, expiry, and reminders
- Frontend web application (Vite + Preact)
- Python test client demonstrating REST, WebSocket, and resumable flows
- Database migrations (custom runner and goose support)
- Comprehensive API documentation (REST, WebSocket, flow checkpoints)
- Integration and E2E test suites

### Changed

- Initial release

### Security

- JWT authentication for REST and WebSocket transports
- File policy validation (size, MIME type, extensions)
- JSON Schema `$ref` URL allowlist to prevent SSRF attacks
