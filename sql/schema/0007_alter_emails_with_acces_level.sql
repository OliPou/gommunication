-- +goose Up
ALTER TABLE emails ADD COLUMN access_level TEXT NOT NULL DEFAULT 'basic';

-- +goose Down
ALTER TABLE emails DROP COLUMN access_level;