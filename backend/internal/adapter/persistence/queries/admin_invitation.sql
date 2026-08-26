-- name: ListPendingInvitations :many
-- 全社横断で pending の招待を返す（SuperAdmin 用）。物理削除はせず status のみ更新するため
-- accepted / canceled は WHERE で除外する。created_at は一意でないため id DESC をタイブレークに置く。
SELECT id, company_id, email, role, name, status, token, expires_at, created_at
FROM invitations
WHERE status = sqlc.arg(status)
ORDER BY created_at DESC, id DESC;

-- name: ListPendingInvitationsByCompany :many
-- 自社の pending 招待のみ返す（CompanyAdmin 用）。
SELECT id, company_id, email, role, name, status, token, expires_at, created_at
FROM invitations
WHERE company_id = sqlc.arg(company_id) AND status = sqlc.arg(status)
ORDER BY created_at DESC, id DESC;

-- name: FindPendingInvitationByEmail :one
-- pending の招待を email で引く（受諾フロー判定用）。突き合わせは domain.NormalizeEmail と同じ
-- 正規形どうしで行う。引数は Go 側で畳み、列は users の一意索引と同じ SQL 式
-- lower(btrim(email, EmailTrimCutset)) で畳む。expires は問わない（pending なら期限切れでも返す）。
-- 1 件しか返さないため、順序が揺れると「どの招待が受理されるか」が変わる。created_at DESC, id DESC で固定。
SELECT id, company_id, email, role, name, status, token, expires_at, created_at
FROM invitations
WHERE lower(btrim(email, E'\t\n\x0B\f\r ')) = sqlc.arg(email_normal)
  AND status = sqlc.arg(status)
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: FindInvitationByID :one
-- ID 一致の招待を返す（会社スコープの認可判定に使う。status は問わない）。
SELECT id, company_id, email, role, name, status, token, expires_at, created_at
FROM invitations
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: FindPendingInvitationByToken :one
-- token 一致 かつ pending かつ 未期限切れの招待のみ返す。期限比較は DB 関数でなく
-- Go の UTC 現在時刻をバインドする（DB エンジン非依存 / ローカル TZ 設定に左右されない）。
-- token は UNIQUE なので 1 行だが、順序を固定するため id ASC を置く。
SELECT id, company_id, email, role, name, status, token, expires_at, created_at
FROM invitations
WHERE token = sqlc.arg(token)
  AND status = sqlc.arg(status)
  AND expires_at > sqlc.arg(now)
ORDER BY id ASC
LIMIT 1;

-- name: InsertInvitation :one
-- 招待行を保存する。email は呼び出し側が正規形へ畳んで渡す。id は採番列なので省き
-- RETURNING で id / created_at を書き戻す。created_at は DB 既定値が無いため呼び出し側が渡す
-- （GORM autoCreateTime 相当。ゼロなら now を入れる）。updated_at 列は持たない。
-- token は未設定を NULL にして UNIQUE を避けるため nullable。
INSERT INTO invitations
  (company_id, email, role, name, status, token, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at;

-- name: UpdateInvitationStatus :exec
-- 招待の status のみ更新する（accepted / canceled への遷移。物理削除はしない）。
-- 対象は id 一致の 1 行だけで、他の列は触らない。
UPDATE invitations SET
  status = sqlc.arg(status)
WHERE id = sqlc.arg(id);
