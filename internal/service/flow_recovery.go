package service

import (
	"context"
	"fmt"

	"pxbox/internal/model"
	"go.uber.org/zap"
)

// RecoverFlows recovers suspended/running flows on application start
func (s *FlowService) RecoverFlows(ctx context.Context, log *zap.Logger) error {
	// Get all running and suspended flows
	flows, err := s.queries.GetFlowsByStatus(ctx, []string{
		string(model.FlowStatusRunning),
		string(model.FlowStatusSuspended),
	})
	if err != nil {
		return fmt.Errorf("failed to get flows: %w", err)
	}

	log.Info("Recovering flows", zap.Int("count", len(flows)))

	for _, flow := range flows {
		flowModel := dbFlowToModel(flow)
		
		// Check if flow is waiting for a request that has been answered
		if flowModel.Status == model.FlowStatusSuspended {
			// Check pending requests in cursor
			pending, _ := flowModel.Cursor["pending"].([]interface{})
			if pending != nil {
				allAnswered := true
				for _, p := range pending {
					reqData, _ := p.(map[string]interface{})
					requestID, _ := reqData["requestId"].(string)
					if requestID == "" {
						continue
					}

					// Check request status
					req, err := s.requestSvc.GetRequest(ctx, requestID)
					if err != nil {
						log.Warn("Failed to get request during recovery",
							zap.String("flowId", flowModel.ID),
							zap.String("requestId", requestID),
							zap.Error(err),
						)
						allAnswered = false
						continue
					}

					if req.Status == model.StatusAnswered {
						// Request was answered, resume flow with response data
						// Get response
						// Note: We'd need to get the response, but for now we'll just resume
						// The actual response data should be in the lastEvent
						if err := s.ResumeFlow(ctx, flowModel.ID, "request.answered", map[string]interface{}{
							"requestId": requestID,
						}); err != nil {
							log.Error("Failed to resume flow after recovery",
								zap.String("flowId", flowModel.ID),
								zap.String("requestId", requestID),
								zap.Error(err),
							)
						} else {
							log.Info("Resumed flow after recovery",
								zap.String("flowId", flowModel.ID),
								zap.String("requestId", requestID),
							)
						}
						allAnswered = false // Don't process other requests
						break
					} else if req.Status == model.StatusPending || req.Status == model.StatusClaimed {
						allAnswered = false
					}
				}

				if allAnswered {
					// All pending requests are answered or cancelled
					// Try to tick the flow
					if err := s.TickFlow(ctx, flowModel.ID); err != nil {
						log.Error("Failed to tick flow during recovery",
							zap.String("flowId", flowModel.ID),
							zap.Error(err),
						)
					}
				}
			}
		} else if flowModel.Status == model.FlowStatusRunning {
			// Try to tick running flows
			if err := s.TickFlow(ctx, flowModel.ID); err != nil {
				log.Error("Failed to tick flow during recovery",
					zap.String("flowId", flowModel.ID),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

