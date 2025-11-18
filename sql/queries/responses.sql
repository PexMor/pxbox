-- name: CreateResponse :one
INSERT INTO responses (id, request_id, answered_by, payload, files)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, request_id, answered_at, answered_by, payload, files, signature_jws;

-- name: GetResponseByID :one
SELECT id, request_id, answered_at, answered_by, payload, files, signature_jws
FROM responses
WHERE id = $1;

-- name: GetResponseByRequestID :one
SELECT id, request_id, answered_at, answered_by, payload, files, signature_jws
FROM responses
WHERE request_id = $1
ORDER BY answered_at DESC
LIMIT 1;

