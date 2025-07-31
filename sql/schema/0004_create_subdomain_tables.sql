-- +goose Up
CREATE TABLE available_subdomains (
    available_subdomain_uuid UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE subdomain_ownerships (
    subdomain_ownership_uuid UUID PRIMARY KEY,
    subdomain_id UUID NOT NULL REFERENCES available_subdomains(available_subdomain_uuid) ON DELETE CASCADE,
    api_key TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(subdomain_id)  
);

-- +goose Down
DROP TABLE IF EXISTS subdomain_ownerships;
DROP TABLE IF EXISTS available_subdomains;