CREATE TABLE activities (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    type         VARCHAR(50) NOT NULL,
    reference_id BIGINT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT now()
);