package api

import (
	"encoding/json"
	"net/http"

	"pxbox/internal/model"
	"pxbox/internal/service"

	"github.com/go-chi/chi/v5"
)

type CreateEntityRequest struct {
	Kind   string                 `json:"kind"`
	Handle string                 `json:"handle"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

func (d Dependencies) createEntity(w http.ResponseWriter, r *http.Request) {
	var req CreateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", d.Log)
		return
	}

	// Validate kind
	kind := model.EntityKind(req.Kind)
	if kind != model.EntityKindUser && kind != model.EntityKindGroup && 
		kind != model.EntityKindRole && kind != model.EntityKindBot {
		WriteError(w, http.StatusBadRequest, "invalid_kind", "Invalid entity kind. Must be: user, group, role, or bot", d.Log)
		return
	}

	entitySvc := service.NewEntityService(d.DB.Queries)

	entity, err := entitySvc.CreateEntity(r.Context(), kind, req.Handle, req.Meta)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "create_failed", err.Error(), d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entity)
}

func (d Dependencies) getEntity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	entitySvc := service.NewEntityService(d.DB.Queries)
	
	entity, err := entitySvc.ResolveEntity(r.Context(), id, "")
	if err != nil {
		WriteError(w, http.StatusNotFound, "not_found", "Entity not found", d.Log)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

