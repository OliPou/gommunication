-- +goose Up
ALTER TABLE emails ADD COLUMN reply_to TEXT;

-- +goose Down
ALTER TABLE emails DROP COLUMN reply_to;