package service

import (
	"context"
	"fmt"

	"pxbox/internal/db"
	"pxbox/internal/model"
)

type EntityService struct {
	queries *db.Queries
}

func NewEntityService(queries *db.Queries) *EntityService {
	return &EntityService{queries: queries}
}

// ResolveEntity resolves an entity by ID or handle
func (s *EntityService) ResolveEntity(ctx context.Context, id, handle string) (*model.Entity, error) {
	if id != "" {
		e, err := s.queries.GetEntityByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("entity not found by ID: %w", err)
		}
		return dbEntityToModel(e), nil
	}

	if handle != "" {
		e, err := s.queries.GetEntityByHandle(ctx, handle)
		if err != nil {
			return nil, fmt.Errorf("entity not found by handle: %w", err)
		}
		return dbEntityToModel(e), nil
	}

	return nil, fmt.Errorf("either id or handle must be provided")
}

// CreateEntity creates a new entity
func (s *EntityService) CreateEntity(ctx context.Context, kind model.EntityKind, handle string, meta map[string]interface{}) (*model.Entity, error) {
	if meta == nil {
		meta = make(map[string]interface{})
	}

	e, err := s.queries.CreateEntity(ctx, string(kind), handle, meta)
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}

	return dbEntityToModel(e), nil
}

func dbEntityToModel(e db.Entity) *model.Entity {
	handle := ""
	if e.Handle != nil {
		handle = *e.Handle
	}
	return &model.Entity{
		ID:        e.ID,
		Kind:      model.EntityKind(e.Kind),
		Handle:    handle,
		Meta:      e.Meta,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

