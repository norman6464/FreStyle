-- name: ListAiChatSessionsByUserID :many
-- 自分の AI チャットセッションを新しい順で返す。
-- created_at は一意でない。同時刻セッションの順序を固定するため id をタイブレークに置く。
SELECT id, user_id, title, session_type, scenario_id, created_at, updated_at
FROM ai_chat_sessions
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC, id DESC;

-- name: GetAiChatSessionByID :one
-- 内部 ID で 1 件取得（存在しなければ sql.ErrNoRows）。所有者検証は usecase 側。
SELECT id, user_id, title, session_type, scenario_id, created_at, updated_at
FROM ai_chat_sessions
WHERE id = sqlc.arg(id);

-- name: InsertAiChatSession :one
-- セッションを 1 件作成する。id は採番列なので省き RETURNING で id / 時刻を書き戻す。
-- created_at / updated_at は DB 既定値が無いため呼び出し側が値を渡す（GORM の autoTime 相当。
-- ゼロなら呼び出し側が now() を入れる）。scenario_id は NULL 可（練習シナリオ紐付けが無いセッション）。
INSERT INTO ai_chat_sessions (user_id, title, session_type, scenario_id, created_at, updated_at)
VALUES (
  sqlc.arg(user_id),
  sqlc.arg(title),
  sqlc.arg(session_type),
  sqlc.narg(scenario_id),
  sqlc.arg(created_at),
  sqlc.arg(updated_at)
)
RETURNING id, created_at, updated_at;

-- name: UpdateAiChatSessionTitle :execrows
-- セッションのタイトルを更新する。書くのは title と updated_at の 2 列だけで、
-- user_id / session_type / scenario_id / created_at は不変。
-- updated_at は DB 既定値が無いため now() を明示する（autoUpdateTime 相当）。
--
-- :exec ではなく :execrows にしている理由:
--   :exec は「SQL がエラーなく流れたか」しか返さない。UPDATE は 1 行も一致しなくても
--   成功なので、存在しない id を渡しても呼び出し側には成功として見える。
--   :execrows は実際に書き換わった行数（RowsAffected）を返すので、repository が 0 行を
--   「対象なし」として domain.ErrNotFound に翻訳できる。
UPDATE ai_chat_sessions SET
  title      = sqlc.arg(title),
  updated_at = now()
WHERE id = sqlc.arg(id);

-- name: DeleteAiChatSession :execrows
-- セッションを物理削除する（ai_chat_sessions は soft delete 列を持たない）。
-- UpdateAiChatSessionTitle と同じ理由で :execrows（0 行削除を成功と区別するため）。
DELETE FROM ai_chat_sessions WHERE id = sqlc.arg(id);
