#!/usr/bin/env python3
"""
Automated test for flow suspend/resume recovery

This test simulates:
1. First run: Create flow, create request, save state, exit
2. Second run: Load state, check if request was answered, resume if answered
3. Multiple runs: Can be run repeatedly to test recovery

Usage:
    uv run python-client/test_flow_recovery.py [--auto-answer]
    
    --auto-answer: Automatically answer the request (for testing without manual UI interaction)
"""

import sys
import time
import json
import uuid
import argparse
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from pxbox_client import PxBoxClient

BASE_URL = "http://localhost:8082"
STATE_FILE = "flow_state.json"

class Colors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_header(text):
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*70}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{text}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'='*70}{Colors.ENDC}\n")

def print_success(text):
    print(f"{Colors.OKGREEN}✓ {text}{Colors.ENDC}")

def print_info(text):
    print(f"{Colors.OKCYAN}ℹ {text}{Colors.ENDC}")

def print_error(text):
    print(f"{Colors.FAIL}✗ {text}{Colors.ENDC}")

def print_warning(text):
    print(f"{Colors.WARNING}⚠ {text}{Colors.ENDC}")

def test_flow_recovery(auto_answer: bool = False):
    """Test flow suspend/resume recovery"""
    print_header("FLOW RECOVERY TEST")
    
    client = PxBoxClient(base_url=BASE_URL)
    state_file = Path(STATE_FILE)
    
    # Step 1: Try to recover existing state
    if state_file.exists():
        print_info("Found existing flow state - attempting recovery...")
        try:
            state = client.load_flow_state(str(state_file))
            flow_id = state.get("flowId")
            entity_id = state.get("entityId")
            cursor = state.get("cursor", {})
            pending_request_id = cursor.get("pendingRequestId")
            
            print_success(f"Loaded state:")
            print(f"  Flow ID: {flow_id}")
            print(f"  Entity ID: {entity_id}")
            print(f"  Pending Request: {pending_request_id}")
            
            # Verify flow still exists
            try:
                flow = client.get_flow(flow_id)
                print_success(f"Flow exists: {flow.get('status')}")
            except Exception as e:
                print_error(f"Flow not found: {e}")
                print_info("Creating new flow...")
                flow_id = None
                entity_id = None
                pending_request_id = None
            
            # Check if pending request was answered
            if pending_request_id:
                print_info(f"Checking status of pending request: {pending_request_id}")
                try:
                    req = client.get_request(pending_request_id)
                    request_status = req.get("status")
                    print(f"  Request status: {request_status}")
                    
                    if request_status == "ANSWERED":
                        print_success("Request was answered!")
                        
                        # Get response data
                        try:
                            response = client.get_response_by_request(pending_request_id)
                            response_payload = response.get("payload", {})
                            print_success(f"Retrieved response: {json.dumps(response_payload, indent=2)}")
                            
                            # Resume flow with response
                            print_info("Resuming flow with response data...")
                            resume_result = client.resume_flow(
                                flow_id=flow_id,
                                event="request.answered",
                                data={
                                    "requestId": pending_request_id,
                                    "payload": response_payload
                                }
                            )
                            print_success(f"Flow resumed: {resume_result}")
                            
                            # Update cursor to mark as completed
                            flow = client.get_flow(flow_id)
                            cursor = flow.get("cursor", {})
                            cursor["lastRequestId"] = pending_request_id
                            cursor["lastResponse"] = response_payload
                            cursor["step"] = "completed"
                            
                            # Save updated state
                            client.save_flow_state(flow_id, entity_id=entity_id, cursor=cursor)
                            print_success("Flow completed and state saved")
                            
                            # Clean up state file for next test run
                            state_file.unlink()
                            print_info("State file cleaned up - ready for next test run")
                            
                            return True
                            
                        except Exception as e:
                            print_error(f"Failed to retrieve response: {e}")
                            return False
                    else:
                        print_warning(f"Request still {request_status} - waiting for user to answer")
                        print_info(f"To answer: http://localhost:5173?entityId={entity_id}")
                        print_info(f"Request ID: {pending_request_id}")
                        print_info("Run this test again after answering to verify recovery")
                        return True  # This is expected - waiting for answer
                        
                except Exception as e:
                    print_error(f"Failed to check request status: {e}")
                    print_info("Request may have been deleted - creating new flow...")
                    flow_id = None
                    entity_id = None
                    pending_request_id = None
            else:
                print_info("No pending request in state - creating new flow...")
                flow_id = None
                entity_id = None
                
        except Exception as e:
            print_error(f"Failed to load state: {e}")
            print_info("Creating new flow...")
            flow_id = None
            entity_id = None
            pending_request_id = None
    else:
        print_info("No existing state found - creating new flow...")
        flow_id = None
        entity_id = None
        pending_request_id = None
    
    # Step 2: Create new flow if needed
    if not flow_id:
        print_header("CREATING NEW FLOW")
        
        # Get or create entity
        if not entity_id:
            # Try to find existing entity from inquiries
            try:
                inquiries = client.list_inquiries(status="PENDING")
                items = inquiries.get("items", [])
                if items:
                    entity_id = items[0].get("entityId")
                    print_info(f"Found existing entity from pending inquiries: {entity_id}")
            except Exception:
                pass
            
            if not entity_id:
                try:
                    entity = client.create_entity(
                        kind="user",
                        handle=f"test-flow-user-{uuid.uuid4().hex[:8]}",
                        meta={"name": "Test Flow User", "test": True}
                    )
                    entity_id = entity["id"]
                    print_success(f"Created new entity: {entity_id}")
                except ValueError:
                    entity = client.create_entity(
                        kind="user",
                        handle=f"test-flow-user-{uuid.uuid4().hex[:8]}",
                        meta={"name": "Test Flow User", "test": True}
                    )
                    entity_id = entity["id"]
                    print_success(f"Created new entity with unique handle: {entity_id}")
        
        # Create flow
        print_info("Creating flow...")
        flow = client.create_flow(
            kind="test-flow",
            owner_entity=entity_id,
            cursor={"step": "requesting-input", "data": {}}
        )
        flow_id = flow["id"]
        print_success(f"Created flow: {flow_id}")
        
        # Create request
        print_info("Creating request...")
        schema = {
            "type": "object",
            "properties": {
                "answer": {"type": "string", "title": "Your Answer"},
                "test": {"type": "boolean", "title": "Test Field", "default": True}
            },
            "required": ["answer"]
        }
        request = client.create_request(
            entity_id=entity_id,
            schema=schema,
            ui_hints={"answer": {"ui:help": "Enter your test answer"}}
        )
        request_id = request["requestId"]
        print_success(f"Created request: {request_id}")
        
        # Update flow cursor
        flow = client.get_flow(flow_id)
        cursor = flow.get("cursor", {})
        cursor["pendingRequestId"] = request_id
        cursor["step"] = "waiting-for-user-response"
        cursor["createdAt"] = time.time()
        
        # Save state
        print_info("Saving flow state...")
        client.save_flow_state(flow_id, entity_id=entity_id, cursor=cursor)
        print_success(f"State saved to {STATE_FILE}")
        
        # Auto-answer if requested (for automated testing)
        if auto_answer:
            print_info("Auto-answering request for testing...")
            time.sleep(1)  # Small delay
            try:
                response_result = client.post_response(
                    request_id=request_id,
                    entity_id=entity_id,
                    payload={"answer": f"Auto-answered test response {uuid.uuid4().hex[:8]}", "test": True}
                )
                print_success(f"Request answered: {response_result.get('responseId')}")
                print_info("Run test again to verify recovery of answered request")
            except Exception as e:
                print_error(f"Failed to auto-answer: {e}")
        else:
            print_info(f"\nTo answer this request:")
            print(f"  1. Open: http://localhost:5173?entityId={entity_id}")
            print(f"  2. Find request: {request_id}")
            print(f"  3. Submit your answer")
            print(f"\nThen run this test again to verify recovery")
        
        return True
    
    return True

def main():
    parser = argparse.ArgumentParser(description="Test flow suspend/resume recovery")
    parser.add_argument("--auto-answer", action="store_true", 
                       help="Automatically answer the request (for automated testing)")
    args = parser.parse_args()
    
    try:
        # Check backend is available
        import requests
        resp = requests.get(f"{BASE_URL}/healthz", timeout=2)
        resp.raise_for_status()
    except Exception as e:
        print_error(f"Cannot connect to backend at {BASE_URL}: {e}")
        print_info("Make sure docker-compose is running: docker-compose up -d")
        sys.exit(1)
    
    success = test_flow_recovery(auto_answer=args.auto_answer)
    
    if success:
        print_header("TEST SUMMARY")
        print_success("Flow recovery test completed successfully")
        print_info("Run multiple times to test suspend/resume cycle")
        sys.exit(0)
    else:
        print_header("TEST SUMMARY")
        print_error("Flow recovery test failed")
        sys.exit(1)

if __name__ == "__main__":
    main()

