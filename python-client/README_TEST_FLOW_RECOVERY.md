# Flow Recovery Test

Automated test for verifying flow suspend/resume recovery functionality.

## Purpose

This test verifies that:
1. Flow state can be saved and recovered after application restart
2. Pending requests are properly tracked in flow state
3. Answered requests are detected and flow can resume with response data
4. Multiple suspend/resume cycles work correctly

## Usage

### Automated Testing (with auto-answer)

```bash
# Run test with auto-answer (for CI/CD)
uv run python-client/test_flow_recovery.py --auto-answer

# Run multiple cycles to verify suspend/resume
for i in 1 2 3; do
  echo "=== Run $i ==="
  uv run python-client/test_flow_recovery.py --auto-answer
done
```

### Manual Testing (wait for user to answer)

```bash
# First run: Creates flow and request, saves state
uv run python-client/test_flow_recovery.py

# Answer the request via web UI:
# http://localhost:5173?entityId=<entity_id>

# Second run: Recovers state, detects answered request, resumes flow
uv run python-client/test_flow_recovery.py
```

## Test Flow

### First Run (No State)
1. Creates new entity (or reuses from pending inquiries)
2. Creates new flow
3. Creates request (inquiry) for user to answer
4. Saves flow state to `flow_state.json`
5. If `--auto-answer`: Automatically answers the request
6. Exits

### Subsequent Run (With State)
1. Loads saved flow state from `flow_state.json`
2. Verifies flow still exists
3. Checks status of pending request:
   - **If ANSWERED**: Retrieves response, resumes flow, cleans up state
   - **If PENDING**: Shows instructions to answer, exits
4. If flow/request not found: Creates new flow

## State File

The test saves state to `flow_state.json`:

```json
{
  "flowId": "806c94fc-a1ce-445a-bbdd-3eb7a4eac49a",
  "entityId": "fc02d90f-7ad1-40aa-8379-489de0e6c1c5",
  "cursor": {
    "pendingRequestId": "01KACHAT0SDW3CHJJYRFA36V38",
    "step": "waiting-for-user-response"
  },
  "status": "RUNNING",
  "savedAt": 1763504769.598932
}
```

## Expected Behavior

### Successful Recovery Cycle

1. **Run 1**: Creates flow + request → saves state → (auto-answers if `--auto-answer`)
2. **Run 2**: Loads state → finds answered request → retrieves response → resumes flow → cleans up state
3. **Run 3**: No state → creates new flow (cycle repeats)

### Manual Answer Flow

1. **Run 1**: Creates flow + request → saves state → exits
2. **User**: Answers request via web UI
3. **Run 2**: Loads state → finds answered request → retrieves response → resumes flow → cleans up state

## Integration with CI/CD

```bash
#!/bin/bash
# Example CI test script

set -e

echo "Testing flow recovery..."

# Clean up any existing state
rm -f python-client/flow_state.json

# Run test cycle
uv run python-client/test_flow_recovery.py --auto-answer
uv run python-client/test_flow_recovery.py

# Verify state was cleaned up after recovery
if [ -f python-client/flow_state.json ]; then
  echo "ERROR: State file should be cleaned up after recovery"
  exit 1
fi

echo "✓ Flow recovery test passed"
```

## Troubleshooting

### State file not found
- **First run**: This is expected - test will create new flow
- **Subsequent runs**: Check file permissions, ensure test completed successfully

### Request not found
- Request may have been deleted manually
- Test will create new flow automatically

### Flow not found
- Flow may have been deleted manually
- Test will create new flow automatically

### Response endpoint not available
- Ensure backend is running and includes GET `/v1/requests/{id}/response` endpoint
- Check backend logs for errors

## Related Files

- `pxbox_client.py`: Client library with `save_flow_state()` and `load_flow_state()` methods
- `demo_flow_resume()`: Interactive demo (less automated, more verbose)
- `flow_state.json`: Saved state file (created by test, cleaned up after recovery)

