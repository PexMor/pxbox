package service

import (
	"context"
	"fmt"

	"pxbox/internal/db"
	"pxbox/internal/model"
)

type FlowService struct {
	queries    *db.Queries
	bus        EventBus
	requestSvc *RequestService
	runner     FlowRunner // Flow runner for executing flow steps
}

func NewFlowService(queries *db.Queries, bus EventBus, requestSvc *RequestService) *FlowService {
	fs := &FlowService{
		queries:    queries,
		bus:        bus,
		requestSvc: requestSvc,
	}
	// Set default basic runner
	fs.runner = NewBasicFlowRunner(requestSvc, fs)
	return fs
}

// SetRunner sets a custom flow runner
func (s *FlowService) SetRunner(runner FlowRunner) {
	s.runner = runner
}

type CreateFlowInput struct {
	Kind        string
	OwnerEntity string
	Cursor      map[string]interface{}
}

func (s *FlowService) CreateFlow(ctx context.Context, input CreateFlowInput) (*model.Flow, error) {
	if input.Cursor == nil {
		input.Cursor = make(map[string]interface{})
	}

	flow, err := s.queries.CreateFlow(ctx, db.CreateFlowParams{
		Kind:        input.Kind,
		OwnerEntity: input.OwnerEntity,
		Status:      string(model.FlowStatusRunning),
		Cursor:      input.Cursor,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create flow: %w", err)
	}

	_ = s.bus.PublishEntity(input.OwnerEntity, map[string]interface{}{
		"type":   "flow.created",
		"flowId": flow.ID,
	})

	return dbFlowToModel(flow), nil
}

func (s *FlowService) GetFlow(ctx context.Context, id string) (*model.Flow, error) {
	flow, err := s.queries.GetFlowByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("flow not found: %w", err)
	}
	return dbFlowToModel(flow), nil
}

func (s *FlowService) ResumeFlow(ctx context.Context, flowID string, event string, data map[string]interface{}) error {
	flow, err := s.queries.GetFlowByID(ctx, flowID)
	if err != nil {
		return fmt.Errorf("flow not found: %w", err)
	}

	// Update cursor with event data
	if flow.Cursor == nil {
		flow.Cursor = make(map[string]interface{})
	}
	flow.Cursor["lastEvent"] = map[string]interface{}{
		"type": event,
		"data": data,
	}

	// Update flow status and cursor
	if err := s.queries.UpdateFlowCursor(ctx, flowID, flow.Cursor); err != nil {
		return fmt.Errorf("failed to update cursor: %w", err)
	}

	if err := s.queries.UpdateFlowStatus(ctx, flowID, string(model.FlowStatusRunning)); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Execute flow step if runner is available
	if s.runner != nil {
		flowModel := dbFlowToModel(flow)
		result := s.runner.Run(ctx, flowModel)
		
		// Update cursor with result
		if result.Cursor != nil {
			if err := s.queries.UpdateFlowCursor(ctx, flowID, result.Cursor); err != nil {
				return fmt.Errorf("failed to update cursor after step: %w", err)
			}
		}

		// Handle suspend
		if result.Suspend != nil {
			if err := s.queries.UpdateFlowStatus(ctx, flowID, string(model.FlowStatusSuspended)); err != nil {
				return fmt.Errorf("failed to suspend flow: %w", err)
			}
			_ = s.bus.PublishEntity(flow.OwnerEntity, map[string]interface{}{
				"type":   "flow.suspended",
				"flowId": flowID,
			})
			return nil
		}

		// Handle completion
		if result.Done {
			if err := s.queries.UpdateFlowStatus(ctx, flowID, string(model.FlowStatusCompleted)); err != nil {
				return fmt.Errorf("failed to complete flow: %w", err)
			}
			_ = s.bus.PublishEntity(flow.OwnerEntity, map[string]interface{}{
				"type":   "flow.completed",
				"flowId": flowID,
			})
			return nil
		}

		// Handle error
		if result.Err != nil {
			if err := s.queries.UpdateFlowStatus(ctx, flowID, string(model.FlowStatusFailed)); err != nil {
				return fmt.Errorf("failed to mark flow as failed: %w", err)
			}
			_ = s.bus.PublishEntity(flow.OwnerEntity, map[string]interface{}{
				"type":   "flow.failed",
				"flowId": flowID,
				"error":  result.Err.Error(),
			})
			return result.Err
		}
	}

	_ = s.bus.PublishEntity(flow.OwnerEntity, map[string]interface{}{
		"type":   "flow.updated",
		"flowId": flowID,
		"status": "RUNNING",
	})

	return nil
}

// TickFlow executes a flow step (called by scheduler or recovery)
func (s *FlowService) TickFlow(ctx context.Context, flowID string) error {
	flow, err := s.queries.GetFlowByID(ctx, flowID)
	if err != nil {
		return fmt.Errorf("flow not found: %w", err)
	}

	if flow.Status != string(model.FlowStatusRunning) && flow.Status != string(model.FlowStatusSuspended) {
		return nil // Only process running or suspended flows
	}

	flowModel := dbFlowToModel(flow)
	if s.runner == nil {
		return fmt.Errorf("flow runner not set")
	}

	result := s.runner.Run(ctx, flowModel)

	// Update cursor
	if result.Cursor != nil {
		if err := s.queries.UpdateFlowCursor(ctx, flowID, result.Cursor); err != nil {
			return fmt.Errorf("failed to update cursor: %w", err)
		}
	}

	// Handle suspend
	if result.Suspend != nil {
		if err := s.queries.UpdateFlowStatus(ctx, flowID, string(model.FlowStatusSuspended)); err != nil {
			return fmt.Errorf("failed to suspend flow: %w", err)
		}
		_ = s.bus.PublishEntity(flow.OwnerEntity, map[string]interface{}{
			"type":   "flow.suspended",
			"flowId": flowID,
		})
		return nil
	}

	// Handle completion
	if result.Done {
		if err := s.queries.UpdateFlowStatus(ctx, flowID, string(model.FlowStatusCompleted)); err != nil {
			return fmt.Errorf("failed to complete flow: %w", err)
		}
		_ = s.bus.PublishEntity(flow.OwnerEntity, map[string]interface{}{
			"type":   "flow.completed",
			"flowId": flowID,
		})
		return nil
	}

	// Handle error
	if result.Err != nil {
		if err := s.queries.UpdateFlowStatus(ctx, flowID, string(model.FlowStatusFailed)); err != nil {
			return fmt.Errorf("failed to mark flow as failed: %w", err)
		}
		_ = s.bus.PublishEntity(flow.OwnerEntity, map[string]interface{}{
			"type":   "flow.failed",
			"flowId": flowID,
			"error":  result.Err.Error(),
		})
		return result.Err
	}

	return nil
}

func (s *FlowService) CancelFlow(ctx context.Context, flowID string) error {
	flow, err := s.queries.GetFlowByID(ctx, flowID)
	if err != nil {
		return fmt.Errorf("flow not found: %w", err)
	}

	if err := s.queries.UpdateFlowStatus(ctx, flowID, string(model.FlowStatusCancelled)); err != nil {
		return fmt.Errorf("failed to cancel flow: %w", err)
	}

	// Cancel all open inquiries for this flow
	// TODO: Implement query to get requests by flow_id and cancel them

	_ = s.bus.PublishEntity(flow.OwnerEntity, map[string]interface{}{
		"type":   "flow.updated",
		"flowId": flowID,
		"status": "CANCELLED",
	})

	return nil
}

func (s *FlowService) UpdateFlowCursor(ctx context.Context, flowID string, cursor map[string]interface{}) error {
	return s.queries.UpdateFlowCursor(ctx, flowID, cursor)
}

func dbFlowToModel(f db.Flow) *model.Flow {
	return &model.Flow{
		ID:          f.ID,
		Kind:        f.Kind,
		OwnerEntity: f.OwnerEntity,
		Status:      model.FlowStatus(f.Status),
		Cursor:      f.Cursor,
		LastEventID: f.LastEventID,
		CreatedAt:   f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

