-- name: ListCompanyApplications :many
-- 利用申請一覧（新しい順）。super_admin が確認する。
-- created_at は一意でないため、同時刻の順序を固定する id DESC をタイブレークに付ける。
SELECT * FROM company_applications
ORDER BY created_at DESC, id DESC;
