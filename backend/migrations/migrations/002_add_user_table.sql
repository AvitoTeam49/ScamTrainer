-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS users_schema;

CREATE TABLE IF NOT EXISTS users_schema.users 
(
    id         BIGINT PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    username   TEXT NOT NULL UNIQUE,
    score      INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_score ON users_schema.users (score DESC);

CREATE TABLE IF NOT EXISTS users_schema.user_progress 
(
    user_id             BIGINT PRIMARY KEY REFERENCES users_schema.users(id) ON DELETE CASCADE,
    scenarios_completed INT NOT NULL DEFAULT 0,
    scams_detected      INT NOT NULL DEFAULT 0,
    failed_attempts     INT NOT NULL DEFAULT 0,
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION users_schema.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users_schema.users
    FOR EACH ROW
    EXECUTE FUNCTION users_schema.update_updated_at_column();

CREATE OR REPLACE TRIGGER update_user_progress_updated_at
    BEFORE UPDATE ON users_schema.user_progress
    FOR EACH ROW
    EXECUTE FUNCTION users_schema.update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP SCHEMA IF EXISTS users_schema CASCADE;

-- +goose StatementEnd