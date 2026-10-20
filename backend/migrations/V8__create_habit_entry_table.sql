CREATE TABLE habit_entries (
    id         BIGSERIAL PRIMARY KEY,
    habit_id   BIGINT NOT NULL REFERENCES habits(id),
    date       DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (habit_id, date)
);