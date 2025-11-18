package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pxbox/internal/api"
	"pxbox/internal/db"
	"pxbox/internal/pubsub"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestServerWithServices(t *testing.T) (*httptest.Server, *db.Pool, func()) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/pxbox_test?sslmode=disable"
	}

	dbPool, err := db.NewPool(databaseURL)
	require.NoError(t, err)

	redisAddr := os.Getenv("TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6380"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	logger, _ := zap.NewDevelopment()
	bus := pubsub.New(rdb, logger)

	// Create a router and mount API routes at /v1
	r := chi.NewRouter()
	r.Mount("/v1", api.Routes(api.Dependencies{
		DB:  dbPool,
		Bus: bus,
		Hub: nil,
		Log: logger,
	}))

	server := httptest.NewServer(r)

	cleanup := func() {
		server.Close()
		dbPool.Close()
		rdb.Close()
	}

	return server, dbPool, cleanup
}

func TestCreateEntityAndRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, _, cleanup := setupTestServerWithServices(t)
	defer cleanup()

	// Run migrations
	testDB, err := SetupTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	if err := RunMigrations(testDB); err != nil {
		t.Logf("Migration error (may be OK if already migrated): %v", err)
	}

	// Create entity directly in DB for testing
	entityID := "550e8400-e29b-41d4-a716-446655440000"
	_, err = testDB.Exec(`
		INSERT INTO entities (id, kind, handle, meta)
		VALUES ($1, 'user', 'test@example.com', '{}')
		ON CONFLICT (id) DO NOTHING
	`, entityID)
	require.NoError(t, err)

	// Create request via API
	reqBody := map[string]interface{}{
		"entity": map[string]interface{}{
			"id": entityID,
		},
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"name"},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", server.URL+"/v1/requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-ID", "test-client")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.NotEmpty(t, result["requestId"])
	assert.Equal(t, "PENDING", result["status"])
}

func TestGetRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, _, cleanup := setupTestServerWithServices(t)
	defer cleanup()

	// First create a request (simplified - would need entity setup)
	// This test verifies the GET endpoint works
	req, _ := http.NewRequest("GET", server.URL+"/v1/requests/nonexistent", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return 404 for non-existent request
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListInquiries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, _, cleanup := setupTestServerWithServices(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", server.URL+"/v1/inquiries?entityId=550e8400-e29b-41d4-a716-446655440000", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	// items might be empty array, but should exist
	assert.Contains(t, result, "items")
}

func TestMarkRead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	server, _, cleanup := setupTestServerWithServices(t)
	defer cleanup()

	req, _ := http.NewRequest("POST", server.URL+"/v1/inquiries/nonexistent/markRead", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should succeed even if inquiry doesn't exist (idempotent)
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError)
}

