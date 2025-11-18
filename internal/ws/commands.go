package ws

import (
	"context"
	"encoding/json"
	"time"

	"pxbox/internal/service"

	"go.uber.org/zap"
)

// CommandHandler handles WebSocket commands
type CommandHandler struct {
	requestSvc *service.RequestService
	flowSvc    *service.FlowService
	log        *zap.Logger
}

func NewCommandHandler(requestSvc *service.RequestService, flowSvc *service.FlowService, log *zap.Logger) *CommandHandler {
	return &CommandHandler{
		requestSvc: requestSvc,
		flowSvc:    flowSvc,
		log:        log,
	}
}

// HandleCommand processes a WebSocket command
func (h *CommandHandler) HandleCommand(ctx context.Context, conn *Conn, cmd map[string]interface{}) {
	op, _ := cmd["op"].(string)
	data, _ := cmd["data"].(map[string]interface{})
	msgID, _ := cmd["id"].(string)

	switch op {
	case "createRequest":
		h.handleCreateRequest(ctx, conn, msgID, data)
	case "getRequest":
		h.handleGetRequest(ctx, conn, msgID, data)
	case "claimRequest":
		h.handleClaimRequest(ctx, conn, msgID, data)
	case "postResponse":
		h.handlePostResponse(ctx, conn, msgID, data)
	case "cancelRequest":
		h.handleCancelRequest(ctx, conn, msgID, data)
	case "createFlow":
		h.handleCreateFlow(ctx, conn, msgID, data)
	case "resumeFlow":
		h.handleResumeFlow(ctx, conn, msgID, data)
	case "cancelFlow":
		h.handleCancelFlow(ctx, conn, msgID, data)
	default:
		h.sendError(conn, msgID, "unknown_command", "Unknown command: "+op)
	}
}

func (h *CommandHandler) handleCreateRequest(ctx context.Context, conn *Conn, msgID string, data map[string]interface{}) {
	// Parse entity
	entityData, _ := data["entity"].(map[string]interface{})
	if entityData == nil {
		h.sendError(conn, msgID, "invalid_input", "entity required")
		return
	}

	// Parse schema
	schema, _ := data["schema"].(map[string]interface{})
	if schema == nil {
		h.sendError(conn, msgID, "invalid_input", "schema required")
		return
	}

	// Build CreateRequestInput
	input := service.CreateRequestInput{
		Schema:    schema,
		CreatedBy: conn.userID, // Use connection's user ID
	}

	// Parse entity ID/handle
	if entityID, ok := entityData["id"].(string); ok {
		input.Entity.ID = entityID
	}
	if handle, ok := entityData["handle"].(string); ok {
		input.Entity.Handle = handle
	}
	if input.Entity.ID == "" && input.Entity.Handle == "" {
		h.sendError(conn, msgID, "invalid_input", "entity.id or entity.handle required")
		return
	}

	// Parse optional fields
	if uiHints, ok := data["uiHints"].(map[string]interface{}); ok {
		input.UIHints = uiHints
	}
	if prefill, ok := data["prefill"].(map[string]interface{}); ok {
		input.Prefill = prefill
	}
	if callbackURL, ok := data["callbackUrl"].(string); ok {
		input.CallbackURL = &callbackURL
	}
	if filesPolicy, ok := data["filesPolicy"].(map[string]interface{}); ok {
		input.FilesPolicy = filesPolicy
	}

	// Parse time fields
	if expiresAtStr, ok := data["expiresAt"].(string); ok && expiresAtStr != "" {
		if t, err := time.Parse(time.RFC3339, expiresAtStr); err == nil {
			input.ExpiresAt = &t
		}
	}
	if deadlineAtStr, ok := data["deadlineAt"].(string); ok && deadlineAtStr != "" {
		if t, err := time.Parse(time.RFC3339, deadlineAtStr); err == nil {
			input.DeadlineAt = &t
		}
	}
	if attentionAtStr, ok := data["attentionAt"].(string); ok && attentionAtStr != "" {
		if t, err := time.Parse(time.RFC3339, attentionAtStr); err == nil {
			input.AttentionAt = &t
		}
	}

	// Create request
	req, err := h.requestSvc.CreateRequest(ctx, input)
	if err != nil {
		h.sendError(conn, msgID, "create_failed", err.Error())
		return
	}

	h.sendResponse(conn, msgID, map[string]interface{}{
		"type": "response",
		"data": map[string]interface{}{
			"requestId": req.ID,
			"status":    req.Status,
			"entityId":  req.EntityID,
		},
	})
}

func (h *CommandHandler) handleGetRequest(ctx context.Context, conn *Conn, msgID string, data map[string]interface{}) {
	requestID, _ := data["requestId"].(string)
	if requestID == "" {
		h.sendError(conn, msgID, "invalid_input", "requestId required")
		return
	}

	req, err := h.requestSvc.GetRequest(ctx, requestID)
	if err != nil {
		h.sendError(conn, msgID, "not_found", err.Error())
		return
	}

	h.sendResponse(conn, msgID, map[string]interface{}{
		"type": "response",
		"data": req,
	})
}

func (h *CommandHandler) handleClaimRequest(ctx context.Context, conn *Conn, msgID string, data map[string]interface{}) {
	requestID, _ := data["requestId"].(string)
	if requestID == "" {
		h.sendError(conn, msgID, "invalid_input", "requestId required")
		return
	}

	if err := h.requestSvc.ClaimRequest(ctx, requestID); err != nil {
		h.sendError(conn, msgID, "claim_failed", err.Error())
		return
	}

	h.sendResponse(conn, msgID, map[string]interface{}{
		"type": "response",
		"data": map[string]string{"status": "CLAIMED"},
	})
}

func (h *CommandHandler) handlePostResponse(ctx context.Context, conn *Conn, msgID string, data map[string]interface{}) {
	requestID, _ := data["requestId"].(string)
	payload, _ := data["payload"].(map[string]interface{})
	files, _ := data["files"].([]interface{})

	if requestID == "" || payload == nil {
		h.sendError(conn, msgID, "invalid_input", "requestId and payload required")
		return
	}

	// Convert files to []map[string]interface{}
	var filesList []map[string]interface{}
	for _, f := range files {
		if fm, ok := f.(map[string]interface{}); ok {
			filesList = append(filesList, fm)
		}
	}

	// TODO: Get answeredBy from connection context
	answeredBy := conn.userID
	resp, err := h.requestSvc.PostResponse(ctx, requestID, answeredBy, payload, filesList)
	if err != nil {
		h.sendError(conn, msgID, "validation_failed", err.Error())
		return
	}

	h.sendResponse(conn, msgID, map[string]interface{}{
		"type": "response",
		"data": map[string]interface{}{
			"responseId": resp.ID,
			"status":     "ANSWERED",
		},
	})
}

func (h *CommandHandler) handleCancelRequest(ctx context.Context, conn *Conn, msgID string, data map[string]interface{}) {
	requestID, _ := data["requestId"].(string)
	if requestID == "" {
		h.sendError(conn, msgID, "invalid_input", "requestId required")
		return
	}

	if err := h.requestSvc.CancelRequest(ctx, requestID); err != nil {
		h.sendError(conn, msgID, "cancel_failed", err.Error())
		return
	}

	h.sendResponse(conn, msgID, map[string]interface{}{
		"type": "response",
		"data": map[string]string{"status": "CANCELLED"},
	})
}

func (h *CommandHandler) handleCreateFlow(ctx context.Context, conn *Conn, msgID string, data map[string]interface{}) {
	kind, _ := data["kind"].(string)
	ownerEntity, _ := data["ownerEntity"].(string)
	cursor, _ := data["cursor"].(map[string]interface{})

	if kind == "" || ownerEntity == "" {
		h.sendError(conn, msgID, "invalid_input", "kind and ownerEntity required")
		return
	}

	flow, err := h.flowSvc.CreateFlow(ctx, service.CreateFlowInput{
		Kind:        kind,
		OwnerEntity: ownerEntity,
		Cursor:      cursor,
	})
	if err != nil {
		h.sendError(conn, msgID, "create_failed", err.Error())
		return
	}

	h.sendResponse(conn, msgID, map[string]interface{}{
		"type": "response",
		"data": flow,
	})
}

func (h *CommandHandler) handleResumeFlow(ctx context.Context, conn *Conn, msgID string, data map[string]interface{}) {
	flowID, _ := data["flowId"].(string)
	event, _ := data["event"].(string)
	eventData, _ := data["data"].(map[string]interface{})

	if flowID == "" || event == "" {
		h.sendError(conn, msgID, "invalid_input", "flowId and event required")
		return
	}

	if err := h.flowSvc.ResumeFlow(ctx, flowID, event, eventData); err != nil {
		h.sendError(conn, msgID, "resume_failed", err.Error())
		return
	}

	h.sendResponse(conn, msgID, map[string]interface{}{
		"type": "response",
		"data": map[string]string{"status": "RUNNING"},
	})
}

func (h *CommandHandler) handleCancelFlow(ctx context.Context, conn *Conn, msgID string, data map[string]interface{}) {
	flowID, _ := data["flowId"].(string)
	if flowID == "" {
		h.sendError(conn, msgID, "invalid_input", "flowId required")
		return
	}

	if err := h.flowSvc.CancelFlow(ctx, flowID); err != nil {
		h.sendError(conn, msgID, "cancel_failed", err.Error())
		return
	}

	h.sendResponse(conn, msgID, map[string]interface{}{
		"type": "response",
		"data": map[string]string{"status": "CANCELLED"},
	})
}

func (h *CommandHandler) sendResponse(conn *Conn, msgID string, response map[string]interface{}) {
	if msgID != "" {
		response["id"] = msgID
	}
	msg, _ := json.Marshal(response)
	select {
	case conn.send <- msg:
	default:
		h.log.Warn("Failed to send response, channel full")
	}
}

func (h *CommandHandler) sendError(conn *Conn, msgID, code, message string) {
	err := map[string]interface{}{
		"type":    "error",
		"code":    code,
		"message": message,
	}
	if msgID != "" {
		err["id"] = msgID
	}
	msg, _ := json.Marshal(err)
	select {
	case conn.send <- msg:
	default:
		h.log.Warn("Failed to send error, channel full")
	}
}

