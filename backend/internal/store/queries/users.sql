-- name: CreateUser :one
INSERT INTO users (email, password_hash, created_at, updated_at)
VALUES (sqlc.arg(email), sqlc.arg(password_hash), sqlc.arg(now), sqlc.arg(now))
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UserExistsByEmail :one
SELECT EXISTS (SELECT 1 FROM users WHERE email = $1);

-- name: EnsurePreferences :exec
INSERT INTO user_preferences (user_id) VALUES ($1)
ON CONFLICT (user_id) DO NOTHING;

-- name: GetPreferences :one
SELECT * FROM user_preferences WHERE user_id = $1;

-- name: UpsertPreferences :one
INSERT INTO user_preferences (
    user_id, goals_enabled, tasks_enabled, habits_enabled,
    fitness_enabled, learning_enabled, finance_enabled
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id) DO UPDATE SET
    goals_enabled    = EXCLUDED.goals_enabled,
    tasks_enabled    = EXCLUDED.tasks_enabled,
    habits_enabled   = EXCLUDED.habits_enabled,
    fitness_enabled  = EXCLUDED.fitness_enabled,
    learning_enabled = EXCLUDED.learning_enabled,
    finance_enabled  = EXCLUDED.finance_enabled
RETURNING *;
