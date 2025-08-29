-- +goose Up
ALTER TABLE subdomain_ownerships RENAME COLUMN api_key TO business_unit;

-- +goose Down
ALTER TABLE subdomain_ownerships RENAME COLUMN business_unit TO api_key;