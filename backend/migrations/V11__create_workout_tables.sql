CREATE TABLE workouts (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    name       VARCHAR(255) NOT NULL,
    date       DATE NOT NULL,
    notes      TEXT,
    completed  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE workout_exercises (
    id          BIGSERIAL PRIMARY KEY,
    workout_id  BIGINT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id BIGINT NOT NULL REFERENCES exercises(id),
    order_index INT NOT NULL DEFAULT 0
);

CREATE TABLE workout_sets (
    id                  BIGSERIAL PRIMARY KEY,
    workout_exercise_id BIGINT NOT NULL REFERENCES workout_exercises(id) ON DELETE CASCADE,
    set_number          INT NOT NULL,
    weight_kg           DECIMAL(6,2) NOT NULL,
    reps                INT NOT NULL,
    completed_at        TIMESTAMP NOT NULL DEFAULT now()
);