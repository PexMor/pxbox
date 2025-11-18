package test

import (
	"context"
	"os"
	"testing"
	"time"

	"pxbox/internal/db"
	"pxbox/internal/jobs"
	"pxbox/internal/pubsub"
	"pxbox/internal/schema"
	"pxbox/internal/service"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDeadlineNotificationJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: getRedisAddr(),
	})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Setup database
	dbPool := setupTestDB(t)
	defer dbPool.Close()

	// Setup logger
	logger := zap.NewNop()

	// Setup pub/sub bus
	bus := pubsub.New(rdb, logger)

	// Create job server
	redisAddr := getRedisAddr()
	jobServer, jobClient := jobs.NewJobServer(redisAddr, dbPool, bus, logger)
	defer jobServer.Stop()

	// Create a request with deadline in the past (for immediate notification)
	entityID := createTestEntity(t, dbPool, "test-entity")
	deadline := time.Now().Add(-2 * time.Hour)
	requestID := createTestRequestWithDeadline(t, dbPool, entityID, deadline)

	// Schedule deadline notification job (should execute immediately since deadline is in the past)
	err := jobs.ScheduleDeadlineNotification(jobClient, requestID, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	// Start job server in background
	go func() {
		if err := jobServer.Start(); err != nil {
			t.Logf("Job server error: %v", err)
		}
	}()

	// Wait a bit for job to process
	time.Sleep(2 * time.Second)

	// Verify request still exists and is PENDING
	req, err := dbPool.Queries.GetRequestByID(ctx, requestID)
	require.NoError(t, err)
	assert.Equal(t, "PENDING", req.Status)
}

func TestDeadlineExpiryJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: getRedisAddr(),
	})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Setup database
	dbPool := setupTestDB(t)
	defer dbPool.Close()

	// Setup logger
	logger := zap.NewNop()

	// Setup pub/sub bus
	bus := pubsub.New(rdb, logger)

	// Create job server
	redisAddr := getRedisAddr()
	jobServer, jobClient := jobs.NewJobServer(redisAddr, dbPool, bus, logger)
	defer jobServer.Stop()

	// Create a request with deadline in the past
	entityID := createTestEntity(t, dbPool, "test-entity")
	deadline := time.Now().Add(-1 * time.Hour)
	requestID := createTestRequestWithDeadline(t, dbPool, entityID, deadline)

	// Schedule expiry job (should execute immediately)
	err := jobs.ScheduleDeadlineExpiry(jobClient, requestID, deadline)
	require.NoError(t, err)

	// Start job server in background
	go func() {
		if err := jobServer.Start(); err != nil {
			t.Logf("Job server error: %v", err)
		}
	}()

	// Wait for job to process
	time.Sleep(2 * time.Second)

	// Verify request status changed to EXPIRED
	req, err := dbPool.Queries.GetRequestByID(ctx, requestID)
	require.NoError(t, err)
	assert.Equal(t, "EXPIRED", req.Status)
}

func TestAutoCancelJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: getRedisAddr(),
	})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Setup database
	dbPool := setupTestDB(t)
	defer dbPool.Close()

	// Setup logger
	logger := zap.NewNop()

	// Setup pub/sub bus
	bus := pubsub.New(rdb, logger)

	// Create job server
	redisAddr := getRedisAddr()
	jobServer, jobClient := jobs.NewJobServer(redisAddr, dbPool, bus, logger)
	defer jobServer.Stop()

	// Create a request
	entityID := createTestEntity(t, dbPool, "test-entity")
	deadline := time.Now().Add(1 * time.Hour)
	requestID := createTestRequestWithDeadline(t, dbPool, entityID, deadline)

	// Schedule auto-cancel job with short grace period
	err := jobs.ScheduleAutoCancel(jobClient, requestID, 1*time.Second)
	require.NoError(t, err)

	// Start job server in background
	go func() {
		if err := jobServer.Start(); err != nil {
			t.Logf("Job server error: %v", err)
		}
	}()

	// Wait for job to process
	time.Sleep(2 * time.Second)

	// Verify request status changed to CANCELLED
	req, err := dbPool.Queries.GetRequestByID(ctx, requestID)
	require.NoError(t, err)
	assert.Equal(t, "CANCELLED", req.Status)
}

func TestAttentionNotificationJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: getRedisAddr(),
	})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Setup database
	dbPool := setupTestDB(t)
	defer dbPool.Close()

	// Setup logger
	logger := zap.NewNop()

	// Setup pub/sub bus
	bus := pubsub.New(rdb, logger)

	// Create job server
	redisAddr := getRedisAddr()
	jobServer, jobClient := jobs.NewJobServer(redisAddr, dbPool, bus, logger)
	defer jobServer.Stop()

	// Create a request with attention time in the past
	entityID := createTestEntity(t, dbPool, "test-entity")
	attentionAt := time.Now().Add(-1 * time.Hour)
	requestID := createTestRequestWithAttention(t, dbPool, entityID, attentionAt)

	// Schedule attention notification job
	err := jobs.ScheduleAttentionNotification(jobClient, requestID, attentionAt)
	require.NoError(t, err)

	// Start job server in background
	go func() {
		if err := jobServer.Start(); err != nil {
			t.Logf("Job server error: %v", err)
		}
	}()

	// Wait for job to process
	time.Sleep(2 * time.Second)

	// Verify request still exists and is PENDING
	req, err := dbPool.Queries.GetRequestByID(ctx, requestID)
	require.NoError(t, err)
	assert.Equal(t, "PENDING", req.Status)
}

// Helper functions

func setupTestDB(t *testing.T) *db.Pool {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/pxbox_test?sslmode=disable"
	}

	dbPool, err := db.NewPool(databaseURL)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil
	}

	return dbPool
}

func createTestEntity(t *testing.T, dbPool *db.Pool, handle string) string {
	ctx := context.Background()
	entity, err := dbPool.Queries.CreateEntity(ctx, "user", handle, map[string]interface{}{})
	require.NoError(t, err)
	return entity.ID
}

func createTestRequestWithDeadline(t *testing.T, dbPool *db.Pool, entityID string, deadline time.Time) string {
	ctx := context.Background()
	
	// Use service layer to create request properly
	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(dbPool.Queries)
	bus := pubsub.New(redis.NewClient(&redis.Options{Addr: getRedisAddr()}), zap.NewNop())
	requestSvc := service.NewRequestService(dbPool.Queries, schemaComp, entitySvc, bus)
	
	result, err := requestSvc.CreateRequest(ctx, service.CreateRequestInput{
		Entity:     entityID,
		Schema:     `{"type":"object","properties":{"name":{"type":"string"}}}`,
		DeadlineAt: &deadline,
		CreatedBy:  "test",
	})
	require.NoError(t, err)
	
	return result.ID
}

func createTestRequestWithAttention(t *testing.T, dbPool *db.Pool, entityID string, attentionAt time.Time) string {
	ctx := context.Background()
	
	// Use service layer to create request properly
	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(dbPool.Queries)
	bus := pubsub.New(redis.NewClient(&redis.Options{Addr: getRedisAddr()}), zap.NewNop())
	requestSvc := service.NewRequestService(dbPool.Queries, schemaComp, entitySvc, bus)
	
	result, err := requestSvc.CreateRequest(ctx, service.CreateRequestInput{
		Entity:      entityID,
		Schema:      `{"type":"object","properties":{"name":{"type":"string"}}}`,
		AttentionAt: &attentionAt,
		CreatedBy:   "test",
	})
	require.NoError(t, err)
	
	return result.ID
}

func getRedisAddr() string {
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6380"
	}
	return addr
}

