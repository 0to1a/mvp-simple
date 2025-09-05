-- +goose Up
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP;

-- +goose Down
ALTER TABLE users DROP COLUMN deleted_at;

