-- name: CreateUser :one
INSERT INTO users
  (name, email)
VALUES
  ($1, $2)
RETURNING id, name, email, created_at, updated_at, deleted_at;

-- name: CreateUserWithAuthProvider :one
INSERT INTO users
  (name, email, auth_provider, auth_provider_id)
VALUES
  ($1, $2, $3, $4)
RETURNING id, name, email, created_at, updated_at, deleted_at, auth_provider, auth_provider_id;

-- name: GetUserByAuthProviderID :one
SELECT id, name, email, created_at, updated_at, deleted_at, auth_provider, auth_provider_id
FROM users
WHERE auth_provider_id = $1 AND auth_provider = $2;