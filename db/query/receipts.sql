-- name: CreateFoodReceipt :one
INSERT INTO food_receipts (
	store_name
) VALUES (
	$1
) RETURNING *;

-- name: GetFoodReceipt :one
SELECT * FROM food_receipts
WHERE id = $1 LIMIT 1;

-- name: CreateFoodContent :one
INSERT INTO food_contents (
	name,
	calories,
	lipid,
	carbohydrate,
	Protein
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetFoodContent :one
SELECT * FROM food_contents
WHERE id = $1 LIMIT 1;

-- name: CreateFoodReceiptContent :one
INSERT INTO food_receipt_contents (
	food_receipt_id,
	food_content_id,
	amount
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: ListFoodReceiptContents :many
SELECT
	food_receipt_contents.food_receipt_id AS food_receipt_id,
	food_receipt_contents.food_content_id AS food_content_id,
	food_receipt_contents.amount AS amount,
	food_contents.name AS name,
	food_contents.calories AS calories,
	food_contents.lipid AS lipid,
	food_contents.carbohydrate AS carbohydrate,
	food_contents.protein AS protein
FROM food_receipt_contents
INNER JOIN food_contents ON food_receipt_contents.food_content_id = food_contents.id
WHERE food_receipt_contents.food_receipt_id = $1;
