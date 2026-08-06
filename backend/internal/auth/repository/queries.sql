-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, created_at
FROM auth.users
WHERE email = sqlc.arg(email);


-- name: CreateUser :one
INSERT INTO auth.users (email, password_hash, role)
VALUES (sqlc.arg(email), sqlc.arg(password_hash), sqlc.arg(role))
RETURNING id, email, password_hash, role, created_at;