-- name: GetSessionNoteBySessionID :one
-- セッション ID でメモを 1 件取得（所有者検証は usecase 側）。
-- session_id に一意制約は無い（domain は index）。万一同一 session_id が複数あっても結果が
-- 決定的になるよう id ASC で最古を返す（GORM First の既定挙動と一致）。
SELECT * FROM session_notes
WHERE session_id = $1
ORDER BY id ASC;
