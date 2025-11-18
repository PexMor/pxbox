# PxBox REST API Documentation

Base URL: `http://localhost:8080/v1`

## Authentication

Authentication is done via JWT tokens in the `Authorization` header:

```
Authorization: Bearer <token>
```

For development, you can also use the `X-Entity-ID` header:

```
X-Entity-ID: <entity-id>
```

## Endpoints

### Requests

#### Create Request

`POST /requests`

Create a new data-entry request.

**Request Body:**

```json
{
  "entity": {
    "id": "entity-id",
    "handle": "entity-handle"
  },
  "schema": {
    "type": "object",
    "properties": {
      "name": { "type": "string" },
      "email": { "type": "string", "format": "email" }
    },
    "required": ["name", "email"]
  },
  "uiHints": {
    "name": { "title": "Full Name", "description": "Enter your full name" }
  },
  "prefill": {
    "name": "John Doe"
  },
  "deadlineAt": "2024-12-31T23:59:59Z",
  "attentionAt": "2024-12-30T00:00:00Z",
  "callbackUrl": "https://example.com/webhook",
  "filesPolicy": {
    "maxFileMB": 10,
    "maxTotalMB": 50,
    "mime": ["image/*", "application/pdf"],
    "extensions": ["jpg", "png", "pdf"]
  }
}
```

**Response:** `201 Created`

```json
{
  "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
  "status": "PENDING",
  "entityId": "entity-id",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

#### Get Request

`GET /requests/{id}`

Get request details.

**Response:** `200 OK`

```json
{
  "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
  "status": "PENDING",
  "entityId": "entity-id",
  "schema": {...},
  "createdAt": "2024-01-01T00:00:00Z"
}
```

#### Claim Request

`POST /requests/{id}/claim`

Claim a pending request.

**Response:** `200 OK`

```json
{
  "status": "CLAIMED"
}
```

#### Post Response

`POST /requests/{id}/response`

Submit a response to a request.

**Request Body:**

```json
{
  "payload": {
    "name": "John Doe",
    "email": "john@example.com"
  },
  "files": [
    {
      "name": "photo.jpg",
      "url": "https://storage.example.com/files/photo.jpg",
      "size": 1024000,
      "mime": "image/jpeg",
      "sha256": "abc123..."
    }
  ]
}
```

**Response:** `200 OK`

```json
{
  "responseId": "01ARZ3NDEKTSV4RRFFQ69G5FAW",
  "status": "ANSWERED"
}
```

#### Cancel Request

`POST /requests/{id}/cancel`

Cancel a request.

**Response:** `200 OK`

```json
{
  "status": "CANCELLED"
}
```

### Entities

#### Create Entity

`POST /entities`

Create a new entity (user, group, role, or bot).

**Request Body:**

```json
{
  "kind": "user",
  "handle": "test-user",
  "meta": {
    "name": "Test User",
    "email": "user@example.com"
  }
}
```

**Response:** `201 Created`

```json
{
  "id": "ebc9c667-69c7-4a00-b002-411f6cbfc456",
  "kind": "user",
  "handle": "test-user",
  "meta": {
    "name": "Test User",
    "email": "user@example.com"
  },
  "createdAt": "2024-01-01T00:00:00Z"
}
```

#### Get Entity

`GET /entities/{id}`

Get entity details by ID.

**Response:** `200 OK`

```json
{
  "id": "ebc9c667-69c7-4a00-b002-411f6cbfc456",
  "kind": "user",
  "handle": "test-user",
  "meta": {...},
  "createdAt": "2024-01-01T00:00:00Z"
}
```

#### Get Entity Queue

`GET /entities/{id}/queue?status=PENDING&limit=20&offset=0`

Get pending inquiries for an entity.

**Query Parameters:**

- `status` (optional): Filter by status (PENDING, CLAIMED, ANSWERED, etc.)
- `limit` (optional, default: 20): Maximum number of results
- `offset` (optional, default: 0): Pagination offset

**Response:** `200 OK`

```json
{
  "requests": [
    {
      "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
      "status": "PENDING",
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

### Inquiries

#### List Inquiries

`GET /inquiries?entityId=<id>&status=PENDING&sortBy=deadline`

List inquiries with filters.

**Query Parameters:**

- `entityId` (optional): Filter by entity ID
- `status` (optional): Filter by status
- `sortBy` (optional): Sort by `deadline` or `created`
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response:** `200 OK`

```json
{
  "inquiries": [...],
  "total": 10
}
```

#### Mark Inquiry as Read

`POST /inquiries/{id}/markRead`

Mark an inquiry as read.

**Response:** `200 OK`

```json
{
  "status": "success"
}
```

#### Snooze Inquiry

`POST /inquiries/{id}/snooze`

Snooze an inquiry until a specific time.

**Request Body:**

```json
{
  "remindAt": "2024-01-02T00:00:00Z"
}
```

**Response:** `200 OK`

```json
{
  "status": "success"
}
```

#### Cancel Inquiry

`POST /inquiries/{id}/cancel`

Cancel an inquiry.

**Response:** `200 OK`

```json
{
  "status": "success"
}
```

#### Delete Inquiry

`DELETE /inquiries/{id}`

Soft delete an inquiry.

**Response:** `200 OK`

```json
{
  "status": "success"
}
```

### Flows

#### Create Flow

`POST /flows`

Create a new flow.

**Request Body:**

```json
{
  "kind": "user-onboarding",
  "ownerEntity": "entity-id",
  "cursor": {
    "step": "collect-email",
    "data": {}
  }
}
```

**Response:** `201 Created`

```json
{
  "id": "01ARZ3NDEKTSV4RRFFQ69G5FAX",
  "status": "RUNNING",
  "cursor": {...}
}
```

#### Get Flow

`GET /flows/{id}`

Get flow details.

**Response:** `200 OK`

```json
{
  "id": "01ARZ3NDEKTSV4RRFFQ69G5FAX",
  "status": "SUSPENDED",
  "cursor": {...}
}
```

#### Resume Flow

`POST /flows/{id}/resume`

Resume a suspended flow.

**Request Body:**

```json
{
  "event": "email-collected",
  "data": {
    "email": "user@example.com"
  }
}
```

**Response:** `200 OK`

```json
{
  "status": "RUNNING"
}
```

#### Cancel Flow

`POST /flows/{id}/cancel`

Cancel a flow.

**Response:** `200 OK`

```json
{
  "status": "CANCELLED"
}
```

### Files

#### Sign File Upload

`POST /files/sign?name=photo.jpg&contentType=image/jpeg&requestId=<id>&size=1024000`

Get presigned URLs for file upload.

**Query Parameters:**

- `name` (required): File name
- `contentType` (required): MIME type
- `requestId` (optional): Request ID for policy validation
- `size` (optional): File size in bytes for validation

**Response:** `200 OK`

```json
{
  "putUrl": "https://storage.example.com/files/photo.jpg?signature=...",
  "getUrl": "https://storage.example.com/files/photo.jpg?signature=..."
}
```

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "error_code",
    "message": "Human-readable error message"
  }
}
```

**Common Error Codes:**

- `invalid_request`: Invalid request parameters
- `not_found`: Resource not found
- `validation_failed`: Schema validation failed
- `policy_violation`: File policy violation
- `unauthorized`: Authentication required

## Status Codes

- `200 OK`: Success
- `201 Created`: Resource created
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Authentication required
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error
