-- name: ListTransactions :many
SELECT * FROM transactions WHERE user_id = $1 ORDER BY date DESC, id DESC;

-- name: CreateTransaction :one
INSERT INTO transactions (user_id, amount, description, category, type, date, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DeleteTransaction :execrows
DELETE FROM transactions WHERE id = $1 AND user_id = $2;

-- name: SumTransactionsBetween :one
SELECT COALESCE(SUM(amount) FILTER (WHERE type = 'INCOME'), 0)::numeric  AS income,
       COALESCE(SUM(amount) FILTER (WHERE type = 'EXPENSE'), 0)::numeric AS expenses
FROM transactions
WHERE user_id = sqlc.arg(user_id)
  AND date >= sqlc.arg(from_date) AND date <= sqlc.arg(to_date);

-- name: MonthlyTransactionSums :many
SELECT EXTRACT(YEAR FROM date)::int  AS year,
       EXTRACT(MONTH FROM date)::int AS month,
       COALESCE(SUM(amount) FILTER (WHERE type = 'INCOME'), 0)::numeric  AS income,
       COALESCE(SUM(amount) FILTER (WHERE type = 'EXPENSE'), 0)::numeric AS expenses
FROM transactions
WHERE user_id = sqlc.arg(user_id)
  AND date >= sqlc.arg(from_date) AND date <= sqlc.arg(to_date)
GROUP BY 1, 2;
