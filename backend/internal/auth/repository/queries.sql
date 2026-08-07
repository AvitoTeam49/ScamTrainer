-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at
FROM auth.users
WHERE email = sqlc.arg(email);


-- name: CreateUser :one
INSERT INTO auth.users (email, password_hash)
VALUES (sqlc.arg(email), sqlc.arg(password_hash))
RETURNING id, email, password_hash, created_at;