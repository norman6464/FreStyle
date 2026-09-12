-- name: GetProfileByUserID :one
-- user_id でプロフィールを 1 件取得（無ければ usecase 側で空表示にフォールバック）。
SELECT * FROM profiles
WHERE user_id = $1;

-- name: UpsertProfile :one
-- user_id 単位のプロフィール upsert。GORM Save（PK=user_id の INSERT-or-UPDATE）を
-- ON CONFLICT (user_id) DO UPDATE で置き換える。updated_at に DB 既定値は無いので
-- （GORM autoUpdateTime 依存）now() をクエリ側で明示し、更新時も now() へ進める。
-- RETURNING で updated_at を書き戻す（GORM Save 相当）。
--
-- status_emoji / status_expires_at には触れない（UpsertProfileStatus の専管。列を
-- INSERT の列挙から外しているので、新規行では DB 既定値（空文字／NULL）が入り、既存行では
-- ON CONFLICT の対象外なので値が保たれる）。
INSERT INTO profiles (user_id, bio, avatar_url, status_text, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (user_id) DO UPDATE SET
  bio         = EXCLUDED.bio,
  avatar_url  = EXCLUDED.avatar_url,
  status_text = EXCLUDED.status_text,
  updated_at  = now()
RETURNING updated_at;

-- name: UpsertProfileStatus :one
-- 一言ステータス（絵文字・テキスト・失効時刻）だけの upsert（段 14。PUT /me/status 用）。
-- bio / avatar_url には触れない（UpsertProfile の専管。同じ理由で列挙から外す）。
INSERT INTO profiles (user_id, status_emoji, status_text, status_expires_at, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (user_id) DO UPDATE SET
  status_emoji      = EXCLUDED.status_emoji,
  status_text       = EXCLUDED.status_text,
  status_expires_at = EXCLUDED.status_expires_at,
  updated_at        = now()
RETURNING user_id, bio, avatar_url, status_text, status_emoji, status_expires_at, updated_at;
