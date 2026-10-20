CREATE TABLE exercises (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(255) NOT NULL UNIQUE,
    muscle_group VARCHAR(50) NOT NULL
);