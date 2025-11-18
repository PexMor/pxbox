package api

import (
	"encoding/json"
	"net/http"

	"pxbox/internal/schema"
	"pxbox/internal/service"

	"github.com/go-chi/chi/v5"
)

type CreateFlowRequest struct {
	Kind        string                 `json:"kind"`
	OwnerEntity string                 `json:"ownerEntity"`
	Cursor      map[string]interface{} `json:"cursor,omitempty"`
}

func (d Dependencies) createFlow(w http.ResponseWriter, r *http.Request) {
	var req CreateFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", d.Log)
		return
	}

	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)
	flowSvc := service.NewFlowService(d.DB.Queries, d.Bus, requestSvc)

	flow, err := flowSvc.CreateFlow(r.Context(), service.CreateFlowInput{
		Kind:        req.Kind,
		OwnerEntity: req.OwnerEntity,
		Cursor:      req.Cursor,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "create_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(flow)
}

func (d Dependencies) getFlow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)
	flowSvc := service.NewFlowService(d.DB.Queries, d.Bus, requestSvc)

	flow, err := flowSvc.GetFlow(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "not_found", "Flow not found", d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flow)
}

type ResumeFlowRequest struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func (d Dependencies) resumeFlow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req ResumeFlowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", d.Log)
		return
	}

	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)
	flowSvc := service.NewFlowService(d.DB.Queries, d.Bus, requestSvc)

	if err := flowSvc.ResumeFlow(r.Context(), id, req.Event, req.Data); err != nil {
		WriteError(w, http.StatusInternalServerError, "resume_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "RUNNING"})
}

func (d Dependencies) cancelFlow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	schemaComp := schema.NewCompilerWithCache(64)
	entitySvc := service.NewEntityService(d.DB.Queries)
	requestSvc := service.NewRequestService(d.DB.Queries, schemaComp, entitySvc, d.Bus)
	flowSvc := service.NewFlowService(d.DB.Queries, d.Bus, requestSvc)

	if err := flowSvc.CancelFlow(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "cancel_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "CANCELLED"})
}

