-- name: GetEntityByID :one
SELECT id, kind, handle, meta, created_at
FROM entities
WHERE id = $1;

-- name: GetEntityByHandle :one
SELECT id, kind, handle, meta, created_at
FROM entities
WHERE handle = $1;

-- name: CreateEntity :one
INSERT INTO entities (kind, handle, meta)
VALUES ($1, $2, $3)
RETURNING id, kind, handle, meta, created_at;

