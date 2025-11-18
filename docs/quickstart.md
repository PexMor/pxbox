# Quick Start Guide

This guide will help you get PxBox running quickly using Docker Compose, demonstrate flows with Python, and show you how to use the web UI.

## Prerequisites

- Docker and Docker Compose installed
- Python 3.8+ (for Python examples)
- Node.js 18+ and Yarn (for frontend, optional)

## Step 1: Start Backend with Docker Compose

The easiest way to get started is using Docker Compose, which sets up PostgreSQL, Redis, and the broker service automatically.

### Start Services

```bash
# Clone the repository (if you haven't already)
git clone <repository-url>
cd pxbox

# Start all services
docker-compose up -d
```

This starts:

- **PostgreSQL** on port `5432`
- **Redis** on port `6379`
- **pxbox-api** on port `8082` (mapped from container port 8080)

> **Note**: If port 8082 is already in use, you can change it in `docker-compose.yaml` by modifying the port mapping.

The pxbox-api automatically runs migrations on startup, so you don't need to run them manually.

### Verify Services

Check that all services are running:

```bash
docker-compose ps
```

You should see all three services with status "Up". The pxbox-api logs will show migration completion and server startup.

### Test the API

```bash
# Health check
curl http://localhost:8082/healthz

# Should return: {"status":"ok"}
```

> **Note**: The pxbox-api runs on port `8082` externally (mapped from container port 8080). If you need to use port 8080, stop any conflicting services first or modify the port mapping in `docker-compose.yaml`.

## Step 2: Create a Test Entity

Before creating requests, you need an entity (user/group/bot) to receive them. You can create one via the API:

```bash
curl -X POST http://localhost:8082/v1/entities \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "user",
    "handle": "test-user",
    "meta": {"name": "Test User"}
  }'
```

Save the `id` from the response - you'll need it for the examples below.

## Step 3: Python Flow Example

Let's create a complete flow demonstration that shows how to:

1. Create a flow
2. Create a request within the flow
3. Suspend and save state
4. Resume after restart

### Install Python Dependencies

```bash
cd python-client
pip install -r requirements.txt
```

### Complete Flow Example

Create a file `example_flow.py`:

```python
#!/usr/bin/env python3
"""
Complete flow example demonstrating request creation, suspension, and resumption.
"""

import json
import time
from pxbox_client import PxBoxClient

# Initialize client
client = PxBoxClient(base_url="http://localhost:8082")

# Use the entity ID from Step 2, or create a new one
ENTITY_ID = "550e8400-e29b-41d4-a716-446655440000"  # Replace with your entity ID

def step1_create_flow():
    """Step 1: Create a flow"""
    print("=== Step 1: Creating Flow ===")

    flow = client.create_flow(
        kind="shipping-address-flow",
        owner_entity=ENTITY_ID,
        cursor={
            "step": "requesting-address",
            "requestId": None,
        }
    )

    flow_id = flow["id"]
    print(f"✓ Created flow: {flow_id}")
    print(f"  Status: {flow['status']}")
    print(f"  Cursor: {json.dumps(flow.get('cursor', {}), indent=2)}")

    return flow_id

def step2_create_request(flow_id):
    """Step 2: Create a request within the flow"""
    print("\n=== Step 2: Creating Request ===")

    schema = {
        "type": "object",
        "properties": {
            "street": {"type": "string", "title": "Street Address"},
            "city": {"type": "string", "title": "City"},
            "zip": {"type": "string", "title": "ZIP Code"},
            "country": {"type": "string", "title": "Country"}
        },
        "required": ["street", "city", "zip", "country"]
    }

    ui_hints = {
        "street": {
            "title": "Street Address",
            "description": "Enter your street address"
        },
        "country": {
            "enum": ["US", "CA", "UK", "DE", "FR"],
            "enumNames": ["United States", "Canada", "United Kingdom", "Germany", "France"]
        }
    }

    request = client.create_request(
        entity_id=ENTITY_ID,
        schema=schema,
        ui_hints=ui_hints
    )

    request_id = request["requestId"]
    print(f"✓ Created request: {request_id}")
    print(f"  Status: {request['status']}")
    print(f"  Entity: {request['entityId']}")

    # Update flow cursor with request ID
    flow = client.get_flow(flow_id)
    cursor = flow.get("cursor", {})
    cursor["requestId"] = request_id
    cursor["step"] = "waiting-for-response"

    # Resume flow to update cursor (in real app, this would be automatic)
    client.resume_flow(
        flow_id=flow_id,
        event="request.created",
        data={"requestId": request_id}
    )

    return request_id, flow_id

def step3_suspend_and_save(flow_id):
    """Step 3: Suspend flow and save state"""
    print("\n=== Step 3: Suspending Flow and Saving State ===")

    # Get current flow state
    flow = client.get_flow(flow_id)
    print(f"✓ Flow status: {flow['status']}")
    print(f"  Cursor: {json.dumps(flow.get('cursor', {}), indent=2)}")

    # Save state to file
    client.save_flow_state(flow_id, "flow_state.json")
    print("✓ State saved to flow_state.json")
    print("\n💡 At this point, you could:")
    print("   - Stop the application")
    print("   - Restart later")
    print("   - Resume from saved state")

    return flow_id

def step4_resume_flow():
    """Step 4: Resume flow from saved state (simulating restart)"""
    print("\n=== Step 4: Resuming Flow (After Restart) ===")

    # Load saved state
    state = client.load_flow_state("flow_state.json")
    flow_id = state["flowId"]

    print(f"✓ Loaded flow state: {flow_id}")
    print(f"  Previous cursor: {json.dumps(state.get('cursor', {}), indent=2)}")

    # Get current flow status
    flow = client.get_flow(flow_id)
    print(f"  Current status: {flow['status']}")

    # Resume flow with a simulated response event
    print("\n📝 Simulating response submission...")
    resume_result = client.resume_flow(
        flow_id=flow_id,
        event="request.answered",
        data={
            "requestId": state["cursor"].get("requestId"),
            "payload": {
                "street": "123 Main St",
                "city": "San Francisco",
                "zip": "94102",
                "country": "US"
            }
        }
    )

    print(f"✓ Flow resumed")
    print(f"  Result: {json.dumps(resume_result, indent=2)}")

    # Check final flow status
    final_flow = client.get_flow(flow_id)
    print(f"\n✓ Final flow status: {final_flow['status']}")
    print(f"  Final cursor: {json.dumps(final_flow.get('cursor', {}), indent=2)}")

def main():
    """Run complete flow demonstration"""
    print("🚀 PxBox Flow Demonstration\n")
    print("This example shows:")
    print("1. Creating a flow")
    print("2. Creating a request within the flow")
    print("3. Suspending and saving state")
    print("4. Resuming after restart\n")

    try:
        # Step 1: Create flow
        flow_id = step1_create_flow()

        # Step 2: Create request
        request_id, flow_id = step2_create_request(flow_id)

        print(f"\n📋 Request {request_id} is now PENDING")
        print("   You can view and answer it in the web UI (see Step 4)")
        print("   Or wait and continue to see the resume flow...")

        # Step 3: Suspend and save
        step3_suspend_and_save(flow_id)

        # Step 4: Resume (simulating restart)
        input("\n⏸️  Press Enter to simulate restart and resume flow...")
        step4_resume_flow()

        print("\n✅ Flow demonstration complete!")

    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()
```

### Run the Example

```bash
# Make sure backend is running (Step 1)
# Replace ENTITY_ID in the script with your entity ID from Step 2

python example_flow.py
```

## Step 4: Using the Web UI

The web UI provides a user-friendly interface for viewing and responding to requests.

### Start the Frontend

```bash
# From project root
cd frontend

# Create empty yarn.lock to make this a separate project (if needed)
touch yarn.lock

# Install dependencies (first time only)
yarn install

# Start development server
yarn dev
```

> **Note**: If you encounter workspace detection errors, ensure `yarn.lock` exists in the `frontend/` directory. This makes the frontend a separate project from the root.

The frontend will be available at `http://localhost:5173` (Vite default port).

### Using the Web UI

1. **View Inbox**: The inbox shows all pending requests for your entity

   - Filter by "All", "Needs Attention", or "Due Soon"
   - Click on a request to view details

2. **Answer a Request**:

   - Click on a request from the inbox
   - Fill out the form (generated from JSON Schema)
   - Upload files if the request allows them
   - Click "Submit" to send your response

3. **Real-time Updates**:
   - New requests appear automatically via WebSocket
   - Status updates are reflected in real-time

### Example: Answering the Request from Python Flow

1. Run the Python flow example (Step 3) to create a request
2. Open the web UI at `http://localhost:5173`
3. You should see the shipping address request in your inbox
4. Click on it and fill out the form
5. Submit your response
6. The Python flow can then resume with the response data

## Step 5: Complete End-to-End Example

A complete end-to-end example is provided in `python-client/example.py`. This example creates a request and waits for a response via the web UI.

### Running the Example

**Prerequisites:**
1. Backend running via `docker-compose up -d` (port 8082)
2. Frontend running via `yarn dev` (port 5173)
3. An entity created (from Step 2)

**Run the example:**

```bash
# From project root
cd python-client

# Update ENTITY_ID in example.py if needed (or use the default)
python example.py
```

The script will:
1. Create a data-entry request with a simple form (name, email, message)
2. Display the request ID and prompt you to open the web UI
3. Poll for the response status (waits up to 5 minutes)
4. Display the response when submitted

**What to do:**
1. Run the Python script
2. Open `http://localhost:5173` in your browser
3. Find the request in your inbox
4. Fill out and submit the form
5. The script will detect the response and complete

See `python-client/example.py` for the complete source code with detailed comments.

## Troubleshooting

### Backend Not Starting

```bash
# Check logs
docker-compose logs pxbox-api

# Check if ports are in use
docker-compose ps

# Restart services
docker-compose restart
```

### Web UI Not Connecting

- Ensure pxbox-api is running on `http://localhost:8082`
- Check that Vite proxy is configured correctly in `frontend/vite.config.js` (should proxy to `http://localhost:8082`)
- Check browser console for WebSocket connection errors
- Verify WebSocket URL in `frontend/src/hooks/useBrokerWS.jsx` uses port `8082`
- Verify CORS settings if accessing from different origin
- Restart Vite dev server after changing proxy configuration: `yarn dev`

### Yarn Installation Issues

If you get workspace detection errors:

```bash
cd frontend
touch yarn.lock  # Create empty lock file to make it a separate project
yarn install
```

If you get missing Yarn release errors:

```bash
cd frontend
# Remove the yarnPath from .yarnrc.yml or download the release file
# Or switch to node-modules linker (already configured)
yarn install
```

### Python Client Errors

- Verify backend is running: `curl http://localhost:8082/healthz`
- Check entity ID is correct
- Ensure all dependencies are installed: `pip install -r requirements.txt`

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Reset database (⚠️ deletes all data)
docker-compose down -v
docker-compose up -d
```

## Next Steps

- Read the [Architecture Guide](../AGENTS.md) for detailed technical information
- Explore the [REST API Documentation](api.md) for all available endpoints
- Learn about [WebSocket Protocol](websocket.md) for real-time communication
- Understand [Flow Checkpoints](flow-checkpoint.md) for advanced flow management
- Check the [Testing Guide](testing.md) for running tests

## Additional Examples

The `python-client/` directory contains more examples:

- `pxbox_client.py` - Complete client library with all methods
- Run `python pxbox_client.py rest` for REST API demo
- Run `python pxbox_client.py flow` for flow demo
- Run `python pxbox_client.py ws` for WebSocket demo
