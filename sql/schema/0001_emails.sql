-- +goose Up
CREATE TABLE emails(
    transaction_uuid UUID PRIMARY KEY,
    consumer TEXT NOT NULL,
    user_name TEXT NOT NULL,
    email_subject TEXT,
    email_text TEXT,
    html TEXT,
    sender_name TEXT NOT NULL,
    sender_email TEXT NOT NULL,
    recipients_name TEXT NOT NULL,
    recipients_email TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

ALTER TABLE emails ADD COLUMN enable_open_tracking BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN opened BOOLEAN DEFAULT FALSE;

-- +goose Down
DROP TABLE emails;
