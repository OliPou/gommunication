-- +goose Up
CREATE TABLE text_messages(
    message_id UUID PRIMARY KEY,
    consumer TEXT NOT NULL,
    user_name TEXT NOT NULL,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    status TEXT NOT NULL,
    api_key TEXT NOT NULL,
    err_code NUMERIC,
    price NUMERIC NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE text_messages;
