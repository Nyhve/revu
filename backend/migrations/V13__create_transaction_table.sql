CREATE TABLE transactions (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    amount      DECIMAL(12,2) NOT NULL,
    description VARCHAR(255) NOT NULL,
    category    VARCHAR(50) NOT NULL,
    type        VARCHAR(20) NOT NULL,
    date        DATE NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT now()
);