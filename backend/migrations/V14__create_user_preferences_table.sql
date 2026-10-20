CREATE TABLE user_preferences (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL UNIQUE REFERENCES users(id),
    goals_enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    tasks_enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    habits_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    fitness_enabled  BOOLEAN NOT NULL DEFAULT TRUE,
    learning_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    finance_enabled  BOOLEAN NOT NULL DEFAULT TRUE
);