#!/usr/bin/env python3
"""
PxBox Python Test Client

Demonstrates REST API, WebSocket, and resumable flow usage.
"""

import json
import os
import time
import uuid
from typing import Dict, Optional, Any
from pathlib import Path

import requests

# Optional dotenv support
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass

# Optional WebSocket support
try:
    import websockets
    import asyncio
    HAS_WEBSOCKETS = True
except ImportError:
    HAS_WEBSOCKETS = False


class PxBoxClient:
    """Client for interacting with PxBox broker"""

    def __init__(self, base_url: str = None, ws_url: str = None, token: str = None):
        self.base_url = base_url or os.getenv("PXBOX_URL", "http://localhost:8082")
        self.ws_url = ws_url or os.getenv("PXBOX_WS_URL", "ws://localhost:8082/v1/ws")
        self.token = token or os.getenv("PXBOX_TOKEN", "")
        self.session = requests.Session()
        if self.token:
            self.session.headers.update({"Authorization": f"Bearer {self.token}"})

    # REST API methods

    def create_request(
        self,
        entity_id: str,
        schema: Dict[str, Any],
        ui_hints: Optional[Dict[str, Any]] = None,
        prefill: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Create a data-entry request"""
        payload = {
            "entity": {"id": entity_id},
            "schema": schema,
        }
        if ui_hints:
            payload["uiHints"] = ui_hints
        if prefill:
            payload["prefill"] = prefill

        resp = self.session.post(
            f"{self.base_url}/v1/requests",
            json=payload,
            headers={"X-Client-ID": "python-client"},
        )
        resp.raise_for_status()
        return resp.json()

    def get_request(self, request_id: str) -> Dict[str, Any]:
        """Get request details"""
        resp = self.session.get(f"{self.base_url}/v1/requests/{request_id}")
        resp.raise_for_status()
        return resp.json()

    def claim_request(self, request_id: str, entity_id: str) -> Dict[str, Any]:
        """Claim a pending request"""
        resp = self.session.post(
            f"{self.base_url}/v1/requests/{request_id}/claim",
            headers={"X-Entity-ID": entity_id},
        )
        resp.raise_for_status()
        return resp.json()

    def post_response(
        self, request_id: str, entity_id: str, payload: Dict[str, Any], files: Optional[list] = None
    ) -> Dict[str, Any]:
        """Submit a response to a request"""
        data = {"payload": payload}
        if files:
            data["files"] = files

        resp = self.session.post(
            f"{self.base_url}/v1/requests/{request_id}/response",
            json=data,
            headers={"X-Entity-ID": entity_id},
        )
        resp.raise_for_status()
        return resp.json()

    def poll_request_status(self, request_id: str, timeout: int = 60) -> Dict[str, Any]:
        """Poll request status until it changes from PENDING"""
        start_time = time.time()
        while time.time() - start_time < timeout:
            req = self.get_request(request_id)
            if req.get("status") != "PENDING":
                return req
            time.sleep(1)
        raise TimeoutError(f"Request {request_id} still PENDING after {timeout}s")

    def cancel_request(self, request_id: str) -> Dict[str, Any]:
        """Cancel a request"""
        resp = self.session.post(f"{self.base_url}/v1/requests/{request_id}/cancel")
        resp.raise_for_status()
        return resp.json()

    def delete_inquiry(self, inquiry_id: str) -> Dict[str, Any]:
        """Soft delete an inquiry"""
        resp = self.session.delete(f"{self.base_url}/v1/inquiries/{inquiry_id}")
        resp.raise_for_status()
        return resp.json()

    def get_response_by_request(self, request_id: str) -> Dict[str, Any]:
        """Get response data for an answered request"""
        resp = self.session.get(f"{self.base_url}/v1/requests/{request_id}/response")
        resp.raise_for_status()
        return resp.json()

    def list_inquiries(self, entity_id: str = None, status: str = None) -> Dict[str, Any]:
        """List inquiries with optional filters
        
        Args:
            entity_id: Filter by entity ID
            status: Filter by status (PENDING, ANSWERED, etc.)
        """
        params = {}
        if entity_id:
            params["entityId"] = entity_id
        if status:
            params["status"] = status
        
        resp = self.session.get(f"{self.base_url}/v1/inquiries", params=params)
        resp.raise_for_status()
        return resp.json()

    # Entity methods

    def create_entity(
        self, kind: str, handle: str, meta: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Create a new entity
        
        Raises:
            requests.HTTPError: If entity creation fails
            ValueError: If duplicate handle error is detected (check error message)
        """
        payload = {"kind": kind, "handle": handle}
        if meta:
            payload["meta"] = meta

        resp = self.session.post(f"{self.base_url}/v1/entities", json=payload)
        if resp.status_code == 500:
            # Check if it's a duplicate handle error
            error_text = resp.text.lower()
            duplicate_detected = False
            try:
                error_data = resp.json()
                error_msg = error_data.get("message", "").lower()
                if "duplicate" in error_msg or "23505" in error_msg or "unique constraint" in error_msg:
                    duplicate_detected = True
            except (KeyError, json.JSONDecodeError):
                # Check error text even if not JSON
                if "duplicate" in error_text or "23505" in error_text or "unique constraint" in error_text:
                    duplicate_detected = True
            
            if duplicate_detected:
                raise ValueError(f"Entity with handle '{handle}' already exists")
        
        resp.raise_for_status()
        return resp.json()

    def get_entity(self, entity_id: str) -> Dict[str, Any]:
        """Get entity details by ID"""
        resp = self.session.get(f"{self.base_url}/v1/entities/{entity_id}")
        resp.raise_for_status()
        return resp.json()

    # Flow methods

    def create_flow(
        self, kind: str, owner_entity: str, cursor: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Create a flow"""
        payload = {"kind": kind, "ownerEntity": owner_entity}
        if cursor:
            payload["cursor"] = cursor

        resp = self.session.post(f"{self.base_url}/v1/flows", json=payload)
        resp.raise_for_status()
        return resp.json()

    def get_flow(self, flow_id: str) -> Dict[str, Any]:
        """Get flow details"""
        resp = self.session.get(f"{self.base_url}/v1/flows/{flow_id}")
        resp.raise_for_status()
        return resp.json()

    def resume_flow(
        self, flow_id: str, event: str, data: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Resume a suspended flow"""
        payload = {"event": event}
        if data:
            payload["data"] = data

        resp = self.session.post(f"{self.base_url}/v1/flows/{flow_id}/resume", json=payload)
        resp.raise_for_status()
        return resp.json()

    def save_flow_state(self, flow_id: str, state_file: str = "flow_state.json", entity_id: str = None, cursor: Dict[str, Any] = None):
        """Save flow state to file for resumption
        
        Args:
            flow_id: Flow ID to save
            state_file: Path to state file (default: flow_state.json)
            entity_id: Entity ID to save in state
            cursor: Optional cursor to save (if None, fetches current cursor from flow)
        """
        if cursor is None:
            flow = self.get_flow(flow_id)
            cursor = flow.get("cursor", {})
            status = flow.get("status")
        else:
            # If cursor is provided, still get status from flow
            flow = self.get_flow(flow_id)
            status = flow.get("status")
        
        state = {
            "flowId": flow_id,
            "cursor": cursor,
            "status": status,
            "savedAt": time.time(),
        }
        if entity_id:
            state["entityId"] = entity_id
        with open(state_file, "w") as f:
            json.dump(state, f, indent=2)
        print(f"Saved flow state to {state_file}")

    def load_flow_state(self, state_file: str = "flow_state.json") -> Dict[str, Any]:
        """Load flow state from file"""
        with open(state_file, "r") as f:
            return json.load(f)

    # WebSocket methods

    async def ws_connect(self):
        """Connect to WebSocket"""
        if not HAS_WEBSOCKETS:
            raise ImportError("websockets package is required for WebSocket support. Install with: pip install websockets")
        additional_headers = None
        if self.token:
            additional_headers = {"Authorization": f"Bearer {self.token}"}
        # websockets 11.0+ uses additional_headers parameter
        return await websockets.connect(self.ws_url, additional_headers=additional_headers)

    async def ws_subscribe(self, ws, channel: str):
        """Subscribe to a WebSocket channel"""
        await ws.send(json.dumps({"type": "subscribe", "channel": channel}))

    async def ws_send_command(self, ws, op: str, data: Dict[str, Any], msg_id: Optional[str] = None):
        """Send a WebSocket command"""
        if msg_id is None:
            msg_id = str(uuid.uuid4())
        message = {"type": "cmd", "op": op, "data": data, "id": msg_id}
        await ws.send(json.dumps(message))
        return msg_id

    async def ws_listen(self, ws, timeout: int = 10):
        """Listen for WebSocket messages"""
        try:
            message = await asyncio.wait_for(ws.recv(), timeout=timeout)
            return json.loads(message)
        except asyncio.TimeoutError:
            return None


def demo_rest_api():
    """Demonstrate REST API usage"""
    print("=== REST API Demo ===")
    client = PxBoxClient()

    # Try to reuse existing demo entity or create new one
    default_entity_id = "550e8400-e29b-41d4-a716-446655440000"
    entity_id = None
    
    # First try the default entity
    try:
        entity = client.get_entity(default_entity_id)
        entity_id = default_entity_id
        print(f"Using existing entity: {entity_id}")
    except Exception:
        # Try to find an existing demo entity
        try:
            resp = requests.get(f"{client.base_url}/v1/inquiries", params={"status": "PENDING"})
            if resp.status_code == 200:
                inquiries = resp.json().get("items", [])
                if inquiries:
                    # Use entity from first pending inquiry
                    entity_id = inquiries[0].get("entityId")
                    try:
                        client.get_entity(entity_id)
                        print(f"Reusing entity from existing inquiry: {entity_id}")
                    except Exception:
                        entity_id = None
        except Exception:
            pass
        
        # Create new entity if needed
        if not entity_id:
            print("Creating new entity...")
            try:
                entity = client.create_entity(
                    kind="user",
                    handle=f"demo-user-{uuid.uuid4().hex[:8]}",
                    meta={"name": "Demo User"}
                )
                entity_id = entity["id"]
                print(f"Created entity: {entity_id}")
            except ValueError as e:
                # Handle duplicate handle
                if "already exists" in str(e).lower():
                    # Try to create with unique handle
                    entity = client.create_entity(
                        kind="user",
                        handle=f"demo-user-{uuid.uuid4().hex[:8]}",
                        meta={"name": "Demo User"}
                    )
                    entity_id = entity["id"]
                    print(f"Created entity with unique handle: {entity_id}")
                else:
                    raise

    # Create request
    schema = {
        "type": "object",
        "properties": {
            "fullName": {"type": "string", "title": "Full Name"},
            "email": {"type": "string", "format": "email"},
        },
        "required": ["fullName", "email"],
    }

    print("Creating request...")
    req = client.create_request(
        entity_id=entity_id,
        schema=schema,
        ui_hints={"fullName": {"ui:help": "Enter your legal name"}},
        prefill={"email": "user@example.com"},
    )
    request_id = req["requestId"]
    print(f"Created request: {request_id}")

    # Get request details
    print(f"\nGetting request {request_id}...")
    req_details = client.get_request(request_id)
    print(f"Request details:")
    print(f"  ID: {req_details.get('id')}")
    print(f"  Status: {req_details.get('status')}")
    print(f"  Entity ID: {req_details.get('entityId')}")
    print(f"  Created By: {req_details.get('createdBy', 'N/A')}")
    # Schema is stored as schemaPayload in the database
    schema_payload = req_details.get('schemaPayload') or req_details.get('schema', {})
    if schema_payload:
        print(f"  Schema: {json.dumps(schema_payload, indent=2)}")
    ui_hints = req_details.get('uiHints') or req_details.get('UIHints', {})
    if ui_hints:
        print(f"  UI Hints: {json.dumps(ui_hints, indent=2)}")
    prefill = req_details.get('prefill') or req_details.get('Prefill', {})
    if prefill:
        print(f"  Prefill: {json.dumps(prefill, indent=2)}")
    if req_details.get('deadlineAt'):
        print(f"  Deadline: {req_details.get('deadlineAt')}")
    if req_details.get('createdAt'):
        print(f"  Created At: {req_details.get('createdAt')}")

    # Demonstrate waiting for user response (via web UI)
    print("\n=== Waiting for User Response ===")
    print("📋 The request is now PENDING and waiting for a user to answer via web UI.")
    print(f"\n   To answer this request:")
    print(f"   1. Open the web UI: http://localhost:5173?entityId={entity_id}")
    print(f"   2. Find request {request_id} in your inbox")
    print(f"   3. Fill out the form and submit")
    print(f"\n   The application can:")
    print(f"   - Poll for status changes: client.poll_request_status('{request_id}')")
    print(f"   - Suspend and resume using flows (see flow demo)")
    print(f"   - Recover after restart from saved state")
    
    # Show polling example (but don't actually wait in demo)
    print("\n📊 Polling example (short timeout for demo):")
    print("  In a real application, you would:")
    print("  1. Poll: updated = client.poll_request_status(request_id, timeout=300)")
    print("  2. Check status: if updated['status'] == 'ANSWERED':")
    print("  3. Get response data: response = client.get_response_by_request(request_id)")
    print("  4. Continue with: response['payload']")
    try:
        updated = client.poll_request_status(request_id, timeout=2)
        print(f"\n✓ Status changed to: {updated['status']}")
        if updated.get('status') == 'ANSWERED':
            print("  ✓ User has answered the request!")
            print("  Application can now retrieve response data:")
            try:
                response = client.get_response_by_request(request_id)
                print(f"  Response payload: {json.dumps(response.get('payload', {}), indent=2)}")
                print(f"  Answered by: {response.get('answeredBy')}")
                print(f"  Answered at: {response.get('answeredAt')}")
            except Exception as e:
                print(f"  (Response retrieval: {e})")
    except TimeoutError:
        print("  Request still PENDING (expected - user hasn't answered yet)")
        print("\n💡 In a real application:")
        print("   - Poll with longer timeout: client.poll_request_status(request_id, timeout=300)")
        print("   - Or use WebSocket events for real-time updates")
        print("   - Or use flows to suspend and resume after restart")
        print("   - Application waits until user answers via web UI, then retrieves response data")


def demo_flow_resume():
    """Demonstrate resumable flow - create request, suspend, wait for user, resume after restart"""
    print("\n=== Resumable Flow Demo ===")
    print("This demonstrates:")
    print("  1. Create flow and request")
    print("  2. Suspend flow waiting for user response")
    print("  3. Save state and simulate restart")
    print("  4. Recover and continue waiting")
    print("  5. Resume when user answers via web UI")
    print("\n💡 Tip: Use test_flow_recovery.py for automated testing\n")
    
    client = PxBoxClient()
    state_file = "flow_state.json"

    # Try to recover existing state first
    try:
        state = client.load_flow_state(state_file)
        flow_id = state.get("flowId")
        entity_id = state.get("entityId")
        cursor = state.get("cursor", {})
        pending_request_id = cursor.get("pendingRequestId")
        
        print("✓ Found existing flow state - attempting recovery...")
        print(f"  Flow ID: {flow_id}")
        print(f"  Entity ID: {entity_id}")
        print(f"  Pending Request: {pending_request_id}")
        
        # Verify flow exists
        try:
            flow = client.get_flow(flow_id)
            print(f"  Flow status: {flow.get('status')}")
        except Exception as e:
            print(f"  ⚠ Flow not found: {e}")
            print("  Creating new flow...")
            flow_id = None
            entity_id = None
            pending_request_id = None
        
        # Check if pending request was answered
        if pending_request_id:
            try:
                req = client.get_request(pending_request_id)
                request_status = req.get("status")
                print(f"  Request {pending_request_id} status: {request_status}")
                
                if request_status == "ANSWERED":
                    print("  ✓ Request was answered!")
                    
                    # Get response data
                    try:
                        response = client.get_response_by_request(pending_request_id)
                        response_payload = response.get("payload", {})
                        print(f"  Response: {json.dumps(response_payload, indent=2)}")
                        
                        # Resume flow
                        print("\n  Resuming flow...")
                        resume_result = client.resume_flow(
                            flow_id=flow_id,
                            event="request.answered",
                            data={"requestId": pending_request_id, "payload": response_payload}
                        )
                        print(f"  ✓ Flow resumed: {resume_result}")
                        
                        # Update and save completed state
                        flow = client.get_flow(flow_id)
                        cursor = flow.get("cursor", {})
                        cursor["lastRequestId"] = pending_request_id
                        cursor["lastResponse"] = response_payload
                        cursor["step"] = "completed"
                        client.save_flow_state(flow_id, entity_id=entity_id, cursor=cursor)
                        
                        # Clean up for next run
                        import os
                        if os.path.exists(state_file):
                            os.remove(state_file)
                        print("  ✓ State cleaned up - ready for next test")
                        return
                    except Exception as e:
                        print(f"  ⚠ Failed to retrieve response: {e}")
                else:
                    print(f"  Request still {request_status} - waiting for answer")
                    print(f"\n  To answer: http://localhost:5173?entityId={entity_id}")
                    print(f"  Request ID: {pending_request_id}")
                    return
            except Exception as e:
                print(f"  ⚠ Request check failed: {e}")
                print("  Creating new flow...")
                flow_id = None
                entity_id = None
    except FileNotFoundError:
        print("No saved state found - creating new flow...")
        flow_id = None
        entity_id = None
        pending_request_id = None
    except Exception as e:
        print(f"Error loading state: {e}")
        flow_id = None
        entity_id = None
        pending_request_id = None

    # Create new flow if needed
    if not flow_id:
        print("\n=== Creating New Flow ===")
        
        # Get or create entity
        if not entity_id:
            # Try to find existing entity from inquiries
            try:
                inquiries = client.list_inquiries(status="PENDING")
                items = inquiries.get("items", [])
                if items:
                    entity_id = items[0].get("entityId")
                    print(f"Using existing entity from inquiries: {entity_id}")
            except Exception:
                pass
            
            if not entity_id:
                try:
                    entity = client.create_entity(
                        kind="user",
                        handle=f"demo-flow-user-{uuid.uuid4().hex[:8]}",
                        meta={"name": "Demo Flow User"}
                    )
                    entity_id = entity["id"]
                    print(f"Created entity: {entity_id}")
                except ValueError:
                    entity = client.create_entity(
                        kind="user",
                        handle=f"demo-flow-user-{uuid.uuid4().hex[:8]}",
                        meta={"name": "Demo Flow User"}
                    )
                    entity_id = entity["id"]
                    print(f"Created entity with unique handle: {entity_id}")

        # Create flow
        print("\nStep 1: Creating flow...")
        flow = client.create_flow(
            kind="demo-flow",
            owner_entity=entity_id,
            cursor={"step": "requesting-input", "data": {}}
        )
        flow_id = flow["id"]
        print(f"✓ Created flow: {flow_id}")
        
        # Create request
        print("\nStep 2: Creating request...")
        schema = {
            "type": "object",
            "properties": {
                "answer": {"type": "string", "title": "Your Answer"}
            },
            "required": ["answer"]
        }
        request = client.create_request(
            entity_id=entity_id,
            schema=schema,
            ui_hints={"answer": {"ui:help": "Enter your response"}}
        )
        request_id = request["requestId"]
        print(f"✓ Created request: {request_id}")
        
        # Update cursor and save state
        cursor = {"pendingRequestId": request_id, "step": "waiting-for-user-response"}
        print("\nStep 3: Saving flow state...")
        client.save_flow_state(flow_id, entity_id=entity_id, cursor=cursor)
        print(f"✓ State saved to {state_file}")
        print(f"\n  To answer: http://localhost:5173?entityId={entity_id}")
        print(f"  Request ID: {request_id}")
        print(f"\n  Run this demo again after answering to verify recovery")


async def demo_websocket():
    """Demonstrate WebSocket usage"""
    print("\n=== WebSocket Demo ===")
    client = PxBoxClient()

    # Try to reuse existing entity
    entity_id = "550e8400-e29b-41d4-a716-446655440000"
    try:
        client.get_entity(entity_id)
        print(f"Using existing entity: {entity_id}")
    except Exception:
        # Try to reuse entity from flow state
        try:
            state = client.load_flow_state()
            entity_id = state.get("entityId")
            if entity_id:
                try:
                    client.get_entity(entity_id)
                    print(f"Reusing entity from flow state: {entity_id}")
                except Exception:
                    entity_id = None
        except FileNotFoundError:
            pass
        
        # Create new entity if needed
        if not entity_id:
            try:
                entity = client.create_entity(
                    kind="user",
                    handle=f"demo-ws-user-{uuid.uuid4().hex[:8]}",
                    meta={"name": "Demo WebSocket User"}
                )
                entity_id = entity["id"]
                print(f"Created entity: {entity_id}")
            except ValueError:
                entity = client.create_entity(
                    kind="user",
                    handle=f"demo-ws-user-{uuid.uuid4().hex[:8]}",
                    meta={"name": "Demo WebSocket User"}
                )
                entity_id = entity["id"]
                print(f"Created entity: {entity_id}")

    try:
        ws = await client.ws_connect()
        print("WebSocket connected")

        # Subscribe to entity inbox
        await client.ws_subscribe(ws, f"entity:{entity_id}")
        print(f"Subscribed to entity:{entity_id}")

        # Send a command
        msg_id = await client.ws_send_command(
            ws, "getRequest", {"requestId": "test123"}
        )
        print(f"Sent command: {msg_id}")

        # Listen for response
        response = await client.ws_listen(ws, timeout=5)
        if response:
            print(f"Received: {json.dumps(response, indent=2)}")

        await ws.close()
    except Exception as e:
        print(f"WebSocket error: {e}")


def cleanup_test_data(include_answered: bool = False):
    """Clean up test inquiries created by demos
    
    Args:
        include_answered: If True, also delete ANSWERED inquiries (default: False)
                          Note: ANSWERED inquiries are auto-answered by demos, so they
                          appear in the entity's inbox even though the user didn't answer them.
    """
    print("=== Cleaning Up Test Data ===")
    if include_answered:
        print("⚠️  Note: ANSWERED inquiries are from demo auto-responses.")
        print("   They appear in inboxes because demos answer as the same entity.\n")
    
    client = PxBoxClient()
    
    try:
        # Get all inquiries
        resp = requests.get(f"{client.base_url}/v1/inquiries")
        resp.raise_for_status()
        inquiries_data = resp.json()
        items = inquiries_data.get("items", [])
        
        # Filter test inquiries (PENDING and optionally ANSWERED)
        test_inquiries = []
        pending_count = 0
        answered_count = 0
        
        for inquiry in items:
            status = inquiry.get("status")
            # Check if it's a demo inquiry (created by python-client)
            created_by = inquiry.get("createdBy", "")
            is_demo = created_by == "python-client" or created_by == ""
            
            if status == "PENDING" and is_demo:
                test_inquiries.append(inquiry["id"])
                pending_count += 1
            elif include_answered and status == "ANSWERED" and is_demo:
                # Only delete ANSWERED if explicitly requested and from demo
                test_inquiries.append(inquiry["id"])
                answered_count += 1
        
        if not test_inquiries:
            print("No test inquiries to clean up")
            return
        
        status_msg = []
        if pending_count > 0:
            status_msg.append(f"{pending_count} PENDING")
        if answered_count > 0:
            status_msg.append(f"{answered_count} ANSWERED")
        
        print(f"Found {len(test_inquiries)} test inquiries to clean up ({', '.join(status_msg)})")
        
        # Delete them
        deleted = 0
        failed = 0
        for inquiry_id in test_inquiries:
            try:
                client.delete_inquiry(inquiry_id)
                deleted += 1
            except Exception as e:
                failed += 1
                if failed <= 3:  # Only show first few errors
                    print(f"  ⚠ Failed to delete {inquiry_id}: {e}")
        
        print(f"✓ Deleted {deleted} test inquiries")
        if failed > 0:
            print(f"  ⚠ {failed} inquiries could not be deleted")
        
    except Exception as e:
        print(f"Error during cleanup: {e}")


if __name__ == "__main__":
    import sys

    if len(sys.argv) > 1:
        demo = sys.argv[1]
        if demo == "rest":
            demo_rest_api()
        elif demo == "flow":
            demo_flow_resume()
        elif demo == "ws":
            asyncio.run(demo_websocket())
        elif demo == "cleanup":
            cleanup_test_data(include_answered=False)
        elif demo == "cleanup-all":
            cleanup_test_data(include_answered=True)
        else:
            print("Usage: python pxbox_client.py [rest|flow|ws|cleanup|cleanup-all]")
            print("  cleanup      - Delete PENDING test inquiries")
            print("  cleanup-all  - Delete PENDING and ANSWERED test inquiries")
    else:
        demo_rest_api()
        demo_flow_resume()
        asyncio.run(demo_websocket())
        print("\n💡 Tip: Run 'python pxbox_client.py cleanup' to remove test inquiries")

