-- name: ListLearningProjects :many
SELECT p.id, p.title, p.description, p.topic, p.created_at,
       COALESCE(SUM(s.duration_minutes), 0)::int AS total_minutes
FROM learning_projects p
LEFT JOIN learning_sessions s ON s.learning_project_id = p.id
WHERE p.user_id = $1
GROUP BY p.id
ORDER BY p.id;

-- name: LearningProjectExists :one
SELECT EXISTS (SELECT 1 FROM learning_projects WHERE id = $1 AND user_id = $2);

-- name: CreateLearningProject :one
INSERT INTO learning_projects (user_id, title, description, topic, created_at, updated_at)
VALUES (sqlc.arg(user_id), sqlc.arg(title), sqlc.narg(description), sqlc.arg(topic),
        sqlc.arg(now), sqlc.arg(now))
RETURNING *;

-- name: DeleteLearningProject :execrows
DELETE FROM learning_projects WHERE id = $1 AND user_id = $2;

-- name: ListLearningSessions :many
SELECT * FROM learning_sessions WHERE learning_project_id = $1 ORDER BY id;

-- name: CreateLearningSession :one
INSERT INTO learning_sessions (learning_project_id, duration_minutes, notes, date, created_at)
VALUES (sqlc.arg(project_id), sqlc.arg(duration_minutes), sqlc.narg(notes), sqlc.arg(date), sqlc.arg(now))
RETURNING *;

-- name: DeleteLearningSession :execrows
DELETE FROM learning_sessions WHERE id = $1 AND learning_project_id = $2;
