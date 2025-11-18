package jobs

import (
	"context"
	"fmt"
	"time"

	"pxbox/internal/db"
	"pxbox/internal/pubsub"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type JobServer struct {
	server *asynq.Server
	client *asynq.Client
	db     *db.Pool
	bus    *pubsub.Bus
	log    *zap.Logger
}

func NewJobServer(redisAddr string, dbPool *db.Pool, bus *pubsub.Bus, log *zap.Logger) (*JobServer, *asynq.Client) {
	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}
	
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":       1,
			},
		},
	)

	client := asynq.NewClient(redisOpt)

	return &JobServer{
		server: server,
		client: client,
		db:     dbPool,
		bus:    bus,
		log:    log,
	}, client
}

func (js *JobServer) Start() error {
	mux := asynq.NewServeMux()
	
	// Register job handlers
	mux.HandleFunc("deadline:notify", js.handleDeadlineNotification)
	mux.HandleFunc("deadline:expire", js.handleDeadlineExpiry)
	mux.HandleFunc("request:autocancel", js.handleAutoCancel)
	mux.HandleFunc("request:attention", js.handleAttentionNotification)
	mux.HandleFunc("reminder:snooze", js.handleReminder)

	return js.server.Start(mux)
}

func (js *JobServer) Stop() {
	js.server.Shutdown()
	js.client.Close()
}

// Job handlers

func (js *JobServer) handleDeadlineNotification(ctx context.Context, t *asynq.Task) error {
	requestID := string(t.Payload())
	
	req, err := js.db.Queries.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	// Only notify if still pending
	if req.Status != "PENDING" {
		return nil
	}

	// Publish notification event
	_ = js.bus.PublishEntity(req.EntityID, map[string]interface{}{
		"type":      "request.deadline_approaching",
		"requestId": requestID,
		"deadlineAt": req.DeadlineAt.Format(time.RFC3339),
	})

	js.log.Info("Deadline notification sent", zap.String("request_id", requestID))
	return nil
}

func (js *JobServer) handleDeadlineExpiry(ctx context.Context, t *asynq.Task) error {
	requestID := string(t.Payload())
	
	req, err := js.db.Queries.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	// Only expire if still pending
	if req.Status != "PENDING" {
		return nil
	}

	// Update status to EXPIRED
	if err := js.db.Queries.UpdateRequestStatus(ctx, requestID, "EXPIRED"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Publish expiry event
	_ = js.bus.PublishEntity(req.EntityID, map[string]interface{}{
		"type":      "request.expired",
		"requestId": requestID,
	})

	js.log.Info("Request expired", zap.String("request_id", requestID))
	return nil
}

func (js *JobServer) handleAutoCancel(ctx context.Context, t *asynq.Task) error {
	requestID := string(t.Payload())
	
	req, err := js.db.Queries.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	// Only auto-cancel if still pending
	if req.Status != "PENDING" {
		return nil
	}

	// Cancel the request directly via database
	if err := js.db.Queries.UpdateRequestStatus(ctx, requestID, "CANCELLED"); err != nil {
		return fmt.Errorf("failed to cancel request: %w", err)
	}

	// Publish cancellation event
	_ = js.bus.PublishRequest(requestID, map[string]interface{}{
		"type": "request.cancelled",
		"requestId": requestID,
	})

	_ = js.bus.PublishEntity(req.EntityID, map[string]interface{}{
		"type": "request.cancelled",
		"requestId": requestID,
	})

	js.log.Info("Request auto-cancelled", zap.String("request_id", requestID))
	return nil
}

func (js *JobServer) handleAttentionNotification(ctx context.Context, t *asynq.Task) error {
	requestID := string(t.Payload())
	
	req, err := js.db.Queries.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	// Only notify if still pending
	if req.Status != "PENDING" {
		return nil
	}

	// Publish attention notification event
	_ = js.bus.PublishEntity(req.EntityID, map[string]interface{}{
		"type":      "request.needs_attention",
		"requestId": requestID,
		"attentionAt": req.AttentionAt.Format(time.RFC3339),
	})

	js.log.Info("Attention notification sent", zap.String("request_id", requestID))
	return nil
}

func (js *JobServer) handleReminder(ctx context.Context, t *asynq.Task) error {
	reminderID := string(t.Payload())
	
	// Get reminder details
	reminder, err := js.db.Queries.GetReminderByID(ctx, reminderID)
	if err != nil {
		return fmt.Errorf("failed to get reminder: %w", err)
	}

	// Publish reminder event
	_ = js.bus.PublishEntity(reminder.EntityID, map[string]interface{}{
		"type":      "request.reminder",
		"requestId": reminder.RequestID,
		"reminderId": reminderID,
	})

	js.log.Info("Reminder sent", zap.String("reminder_id", reminderID), zap.String("request_id", reminder.RequestID))
	return nil
}

// Schedule jobs

func ScheduleDeadlineNotification(client *asynq.Client, requestID string, deadlineAt time.Time) error {
	// Schedule notification 1 hour before deadline
	notifyAt := deadlineAt.Add(-1 * time.Hour)
	if notifyAt.Before(time.Now()) {
		return nil // Already past notification time
	}

	task := asynq.NewTask("deadline:notify", []byte(requestID))
	_, err := client.Enqueue(task, asynq.ProcessIn(time.Until(notifyAt)))
	return err
}

func ScheduleDeadlineExpiry(client *asynq.Client, requestID string, deadlineAt time.Time) error {
	if deadlineAt.Before(time.Now()) {
		return nil // Already expired
	}

	task := asynq.NewTask("deadline:expire", []byte(requestID))
	_, err := client.Enqueue(task, asynq.ProcessIn(time.Until(deadlineAt)))
	return err
}

func ScheduleAutoCancel(client *asynq.Client, requestID string, gracePeriod time.Duration) error {
	task := asynq.NewTask("request:autocancel", []byte(requestID))
	_, err := client.Enqueue(task, asynq.ProcessIn(gracePeriod))
	return err
}

func ScheduleAttentionNotification(client *asynq.Client, requestID string, attentionAt time.Time) error {
	if attentionAt.Before(time.Now()) {
		return nil // Already past attention time
	}

	task := asynq.NewTask("request:attention", []byte(requestID))
	_, err := client.Enqueue(task, asynq.ProcessIn(time.Until(attentionAt)))
	return err
}

func ScheduleReminder(client *asynq.Client, reminderID string, remindAt time.Time) error {
	if remindAt.Before(time.Now()) {
		return nil // Already past reminder time
	}

	task := asynq.NewTask("reminder:snooze", []byte(reminderID))
	_, err := client.Enqueue(task, asynq.ProcessIn(time.Until(remindAt)))
	return err
}

