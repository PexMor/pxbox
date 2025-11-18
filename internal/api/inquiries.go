package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"pxbox/internal/schema"
	"pxbox/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (d Dependencies) listInquiries(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("entityId")
	status := r.URL.Query().Get("status")
	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"
	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "" {
		sortBy = "created"
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	var entityIDPtr *string
	if entityID != "" {
		entityIDPtr = &entityID
	}
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	d.Log.Info("ListInquiries called", zap.String("entityID", entityID), zap.Bool("entityIDPtrNil", entityIDPtr == nil), zap.Any("statusPtr", statusPtr))
	if entityIDPtr != nil {
		d.Log.Info("EntityIDPtr value", zap.String("value", *entityIDPtr))
	}

	requests, err := d.DB.Queries.ListInquiries(r.Context(), entityIDPtr, statusPtr, includeDeleted, sortBy, limit, offset)
	if err != nil {
		d.Log.Error("Failed to list inquiries", zap.Error(err), zap.String("entityID", entityID), zap.Any("entityIDPtr", entityIDPtr))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	d.Log.Info("ListInquiries result", zap.Int("count", len(requests)), zap.String("entityID", entityID))

	result := make([]map[string]interface{}, 0)
	for _, req := range requests {
		result = append(result, map[string]interface{}{
			"id":         req.ID,
			"status":     req.Status,
			"createdBy":  req.CreatedBy,
			"entityId":   req.EntityID,
			"createdAt":  req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"deadlineAt": timePtrToString(req.DeadlineAt),
			"readAt":     timePtrToString(req.ReadAt),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": result,
		"total": len(result),
	})
}

func (d Dependencies) markRead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := d.DB.Queries.MarkInquiryRead(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "update_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "read"})
}

type SnoozeRequest struct {
	RemindAt time.Time `json:"remindAt"`
}

func (d Dependencies) snooze(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req SnoozeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", d.Log)
		return
	}

	// Get entity ID from auth context (TODO: implement auth)
	entityID := r.Header.Get("X-Entity-ID")
	if entityID == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", d.Log)
		return
	}

	// Create reminder
	_, err := d.DB.Queries.CreateReminder(r.Context(), id, entityID, req.RemindAt)
	if err != nil {
		d.Log.Error("Failed to create reminder", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "snoozed",
		"remindAt": req.RemindAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (d Dependencies) cancelInquiry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)
	if d.JobClient != nil {
		requestSvc.SetJobClient(d.JobClient)
	}

	if err := requestSvc.CancelRequest(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "cancel_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "CANCELLED"})
}

func (d Dependencies) deleteInquiry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := d.DB.Queries.SoftDeleteInquiry(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "delete_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

