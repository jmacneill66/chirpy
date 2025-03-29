-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetChirpsForUsers :one
SELECT id, created_at, updated_at, body, user_id 
FROM chirps 
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, email, hashed_password, is_chirpy_red
FROM users
WHERE email = $1
LIMIT 1;

-- name: UpdateUser :exec
UPDATE users 
SET email = $1, hashed_password = $2, updated_at = NOW()
WHERE id = $3;

-- name: GetUserByID :one
SELECT id, email, hashed_password, created_at, updated_at, is_chirpy_red
FROM users
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1 AND user_id = $2;

-- name: GetChirpByID :one
SELECT id, user_id, body FROM chirps WHERE id = $1;

-- name: UpgradeUserToChirpyRed :execrows
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1;
