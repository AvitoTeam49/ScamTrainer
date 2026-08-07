-- name: CreateUser :one
WITH new_user AS (
    INSERT INTO users_schema.users (id, username)
    VALUES (
        sqlc.arg(id), 
        sqlc.arg(username)
    )
    RETURNING id, username, score, created_at, updated_at
),
new_progress AS (
    INSERT INTO users_schema.user_progress (user_id)
    SELECT id FROM new_user
)
SELECT id, username, score, created_at, updated_at 
FROM new_user;

-- name: GetUserByID :one
SELECT id, username, score, created_at, updated_at
FROM users_schema.users
WHERE id = sqlc.arg(id);

-- name: GetUserProgress :one
SELECT user_id, scenarios_completed, scams_detected, failed_attempts, updated_at
FROM users_schema.user_progress
WHERE user_id = sqlc.arg(user_id);

-- name: UpdateUserScore :one
UPDATE users_schema.users
SET score = score + sqlc.arg(score_delta)
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: GetLeaderboard :many
SELECT id, username, score
FROM users_schema.users
ORDER BY score DESC, id ASC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(offs);

-- name: CountUsers :one
SELECT COUNT(*) 
FROM users_schema.users;