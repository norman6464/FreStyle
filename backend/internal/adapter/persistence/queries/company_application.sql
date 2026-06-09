-- name: ListCompanyApplications :many
-- 利用申請一覧（新しい順）。super_admin が確認する。
SELECT * FROM company_applications
ORDER BY created_at DESC;
