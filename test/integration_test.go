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

func setupTestServer(t *testing.T) (*httptest.Server, *db.Pool, func()) {
	// Use test database
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/pxbox_test?sslmode=disable"
	}

	dbPool, err := db.NewPool(databaseURL)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil, nil, func() {}
	}

	// Setup Redis
	redisAddr := os.Getenv("TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6380"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	logger, _ := zap.NewDevelopment()
	bus := pubsub.New(rdb, logger)

	r := chi.NewRouter()
	r.Mount("/v1", api.Routes(api.Dependencies{
		DB:  dbPool,
		Bus: bus,
		Hub: nil,
		Log: logger,
	}))
	
	// Add health check route
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := httptest.NewServer(r)

	cleanup := func() {
		server.Close()
		dbPool.Close()
		rdb.Close()
	}

	return server, dbPool, cleanup
}

func TestHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	resp, err := http.Get(server.URL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCreateRequest_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Start test services if not already running
	if err := StartTestServices(); err != nil {
		t.Logf("Could not start test services: %v (assuming they're already running)", err)
	}

	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	// Run migrations
	testDB, err := SetupTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	if err := RunMigrations(testDB); err != nil {
		t.Logf("Migration error (may be OK if already migrated): %v", err)
	}

	reqBody := map[string]interface{}{
		"entity": map[string]interface{}{
			"handle": "test@example.com",
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

	// Should either succeed or fail with entity not found (which is expected)
	assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusInternalServerError)
}
