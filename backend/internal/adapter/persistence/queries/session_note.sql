-- name: GetSessionNoteBySessionID :one
-- セッション ID でメモを 1 件取得（所有者検証は usecase 側）。
-- session_id に一意制約は無い（domain は index）。万一同一 session_id が複数あっても結果が
-- 決定的になるよう id ASC で最古を返す（GORM First の既定挙動と一致）。
SELECT * FROM session_notes
WHERE session_id = $1
ORDER BY id ASC;

-- name: InsertSessionNote :one
-- セッションメモを 1 件作成する。id は採番列（bigserial）に任せ、created_at / updated_at は
-- DB 既定値が無いため now() を明示する。RETURNING で id / created_at / updated_at を書き戻す。
--
-- 注意: repository の口は Upsert だが、session_id には一意制約が無いため ON CONFLICT
-- (session_id) は張れない（GORM Save(ID=0) も常に INSERT する現行挙動）。ここでは現行の
-- 振る舞いを厳密に保つため INSERT のみとする。session_id 単位の真の upsert にするには
-- session_id への一意制約と既存重複行の解消が別途必要（本移行の対象外）。
INSERT INTO session_notes (session_id, user_id, content, created_at, updated_at)
VALUES ($1, $2, $3, now(), now())
RETURNING id, created_at, updated_at;
