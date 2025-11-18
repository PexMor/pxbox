## Why

Build a universal data-entry broker (PxBox) that enables any application, CLI, or server to request structured data input from users or automated agents. The system decouples data collection from user interaction, supporting asynchronous, durable workflows where human input can be requested, paused, and resumed seamlessly—even across application restarts.

## What Changes

- **Core broker system**: Request/response lifecycle with JSON Schema validation
- **Dual transport**: WebSocket (primary, bidirectional) and REST (polling fallback) APIs
- **Durable flows**: Suspend/resume workflows that survive application restarts
- **User inbox**: CRUD operations for managing pending inquiries with deadlines and notifications
- **File storage**: Support for file uploads (local filesystem initially, S3/GCS extensible)
- **Frontend**: Vite + Preact web application for responders
- **Test client**: Python demonstration app showing REST, WebSocket, and resumable flows

## Impact

- **Affected specs**: All new capabilities (request-management, websocket-transport, rest-transport, file-storage, flow-management, user-inbox, notifications, frontend-form-rendering, python-test-client)
- **Affected code**: New codebase (Go backend, Preact frontend, Python test client)
- **Database**: PostgreSQL schema for entities, requests, responses, flows, reminders
- **Infrastructure**: Redis for pub/sub and job queue, optional object storage
