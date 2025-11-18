package service

import (
	"context"
	"fmt"
	"time"

	"pxbox/internal/model"
)

// StepResult represents the result of executing a flow step
type StepResult struct {
	Cursor  map[string]interface{} `json:"cursor"`
	Suspend *Suspend               `json:"suspend,omitempty"`
	Done    bool                   `json:"done"`
	Err     error                  `json:"error,omitempty"`
}

// Suspend represents a flow suspension point
type Suspend struct {
	Event      string     `json:"event"`       // Event type to wait for (e.g., "request.answered")
	RequestID  *string    `json:"requestId,omitempty"` // Specific request to wait for
	DeadlineAt *time.Time `json:"deadlineAt,omitempty"` // Optional deadline
	OnTimeout  string     `json:"onTimeout,omitempty"`  // Label/branch for timeout handling
}

// FlowRunner defines the interface for executing flow steps
type FlowRunner interface {
	// Run executes a flow step and returns the result
	Run(ctx context.Context, flow *model.Flow) StepResult
}

// BasicFlowRunner provides a basic implementation of FlowRunner
type BasicFlowRunner struct {
	requestSvc *RequestService
	flowSvc    *FlowService
}

// NewBasicFlowRunner creates a new basic flow runner
func NewBasicFlowRunner(requestSvc *RequestService, flowSvc *FlowService) *BasicFlowRunner {
	return &BasicFlowRunner{
		requestSvc: requestSvc,
		flowSvc:    flowSvc,
	}
}

// Run executes a flow step based on the cursor state
func (r *BasicFlowRunner) Run(ctx context.Context, flow *model.Flow) StepResult {
	// Extract step from cursor
	step, _ := flow.Cursor["step"].(string)
	if step == "" {
		// Initialize flow if no step set
		flow.Cursor["step"] = "init"
		return StepResult{
			Cursor: flow.Cursor,
			Done:   false,
		}
	}

	// Basic step execution logic
	// This is a minimal implementation - actual flow logic should be implemented
	// by flow-specific runners that embed BasicFlowRunner
	switch step {
	case "init":
		// Initial step - flow can proceed or suspend
		return StepResult{
			Cursor: flow.Cursor,
			Done:   false,
		}
	case "complete":
		// Flow is complete
		return StepResult{
			Cursor: flow.Cursor,
			Done:   true,
		}
	default:
		// Unknown step - continue with current cursor
		return StepResult{
			Cursor: flow.Cursor,
			Done:   false,
		}
	}
}

// AwaitInput creates a request and suspends the flow until it's answered
func (r *BasicFlowRunner) AwaitInput(ctx context.Context, flow *model.Flow, input CreateRequestInput) (*model.Request, *Suspend, error) {
	// Note: FlowID is set when creating the request via the service

	// Create request
	req, err := r.requestSvc.CreateRequest(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Create suspend point
	suspend := &Suspend{
		Event:     "request.answered",
		RequestID: &req.ID,
	}

	// Set deadline if provided
	if input.DeadlineAt != nil {
		suspend.DeadlineAt = input.DeadlineAt
		suspend.OnTimeout = "timeout"
	}

	// Update cursor with pending request
	if flow.Cursor == nil {
		flow.Cursor = make(map[string]interface{})
	}
	pending, _ := flow.Cursor["pending"].([]interface{})
	if pending == nil {
		pending = []interface{}{}
	}
	pending = append(pending, map[string]interface{}{
		"requestId": req.ID,
		"type":      "input",
		"status":    "PENDING",
	})
	flow.Cursor["pending"] = pending

	return req, suspend, nil
}

// GetLastEvent extracts the last event from the cursor
func GetLastEvent(cursor map[string]interface{}) map[string]interface{} {
	if cursor == nil {
		return nil
	}
	lastEvent, _ := cursor["lastEvent"].(map[string]interface{})
	return lastEvent
}

// IsEventType checks if the last event matches the expected type
func IsEventType(cursor map[string]interface{}, eventType string) bool {
	lastEvent := GetLastEvent(cursor)
	if lastEvent == nil {
		return false
	}
	evType, _ := lastEvent["type"].(string)
	return evType == eventType
}

// GetEventData extracts data from the last event
func GetEventData(cursor map[string]interface{}) map[string]interface{} {
	lastEvent := GetLastEvent(cursor)
	if lastEvent == nil {
		return nil
	}
	data, _ := lastEvent["data"].(map[string]interface{})
	return data
}

