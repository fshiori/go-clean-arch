-- name: GetUserByID :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE id = ?
LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE email = ?
LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO users (email, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?);

-- name: UpdateUser :execresult
UPDATE users
SET email = ?, password_hash = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteUser :execresult
DELETE FROM users
WHERE id = ?;
