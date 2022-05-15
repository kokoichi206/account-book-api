-- name: CreateSession :one
INSERT INTO sessions (
	id,
	user_id,
	user_agent,
	client_ip,
	expires_at
) VALUES (
	$1, $2, $3, $4, $5
) RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1 LIMIT 1;

-- name: UpdateSession :exec
UPDATE sessions
SET expires_at = $1
WHERE id = $2
RETURNING *;

-- name: DeleteSession :exec
UPDATE sessions
SET expires_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;
