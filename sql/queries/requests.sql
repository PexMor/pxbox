-- name: CreateRequest :one
INSERT INTO requests (
    id, created_by, entity_id, status, schema_kind, schema_payload,
    ui_hints, prefill, expires_at, deadline_at, attention_at,
    autocancel_grace, callback_url, callback_secret, files_policy, flow_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING id, created_by, entity_id, status, schema_kind, schema_payload,
          ui_hints, prefill, expires_at, deadline_at, attention_at,
          autocancel_grace, callback_url, callback_secret, files_policy,
          flow_id, deleted_at, read_at, created_at, updated_at;

-- name: GetRequestByID :one
SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
       ui_hints, prefill, expires_at, deadline_at, attention_at,
       autocancel_grace, callback_url, callback_secret, files_policy,
       flow_id, deleted_at, read_at, created_at, updated_at
FROM requests
WHERE id = $1;

-- name: UpdateRequestStatus :exec
UPDATE requests
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: ClaimRequest :exec
UPDATE requests
SET status = 'CLAIMED', updated_at = NOW()
WHERE id = $1 AND status = 'PENDING';

-- name: GetEntityQueue :many
SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
       ui_hints, prefill, expires_at, deadline_at, attention_at,
       autocancel_grace, callback_url, callback_secret, files_policy,
       flow_id, deleted_at, read_at, created_at, updated_at
FROM requests
WHERE entity_id = $1
  AND ($2::text IS NULL OR status = $2)
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

