-- name: CreateExpense :one
INSERT INTO expenses (
	user_id,
	category_id,
	amount,
	food_receipt_id,
	comment
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
) RETURNING *;

-- name: ListExpenses :many
SELECT
	expenses.id AS id,
	expenses.user_id AS user_id,
	expenses.category_id AS category_id,
	expenses.amount AS amount,
	CASE
		WHEN food_receipts.store_name IS NULL then ''
		ELSE food_receipts.store_name
	END AS store_name,
	expenses.comment AS comment,
	expenses.created_at AS created_at
FROM expenses
LEFT OUTER JOIN food_receipts ON expenses.food_receipt_id = food_receipts.id
INNER JOIN categories ON expenses.category_id = categories.id
WHERE expenses.user_id = $1;
