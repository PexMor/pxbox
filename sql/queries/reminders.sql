-- name: CreateReminder :one
INSERT INTO reminders (request_id, entity_id, remind_at)
VALUES ($1, $2, $3)
RETURNING id, request_id, entity_id, remind_at, created_at;

-- name: GetRemindersByEntity :many
SELECT id, request_id, entity_id, remind_at, created_at
FROM reminders
WHERE entity_id = $1
  AND remind_at > NOW()
ORDER BY remind_at ASC;

-- name: DeleteReminder :exec
DELETE FROM reminders
WHERE id = $1;

