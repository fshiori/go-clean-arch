-- name: GetOrderByID :one
SELECT id, user_id, total_amount, status, items, created_at, updated_at
FROM orders
WHERE id = ?
LIMIT 1;

-- name: GetOrdersByUserID :many
SELECT id, user_id, total_amount, status, items, created_at, updated_at
FROM orders
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: CreateOrder :execresult
INSERT INTO orders (user_id, total_amount, status, items, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateOrder :execresult
UPDATE orders
SET user_id = ?, total_amount = ?, status = ?, items = ?, updated_at = ?
WHERE id = ?;

-- name: UpdateOrderStatus :execresult
UPDATE orders
SET status = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteOrder :execresult
DELETE FROM orders
WHERE id = ?;
