-- name: ListCompanyApplications :many
-- 利用申請一覧（新しい順）。super_admin が確認する。
-- created_at は一意でないため、同時刻の順序を固定する id DESC をタイブレークに付ける。
SELECT * FROM company_applications
ORDER BY created_at DESC, id DESC;

-- name: InsertCompanyApplication :one
-- 利用申請を 1 件作成する。created_at / updated_at は DB 既定値が無いため呼び出し側が
-- 値を渡す（GORM autoCreateTime/autoUpdateTime 相当。ゼロなら呼び出し側で now() を入れる）。
-- RETURNING で id / created_at / updated_at を書き戻す。
INSERT INTO company_applications
  (company_name, applicant_name, email, message, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at;

-- name: UpdateCompanyApplicationStatus :execrows
-- 申請の status を更新する（super_admin 専用）。updated_at は now() へ進める。
--
-- :exec ではなく :execrows にしている理由:
--   :exec は「SQL がエラーなく流れたか」しか返さない。UPDATE は 1 行も一致しなくても
--   成功なので、存在しない id を渡しても呼び出し側には成功として見える。承認/却下の記録が
--   どこにも残っていないのに管理画面では処理済みに見える、という取り違えが起きる。
--   :execrows は実際に書き換わった行数（RowsAffected）を返すので 0 行を not-found にできる。
UPDATE company_applications SET
  status     = $2,
  updated_at = now()
WHERE id = $1;
