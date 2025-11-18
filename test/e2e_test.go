package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// E2E test that exercises the full request/response flow via API
func TestE2E_RequestResponseFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		t.Skip("Skipping E2E test: TEST_API_URL not set (requires docker-compose)")
	}
	
	// Quick health check
	healthResp, err := http.Get(baseURL + "/healthz")
	if err != nil || healthResp.StatusCode != http.StatusOK {
		t.Skip("Skipping E2E test: server not available")
	}
	healthResp.Body.Close()

	// Setup: Create entity
	entityID := "550e8400-e29b-41d4-a716-446655440000"
	testDB, err := SetupTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	_, err = testDB.Exec(`
		INSERT INTO entities (id, kind, handle, meta)
		VALUES ($1, 'user', 'test@example.com', '{}')
		ON CONFLICT (id) DO UPDATE SET handle = EXCLUDED.handle
	`, entityID)
	require.NoError(t, err)

	// Step 1: Create request
	createReq := map[string]interface{}{
		"entity": map[string]interface{}{
			"id": entityID,
		},
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"fullName": map[string]interface{}{
					"type": "string",
					"title": "Full Name",
				},
				"email": map[string]interface{}{
					"type": "string",
					"format": "email",
				},
			},
			"required": []string{"fullName", "email"},
		},
		"uiHints": map[string]interface{}{
			"fullName": map[string]interface{}{
				"ui:help": "Enter your legal name",
			},
		},
		"prefill": map[string]interface{}{
			"email": "test@example.com",
		},
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", baseURL+"/v1/requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-ID", "test-client")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	// Accept both 200 and 201 as success (some proxies/containers return 200)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 201 or 200, got %d", resp.StatusCode)
	}

	var createResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResult)
	resp.Body.Close()

	// Handle both success formats
	requestID, ok := createResult["requestId"].(string)
	if !ok {
		// Try alternative field name or check for error
		if errMsg, hasErr := createResult["error"].(string); hasErr {
			t.Fatalf("Request creation failed: %s", errMsg)
		}
		t.Fatalf("Invalid response format: %+v", createResult)
	}
	require.NotEmpty(t, requestID)

	// Step 2: Get request
	getReq, _ := http.NewRequest("GET", baseURL+"/v1/requests/"+requestID, nil)
	getResp, err := http.DefaultClient.Do(getReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var getResult map[string]interface{}
	json.NewDecoder(getResp.Body).Decode(&getResult)
	getResp.Body.Close()

	assert.Equal(t, "PENDING", getResult["status"])
	assert.NotNil(t, getResult["schemaPayload"])

	// Step 3: Claim request
	claimReq, _ := http.NewRequest("POST", baseURL+"/v1/requests/"+requestID+"/claim", nil)
	claimReq.Header.Set("X-Entity-ID", entityID)
	claimResp, err := http.DefaultClient.Do(claimReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, claimResp.StatusCode)
	claimResp.Body.Close()

	// Step 4: Submit response
	responseData := map[string]interface{}{
		"payload": map[string]interface{}{
			"fullName": "John Doe",
			"email":    "john@example.com",
		},
	}

	responseBody, _ := json.Marshal(responseData)
	postReq, _ := http.NewRequest("POST", baseURL+"/v1/requests/"+requestID+"/response", bytes.NewReader(responseBody))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Entity-ID", entityID)

	postResp, err := http.DefaultClient.Do(postReq)
	require.NoError(t, err)
	
	if postResp.StatusCode != http.StatusCreated {
		// Read error message for debugging (may be plain text or JSON)
		errorBytes := make([]byte, 1024)
		n, _ := postResp.Body.Read(errorBytes)
		errorMsg := string(errorBytes[:n])
		postResp.Body.Close()
		t.Fatalf("Expected 201, got %d. Error body: %s", postResp.StatusCode, errorMsg)
	}

	var postResult map[string]interface{}
	json.NewDecoder(postResp.Body).Decode(&postResult)
	postResp.Body.Close()

	assert.Equal(t, "ANSWERED", postResult["status"])
	assert.NotEmpty(t, postResult["responseId"])

	// Step 5: Verify request status changed
	getReq2, _ := http.NewRequest("GET", baseURL+"/v1/requests/"+requestID, nil)
	getResp2, err := http.DefaultClient.Do(getReq2)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getResp2.StatusCode)

	var getResult2 map[string]interface{}
	json.NewDecoder(getResp2.Body).Decode(&getResult2)
	getResp2.Body.Close()

	assert.Equal(t, "ANSWERED", getResult2["status"])
}

func TestE2E_FlowSuspendResume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	baseURL := os.Getenv("TEST_API_URL")
	if baseURL == "" {
		t.Skip("Skipping E2E test: TEST_API_URL not set (requires docker-compose)")
	}
	
	// Quick health check
	healthResp, err := http.Get(baseURL + "/healthz")
	if err != nil || healthResp.StatusCode != http.StatusOK {
		t.Skip("Skipping E2E test: server not available")
	}
	healthResp.Body.Close()

	entityID := "550e8400-e29b-41d4-a716-446655440000"
	testDB, err := SetupTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	// Ensure entity exists
	_, err = testDB.Exec(`
		INSERT INTO entities (id, kind, handle, meta)
		VALUES ($1, 'user', 'test@example.com', '{}')
		ON CONFLICT (id) DO UPDATE SET handle = EXCLUDED.handle
	`, entityID)
	require.NoError(t, err)

	// Create flow
	flowReq := map[string]interface{}{
		"kind":        "test-flow",
		"ownerEntity": entityID,
		"cursor": map[string]interface{}{
			"step": "waiting-input",
		},
	}

	body, _ := json.Marshal(flowReq)
	req, _ := http.NewRequest("POST", baseURL+"/v1/flows", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	// Accept both 200 and 201 as success (some proxies/containers return 200)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 201 or 200, got %d", resp.StatusCode)
	}

	var flowResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&flowResult)
	resp.Body.Close()

	flowID := flowResult["id"].(string)
	require.NotEmpty(t, flowID)

	// Resume flow with event
	resumeReq := map[string]interface{}{
		"event": "request.answered",
		"data": map[string]interface{}{
			"payload": map[string]interface{}{
				"answer": "test",
			},
		},
	}

	resumeBody, _ := json.Marshal(resumeReq)
	resumeHTTPReq, _ := http.NewRequest("POST", baseURL+"/v1/flows/"+flowID+"/resume", bytes.NewReader(resumeBody))
	resumeHTTPReq.Header.Set("Content-Type", "application/json")

	resumeResp, err := http.DefaultClient.Do(resumeHTTPReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resumeResp.StatusCode)

	var resumeResult map[string]interface{}
	json.NewDecoder(resumeResp.Body).Decode(&resumeResult)
	resumeResp.Body.Close()

	assert.Equal(t, "RUNNING", resumeResult["status"])
}

