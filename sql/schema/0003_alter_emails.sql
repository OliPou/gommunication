-- +goose Up
ALTER TABLE emails ADD COLUMN enable_open_tracking BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN opened BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE emails DROP COLUMN enable_open_tracking;
ALTER TABLE emails DROP COLUMN opened;
