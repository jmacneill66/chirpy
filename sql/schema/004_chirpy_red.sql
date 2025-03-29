-- +goose Up

-- Add hashed_password column with a default value
ALTER TABLE users ADD COLUMN is_chirpy_red BOOLEAN DEFAULT FALSE;

-- +goose Down
