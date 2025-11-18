#!/usr/bin/env python3
"""
Comprehensive API flow test simulating both client and user roles

This test simulates:
1. Client role: Creating entities and requests
2. User role: Viewing inquiries, claiming requests, submitting responses

Run with: python test_api_flow.py
"""

import sys
import time
import uuid
import requests
from pathlib import Path

# Add parent directory to path to import pxbox_client
sys.path.insert(0, str(Path(__file__).parent))

from pxbox_client import PxBoxClient

BASE_URL = "http://localhost:8082"

class Colors:
    """ANSI color codes for terminal output"""
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKCYAN = '\033[96m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'

def print_header(text):
    print(f"\n{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{text}{Colors.ENDC}")
    print(f"{Colors.HEADER}{Colors.BOLD}{'='*60}{Colors.ENDC}\n")

def print_success(text):
    print(f"{Colors.OKGREEN}✓ {text}{Colors.ENDC}")

def print_info(text):
    print(f"{Colors.OKCYAN}ℹ {text}{Colors.ENDC}")

def print_error(text):
    print(f"{Colors.FAIL}✗ {text}{Colors.ENDC}")

def print_warning(text):
    print(f"{Colors.WARNING}⚠ {text}{Colors.ENDC}")

def test_client_role(client):
    """Test client role: Creating entities and requests"""
    print_header("CLIENT ROLE: Creating Requests")
    
    # Create client entity
    print_info("Creating client entity...")
    try:
        client_entity = client.create_entity(
            kind="bot",
            handle=f"test-client-{uuid.uuid4().hex[:8]}",
            meta={"name": "Test Client", "role": "client"}
        )
        client_entity_id = client_entity['id']
        print_success(f"Created client entity: {client_entity_id}")
    except ValueError:
        # Handle duplicate
        client_entity = client.create_entity(
            kind="bot",
            handle=f"test-client-{uuid.uuid4().hex[:8]}",
            meta={"name": "Test Client", "role": "client"}
        )
        client_entity_id = client_entity['id']
        print_success(f"Created client entity: {client_entity_id}")
    
    # Create user entity (the one who will receive requests)
    print_info("Creating user entity...")
    try:
        user_entity = client.create_entity(
            kind="user",
            handle=f"test-user-{uuid.uuid4().hex[:8]}",
            meta={"name": "Test User", "role": "user"}
        )
        user_entity_id = user_entity['id']
        print_success(f"Created user entity: {user_entity_id}")
    except ValueError:
        user_entity = client.create_entity(
            kind="user",
            handle=f"test-user-{uuid.uuid4().hex[:8]}",
            meta={"name": "Test User", "role": "user"}
        )
        user_entity_id = user_entity['id']
        print_success(f"Created user entity: {user_entity_id}")
    
    # Create multiple requests with different schemas
    requests_created = []
    
    # Request 1: Simple contact form
    print_info("Creating request 1: Contact form...")
    req1 = client.create_request(
        entity_id=user_entity_id,
        schema={
            "type": "object",
            "properties": {
                "name": {"type": "string", "title": "Full Name"},
                "email": {"type": "string", "format": "email", "title": "Email Address"},
                "phone": {"type": "string", "title": "Phone Number"},
                "message": {"type": "string", "title": "Message", "ui:widget": "textarea"}
            },
            "required": ["name", "email", "message"]
        },
        ui_hints={
            "name": {"description": "Enter your first and last name"},
            "email": {"description": "We'll never share your email"},
            "message": {"description": "Tell us how we can help"}
        }
    )
    requests_created.append({
        "id": req1["requestId"],
        "entity_id": user_entity_id,
        "type": "contact"
    })
    print_success(f"Created request: {req1['requestId']}")
    
    # Request 2: Survey form
    print_info("Creating request 2: Survey form...")
    req2 = client.create_request(
        entity_id=user_entity_id,
        schema={
            "type": "object",
            "properties": {
                "satisfaction": {
                    "type": "integer",
                    "title": "Satisfaction Level",
                    "minimum": 1,
                    "maximum": 5
                },
                "feedback": {"type": "string", "title": "Additional Feedback", "ui:widget": "textarea"},
                "recommend": {"type": "boolean", "title": "Would you recommend us?"}
            },
            "required": ["satisfaction"]
        }
    )
    requests_created.append({
        "id": req2["requestId"],
        "entity_id": user_entity_id,
        "type": "survey"
    })
    print_success(f"Created request: {req2['requestId']}")
    
    # Request 3: With deadline
    print_info("Creating request 3: With deadline...")
    from datetime import datetime, timedelta, timezone
    deadline = datetime.now(timezone.utc) + timedelta(days=1)
    # Use direct API call for deadline since client method may not support it
    req3_resp = requests.post(
        f"{BASE_URL}/v1/requests",
        json={
            "entity": {"id": user_entity_id},
            "schema": {
                "type": "object",
                "properties": {
                    "response": {"type": "string", "title": "Your Response"}
                },
                "required": ["response"]
            },
            "deadlineAt": deadline.isoformat()
        }
    )
    req3_resp.raise_for_status()
    req3 = req3_resp.json()
    requests_created.append({
        "id": req3.get("requestId") or req3.get("id"),
        "entity_id": user_entity_id,
        "type": "deadline"
    })
    request_id = req3.get("requestId") or req3.get("id")
    print_success(f"Created request with deadline: {request_id}")
    
    return {
        "client_entity_id": client_entity_id,
        "user_entity_id": user_entity_id,
        "requests": requests_created
    }

def test_user_role(user_entity_id, requests_data):
    """Test user role: Viewing inquiries, claiming, and responding"""
    print_header("USER ROLE: Viewing and Responding to Requests")
    
    # Simulate user opening inbox - fetch inquiries
    print_info(f"Fetching inquiries for entity {user_entity_id}...")
    resp = requests.get(f"{BASE_URL}/v1/inquiries", params={"entityId": user_entity_id})
    resp.raise_for_status()
    inquiries_data = resp.json()
    
    items = inquiries_data.get("items") or []
    total = inquiries_data.get("total", 0)
    
    if total == 0:
        print_error(f"No inquiries found! Expected {len(requests_data)} requests")
        print_warning("This indicates the ListInquiries query is not working correctly")
        return False
    
    print_success(f"Found {total} inquiries")
    
    if len(items) != len(requests_data):
        print_warning(f"Expected {len(requests_data)} inquiries, got {len(items)}")
    
    # Display inquiries
    print_info("\nInquiries list:")
    for idx, inquiry in enumerate(items, 1):
        print(f"  {idx}. Request {inquiry['id']} - Status: {inquiry['status']}")
        if inquiry.get('deadlineAt'):
            print(f"     Deadline: {inquiry['deadlineAt']}")
    
    # Test: Get first request details
    if items:
        first_request_id = items[0]['id']
        print_info(f"\nFetching details for request {first_request_id}...")
        resp = requests.get(f"{BASE_URL}/v1/requests/{first_request_id}")
        resp.raise_for_status()
        request_details = resp.json()
        print_success(f"Retrieved request details")
        print(f"  Entity ID: {request_details.get('entityId')}")
        print(f"  Status: {request_details.get('status')}")
        print(f"  Schema Kind: {request_details.get('schemaKind')}")
        
        # Test: Claim request
        print_info(f"\nClaiming request {first_request_id}...")
        resp = requests.post(
            f"{BASE_URL}/v1/requests/{first_request_id}/claim",
            headers={"X-Entity-ID": user_entity_id}
        )
        resp.raise_for_status()
        claim_result = resp.json()
        print_success(f"Request claimed: {claim_result.get('status')}")
        
        # Verify status changed
        resp = requests.get(f"{BASE_URL}/v1/requests/{first_request_id}")
        resp.raise_for_status()
        updated_request = resp.json()
        if updated_request.get('status') == 'CLAIMED':
            print_success("Request status updated to CLAIMED")
        else:
            print_error(f"Expected status CLAIMED, got {updated_request.get('status')}")
        
        # Test: Submit response
        print_info(f"\nSubmitting response for request {first_request_id}...")
        # Get the request schema to determine what fields are required
        resp = requests.get(f"{BASE_URL}/v1/requests/{first_request_id}")
        resp.raise_for_status()
        request_details = resp.json()
        schema = request_details.get('schemaPayload', {}) or request_details.get('schema', {})
        properties = schema.get('properties', {})
        
        # Build response payload based on schema
        required_fields = schema.get('required', [])
        response_payload = {}
        
        # Fill required fields first
        for field in required_fields:
            if field == 'name':
                response_payload['name'] = "Test User"
            elif field == 'email':
                response_payload['email'] = "test@example.com"
            elif field == 'message':
                response_payload['message'] = "This is a test response from the API test"
            elif field == 'response':
                response_payload['response'] = "This is a test response"
            elif field == 'satisfaction':
                response_payload['satisfaction'] = 5
            elif field == 'recommend':
                response_payload['recommend'] = True
        
        # Fill optional fields
        if 'name' in properties and 'name' not in response_payload:
            response_payload['name'] = "Test User"
        if 'email' in properties and 'email' not in response_payload:
            response_payload['email'] = "test@example.com"
        if 'message' in properties and 'message' not in response_payload:
            response_payload['message'] = "This is a test response from the API test"
        if 'feedback' in properties and 'feedback' not in response_payload:
            response_payload['feedback'] = "Great service!"
        
        resp = requests.post(
            f"{BASE_URL}/v1/requests/{first_request_id}/response",
            headers={
                "Content-Type": "application/json",
                "X-Entity-ID": user_entity_id
            },
            json={
                "payload": response_payload
            }
        )
        resp.raise_for_status()
        response_result = resp.json()
        print_success(f"Response submitted: {response_result.get('responseId')}")
        
        # Verify request is answered
        resp = requests.get(f"{BASE_URL}/v1/requests/{first_request_id}")
        resp.raise_for_status()
        final_request = resp.json()
        if final_request.get('status') == 'ANSWERED':
            print_success("Request status updated to ANSWERED")
        else:
            print_error(f"Expected status ANSWERED, got {final_request.get('status')}")
        
        return True
    else:
        print_error("No inquiries to process")
        return False

def test_inquiry_filters(user_entity_id):
    """Test inquiry filtering and grouping"""
    print_header("TESTING: Inquiry Filters")
    
    # Test: Get all inquiries
    print_info("Testing: Get all inquiries...")
    resp = requests.get(f"{BASE_URL}/v1/inquiries", params={"entityId": user_entity_id})
    resp.raise_for_status()
    all_data = resp.json()
    all_count = all_data.get("total", 0)
    print_success(f"All inquiries: {all_count}")
    
    # Test: Filter by status
    print_info("Testing: Filter by status=PENDING...")
    resp = requests.get(
        f"{BASE_URL}/v1/inquiries",
        params={"entityId": user_entity_id, "status": "PENDING"}
    )
    resp.raise_for_status()
    pending_data = resp.json()
    pending_count = pending_data.get("total", 0)
    print_success(f"Pending inquiries: {pending_count}")
    
    # Test: Filter by status=ANSWERED
    print_info("Testing: Filter by status=ANSWERED...")
    resp = requests.get(
        f"{BASE_URL}/v1/inquiries",
        params={"entityId": user_entity_id, "status": "ANSWERED"}
    )
    resp.raise_for_status()
    answered_data = resp.json()
    answered_count = answered_data.get("total", 0)
    print_success(f"Answered inquiries: {answered_count}")
    
    return True

def main():
    """Run all tests"""
    print_header("PxBox API Flow Test")
    print_info("Testing both client and user roles through REST API")
    print_info(f"Backend URL: {BASE_URL}\n")
    
    # Check backend is running
    try:
        resp = requests.get(f"{BASE_URL}/v1/inquiries")
        resp.raise_for_status()
    except requests.exceptions.ConnectionError:
        print_error(f"Cannot connect to backend at {BASE_URL}")
        print_info("Make sure docker-compose is running: docker-compose up -d")
        sys.exit(1)
    except Exception as e:
        print_error(f"Backend error: {e}")
        sys.exit(1)
    
    client = PxBoxClient(base_url=BASE_URL)
    
    try:
        # Test client role
        test_data = test_client_role(client)
        
        # Small delay to ensure requests are persisted
        time.sleep(1)
        
        # Test user role
        user_success = test_user_role(test_data["user_entity_id"], test_data["requests"])
        
        # Test filters
        filter_success = test_inquiry_filters(test_data["user_entity_id"])
        
        # Summary
        print_header("TEST SUMMARY")
        print_success(f"Client entity: {test_data['client_entity_id']}")
        print_success(f"User entity: {test_data['user_entity_id']}")
        print_success(f"Requests created: {len(test_data['requests'])}")
        print(f"\nUser role test: {'PASSED' if user_success else 'FAILED'}")
        print(f"Filter test: {'PASSED' if filter_success else 'FAILED'}")
        
        if user_success and filter_success:
            print_success("\n✅ All API tests passed!")
            print_info(f"\n📋 Test with UI: http://localhost:5173?entityId={test_data['user_entity_id']}")
            return 0
        else:
            print_error("\n❌ Some tests failed")
            return 1
            
    except Exception as e:
        print_error(f"\nTest failed with error: {e}")
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    sys.exit(main())

