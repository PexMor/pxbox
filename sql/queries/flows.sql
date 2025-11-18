-- name: CreateFlow :one
INSERT INTO flows (kind, owner_entity, status, cursor, last_event_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, kind, owner_entity, status, cursor, last_event_id, created_at, updated_at;

-- name: GetFlowByID :one
SELECT id, kind, owner_entity, status, cursor, last_event_id, created_at, updated_at
FROM flows
WHERE id = $1;

-- name: UpdateFlowStatus :exec
UPDATE flows
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateFlowCursor :exec
UPDATE flows
SET cursor = $2, updated_at = NOW()
WHERE id = $1;

-- name: GetRunningFlows :many
SELECT id, kind, owner_entity, status, cursor, last_event_id, created_at, updated_at
FROM flows
WHERE status IN ('RUNNING', 'WAITING_INPUT');

