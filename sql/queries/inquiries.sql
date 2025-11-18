-- name: ListInquiries :many
SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
       ui_hints, prefill, expires_at, deadline_at, attention_at,
       autocancel_grace, callback_url, callback_secret, files_policy,
       flow_id, deleted_at, read_at, created_at, updated_at
FROM requests
WHERE ($1::uuid IS NULL OR entity_id = $1)
  AND ($2::text IS NULL OR status = $2)
  AND ($3::boolean IS NULL OR ($3 = true AND deleted_at IS NULL) OR ($3 = false AND deleted_at IS NOT NULL))
  AND deleted_at IS NULL
ORDER BY 
  CASE WHEN $4::text = 'deadline' THEN deadline_at END ASC NULLS LAST,
  CASE WHEN $4::text = 'created' THEN created_at END DESC
LIMIT $5 OFFSET $6;

-- name: MarkInquiryRead :exec
UPDATE requests
SET read_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteInquiry :exec
UPDATE requests
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: GetInquiryByID :one
SELECT id, created_by, entity_id, status, schema_kind, schema_payload,
       ui_hints, prefill, expires_at, deadline_at, attention_at,
       autocancel_grace, callback_url, callback_secret, files_policy,
       flow_id, deleted_at, read_at, created_at, updated_at
FROM requests
WHERE id = $1;

