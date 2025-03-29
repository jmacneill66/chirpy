-- name: GetAllChirps :many
SELECT id, created_at, updated_at, body, user_id 
FROM chirps 
ORDER BY created_at ASC;

-- name: GetChirpsByAuthor :many
SELECT id, user_id, body, created_at 
FROM chirps 
WHERE user_id = $1 
ORDER BY created_at ASC;
