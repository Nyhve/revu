-- name: ListGoals :many
SELECT * FROM goals WHERE user_id = $1 ORDER BY id;

-- name: GetGoal :one
SELECT * FROM goals WHERE id = $1 AND user_id = $2;

-- name: CreateGoal :one
INSERT INTO goals (title, description, target_date, status, user_id, created_at, updated_at)
VALUES (sqlc.arg(title), sqlc.narg(description), sqlc.narg(target_date), sqlc.arg(status),
        sqlc.arg(user_id), sqlc.arg(now), sqlc.arg(now))
RETURNING *;

-- name: UpdateGoal :one
UPDATE goals
SET title = sqlc.arg(title),
    description = sqlc.narg(description),
    target_date = sqlc.narg(target_date),
    status = sqlc.arg(status),
    updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING *;

-- name: DeleteGoal :execrows
DELETE FROM goals WHERE id = $1 AND user_id = $2;
