-- name: ListCompanies :many
-- 企業一覧（名前昇順）。
SELECT * FROM companies
ORDER BY name ASC;

-- name: GetCompanyByID :one
-- ID で企業を 1 件取得。
SELECT * FROM companies
WHERE id = $1;
