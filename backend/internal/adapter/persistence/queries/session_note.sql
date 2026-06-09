-- name: GetSessionNoteBySessionID :one
-- セッション ID でメモを 1 件取得（所有者検証は usecase 側）。
SELECT * FROM session_notes
WHERE session_id = $1;
