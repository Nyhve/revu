-- name: CreateActivity :exec
INSERT INTO activities (user_id, type, reference_id, created_at)
VALUES ($1, $2, $3, $4);

-- name: ListActivities :many
SELECT * FROM activities WHERE user_id = $1 ORDER BY created_at DESC, id DESC;

-- name: ListActivityTimesSince :many
SELECT created_at FROM activities
WHERE user_id = $1 AND created_at >= $2
ORDER BY created_at;
