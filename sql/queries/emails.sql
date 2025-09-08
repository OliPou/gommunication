-- name: CreateEmail :one
INSERT INTO emails (
    transaction_uuid, consumer, user_name, email_subject, email_text, html,
    sender_name, sender_email, recipients_name, recipients_email,
    status, created_at,
    enable_open_tracking, opened, reply_to, access_level
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;
-- name: GetEmailByTransactionUUID :one
SELECT
    transaction_uuid, consumer, user_name, email_subject, email_text, html,
    sender_name, sender_email, recipients_name, recipients_email,
    status, created_at,
    enable_open_tracking, opened, reply_to, access_level
FROM emails
WHERE transaction_uuid = $1 AND consumer = $2;
-- name: GetEmailConsumer :many
SELECT
    transaction_uuid, consumer, user_name, email_subject, email_text, html,
    sender_name, sender_email, recipients_name, recipients_email,
    status, created_at, enable_open_tracking, opened, reply_to, access_level
FROM emails
WHERE consumer = $1
ORDER BY created_at DESC;
-- name: GetEmailsByConsumerAndBusinessUnit :many
SELECT
    e.transaction_uuid, e.consumer, e.user_name, e.email_subject, e.email_text, e.html,
    e.sender_name, e.sender_email, e.recipients_name, e.recipients_email,
    e.status, e.created_at, e.enable_open_tracking, e.opened, e.reply_to, e.access_level
FROM emails e
WHERE e.consumer = $1
  AND EXISTS (
    SELECT 1
    FROM available_subdomains s
    LEFT JOIN subdomain_ownerships o
      ON o.subdomain_id = s.available_subdomain_uuid
     AND o.business_unit = $2
    WHERE s.name = split_part(e.sender_email, '@', 2)
      AND (o.business_unit IS NOT NULL OR s.is_default = TRUE)
  )
ORDER BY e.created_at DESC;
-- name: UpdateEmailStatus :exec
UPDATE emails
SET status = $2
WHERE transaction_uuid = $1;

-- name: UpdateEmailOpened :exec
UPDATE emails
SET opened = $2
WHERE transaction_uuid = $1
RETURNING *;

