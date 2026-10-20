-- name: ListTasks :many
SELECT * FROM tasks WHERE user_id = $1 ORDER BY id;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1 AND user_id = $2;

-- name: CreateTask :one
INSERT INTO tasks (title, description, due_date, completed, user_id, created_at, updated_at)
VALUES (sqlc.arg(title), sqlc.narg(description), sqlc.narg(due_date), FALSE,
        sqlc.arg(user_id), sqlc.arg(now), sqlc.arg(now))
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET title = sqlc.arg(title),
    description = sqlc.narg(description),
    due_date = sqlc.narg(due_date),
    completed = sqlc.arg(completed),
    updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING *;

-- name: DeleteTask :execrows
DELETE FROM tasks WHERE id = $1 AND user_id = $2;
