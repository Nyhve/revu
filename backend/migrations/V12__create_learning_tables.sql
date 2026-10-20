CREATE TABLE learning_projects (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    topic       VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE learning_sessions (
    id                  BIGSERIAL PRIMARY KEY,
    learning_project_id BIGINT NOT NULL REFERENCES learning_projects(id) ON DELETE CASCADE,
    duration_minutes    INT NOT NULL,
    notes               TEXT,
    date                DATE NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT now()
);