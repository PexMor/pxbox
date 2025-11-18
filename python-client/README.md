# PxBox Python Test Client

Python client demonstrating REST API, WebSocket, and resumable flow usage.

## Quick Start

**Prerequisites:**
- Backend running via `docker-compose up -d` (port 8082)
- Frontend running via `yarn dev` (port 5173)

**Run the end-to-end example:**

```bash
# Install dependencies
pip install -r requirements.txt

# Run the complete example (creates request, waits for web UI response)
python example.py
```

This will create a request and wait for you to answer it in the web UI at `http://localhost:5173`.

## Setup

```bash
# Install dependencies
pip install -r requirements.txt

# Set environment variables (optional)
export PXBOX_URL=http://localhost:8082
export PXBOX_WS_URL=ws://localhost:8082/v1/ws
export PXBOX_TOKEN=your-jwt-token
```

## Usage

```bash
# Run complete end-to-end example (recommended)
python example.py

# Run all demos
python pxbox_client.py

# Run specific demo
python pxbox_client.py rest   # REST API demo
python pxbox_client.py flow   # Resumable flow demo
python pxbox_client.py ws     # WebSocket demo
```

## Examples

### REST API

```python
from pxbox_client import PxBoxClient

client = PxBoxClient()

# Create request
req = client.create_request(
    entity_id="550e8400-e29b-41d4-a716-446655440000",
    schema={"type": "object", "properties": {"name": {"type": "string"}}}
)

# Poll for status
status = client.poll_request_status(req["requestId"])
```

### Resumable Flow

```python
# Create flow
flow = client.create_flow("my-flow", entity_id, cursor={"step": "wait"})

# Save state
client.save_flow_state(flow["id"])

# Later, after restart:
state = client.load_flow_state()
client.resume_flow(state["flowId"], "request.answered", {"data": {}})
```

### WebSocket

```python
import asyncio
from pxbox_client import PxBoxClient

async def main():
    client = PxBoxClient()
    ws = await client.ws_connect()
    await client.ws_subscribe(ws, "entity:550e8400-e29b-41d4-a716-446655440000")
    
    # Listen for events
    while True:
        msg = await client.ws_listen(ws)
        print(msg)

asyncio.run(main())
```

