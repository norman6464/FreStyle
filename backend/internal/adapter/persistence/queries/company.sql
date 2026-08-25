-- name: ListCompanies :many
-- 企業一覧（名前昇順）。name に一意制約は無いので同名企業の順序を固定する id ASC を付ける。
SELECT * FROM companies
ORDER BY name ASC, id ASC;

-- name: GetCompanyByID :one
-- ID で企業を 1 件取得。
SELECT * FROM companies
WHERE id = $1;
