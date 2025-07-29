-- name: CreateEmail :one
INSERT INTO emails (
    transaction_uuid, consumer, user_name, email_subject, email_text, html,
    sender_name, sender_email, recipients_name, recipients_email,
    status, created_at,
    enable_open_tracking, opened
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING *;
-- name: GetEmailByTransactionUUID :one
SELECT
    transaction_uuid, consumer, user_name, email_subject, email_text, html,
    sender_name, sender_email, recipients_name, recipients_email,
    status, created_at,
    enable_open_tracking, opened
FROM emails
WHERE transaction_uuid = $1 AND consumer = $2;
-- name: GetEmailConsumer :many
SELECT
    transaction_uuid, consumer, user_name, email_subject, email_text, html,
    sender_name, sender_email, recipients_name, recipients_email,
    status, created_at, enable_open_tracking, opened
FROM emails
WHERE consumer = $1
ORDER BY created_at DESC;
-- name: UpdateEmailStatus :exec
UPDATE emails
SET status = $2
WHERE transaction_uuid = $1;
