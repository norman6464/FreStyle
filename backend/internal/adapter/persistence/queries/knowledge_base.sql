-- ノート（workspaces / spaces / pages / blocks / page_paths / page_snapshots）のクエリ。
--
-- 作法（このファイル全体の前提）:
--   - すべての SELECT / UPDATE / DELETE の WHERE に workspace_id を含める。
--     DB の複合 FK が守るのは「親子の整合」までで、テナント越えの読み書きは
--     クエリレベルで塞ぐ（page_snapshots のように workspace_id を持たない表は pages と JOIN する）。
--   - UPDATE 文には必ず updated_at = now() を明示する。GORM を通さないため自動更新が無く、
--     忘れると snapshot の鮮度判定（built_at との比較）が将来壊れる。
--   - 並び順（position）は COLLATE "C" の列なので ORDER BY はバイト順になり、
--     fracindex（Go 側のバイト比較）と一致する。

-- name: GetWorkspaceByID :one
-- ワークスペースの存在確認（テナント検証の入口）。
SELECT * FROM workspaces
WHERE id = $1;

-- name: GetWorkspaceBySlug :one
-- URL に出る slug からワークスペースを引く（HTTP 層のテナント解決の入口）。
-- slug はグローバルに一意（uq_workspaces_slug）なので workspace_id での絞り込みは要らない。
SELECT * FROM workspaces
WHERE slug = $1;

-- name: InsertWorkspace :one
-- ワークスペースの作成。slug はグローバルに一意（uq_workspaces_slug）なので、
-- 重複は一意制約違反として返り、repository が「その slug は使用済み」へ翻訳する。
-- personal_owner_user_id は個人ワークスペースだけ非 NULL（uq_workspaces_personal_owner で
-- 1 人 1 つ）。通常のチームワークスペースは NULL のまま渡す。
INSERT INTO workspaces (id, slug, name, personal_owner_user_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPersonalWorkspaceByOwner :one
-- 個人ワークスペースを持ち主から引く。サインアップの「作る前に既に在るか見る」に使う
-- （uq_workspaces_personal_owner が 1 人 1 つを守るので、あれば必ず 1 行）。
SELECT * FROM workspaces
WHERE personal_owner_user_id = $1;

-- name: InsertSpace :one
-- スペースの作成。key はワークスペース内で一意（uq_spaces_workspace_key）。
-- visibility の値は domain.SpaceVisibility が正（'workspace' / 'private'）。
INSERT INTO spaces (id, workspace_id, "key", name, visibility)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteWorkspace :execrows
-- ワークスペースを消す。**そこに所属している人がいるものは消さない**（WHERE で弾く）。
--
-- 配下（spaces / pages / blocks / page_paths / principals / grants / 例外 / 共有リンク）は
-- すべて workspaces への FK が ON DELETE CASCADE で連なっているので、この 1 文で消える。
--
-- 人が居るワークスペースを守るのは、そこに全員のノートが入るため。1 人の操作でみんなの
-- 資産が消えてよいはずがない。users.workspace_id の FK（fk_users_workspace）も同じ削除を
-- 物理的に止めるが、それに任せると理由が制約違反エラーの文面でしか分からない。WHERE で
-- 明示して 0 行で返し、呼び出し側が「無い」と「人が居る」を撃ち分けられるようにする。
--
-- 判定は users を見る。呼び出し側の引数に頼ると、そこを間違えたときに守りが消える。
DELETE FROM workspaces w
WHERE w.id = $1
  AND NOT EXISTS (
      SELECT 1 FROM users u
      WHERE u.workspace_id = w.id
  );

-- name: GetSpace :one
-- スペースの存在確認。workspace_id を含めることで別テナントのスペース ID を弾く。
SELECT * FROM spaces
WHERE workspace_id = $1 AND id = $2;

-- name: InsertPage :one
-- ページの作成。created_at / updated_at / archived_at は DB 既定値に任せ、
-- RETURNING で確定した行を返す（アプリ側の時刻と DB の時刻を二重管理しない）。
INSERT INTO pages (id, workspace_id, space_id, parent_id, "position", title, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPage :one
-- ページを 1 件取得（アーカイブ済みも返す。現役かどうかの判断は usecase 側）。
SELECT * FROM pages
WHERE workspace_id = $1 AND id = $2;

-- name: GetPageAcrossWorkspaces :one
-- ページを **ID だけ** で引く。/p/{pageId} の URL からワークスペースを特定するための、
-- このファイルで唯一 workspace_id を WHERE に持たない読み取り。
-- 引いた直後に必ずその workspace の権限判定を通すこと（判定なしで応答に使わない）。
-- id は uuid の主キーで全テナント一意なので、これ自体が越境にはならない
-- （危ういのは結果の使い方で、それは呼び出し側の usecase が縛る）。
SELECT * FROM pages
WHERE id = $1;

-- name: ListChildPages :many
-- 指定ページ直下の現役の子ページ一覧（position 順 = 表示順）。
SELECT * FROM pages
WHERE workspace_id = $1 AND parent_id = $2 AND archived_at IS NULL
ORDER BY "position";

-- name: UpdateSpaceName :execrows
-- スペースの表示名だけを変える。key は URL・権限の参照に使われるので触らない。
-- 0 件なら対象が存在しない（別ワークスペースの ID もここで 0 件になる —
-- workspace_id を WHERE に含める作法はこのファイルの先頭コメントのとおり）。
UPDATE spaces SET name = $3, updated_at = now()
WHERE workspace_id = $1 AND id = $2;

-- name: ListActivePagesBySpace :many
-- スペース配下の現役ページ全件（ツリー構築用）。position はバイト順なので、
-- 同じ親を持つページ同士はこの並びのまま兄弟順になる（木への組み立ては Go 側）。
SELECT * FROM pages
WHERE workspace_id = $1 AND space_id = $2 AND archived_at IS NULL
ORDER BY "position";

-- name: GetLastActiveSiblingPosition :one
-- 兄弟（同じ親、ルートなら同じスペース直下）の末尾 position。末尾追加の採番
-- fracindex.Between(末尾, "") に使う。parent_id は NULL 可のため IS NOT DISTINCT FROM で比較する。
SELECT "position" FROM pages
WHERE workspace_id = sqlc.arg(workspace_id)
  AND space_id = sqlc.arg(space_id)
  AND parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id)
  AND archived_at IS NULL
ORDER BY "position" DESC
LIMIT 1;

-- name: SiblingPositionsAround :one
-- 「ある兄弟のすぐ前／すぐ後ろに入れる」ための、前後の並び順キーを返す。
--
-- ドラッグで落とした位置を表すのに使う。**クライアントは並び順のキーを持たない**
-- （応答に入れていない。整数部が兄弟の通し番号になるので、飛びから伏せた枚数が読めるため）。
-- 代わりに「どの兄弟の隣か」をページの ID で受け取り、キーの計算はここから先で行う。
--
-- 3 つとも NULL なら、anchor_page_id はその親の現役の子ではない（存在しない・別の親・
-- 別スペース・アーカイブ済みのいずれか）。呼び出し側はまとめて「兄弟ではない」に落とす。
--
-- moving_page_id を必ず除くのは、動かす当人がまだその並びに居るため。除かないと
-- 自分自身との中間値を計算することになり、動かないか、隣とキーが衝突する。
--
-- 伏せられている兄弟も隣人として数える。利用者からは見えないが並びには居るので、
-- 除くとキーが既存の行と衝突する。どのキーになったかは応答に出ないので漏れない。
WITH anchor AS (
    SELECT a."position"
    FROM pages a
    WHERE a.workspace_id = sqlc.arg(workspace_id)
      AND a.id = sqlc.arg(anchor_page_id)
      AND a.space_id = sqlc.arg(space_id)
      AND a.parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id)
      AND a.archived_at IS NULL
      AND a.id <> sqlc.arg(moving_page_id)
)
-- 端が無いことは NULL ではなく空文字で表す。fracindex.Between が「空文字＝そちら側に
-- 隣がいない」という約束なので、そのまま渡せる形に揃える。
-- anchor が兄弟だったかは found で別に返す（空文字だけでは「先頭に入れる」と
-- 「兄弟ではない」を区別できない）。
SELECT
    EXISTS (SELECT 1 FROM anchor) AS found,
    COALESCE((SELECT anchor."position" FROM anchor), '')::text AS anchor_position,
    COALESCE((SELECT max(p."position")
       FROM pages p, anchor
      WHERE p.workspace_id = sqlc.arg(workspace_id)
        AND p.space_id = sqlc.arg(space_id)
        AND p.parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id)
        AND p.archived_at IS NULL
        AND p.id <> sqlc.arg(moving_page_id)
        AND p."position" < anchor."position"), '')::text AS prev_position,
    COALESCE((SELECT min(p."position")
       FROM pages p, anchor
      WHERE p.workspace_id = sqlc.arg(workspace_id)
        AND p.space_id = sqlc.arg(space_id)
        AND p.parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id)
        AND p.archived_at IS NULL
        AND p.id <> sqlc.arg(moving_page_id)
        AND p."position" > anchor."position"), '')::text AS next_position;

-- name: HasActiveSiblingPosition :one
-- 指定 position を持つ現役の兄弟が既にいるか（アーカイブ復帰時の衝突検出用）。
-- 部分 UNIQUE（uq_pages_parent_position / uq_pages_space_position）は現役だけを守るため、
-- 復帰する行自身（excluded_page_id）を除いて衝突を先に調べる。
SELECT EXISTS (
    SELECT 1 FROM pages
    WHERE workspace_id = sqlc.arg(workspace_id)
      AND space_id = sqlc.arg(space_id)
      AND parent_id IS NOT DISTINCT FROM sqlc.narg(parent_id)
      AND "position" = sqlc.arg(position)
      AND archived_at IS NULL
      AND id <> sqlc.arg(excluded_page_id)
) AS conflicted;

-- name: UpdatePageTitle :one
-- タイトル変更。RETURNING で更新後の行を返す。
UPDATE pages
SET title = $3, updated_at = now()
WHERE workspace_id = $1 AND id = $2
RETURNING *;

-- name: SetPagePosition :execrows
-- position の振り直し（アーカイブ復帰で衝突したときの末尾再採番用）。
UPDATE pages
SET "position" = $3, updated_at = now()
WHERE workspace_id = $1 AND id = $2;

-- name: MovePageWithinSpace :execrows
-- 同一スペース内の移動（親と position の付け替え）。space_id が変わらないので
-- 子孫には触らない（子孫の parent_id / space_id はそのままで整合が保たれる）。
UPDATE pages
SET parent_id = sqlc.narg(new_parent_id),
    "position" = sqlc.arg(new_position),
    updated_at = now()
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(page_id);

-- name: MovePageSubtreeToSpace :execrows
-- スペースをまたぐ移動。移動するページ自身（親・position・space_id）と
-- 子孫全員（space_id のみ）を 1 文で更新する。
--
-- 1 文であることは省略できない: fk_pages_parent は「親は同じスペースのページ」を要求し、
-- FK（NO ACTION）は文の終わりに検査される。ページだけ先に動かすと子の FK が、
-- 子孫だけ先に動かすと子孫自身の FK が、それぞれ文末で違反になる。
-- サブツリーの特定は page_paths（ancestor_id = 移動ページ。depth=0 の自分自身も含む）。
-- アーカイブ済みの子孫も FK の対象なので除外しない。
UPDATE pages
SET space_id = sqlc.arg(new_space_id),
    parent_id = CASE WHEN pages.id = sqlc.arg(page_id) THEN sqlc.narg(new_parent_id) ELSE pages.parent_id END,
    "position" = CASE WHEN pages.id = sqlc.arg(page_id) THEN sqlc.arg(new_position) ELSE pages."position" END,
    updated_at = now()
WHERE pages.workspace_id = sqlc.arg(workspace_id)
  AND pages.id IN (
      SELECT pp.page_id FROM page_paths pp
      WHERE pp.workspace_id = sqlc.arg(workspace_id) AND pp.ancestor_id = sqlc.arg(page_id)
  );

-- name: ArchivePageSubtree :execrows
-- サブツリーごとアーカイブ。now() はトランザクションのタイムスタンプなので、
-- 1 回の実行で archived_at が全行同じ値になる（復帰時に「この一括操作で archive された行」を
-- archived_at >= 根の archived_at で特定できる）。既にアーカイブ済みの行は
-- 元の archived_at を保つため触らない。
UPDATE pages
SET archived_at = now(), updated_at = now()
WHERE pages.workspace_id = $1
  AND pages.archived_at IS NULL
  AND pages.id IN (
      SELECT pp.page_id FROM page_paths pp
      WHERE pp.workspace_id = $1 AND pp.ancestor_id = $2
  );

-- name: UnarchivePageSubtree :execrows
-- アーカイブ解除。根の archived_at（archived_since）以降にアーカイブされた行だけを戻す。
-- サブツリー全行を無条件に戻すと、根より前に個別アーカイブされていた子まで復帰し、
-- そのあいだに同じ position で作られた現役の兄弟と部分 UNIQUE が衝突しうる。
-- 「一緒にアーカイブされた一括分だけを戻す」ことで衝突面を根の 1 行に閉じる
-- （根の衝突は usecase が SetPagePosition で末尾へ再採番してから解除する）。
UPDATE pages
SET archived_at = NULL, updated_at = now()
WHERE pages.workspace_id = sqlc.arg(workspace_id)
  AND pages.archived_at >= sqlc.arg(archived_since)
  AND pages.id IN (
      SELECT pp.page_id FROM page_paths pp
      WHERE pp.workspace_id = sqlc.arg(workspace_id) AND pp.ancestor_id = sqlc.arg(page_id)
  );

-- name: InsertPagePathSelf :exec
-- closure の自己参照行（depth=0）。ページ作成と同じトランザクションで張る。
INSERT INTO page_paths (workspace_id, page_id, ancestor_id, depth)
VALUES ($1, $2, $2, 0);

-- name: InsertPagePathAncestors :exec
-- ページ作成時に親の祖先集合（親自身 depth=0 を含む）を +1 して引き継ぐ。
INSERT INTO page_paths (workspace_id, page_id, ancestor_id, depth)
SELECT pp.workspace_id, sqlc.arg(page_id)::uuid, pp.ancestor_id, pp.depth + 1
FROM page_paths pp
WHERE pp.workspace_id = sqlc.arg(workspace_id) AND pp.page_id = sqlc.arg(parent_id);

-- name: PageHasDescendant :one
-- descendant_id が page_id の子孫（depth=0 の自分自身を含む）かどうか。移動時の循環検出に使う。
SELECT EXISTS (
    SELECT 1 FROM page_paths
    WHERE workspace_id = $1 AND ancestor_id = $2 AND page_id = $3
) AS found;

-- name: DetachPageSubtreePaths :exec
-- 移動時の closure 付け替え（前半）: サブツリー内の各ページと「サブツリー外の祖先」との組を消す。
-- サブツリー内部同士の組（自己参照 depth=0 を含む）は移動後も変わらないため残す。
DELETE FROM page_paths
WHERE page_paths.workspace_id = sqlc.arg(workspace_id)
  AND page_paths.page_id IN (
      SELECT pp.page_id FROM page_paths pp
      WHERE pp.workspace_id = sqlc.arg(workspace_id) AND pp.ancestor_id = sqlc.arg(page_id)
  )
  AND page_paths.ancestor_id NOT IN (
      SELECT pp.page_id FROM page_paths pp
      WHERE pp.workspace_id = sqlc.arg(workspace_id) AND pp.ancestor_id = sqlc.arg(page_id)
  );

-- name: AttachPageSubtreePaths :exec
-- 移動時の closure 付け替え（後半）: 新しい親の祖先集合（親自身を含む）×サブツリー全員の
-- 直積を張る。深さは「サブツリー内での深さ + 親までの深さ + 1」。
-- ルートへの移動（親なし）ではこのクエリは呼ばない（Detach だけで完結する）。
INSERT INTO page_paths (workspace_id, page_id, ancestor_id, depth)
SELECT sub.workspace_id, sub.page_id, sup.ancestor_id, sub.depth + sup.depth + 1
FROM page_paths sub
JOIN page_paths sup
  ON sup.workspace_id = sub.workspace_id AND sup.page_id = sqlc.arg(new_parent_id)
WHERE sub.workspace_id = sqlc.arg(workspace_id) AND sub.ancestor_id = sqlc.arg(page_id);

-- name: ListBlocksByPage :many
-- ページの全ブロック（doc への組み立て用）。position はバイト順なので同じ親を持つ
-- ブロック同士はこの並びのまま兄弟順になる。id は position が親違いで偶然一致したときの
-- 並びを決定的にするためのタイブレーク。
SELECT * FROM blocks
WHERE workspace_id = $1 AND page_id = $2
ORDER BY "position", id;

-- name: DeletePageBlocks :exec
-- ページの全ブロック削除（本文の書き換えは「全消し全入れ」。差分更新は将来の最適化）。
DELETE FROM blocks
WHERE workspace_id = $1 AND page_id = $2;

-- name: InsertBlock :exec
-- ブロック 1 行の挿入（全入れ替えの一括 INSERT で使う）。
INSERT INTO blocks (id, workspace_id, page_id, parent_id, "position", type, attrs, inline)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpsertPageSnapshot :exec
-- snapshot の焼き直し。blocks の全入れ替えと同じトランザクションで呼び、
-- 「snapshot は常に blocks と同期している」を保つ。
--
-- 衝突キー (page_id) に所有者列が入っていないが、この表に限っては安全。page_snapshots は
-- page_id / doc / built_at しか持たない導出キャッシュで、持ち主という概念がそもそも無い
-- （中身はいつでも blocks から焼き直せる）。誰のページなのかを決めるのは pages 側で、
-- 認可もそちらで行う。
--
-- そのため機械向けの免除コメント（-- upsert-owner-scope: …）は付けていない。付けてしまうと、
-- 将来この表に所有者列が入ったときに検査が黙って素通りする。検査は「所有者列を 1 つも
-- 持たない表には何も要求しない」という作りなので、この形のままで通る
-- （検査の実体は internal/adapter/persistence/queries_static_check_test.go）。
INSERT INTO page_snapshots (page_id, doc, built_at)
VALUES ($1, $2, now())
ON CONFLICT (page_id) DO UPDATE SET doc = EXCLUDED.doc, built_at = now();

-- name: GetPageSnapshot :one
-- snapshot の取得。page_snapshots は workspace_id を持たないため、pages と JOIN して
-- テナント検証をクエリレベルで行う（このファイル冒頭の作法）。
SELECT ps.* FROM page_snapshots ps
JOIN pages p ON p.id = ps.page_id
WHERE p.workspace_id = $1 AND ps.page_id = $2;

-- name: ListPageAncestorIDs :many
-- ページの祖先 ID を根から順（depth の大きい順）に返す。自分自身（depth=0）は含まない。
-- パンくず用の骨組みで、**題名や可視性はここでは返さない** — 可視の判定は
-- ListWorkspacePageViewFactsByIDs と domain.ResolvePageView が持つ（判定の写経をしない）。
SELECT pp.ancestor_id
FROM page_paths pp
WHERE pp.workspace_id = $1 AND pp.page_id = $2 AND pp.depth > 0
ORDER BY pp.depth DESC;

-- name: DeletePage :execrows
-- ページを物理削除する。子孫・closure（page_paths）・blocks・snapshot は
-- ON DELETE CASCADE で一緒に消える（fk_pages_parent が CASCADE のため、
-- この 1 文で部分木全体が落ちる）。アーカイブと違い戻せない — 子孫全員の
-- 編集権限の確認（requireSubtreeEditPermission）は呼び出し側の入口が行う。
DELETE FROM pages
WHERE workspace_id = $1 AND id = $2;
