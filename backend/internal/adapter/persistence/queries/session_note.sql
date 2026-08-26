-- name: GetSessionNoteBySessionID :one
-- セッション ID でメモを 1 件取得（所有者検証は usecase 側）。
-- session_id は一意（uq_session_notes_session_id）なので最大 1 行。id ASC は保険。
SELECT * FROM session_notes
WHERE session_id = $1
ORDER BY id ASC;

-- name: UpsertSessionNote :one
-- セッションメモを 1 セッション 1 件で upsert する。session_id の一意制約に当てて
-- ON CONFLICT DO UPDATE し、既存行があれば content と updated_at だけを更新する。
--
-- 保持列: created_at は初回作成時刻を保つため上書きしない。user_id も同じセッションの
-- 所有者は不変なので触らない（DO UPDATE の SET に含めない列は既存値のまま残る）。
-- 時刻列は DB 既定値が無いため、INSERT では now() を明示し、UPDATE では updated_at に now() を入れる。
INSERT INTO session_notes (session_id, user_id, content, created_at, updated_at)
VALUES ($1, $2, $3, now(), now())
ON CONFLICT (session_id) DO UPDATE SET
    content = EXCLUDED.content,
    updated_at = now()
RETURNING id, created_at, updated_at;
