-- page_versions（ページ本文の明示的なスナップショット履歴）のクエリ。
--
-- 採番・間引き・掃除の判定はすべて Go 側（PageVersionRepository.CreateVersionIfDue）が行う。
-- ここに並ぶクエリはその手順 1 つずつの実体で、単独では「版を残すべきか」の意味を持たない。

-- name: LockPageForVersioning :one
-- 版の書き込み（本文保存経路・「版を残す」・復元のすべて）が最初に呼ぶ。この 1 行をロックする
-- ことで、同じページへの同時書き込みはここで直列化され、page_versions の PK (page_id, seq) の
-- 衝突が原理的に起きない（usecase/repository/page_version.go の CreateVersionIfDue ドキュメント
-- 参照）。対象が無ければ呼び出し側は sql.ErrNoRows を repository.ErrPageNotFound へ翻訳する。
SELECT id FROM pages
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(page_id)
FOR UPDATE;

-- name: GetLatestPageVersion :one
-- そのページの直近の版（seq 降順 1 件）。PK (page_id, seq) の btree だけで素引きできる
-- （schema.hcl の page_versions のコメント参照）。版が 1 つも無ければ sql.ErrNoRows —
-- これは「まだ版が無い」という正常な状態で、呼び出し側はエラーとして扱わない。
SELECT * FROM page_versions
WHERE workspace_id = sqlc.arg(workspace_id) AND page_id = sqlc.arg(page_id)
ORDER BY seq DESC
LIMIT 1;

-- name: CreatePageVersion :one
-- 版を 1 件挿入する。seq は呼び出し側（Go）が LockPageForVersioning でロックした後に
-- GetLatestPageVersion の結果 + 1（無ければ 1）で計算して渡す — SQL 側では採番しない
-- （採番を SQL の DEFAULT やシーケンスに任せると、ロックの外で採番が起きて衝突しうる）。
-- created_at も列の DEFAULT now()（トランザクション開始時刻で固定される）には頼らず、
-- 呼び出し側が渡す time.Now() を明示的に書く。DeleteOldPageVersions の cutoff も同じ
-- time.Now() 由来の値を使うため、両者の時刻の出どころを実際に揃えるにはここも Go 側で
-- 決めた値でなければならない（列の DEFAULT はこの経路を通らない他の書き込みのための保険として残す）。
INSERT INTO page_versions (workspace_id, page_id, seq, doc, author_user_id, note, created_at)
VALUES (sqlc.arg(workspace_id), sqlc.arg(page_id), sqlc.arg(seq), sqlc.arg(doc), sqlc.arg(author_user_id), sqlc.narg(note), sqlc.arg(created_at))
RETURNING *;

-- name: DeleteOldPageVersions :exec
-- 30 日保持の掃除。CreatePageVersion と同じトランザクションで、挿入の直後に呼ぶ。
-- cutoff は Go 側の time.Now().Add(-30*24*time.Hour) を渡す（DB の now() には頼らない —
-- 10 分規則の判定も Go 側の time.Now() を使っており、両者の時刻の出どころを揃えるため）。
-- 今挿入した行は created_at が cutoff より新しいので、この DELETE の対象にはならない。
--
-- NOT EXISTS で除いているのは、open な提案（page_suggestions）が base_seq として参照している
-- 版。base_seq の FK は NO_ACTION なので、この DELETE がそれを消そうとすると FK 違反になり、
-- 同じトランザクションで今まさに進行中の本文保存自体が丸ごと失敗する。採用・却下で
-- status が open でなくなった提案の参照はここで無視してよい（対象から外れる＝掃除される）。
DELETE FROM page_versions
WHERE page_versions.workspace_id = sqlc.arg(workspace_id) AND page_versions.page_id = sqlc.arg(page_id)
  AND page_versions.created_at < sqlc.arg(cutoff)
  AND NOT EXISTS (
    SELECT 1 FROM page_suggestions
    WHERE page_suggestions.page_id = page_versions.page_id
      AND page_suggestions.base_seq = page_versions.seq
      AND page_suggestions.status = 'open'
  );

-- name: ListPageVersions :many
-- 版一覧。seq 降順・上限 5000 件（defensive な LIMIT。ページネーションは今回作らない —
-- CodeRabbit指摘だが、design ticketにも無い範囲であり、doc を含まない軽量な行なので
-- 一旦この上限で様子を見る判断とした）。30日保持 × 10分規則の理論上の最大件数
-- （24h/10min × 30日 = 4320）に余裕を持たせた値。これを超える書き込み頻度が実際に
-- 観測されたらページネーションを足す — repository.PageVersionRepository の
-- ListVersions のコメント参照。
SELECT * FROM page_versions
WHERE workspace_id = sqlc.arg(workspace_id) AND page_id = sqlc.arg(page_id)
ORDER BY seq DESC
LIMIT 5000;

-- name: GetPageVersion :one
-- 版 1 件の取得。workspace_id まで絞ることで、他ページ・他テナントの seq を渡されても
-- 見つからない（＝存在しないのと同じ）ようにする。
SELECT * FROM page_versions
WHERE workspace_id = sqlc.arg(workspace_id) AND page_id = sqlc.arg(page_id) AND seq = sqlc.arg(seq);
