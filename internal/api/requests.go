package api

import (
	"encoding/json"
	"net/http"
	"time"

	"pxbox/internal/schema"
	"pxbox/internal/service"

	"github.com/go-chi/chi/v5"
)

type CreateRequestRequest struct {
	Entity struct {
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
}

func (d Dependencies) createRequest(w http.ResponseWriter, r *http.Request) {
	var req CreateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", d.Log)
		return
	}

	// Get created_by from auth context (TODO: implement auth)
	createdBy := r.Header.Get("X-Client-ID")
	if createdBy == "" {
		createdBy = "anonymous"
	}

	// Initialize services
	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)
	if d.JobClient != nil {
		requestSvc.SetJobClient(d.JobClient)
	}

	// Create request
	result, err := requestSvc.CreateRequest(r.Context(), service.CreateRequestInput{
		Entity:      req.Entity,
		Schema:      req.Schema,
		UIHints:     req.UIHints,
		Prefill:     req.Prefill,
		ExpiresAt:   req.ExpiresAt,
		DeadlineAt:  req.DeadlineAt,
		AttentionAt: req.AttentionAt,
		CallbackURL: req.CallbackURL,
		FilesPolicy: req.FilesPolicy,
		CreatedBy:   createdBy,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "create_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requestId": result.ID,
		"status":    result.Status,
	})
}

func (d Dependencies) getRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)

	req, err := requestSvc.GetRequest(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not_found", "Request not found", d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

func (d Dependencies) cancelRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)

	if err := requestSvc.CancelRequest(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "CANCELLED"})
}

func (d Dependencies) claimRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)

	if err := requestSvc.ClaimRequest(r.Context(), id); err != nil {
		WriteError(w, http.StatusConflict, "claim_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "CLAIMED"})
}

func (d Dependencies) postResponse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var body struct {
		Payload map[string]interface{} `json:"payload"`
		Files   []map[string]interface{} `json:"files,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", d.Log)
		return
	}

	// Get answered_by from auth context (TODO: implement auth)
	answeredBy := r.Header.Get("X-Entity-ID")
	if answeredBy == "" {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized", d.Log)
		return
	}

	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)

	resp, err := requestSvc.PostResponse(r.Context(), id, answeredBy, body.Payload, body.Files)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"responseId": resp.ID,
		"status":     "ANSWERED",
	})
}

func (d Dependencies) getResponse(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "id")
	
	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)

	resp, err := requestSvc.GetResponseByRequestID(r.Context(), requestID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not_found", "Response not found", d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (d Dependencies) entityQueue(w http.ResponseWriter, r *http.Request) {
	entityID := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")
	
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	limit := 50
	offset := 0
	// TODO: Parse limit and offset from query params

	requests, err := d.DB.Queries.GetEntityQueue(r.Context(), entityID, statusPtr, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query_failed", err.Error(), d.Log)
		return
	}

	// Convert to model
	var result []map[string]interface{}
	for _, req := range requests {
		result = append(result, map[string]interface{}{
			"id":         req.ID,
			"status":     req.Status,
			"createdAt":  req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			"deadlineAt": timePtrToString(req.DeadlineAt),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": result,
	})
}

func timePtrToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02T15:04:05Z07:00")
	return &s
}

