package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Queries wraps database queries
type Queries struct {
	*pgxpool.Pool
}

// NewQueries creates a new Queries instance
func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{Pool: pool}
}

// Entity queries
func (q *Queries) GetEntityByID(ctx context.Context, id string) (Entity, error) {
	var e Entity
	err := q.Pool.QueryRow(ctx,
		"SELECT id, kind, handle, meta, created_at FROM entities WHERE id = $1",
		id,
	).Scan(&e.ID, &e.Kind, &e.Handle, &e.Meta, &e.CreatedAt)
	return e, err
}

func (q *Queries) GetEntityByHandle(ctx context.Context, handle string) (Entity, error) {
	var e Entity
	err := q.Pool.QueryRow(ctx,
		"SELECT id, kind, handle, meta, created_at FROM entities WHERE handle = $1",
		handle,
	).Scan(&e.ID, &e.Kind, &e.Handle, &e.Meta, &e.CreatedAt)
	return e, err
}

func (q *Queries) CreateEntity(ctx context.Context, kind, handle string, meta map[string]interface{}) (Entity, error) {
	var e Entity
	err := q.Pool.QueryRow(ctx,
		"INSERT INTO entities (kind, handle, meta) VALUES ($1, $2, $3) RETURNING id, kind, handle, meta, created_at",
		kind, handle, meta,
	).Scan(&e.ID, &e.Kind, &e.Handle, &e.Meta, &e.CreatedAt)
	return e, err
}

// Entity represents an entity row
type Entity struct {
	ID        string
	Kind      string
	Handle    *string
	Meta      map[string]interface{}
	CreatedAt time.Time
}

// Request queries
func (q *Queries) CreateRequest(ctx context.Context, req CreateRequestParams) (Request, error) {
	var r Request
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO requests (
			id, created_by, entity_id, status, schema_kind, schema_payload,
			ui_hints, prefill, expires_at, deadline_at, attention_at,
			autocancel_grace, callback_url, callback_secret, files_policy, flow_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_by, entity_id, status, schema_kind, schema_payload,
			ui_hints, prefill, expires_at, deadline_at, attention_at,
			autocancel_grace, callback_url, callback_secret, files_policy,
			flow_id, deleted_at, read_at, created_at, updated_at`,
		req.ID, req.CreatedBy, req.EntityID, req.Status, req.SchemaKind, req.SchemaPayload,
		req.UIHints, req.Prefill, req.ExpiresAt, req.DeadlineAt, req.AttentionAt,
		req.AutocancelGrace, req.CallbackURL, req.CallbackSecret, req.FilesPolicy, req.FlowID,
	).Scan(
		&r.ID, &r.CreatedBy, &r.EntityID, &r.Status, &r.SchemaKind, &r.SchemaPayload,
		&r.UIHints, &r.Prefill, &r.ExpiresAt, &r.DeadlineAt, &r.AttentionAt,
		&r.AutocancelGrace, &r.CallbackURL, &r.CallbackSecret, &r.FilesPolicy, &r.FlowID,
		&r.DeletedAt, &r.ReadAt, &r.CreatedAt, &r.UpdatedAt,
	)
	return r, err
}

type CreateRequestParams struct {
	ID              string
	CreatedBy       string
	EntityID        string
	Status          string
	SchemaKind      string
	SchemaPayload   map[string]interface{}
	UIHints         map[string]interface{}
	Prefill         map[string]interface{}
	ExpiresAt       *time.Time
	DeadlineAt      *time.Time
	AttentionAt     *time.Time
	AutocancelGrace *time.Duration
	CallbackURL     *string
	CallbackSecret  *string
	FilesPolicy     map[string]interface{}
	FlowID          *string
}

func (q *Queries) GetRequestByID(ctx context.Context, id string) (Request, error) {
	var r Request
	err := q.Pool.QueryRow(ctx,
		`SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
			ui_hints, prefill, expires_at, deadline_at, attention_at,
			autocancel_grace, callback_url, callback_secret, files_policy,
			flow_id, deleted_at, read_at, created_at, updated_at
		FROM requests WHERE id = $1`,
		id,
	).Scan(
		&r.ID, &r.CreatedBy, &r.EntityID, &r.Status, &r.SchemaKind, &r.SchemaPayload,
		&r.UIHints, &r.Prefill, &r.ExpiresAt, &r.DeadlineAt, &r.AttentionAt,
		&r.AutocancelGrace, &r.CallbackURL, &r.CallbackSecret, &r.FilesPolicy, &r.FlowID,
		&r.DeletedAt, &r.ReadAt, &r.CreatedAt, &r.UpdatedAt,
	)
	return r, err
}

func (q *Queries) UpdateRequestStatus(ctx context.Context, id, status string) error {
	_, err := q.Pool.Exec(ctx,
		"UPDATE requests SET status = $2, updated_at = NOW() WHERE id = $1",
		id, status,
	)
	return err
}

func (q *Queries) ClaimRequest(ctx context.Context, id string) error {
	result, err := q.Pool.Exec(ctx,
		"UPDATE requests SET status = 'CLAIMED', updated_at = NOW() WHERE id = $1 AND status = 'PENDING'",
		id,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (q *Queries) GetEntityQueue(ctx context.Context, entityID string, status *string, limit, offset int) ([]Request, error) {
	var rows pgx.Rows
	var err error

	if status != nil {
		rows, err = q.Pool.Query(ctx,
			`SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
				ui_hints, prefill, expires_at, deadline_at, attention_at,
				autocancel_grace, callback_url, callback_secret, files_policy,
				flow_id, deleted_at, read_at, created_at, updated_at
			FROM requests
			WHERE entity_id = $1 AND status = $2 AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4`,
			entityID, *status, limit, offset,
		)
	} else {
		rows, err = q.Pool.Query(ctx,
			`SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
				ui_hints, prefill, expires_at, deadline_at, attention_at,
				autocancel_grace, callback_url, callback_secret, files_policy,
				flow_id, deleted_at, read_at, created_at, updated_at
			FROM requests
			WHERE entity_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`,
			entityID, limit, offset,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Request
	for rows.Next() {
		var r Request
		err := rows.Scan(
			&r.ID, &r.CreatedBy, &r.EntityID, &r.Status, &r.SchemaKind, &r.SchemaPayload,
			&r.UIHints, &r.Prefill, &r.ExpiresAt, &r.DeadlineAt, &r.AttentionAt,
			&r.AutocancelGrace, &r.CallbackURL, &r.CallbackSecret, &r.FilesPolicy, &r.FlowID,
			&r.DeletedAt, &r.ReadAt, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, rows.Err()
}

type Request struct {
	ID              string
	CreatedBy       string
	EntityID        string
	Status          string
	SchemaKind      string
	SchemaPayload   map[string]interface{}
	UIHints         map[string]interface{}
	Prefill         map[string]interface{}
	ExpiresAt       *time.Time
	DeadlineAt      *time.Time
	AttentionAt     *time.Time
	AutocancelGrace *time.Duration
	CallbackURL     *string
	CallbackSecret  *string
	FilesPolicy     map[string]interface{}
	FlowID          *string
	DeletedAt       *time.Time
	ReadAt          *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Response queries
func (q *Queries) CreateResponse(ctx context.Context, resp CreateResponseParams) (Response, error) {
	var r Response
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO responses (id, request_id, answered_by, payload, files)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, request_id, answered_at, answered_by, payload, files, signature_jws`,
		resp.ID, resp.RequestID, resp.AnsweredBy, resp.Payload, resp.Files,
	).Scan(
		&r.ID, &r.RequestID, &r.AnsweredAt, &r.AnsweredBy, &r.Payload, &r.Files, &r.SignatureJWS,
	)
	return r, err
}

type CreateResponseParams struct {
	ID         string
	RequestID  string
	AnsweredBy string
	Payload    map[string]interface{}
	Files      []map[string]interface{}
}

type Response struct {
	ID          string
	RequestID   string
	AnsweredAt  time.Time
	AnsweredBy  string
	Payload     map[string]interface{}
	Files       []map[string]interface{}
	SignatureJWS *string
}

func (q *Queries) GetResponseByRequestID(ctx context.Context, requestID string) (Response, error) {
	var r Response
	err := q.Pool.QueryRow(ctx,
		`SELECT id, request_id, answered_at, answered_by, payload, files, signature_jws
		FROM responses
		WHERE request_id = $1
		ORDER BY answered_at DESC
		LIMIT 1`,
		requestID,
	).Scan(
		&r.ID, &r.RequestID, &r.AnsweredAt, &r.AnsweredBy, &r.Payload, &r.Files, &r.SignatureJWS,
	)
	return r, err
}

// Flow queries
func (q *Queries) CreateFlow(ctx context.Context, flow CreateFlowParams) (Flow, error) {
	var f Flow
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO flows (kind, owner_entity, status, cursor, last_event_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, kind, owner_entity, status, cursor, last_event_id, created_at, updated_at`,
		flow.Kind, flow.OwnerEntity, flow.Status, flow.Cursor, flow.LastEventID,
	).Scan(
		&f.ID, &f.Kind, &f.OwnerEntity, &f.Status, &f.Cursor, &f.LastEventID, &f.CreatedAt, &f.UpdatedAt,
	)
	return f, err
}

type CreateFlowParams struct {
	Kind        string
	OwnerEntity string
	Status      string
	Cursor      map[string]interface{}
	LastEventID *string
}

func (q *Queries) GetFlowByID(ctx context.Context, id string) (Flow, error) {
	var f Flow
	err := q.Pool.QueryRow(ctx,
		`SELECT id, kind, owner_entity, status, cursor, last_event_id, created_at, updated_at
		FROM flows WHERE id = $1`,
		id,
	).Scan(
		&f.ID, &f.Kind, &f.OwnerEntity, &f.Status, &f.Cursor, &f.LastEventID, &f.CreatedAt, &f.UpdatedAt,
	)
	return f, err
}

func (q *Queries) UpdateFlowStatus(ctx context.Context, id, status string) error {
	_, err := q.Pool.Exec(ctx,
		"UPDATE flows SET status = $2, updated_at = NOW() WHERE id = $1",
		id, status,
	)
	return err
}

func (q *Queries) UpdateFlowCursor(ctx context.Context, id string, cursor map[string]interface{}) error {
	_, err := q.Pool.Exec(ctx,
		"UPDATE flows SET cursor = $2, updated_at = NOW() WHERE id = $1",
		id, cursor,
	)
	return err
}

type Flow struct {
	ID          string
	Kind        string
	OwnerEntity string
	Status      string
	Cursor      map[string]interface{}
	LastEventID *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Inquiry queries
func (q *Queries) ListInquiries(ctx context.Context, entityID *string, status *string, includeDeleted bool, sortBy string, limit, offset int) ([]Request, error) {
	// Use GetEntityQueue when entityID is provided (it works correctly with string parameter)
	if entityID != nil && *entityID != "" {
		var statusPtr *string
		if status != nil && *status != "" {
			statusPtr = status
		}
		// Call GetEntityQueue which we know works (verified via /v1/entities/{id}/queue endpoint)
		requests, err := q.GetEntityQueue(ctx, *entityID, statusPtr, limit, offset)
		if err != nil {
			return nil, err
		}
		// Return empty slice instead of nil if no results
		if requests == nil {
			return []Request{}, nil
		}
		return requests, nil
	}
	
	// Otherwise, query all requests (no entity filter)
	var rows pgx.Rows
	var err error
	
	var query string
	var args []interface{}
	
	if status != nil && *status != "" {
		query = `SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
			ui_hints, prefill, expires_at, deadline_at, attention_at,
			autocancel_grace, callback_url, callback_secret, files_policy,
			flow_id, deleted_at, read_at, created_at, updated_at
		FROM requests
		WHERE status = $1
		  AND deleted_at IS NULL
		ORDER BY 
		  CASE WHEN $2::text = 'deadline' THEN deadline_at END ASC NULLS LAST,
		  CASE WHEN $2::text = 'created' THEN created_at END DESC
		LIMIT $3 OFFSET $4`
		args = []interface{}{*status, sortBy, limit, offset}
	} else {
		query = `SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
			ui_hints, prefill, expires_at, deadline_at, attention_at,
			autocancel_grace, callback_url, callback_secret, files_policy,
			flow_id, deleted_at, read_at, created_at, updated_at
		FROM requests
		WHERE deleted_at IS NULL
		ORDER BY 
		  CASE WHEN $1::text = 'deadline' THEN deadline_at END ASC NULLS LAST,
		  CASE WHEN $1::text = 'created' THEN created_at END DESC
		LIMIT $2 OFFSET $3`
		args = []interface{}{sortBy, limit, offset}
	}

	rows, err = q.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]Request, 0)
	for rows.Next() {
		var r Request
		err := rows.Scan(
			&r.ID, &r.CreatedBy, &r.EntityID, &r.Status, &r.SchemaKind, &r.SchemaPayload,
			&r.UIHints, &r.Prefill, &r.ExpiresAt, &r.DeadlineAt, &r.AttentionAt,
			&r.AutocancelGrace, &r.CallbackURL, &r.CallbackSecret, &r.FilesPolicy, &r.FlowID,
			&r.DeletedAt, &r.ReadAt, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

type Reminder struct {
	ID        string
	RequestID string
	EntityID  string
	RemindAt  time.Time
	CreatedAt time.Time
}

func (q *Queries) GetReminderByID(ctx context.Context, id string) (Reminder, error) {
	var r Reminder
	err := q.Pool.QueryRow(ctx,
		`SELECT id::text, request_id, entity_id, remind_at, created_at
		FROM reminders WHERE id = $1`,
		id,
	).Scan(&r.ID, &r.RequestID, &r.EntityID, &r.RemindAt, &r.CreatedAt)
	return r, err
}

func (q *Queries) CreateReminder(ctx context.Context, requestID, entityID string, remindAt time.Time) (Reminder, error) {
	var r Reminder
	err := q.Pool.QueryRow(ctx,
		`INSERT INTO reminders (request_id, entity_id, remind_at)
		VALUES ($1, $2, $3)
		RETURNING id::text, request_id, entity_id, remind_at, created_at`,
		requestID, entityID, remindAt,
	).Scan(&r.ID, &r.RequestID, &r.EntityID, &r.RemindAt, &r.CreatedAt)
	return r, err
}

func (q *Queries) MarkInquiryRead(ctx context.Context, id string) error {
	_, err := q.Pool.Exec(ctx,
		"UPDATE requests SET read_at = NOW(), updated_at = NOW() WHERE id = $1",
		id,
	)
	return err
}

func (q *Queries) SoftDeleteInquiry(ctx context.Context, id string) error {
	_, err := q.Pool.Exec(ctx,
		"UPDATE requests SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1",
		id,
	)
	return err
}

func (q *Queries) GetInquiryByID(ctx context.Context, id string) (Request, error) {
	return q.GetRequestByID(ctx, id)
}

// GetFlowsByStatus gets flows by status list
func (q *Queries) GetFlowsByStatus(ctx context.Context, statuses []string) ([]Flow, error) {
	if len(statuses) == 0 {
		return []Flow{}, nil
	}

	query := `SELECT id, kind, owner_entity, status, cursor, last_event_id, created_at, updated_at
		FROM flows
		WHERE status = ANY($1)
		ORDER BY created_at ASC`

	rows, err := q.Pool.Query(ctx, query, statuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []Flow
	for rows.Next() {
		var f Flow
		err := rows.Scan(
			&f.ID, &f.Kind, &f.OwnerEntity, &f.Status, &f.Cursor, &f.LastEventID,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		flows = append(flows, f)
	}
	return flows, rows.Err()
}

