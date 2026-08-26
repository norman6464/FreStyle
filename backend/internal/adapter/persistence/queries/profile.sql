-- name: GetProfileByUserID :one
-- user_id でプロフィールを 1 件取得（無ければ usecase 側で空表示にフォールバック）。
SELECT * FROM profiles
WHERE user_id = $1;

-- name: UpsertProfile :one
-- user_id 単位のプロフィール upsert。GORM Save（PK=user_id の INSERT-or-UPDATE）を
-- ON CONFLICT (user_id) DO UPDATE で置き換える。updated_at に DB 既定値は無いので
-- （GORM autoUpdateTime 依存）now() をクエリ側で明示し、更新時も now() へ進める。
-- RETURNING で updated_at を書き戻す（GORM Save 相当）。
INSERT INTO profiles (user_id, bio, avatar_url, status_message, updated_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (user_id) DO UPDATE SET
  bio            = EXCLUDED.bio,
  avatar_url     = EXCLUDED.avatar_url,
  status_message = EXCLUDED.status_message,
  updated_at     = now()
RETURNING updated_at;
