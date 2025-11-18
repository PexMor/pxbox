# PxBox WebSocket Protocol Documentation

## Connection

Connect to: `ws://localhost:8080/v1/ws`

**Authentication:**

- JWT token via query parameter: `?token=<jwt-token>`
- JWT token via Authorization header: `Authorization: Bearer <token>`
- Development fallback: `?X-Entity-ID=<entity-id>` or `X-Entity-ID` header

## Message Format

All messages are JSON objects:

```json
{
  "type": "cmd|event|ack|subscribe|unsubscribe|resume|ping",
  "id": "message-id",
  "op": "operation-name",
  "channel": "channel-name",
  "seq": 123,
  "since": 100,
  "data": {...}
}
```

## Message Types

### Commands (`type: "cmd"`)

Commands are sent from client to server.

#### Create Request

```json
{
  "type": "cmd",
  "op": "createRequest",
  "id": "cmd-1",
  "data": {
    "entity": {
      "id": "entity-id"
    },
    "schema": {
      "type": "object",
      "properties": {
        "name": { "type": "string" }
      }
    },
    "deadlineAt": "2024-12-31T23:59:59Z"
  }
}
```

**Response:**

```json
{
  "type": "response",
  "id": "cmd-1",
  "data": {
    "requestId": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "status": "PENDING",
    "entityId": "entity-id"
  }
}
```

#### Get Request

```json
{
  "type": "cmd",
  "op": "getRequest",
  "id": "cmd-2",
  "data": {
    "requestId": "01ARZ3NDEKTSV4RRFFQ69G5FAV"
  }
}
```

**Response:**

```json
{
  "type": "response",
  "id": "cmd-2",
  "data": {
    "id": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "status": "PENDING",
    "schema": {...}
  }
}
```

#### Post Response

```json
{
  "type": "cmd",
  "op": "postResponse",
  "id": "cmd-3",
  "data": {
    "requestId": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "payload": {
      "name": "John Doe"
    },
    "files": []
  }
}
```

**Response:**

```json
{
  "type": "response",
  "id": "cmd-3",
  "data": {
    "responseId": "01ARZ3NDEKTSV4RRFFQ69G5FAW",
    "status": "ANSWERED"
  }
}
```

#### Claim Request

```json
{
  "type": "cmd",
  "op": "claimRequest",
  "id": "cmd-4",
  "data": {
    "requestId": "01ARZ3NDEKTSV4RRFFQ69G5FAV"
  }
}
```

#### Cancel Request

```json
{
  "type": "cmd",
  "op": "cancelRequest",
  "id": "cmd-5",
  "data": {
    "requestId": "01ARZ3NDEKTSV4RRFFQ69G5FAV"
  }
}
```

#### Create Flow

```json
{
  "type": "cmd",
  "op": "createFlow",
  "id": "cmd-6",
  "data": {
    "kind": "user-onboarding",
    "ownerEntity": "entity-id",
    "cursor": {}
  }
}
```

#### Resume Flow

```json
{
  "type": "cmd",
  "op": "resumeFlow",
  "id": "cmd-7",
  "data": {
    "flowId": "01ARZ3NDEKTSV4RRFFQ69G5FAX",
    "event": "email-collected",
    "data": {
      "email": "user@example.com"
    }
  }
}
```

#### Cancel Flow

```json
{
  "type": "cmd",
  "op": "cancelFlow",
  "id": "cmd-8",
  "data": {
    "flowId": "01ARZ3NDEKTSV4RRFFQ69G5FAX"
  }
}
```

### Subscriptions (`type: "subscribe"`)

Subscribe to a channel to receive events.

```json
{
  "type": "subscribe",
  "channel": "entity:entity-id"
}
```

**Acknowledgment:**

```json
{
  "type": "ack",
  "ack": "subscribed",
  "channel": "entity:entity-id"
}
```

### Unsubscribe (`type: "unsubscribe"`)

Unsubscribe from a channel.

```json
{
  "type": "unsubscribe",
  "channel": "entity:entity-id"
}
```

### Resume (`type: "resume"`)

Resume receiving events from a specific sequence number.

```json
{
  "type": "resume",
  "channel": "entity:entity-id",
  "since": 100
}
```

### Events (`type: "event"`)

Events are sent from server to client.

```json
{
  "type": "event",
  "channel": "entity:entity-id",
  "seq": 123,
  "data": {
    "type": "request.created",
    "requestId": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
    "entityId": "entity-id"
  }
}
```

**Event Types:**

- `request.created`: New request created
- `request.claimed`: Request claimed
- `request.answered`: Response submitted
- `request.cancelled`: Request cancelled
- `request.expired`: Request expired
- `request.deadline_approaching`: Deadline approaching
- `request.needs_attention`: Request needs attention
- `flow.created`: Flow created
- `flow.suspended`: Flow suspended
- `flow.completed`: Flow completed
- `flow.cancelled`: Flow cancelled

### Acknowledgment (`type: "ack"`)

Acknowledge receipt of an event.

```json
{
  "type": "ack",
  "channel": "entity:entity-id",
  "seq": 123
}
```

### Ping/Pong

**Ping:**

```json
{
  "type": "ping"
}
```

**Pong:**

```json
{
  "type": "ack",
  "ack": "pong"
}
```

### Error Responses

```json
{
  "type": "error",
  "id": "cmd-1",
  "code": "error_code",
  "message": "Error message"
}
```

## Channels

Channels follow the pattern: `<type>:<id>`

- `entity:<entity-id>`: Events for a specific entity
- `request:<request-id>`: Events for a specific request
- `requestor:<client-id>`: Events for a specific requestor

## Sequence Numbers

Each event has a sequence number (`seq`) that increases monotonically per channel. Clients should acknowledge events to enable resume functionality.

## Resume Flow

1. Client connects and subscribes to channel
2. Client receives events with sequence numbers
3. Client acknowledges events: `{type: "ack", channel: "...", seq: N}`
4. If connection is lost, client reconnects
5. Client resumes: `{type: "resume", channel: "...", since: N}`
6. Server replays events from `seq: N+1` onwards

## Example Flow

```javascript
// Connect
const ws = new WebSocket("ws://localhost:8080/v1/ws?token=jwt-token");

// Subscribe
ws.send(
  JSON.stringify({
    type: "subscribe",
    channel: "entity:entity-id",
  })
);

// Send command
ws.send(
  JSON.stringify({
    type: "cmd",
    op: "createRequest",
    id: "cmd-1",
    data: {
      entity: { id: "entity-id" },
      schema: { type: "object", properties: { name: { type: "string" } } },
    },
  })
);

// Receive response
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === "response") {
    console.log("Request created:", msg.data.requestId);
  } else if (msg.type === "event") {
    console.log("Event:", msg.data);
    // Acknowledge
    ws.send(
      JSON.stringify({
        type: "ack",
        channel: msg.channel,
        seq: msg.seq,
      })
    );
  }
};
```
