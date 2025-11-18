package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pxbox/internal/api"
	"pxbox/internal/db"
	"pxbox/internal/jobs"
	"pxbox/internal/pubsub"
	"pxbox/internal/schema"
	"pxbox/internal/service"
	"pxbox/internal/ws"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// Check for migrate command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runMigrations(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		os.Exit(0)
	}

	// Check for goose migrate command
	if len(os.Args) > 1 && os.Args[1] == "goose-migrate" {
		if err := runGooseMigrations(); err != nil {
			log.Fatalf("Goose migration failed: %v", err)
		}
		os.Exit(0)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Check for serve command (default)
	if len(os.Args) > 1 && os.Args[1] != "serve" && os.Args[1] != "migrate" {
		log.Fatalf("Unknown command: %s (use 'serve' or 'migrate')", os.Args[1])
	}

	// Database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/pxbox?sslmode=disable"
	}
	
	dbPool, err := db.NewPool(databaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbPool.Close()

	// Redis connection
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// Pub/sub bus
	bus := pubsub.New(rdb, logger)

	// Background jobs
	jobServer, jobClient := jobs.NewJobServer(redisAddr, dbPool, bus, logger)
	go func() {
		if err := jobServer.Start(); err != nil {
			logger.Fatal("Job server failed", zap.Error(err))
		}
	}()
	defer jobServer.Stop()

	// WebSocket hub
	hub := ws.NewHub(logger)
	// Create adapter to convert pubsub.Streams to ws.StreamsProvider
	streamsAdapter := &wsStreamsAdapter{streams: bus.GetStreams()}
	hub.SetStreamsProvider(streamsAdapter)
	go hub.Run()
	bus.SetWSHub(hub)

	// Initialize services for WebSocket commands
	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(dbPool.Queries)
	requestSvc := service.NewRequestService(dbPool.Queries, schemaComp, entitySvc, bus)
	
	// Set job client for request service if available
	if jobClient != nil {
		jobClientWrapper := service.NewAsynqJobClient(jobClient)
		requestSvc.SetJobClient(jobClientWrapper)
	}
	
	flowSvc := service.NewFlowService(dbPool.Queries, bus, requestSvc)
	
	// Recover flows on startup
	if err := flowSvc.RecoverFlows(context.Background(), logger); err != nil {
		logger.Warn("Failed to recover flows on startup", zap.Error(err))
	}
	
	cmdHandler := ws.NewCommandHandler(requestSvc, flowSvc, logger)
	hub.SetCommandHandler(cmdHandler)

	// HTTP router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	// Timeout middleware - skip for WebSocket upgrades
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Skip timeout for WebSocket upgrade requests
			if req.Header.Get("Upgrade") == "websocket" {
				next.ServeHTTP(w, req)
				return
			}
			middleware.Timeout(60 * time.Second)(next).ServeHTTP(w, req)
		})
	})

	// Mount API routes
	jobClientWrapper := service.NewAsynqJobClient(jobClient)
	r.Mount("/v1", api.Routes(api.Dependencies{
		DB:        dbPool,
		Bus:       bus,
		Hub:       hub,
		Log:       logger,
		JobClient: jobClientWrapper,
	}))

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server
	logger.Info("Starting server", zap.String("addr", addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}

// wsStreamsAdapter adapts pubsub.Streams to ws.StreamsProvider
type wsStreamsAdapter struct {
	streams *pubsub.Streams
}

func (a *wsStreamsAdapter) GetLastSequence(channel, connectionID string) (int64, error) {
	return a.streams.GetLastSequence(channel, connectionID)
}

func (a *wsStreamsAdapter) AcknowledgeSequence(channel, connectionID string, sequence int64) error {
	return a.streams.AcknowledgeSequence(channel, connectionID, sequence)
}

func (a *wsStreamsAdapter) ReplayEvents(channel string, sinceSeq int64, limit int64) ([]ws.StreamEvent, error) {
	events, err := a.streams.ReplayEvents(channel, sinceSeq, limit)
	if err != nil {
		return nil, err
	}
	
	// Convert pubsub.StreamEvent to ws.StreamEvent
	wsEvents := make([]ws.StreamEvent, len(events))
	for i, e := range events {
		wsEvents[i] = ws.StreamEvent{
			Channel:   e.Channel,
			Sequence:  e.Sequence,
			Event:     e.Event,
			Timestamp: e.Timestamp,
		}
	}
	
	return wsEvents, nil
}

