-- name: ListExercises :many
SELECT * FROM exercises ORDER BY id;

-- name: GetExercise :one
SELECT * FROM exercises WHERE id = $1;

-- name: ExerciseExistsByName :one
SELECT EXISTS (SELECT 1 FROM exercises WHERE name = $1);

-- name: CreateExercise :one
INSERT INTO exercises (name, muscle_group) VALUES ($1, $2) RETURNING *;

-- name: ListWorkouts :many
SELECT * FROM workouts WHERE user_id = $1 ORDER BY date DESC, id DESC;

-- name: GetWorkout :one
SELECT * FROM workouts WHERE id = $1 AND user_id = $2;

-- name: GetWorkoutForUpdate :one
SELECT * FROM workouts WHERE id = $1 AND user_id = $2 FOR UPDATE;

-- name: CreateWorkout :one
INSERT INTO workouts (user_id, name, date, notes, completed, created_at, updated_at)
VALUES (sqlc.arg(user_id), sqlc.arg(name), sqlc.arg(date), sqlc.narg(notes), FALSE,
        sqlc.arg(now), sqlc.arg(now))
RETURNING *;

-- name: CompleteWorkout :exec
UPDATE workouts SET completed = TRUE, updated_at = $2 WHERE id = $1;

-- name: DeleteWorkout :execrows
DELETE FROM workouts WHERE id = $1 AND user_id = $2;

-- name: ListWorkoutExercises :many
SELECT we.id, we.workout_id, we.order_index,
       e.id AS exercise_id, e.name AS exercise_name, e.muscle_group AS exercise_muscle_group
FROM workout_exercises we
JOIN exercises e ON e.id = we.exercise_id
WHERE we.workout_id = ANY(sqlc.arg(workout_ids)::bigint[])
ORDER BY we.workout_id, we.order_index, we.id;

-- name: WorkoutExerciseExists :one
SELECT EXISTS (SELECT 1 FROM workout_exercises WHERE id = $1 AND workout_id = $2);

-- name: NextWorkoutExerciseIndex :one
SELECT COALESCE(MAX(order_index) + 1, 0)::int FROM workout_exercises WHERE workout_id = $1;

-- name: CreateWorkoutExercise :exec
INSERT INTO workout_exercises (workout_id, exercise_id, order_index) VALUES ($1, $2, $3);

-- name: DeleteWorkoutExercise :execrows
DELETE FROM workout_exercises WHERE id = $1 AND workout_id = $2;

-- name: RenumberWorkoutExercises :exec
UPDATE workout_exercises we
SET order_index = r.rn - 1
FROM (
    SELECT x.id, (ROW_NUMBER() OVER (ORDER BY x.order_index, x.id))::int AS rn
    FROM workout_exercises x WHERE x.workout_id = $1
) r
WHERE we.id = r.id;

-- name: ListWorkoutSets :many
SELECT * FROM workout_sets
WHERE workout_exercise_id = ANY(sqlc.arg(workout_exercise_ids)::bigint[])
ORDER BY workout_exercise_id, set_number, id;

-- name: NextSetNumber :one
SELECT COALESCE(MAX(set_number) + 1, 1)::int FROM workout_sets WHERE workout_exercise_id = $1;

-- name: CreateWorkoutSet :exec
INSERT INTO workout_sets (workout_exercise_id, set_number, weight_kg, reps, completed_at)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteWorkoutSet :execrows
DELETE FROM workout_sets WHERE id = $1 AND workout_exercise_id = $2;

-- name: RenumberWorkoutSets :exec
UPDATE workout_sets ws
SET set_number = r.rn
FROM (
    SELECT x.id, (ROW_NUMBER() OVER (ORDER BY x.set_number, x.id))::int AS rn
    FROM workout_sets x WHERE x.workout_exercise_id = $1
) r
WHERE ws.id = r.id;
