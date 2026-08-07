-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE auth.users 
(
       id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
       email         TEXT NOT NULL,
       password_hash TEXT NOT NULL,
       created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON auth.users(email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS auth.users;

-- +goose StatementEnd