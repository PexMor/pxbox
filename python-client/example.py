#!/usr/bin/env python3
"""
End-to-end example: Create request and wait for web UI response

This example demonstrates:
1. Creating a data-entry request via REST API
2. Waiting for a user to respond via the web UI
3. Polling for the response status

Prerequisites:
- Backend running via docker-compose (port 8082)
- Frontend running via yarn dev (port 5173)
- An entity created (use Step 2 from quickstart guide)
"""

import time
import sys
from pathlib import Path

# Add parent directory to path to import pxbox_client
sys.path.insert(0, str(Path(__file__).parent))

from pxbox_client import PxBoxClient

# Configuration
BASE_URL = "http://localhost:8082"
ENTITY_HANDLE = "example-user"  # Handle for the entity (will be created if doesn't exist)

def ensure_entity(client, handle):
    """Ensure entity exists, create if it doesn't"""
    import uuid
    import requests
    
    try:
        # Try to create entity
        entity = client.create_entity(
            kind="user",
            handle=handle,
            meta={"name": "Example User", "created_by": "example.py"}
        )
        print(f"✓ Created entity: {entity['id']} (handle: {handle})")
        return entity['id']
    except ValueError as e:
        # Duplicate handle detected
        error_str = str(e).lower()
        if "already exists" in error_str or "duplicate" in error_str:
            print(f"⚠️  Entity with handle '{handle}' already exists")
            print("   Creating entity with unique handle...")
            unique_handle = f"{handle}-{uuid.uuid4().hex[:8]}"
            try:
                entity = client.create_entity(
                    kind="user",
                    handle=unique_handle,
                    meta={"name": "Example User", "created_by": "example.py"}
                )
                print(f"✓ Created entity: {entity['id']} (handle: {unique_handle})")
                return entity['id']
            except Exception as e2:
                print(f"❌ Failed to create entity with unique handle: {e2}")
                raise
        else:
            raise
    except requests.HTTPError as e:
        # Try to get response body for better error detection
        error_str = str(e).lower()
        response_text = ""
        try:
            if hasattr(e.response, 'text'):
                response_text = e.response.text.lower()
        except:
            pass
        
        # Check if it's a duplicate error from the server
        if ("duplicate" in error_str or "23505" in error_str or "unique constraint" in error_str or
            "duplicate" in response_text or "23505" in response_text or "unique constraint" in response_text):
            print(f"⚠️  Entity with handle '{handle}' already exists")
            print("   Creating entity with unique handle...")
            unique_handle = f"{handle}-{uuid.uuid4().hex[:8]}"
            try:
                entity = client.create_entity(
                    kind="user",
                    handle=unique_handle,
                    meta={"name": "Example User", "created_by": "example.py"}
                )
                print(f"✓ Created entity: {entity['id']} (handle: {unique_handle})")
                return entity['id']
            except Exception as e2:
                print(f"❌ Failed to create entity: {e2}")
                raise
        else:
            print(f"❌ Failed to create entity: {e}")
            if response_text:
                print(f"   Server response: {response_text[:200]}")
            print("\n💡 You can create an entity manually:")
            print(f"   curl -X POST {BASE_URL}/v1/entities \\")
            print(f"     -H 'Content-Type: application/json' \\")
            print(f"     -d '{{\"kind\":\"user\",\"handle\":\"{handle}\",\"meta\":{{\"name\":\"Example User\"}}}}'")
            raise
    except Exception as e:
        print(f"❌ Failed to create entity: {e}")
        print("\n💡 You can create an entity manually:")
        print(f"   curl -X POST {BASE_URL}/v1/entities \\")
        print(f"     -H 'Content-Type: application/json' \\")
        print(f"     -d '{{\"kind\":\"user\",\"handle\":\"{handle}\",\"meta\":{{\"name\":\"Example User\"}}}}'")
        raise

def main():
    """Run the end-to-end example"""
    print("🚀 PxBox End-to-End Example\n")
    print("This example will:")
    print("1. Ensure an entity exists (create if needed)")
    print("2. Create a data-entry request")
    print("3. Wait for you to answer it in the web UI")
    print("4. Detect when the response is submitted\n")
    
    # Initialize client
    client = PxBoxClient(base_url=BASE_URL)
    
    # Ensure entity exists
    print("🔍 Checking entity...")
    try:
        entity_id = ensure_entity(client, ENTITY_HANDLE)
    except Exception as e:
        print(f"\n❌ Entity setup failed: {e}")
        print("\nPlease create an entity manually and update ENTITY_ID in this script.")
        return
    
    # Create a simple request
    print("\n📝 Creating request...")
    request = client.create_request(
        entity_id=entity_id,
        schema={
            "type": "object",
            "properties": {
                "name": {
                    "type": "string",
                    "title": "Your Name"
                },
                "email": {
                    "type": "string",
                    "format": "email",
                    "title": "Email Address"
                },
                "message": {
                    "type": "string",
                    "title": "Message",
                    "ui:widget": "textarea"
                }
            },
            "required": ["name", "email"]
        },
        ui_hints={
            "name": {
                "title": "Full Name",
                "description": "Enter your first and last name"
            },
            "email": {
                "title": "Email",
                "description": "We'll never share your email"
            }
        }
    )
    
    request_id = request["requestId"]
    print(f"✓ Created request: {request_id}")
    print(f"\n📋 Open the web UI and answer this request:")
    print(f"   http://localhost:5173?entityId={entity_id}")
    print(f"\n   Entity ID: {entity_id}")
    print(f"   Request ID: {request_id}")
    print(f"\n⏳ Waiting for response...")
    print("   (The script will poll every second for up to 5 minutes)")
    
    # Poll for response
    try:
        final_request = client.poll_request_status(request_id, timeout=300)
        print(f"\n✅ Request answered!")
        print(f"   Status: {final_request['status']}")
        
        # Get the response details
        if final_request['status'] == 'ANSWERED':
            print(f"\n📄 Retrieving response data...")
            try:
                response = client.get_response_by_request(request_id)
                print(f"✅ Response received!")
                print(f"   Response ID: {response.get('id')}")
                print(f"   Answered By: {response.get('answeredBy')}")
                print(f"   Answered At: {response.get('answeredAt')}")
                print(f"\n   User's Response Data:")
                payload = response.get('payload', {})
                for key, value in payload.items():
                    print(f"     {key}: {value}")
                
                # Application can now continue with the response data
                print(f"\n💡 Application can now continue with:")
                print(f"   response_data = {payload}")
            except Exception as e:
                print(f"⚠️  Could not retrieve response payload: {e}")
                print(f"   Request is ANSWERED but response endpoint may not be available")
                print(f"   Check backend logs or database for response data")
        
    except TimeoutError:
        print("\n⏱️  Request still pending (timeout after 5 minutes)")
        print("   You can check status later or answer it in the web UI")
        print(f"   Request ID: {request_id}")
        sys.exit(1)
    
    print("\n✅ Example completed successfully!")

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n⚠️  Interrupted by user")
        sys.exit(0)
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

