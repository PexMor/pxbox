package service

import (
	"context"
	"fmt"
	"time"

	"pxbox/internal/db"
	"pxbox/internal/model"
	"pxbox/internal/schema"
	"pxbox/internal/storage"

	"github.com/oklog/ulid/v2"
)

type RequestService struct {
	queries      *db.Queries
	schemaComp   *schema.Compiler
	entitySvc    *EntityService
	bus          EventBus
	jobClient    JobClient
}

type EventBus interface {
	PublishEntity(entityID string, event map[string]interface{}) error
	PublishRequest(requestID string, event map[string]interface{}) error
	PublishRequestor(clientID string, event map[string]interface{}) error
}

func NewRequestService(queries *db.Queries, schemaComp *schema.Compiler, entitySvc *EntityService, bus EventBus) *RequestService {
	return &RequestService{
		queries:   queries,
		schemaComp: schemaComp,
		entitySvc: entitySvc,
		bus:      bus,
		jobClient: nil, // Will be set if job client is available
	}
}

// SetJobClient sets the job client for scheduling background jobs
func (s *RequestService) SetJobClient(client JobClient) {
	s.jobClient = client
}

type CreateRequestInput struct {
	Entity      struct {
		ID     string `json:"id"`
		Handle string `json:"handle"`
	} `json:"entity"`
	Schema      map[string]interface{} `json:"schema"`
	UIHints     map[string]interface{} `json:"uiHints,omitempty"`
	Prefill     map[string]interface{} `json:"prefill,omitempty"`
	ExpiresAt   *time.Time             `json:"expiresAt,omitempty"`
	DeadlineAt  *time.Time              `json:"deadlineAt,omitempty"`
	AttentionAt *time.Time              `json:"attentionAt,omitempty"`
	CallbackURL *string                 `json:"callbackUrl,omitempty"`
	FilesPolicy map[string]interface{}  `json:"filesPolicy,omitempty"`
	CreatedBy   string
}

func (s *RequestService) CreateRequest(ctx context.Context, input CreateRequestInput) (*model.Request, error) {
	// Resolve entity
	entity, err := s.entitySvc.ResolveEntity(ctx, input.Entity.ID, input.Entity.Handle)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve entity: %w", err)
	}

	// Detect schema kind
	schemaKind := detectSchemaKind(input.Schema)

	// Validate schema if it's a JSON Schema
	if schemaKind == model.SchemaKindJSON || schemaKind == model.SchemaKindRef {
		if err := s.schemaComp.Prepare(ctx, input.Schema); err != nil {
			return nil, fmt.Errorf("invalid schema: %w", err)
		}
	}

	// Generate request ID
	requestID := ulid.Make().String()

	// Set defaults
	if input.UIHints == nil {
		input.UIHints = make(map[string]interface{})
	}

	// Create request in database
	req, err := s.queries.CreateRequest(ctx, db.CreateRequestParams{
		ID:              requestID,
		CreatedBy:       input.CreatedBy,
		EntityID:        entity.ID,
		Status:          string(model.StatusPending),
		SchemaKind:      string(schemaKind),
		SchemaPayload:   input.Schema,
		UIHints:         input.UIHints,
		Prefill:         input.Prefill,
		ExpiresAt:       input.ExpiresAt,
		DeadlineAt:      input.DeadlineAt,
		AttentionAt:     input.AttentionAt,
		CallbackURL:     input.CallbackURL,
		FilesPolicy:     input.FilesPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Publish event
	_ = s.bus.PublishEntity(entity.ID, map[string]interface{}{
		"type":      "request.created",
		"requestId":  requestID,
		"entityId":   entity.ID,
	})

	_ = s.bus.PublishRequestor(input.CreatedBy, map[string]interface{}{
		"type":      "request.created",
		"requestId":  requestID,
	})

	// Schedule background jobs if job client is available
	if s.jobClient != nil {
		// Schedule deadline notification (1h before)
		if req.DeadlineAt != nil {
			_ = s.jobClient.ScheduleDeadlineNotification(requestID, *req.DeadlineAt)
			_ = s.jobClient.ScheduleDeadlineExpiry(requestID, *req.DeadlineAt)
		}

		// Schedule attention notification
		if req.AttentionAt != nil {
			_ = s.jobClient.ScheduleAttentionNotification(requestID, *req.AttentionAt)
		}

		// Schedule auto-cancel if grace period is set
		if req.AutocancelGrace != nil && *req.AutocancelGrace > 0 {
			// Auto-cancel after expiry + grace period
			if req.DeadlineAt != nil {
				cancelAt := req.DeadlineAt.Add(*req.AutocancelGrace)
				_ = s.jobClient.ScheduleAutoCancel(requestID, time.Until(cancelAt))
			}
		}
	}

	return dbRequestToModel(req), nil
}

func (s *RequestService) GetRequest(ctx context.Context, id string) (*model.Request, error) {
	req, err := s.queries.GetRequestByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("request not found: %w", err)
	}
	return dbRequestToModel(req), nil
}

func (s *RequestService) GetResponseByRequestID(ctx context.Context, requestID string) (*model.Response, error) {
	resp, err := s.queries.GetResponseByRequestID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("response not found: %w", err)
	}
	return dbResponseToModel(resp), nil
}

func (s *RequestService) ClaimRequest(ctx context.Context, id string) error {
	err := s.queries.ClaimRequest(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to claim request: %w", err)
	}

	req, _ := s.queries.GetRequestByID(ctx, id)
	_ = s.bus.PublishRequest(id, map[string]interface{}{
		"type": "request.claimed",
		"requestId": id,
	})

	_ = s.bus.PublishEntity(req.EntityID, map[string]interface{}{
		"type": "request.claimed",
		"requestId": id,
	})

	return nil
}

func (s *RequestService) PostResponse(ctx context.Context, requestID string, answeredBy string, payload map[string]interface{}, files []map[string]interface{}) (*model.Response, error) {
	// Get request
	req, err := s.queries.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("request not found: %w", err)
	}

	// If answeredBy is not provided or empty, use the request's entityId
	// (the entity the request was sent to should be the one responding)
	if answeredBy == "" {
		answeredBy = req.EntityID
	}

	// Validate that the answeredBy entity exists
	// Check via database query since EntityService doesn't expose GetEntity
	if _, err := s.queries.GetEntityByID(ctx, answeredBy); err != nil {
		return nil, fmt.Errorf("entity not found: %w", err)
	}

	// Validate payload against schema
	if req.SchemaKind == string(model.SchemaKindJSON) || req.SchemaKind == string(model.SchemaKindRef) {
		if err := s.schemaComp.Validate(ctx, req.SchemaKind, req.SchemaPayload, payload); err != nil {
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}
	}

	// Create response
	responseID := ulid.Make().String()
	// Ensure files is never nil (use empty slice instead)
	// pgx may encode nil slice as null, so we explicitly use empty slice
	filesParam := files
	if filesParam == nil || len(filesParam) == 0 {
		filesParam = []map[string]interface{}{}
	} else {
		// Normalize and validate file metadata
		normalized, err := storage.NormalizeFiles(filesParam)
		if err != nil {
			return nil, fmt.Errorf("invalid file metadata: %w", err)
		}
		filesParam = normalized
	}
	resp, err := s.queries.CreateResponse(ctx, db.CreateResponseParams{
		ID:         responseID,
		RequestID:  requestID,
		AnsweredBy: answeredBy,
		Payload:    payload,
		Files:      filesParam,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create response: %w", err)
	}

	// Update request status
	if err := s.queries.UpdateRequestStatus(ctx, requestID, string(model.StatusAnswered)); err != nil {
		return nil, fmt.Errorf("failed to update request status: %w", err)
	}

	// Publish events
	_ = s.bus.PublishRequest(requestID, map[string]interface{}{
		"type": "request.answered",
		"requestId": requestID,
	})

	_ = s.bus.PublishRequestor(req.CreatedBy, map[string]interface{}{
		"type":      "request.answered",
		"requestId":  requestID,
		"payload":    payload,
		"files":      files,
	})

	return dbResponseToModel(resp), nil
}

func (s *RequestService) CancelRequest(ctx context.Context, id string) error {
	if err := s.queries.UpdateRequestStatus(ctx, id, string(model.StatusCancelled)); err != nil {
		return fmt.Errorf("failed to cancel request: %w", err)
	}

	req, _ := s.queries.GetRequestByID(ctx, id)
	_ = s.bus.PublishRequest(id, map[string]interface{}{
		"type": "request.cancelled",
		"requestId": id,
	})

	_ = s.bus.PublishEntity(req.EntityID, map[string]interface{}{
		"type": "request.cancelled",
		"requestId": id,
	})

	return nil
}

func detectSchemaKind(schema map[string]interface{}) model.SchemaKind {
	if _, ok := schema["$ref"]; ok {
		return model.SchemaKindRef
	}
	if _, ok := schema["example"]; ok {
		return model.SchemaKindExample
	}
	return model.SchemaKindJSON
}

func dbRequestToModel(r db.Request) *model.Request {
	return &model.Request{
		ID:            r.ID,
		CreatedBy:     r.CreatedBy,
		EntityID:      r.EntityID,
		Status:        model.Status(r.Status),
		SchemaKind:    model.SchemaKind(r.SchemaKind),
		SchemaPayload: r.SchemaPayload,
		UIHints:       r.UIHints,
		Prefill:       r.Prefill,
		ExpiresAt:     timePtrToString(r.ExpiresAt),
		DeadlineAt:    timePtrToString(r.DeadlineAt),
		AttentionAt:   timePtrToString(r.AttentionAt),
		CallbackURL:   r.CallbackURL,
		FilesPolicy:   r.FilesPolicy,
		FlowID:        r.FlowID,
		CreatedAt:     r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func dbResponseToModel(r db.Response) *model.Response {
	return &model.Response{
		ID:          r.ID,
		RequestID:   r.RequestID,
		AnsweredBy:  r.AnsweredBy,
		Payload:     r.Payload,
		Files:       r.Files,
		AnsweredAt:  r.AnsweredAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func timePtrToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02T15:04:05Z07:00")
	return &s
}

