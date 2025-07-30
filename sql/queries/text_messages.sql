-- name: CreateTextMessage :one
INSERT INTO text_messages (
    message_id, consumer, user_name, sender, recipient,
    status, api_key, price, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;
-- -- name: GetTextMessagesByTransactionUUID :one
-- SELECT
--     message_id, consumer, user_name, sender, recipient,
--     status, created_at
-- FROM text_messages
-- WHERE message_id = $1 AND consumer = $2;
-- name: GetTextMessages :many
SELECT
    message_id, consumer, user_name, sender, recipient,
    price, status, created_at
FROM text_messages
WHERE consumer = $1
ORDER BY created_at DESC;

-- name: GetTextMessage :one
SELECT
    message_id, consumer, user_name, sender, recipient,
    price, status, created_at
FROM text_messages
WHERE message_id = $1 AND consumer = $2;

-- name: UpdateTextMessageStatus :one
UPDATE text_messages
SET status = $3,
    price = $4,
    err_code = $5
WHERE message_id = $1 AND api_key = $2
RETURNING *;

