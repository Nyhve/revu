-- name: ListHabits :many
SELECT * FROM habits WHERE user_id = $1 ORDER BY id;

-- name: GetHabit :one
SELECT * FROM habits WHERE id = $1 AND user_id = $2;

-- name: CreateHabit :one
INSERT INTO habits (user_id, name, description, frequency, created_at, updated_at)
VALUES (sqlc.arg(user_id), sqlc.arg(name), sqlc.narg(description), sqlc.arg(frequency),
        sqlc.arg(now), sqlc.arg(now))
RETURNING *;

-- name: DeleteHabit :execrows
DELETE FROM habits WHERE id = $1 AND user_id = $2;

-- name: ListHabitEntries :many
SELECT * FROM habit_entries WHERE habit_id = $1 ORDER BY date, id;

-- name: ListHabitEntriesForUserOnDate :many
SELECT e.* FROM habit_entries e
JOIN habits h ON h.id = e.habit_id
WHERE h.user_id = $1 AND e.date = $2
ORDER BY e.id;

-- name: HabitEntryExistsBetween :one
SELECT EXISTS (
    SELECT 1 FROM habit_entries
    WHERE habit_id = sqlc.arg(habit_id)
      AND date >= sqlc.arg(from_date) AND date <= sqlc.arg(to_date)
);

-- name: CreateHabitEntry :one
INSERT INTO habit_entries (habit_id, date, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteHabitEntry :execrows
DELETE FROM habit_entries WHERE id = $1 AND habit_id = $2;

-- name: HabitCompletionCounts :many
SELECT h.id, h.name, h.frequency,
       COUNT(e.id) FILTER (WHERE e.date >= sqlc.arg(from_date))::int AS completed_days
FROM habits h
LEFT JOIN habit_entries e ON e.habit_id = h.id
WHERE h.user_id = sqlc.arg(user_id)
GROUP BY h.id
ORDER BY h.id;
