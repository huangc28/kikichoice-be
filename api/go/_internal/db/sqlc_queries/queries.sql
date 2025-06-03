-- name: GetUser :one
SELECT id, name, email, created_at, updated_at, deleted_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at, updated_at, deleted_at;

-- name: ListUsers :many
SELECT id, name, email, created_at, updated_at, deleted_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC;