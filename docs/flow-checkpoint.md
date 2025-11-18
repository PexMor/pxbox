# Flow Checkpoint Format

## Overview

Flows in PxBox use a checkpoint system to enable suspend/resume functionality. The checkpoint is stored as a JSONB `cursor` field in the `flows` table.

## Cursor Structure

The cursor is a flexible JSON object that can contain any flow-specific state. However, there are some recommended conventions:

### Basic Structure

```json
{
  "step": "current-step-name",
  "data": {
    "collected": {},
    "pending": []
  },
  "metadata": {
    "version": "1.0",
    "lastEvent": "event-name",
    "lastEventAt": "2024-01-01T00:00:00Z"
  }
}
```

### Recommended Fields

- `step`: Current step/state name (string)
- `data`: Flow-specific data (object)
- `metadata`: Checkpoint metadata (object)
  - `version`: Checkpoint format version (string)
  - `lastEvent`: Last processed event type (string)
  - `lastEventAt`: Timestamp of last event (ISO 8601 string)

### Example: User Onboarding Flow

```json
{
  "step": "collect-email",
  "data": {
    "collected": {},
    "pending": [
      {
        "requestId": "01ARZ3NDEKTSV4RRFFQ69G5FAV",
        "type": "email",
        "status": "PENDING"
      }
    ]
  },
  "metadata": {
    "version": "1.0",
    "lastEvent": "flow.started",
    "lastEventAt": "2024-01-01T00:00:00Z"
  }
}
```

After email is collected:

```json
{
  "step": "collect-profile",
  "data": {
    "collected": {
      "email": "user@example.com"
    },
    "pending": [
      {
        "requestId": "01ARZ3NDEKTSV4RRFFQ69G5FAW",
        "type": "profile",
        "status": "PENDING"
      }
    ]
  },
  "metadata": {
    "version": "1.0",
    "lastEvent": "email-collected",
    "lastEventAt": "2024-01-01T01:00:00Z"
  }
}
```

## Resume Flow

When resuming a flow:

1. **Get Flow State:**

   ```http
   GET /flows/{id}
   ```

2. **Check Cursor:**

   - Read the `cursor` field
   - Identify the current `step`
   - Check `pending` requests to see what's waiting

3. **Resume:**

   ```http
   POST /flows/{id}/resume
   {
     "event": "response-received",
     "data": {
       "requestId": "...",
       "response": {...}
     }
   }
   ```

4. **Update Cursor:**
   The system updates the cursor with the new event data:
   ```json
   {
     "step": "next-step",
     "data": {
       "collected": {...},
       "pending": [...]
     },
     "metadata": {
       "lastEvent": "response-received",
       "lastEventAt": "2024-01-01T02:00:00Z"
     }
   }
   ```

## Best Practices

1. **Version Your Checkpoints:**
   Include a version field to handle format migrations:

   ```json
   {
     "metadata": {
       "version": "1.0"
     }
   }
   ```

2. **Store Request IDs:**
   Keep track of pending requests in the cursor:

   ```json
   {
     "data": {
       "pending": [{ "requestId": "...", "type": "...", "status": "PENDING" }]
     }
   }
   ```

3. **Use Step Names:**
   Use descriptive step names for debugging:

   - `collect-email`
   - `verify-email`
   - `collect-profile`
   - `complete`

4. **Handle Errors:**
   Store error state in cursor if needed:

   ```json
   {
     "step": "error",
     "error": {
       "type": "validation-failed",
       "message": "..."
     }
   }
   ```

5. **Keep Cursor Small:**
   Don't store large data in cursor. Store references instead:
   ```json
   {
     "data": {
       "collected": {
         "email": "user@example.com",
         "profileId": "profile-123" // Reference, not full data
       }
     }
   }
   ```

## Flow Status

Flows have the following statuses:

- `RUNNING`: Flow is active and processing
- `SUSPENDED`: Flow is waiting for input
- `COMPLETED`: Flow finished successfully
- `CANCELLED`: Flow was cancelled
- `FAILED`: Flow encountered an error

## Example: Multi-Step Flow

```json
// Step 1: Collect email
{
  "step": "collect-email",
  "data": {
    "pending": [{"requestId": "req-1", "type": "email"}]
  }
}

// Step 2: After email collected, collect profile
{
  "step": "collect-profile",
  "data": {
    "collected": {"email": "user@example.com"},
    "pending": [{"requestId": "req-2", "type": "profile"}]
  }
}

// Step 3: After profile collected, complete
{
  "step": "complete",
  "data": {
    "collected": {
      "email": "user@example.com",
      "profile": {...}
    },
    "pending": []
  }
}
```

## Recovery on Application Restart

When the application restarts:

1. Query all flows with status `RUNNING` or `SUSPENDED`
2. For each flow:
   - Read the cursor
   - Check `pending` requests
   - Verify request statuses
   - Resume flow if needed

Example recovery logic:

```go
flows := getSuspendedFlows()
for _, flow := range flows {
    pending := flow.Cursor["data"].(map[string]interface{})["pending"].([]interface{})
    for _, req := range pending {
        requestID := req.(map[string]interface{})["requestId"].(string)
        reqStatus := getRequestStatus(requestID)
        if reqStatus == "ANSWERED" {
            resumeFlow(flow.ID, "response-received", getResponseData(requestID))
        }
    }
}
```
