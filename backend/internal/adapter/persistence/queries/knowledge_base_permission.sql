-- ナレッジ基盤の権限モデル（principals / principal_members / workspace_grants /
-- space_grants / page_restrictions / share_links）のクエリ。
--
-- 作法（knowledge_base.sql と同じ）:
--   - すべての SELECT / UPDATE / DELETE の WHERE に workspace_id を含める。
--     DB の複合 FK が守るのは「書き込み時に境界を越えられないこと」までで、
--     読み出しのテナント越えはクエリレベルで塞ぐ。
--   - UPDATE 文には必ず updated_at = now() を明示する（GORM を通さないため自動更新が無い）。
--
-- 権限解決の方針:
--   - SQL が返すのは「事実」（既定の役割・経路上の制限の集計）だけで、
--     どう組み合わせるかの規則は domain.ResolvePagePermission が持つ。
--     規則を SQL と Go の 2 か所に書くと、片方だけ直したときに
--     「1 ページを開くと見えるのに一覧には出ない」という直しにくいずれ方をする。
--   - 祖先をたどるのに再帰は使わない。page_paths（closure）を 1 回 JOIN して
--     経路上の制限をまとめて拾う。

-- name: InsertPrincipal :one
-- 主体の作成。使う列は kind で決まり、DB の CHECK が「その kind のときだけ非 NULL」を強制する。
INSERT INTO principals (id, workspace_id, kind, user_id, space_id, page_id, name)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPrincipal :one
-- 主体を 1 件取得。workspace_id を含めることで別テナントの principal ID を弾く。
SELECT * FROM principals
WHERE workspace_id = $1 AND id = $2;

-- name: GetUserPrincipal :one
-- ユーザーの主体（＝ ワークスペース所属そのもの）。無ければ「メンバーではない」。
SELECT * FROM principals
WHERE workspace_id = $1 AND kind = 'user' AND user_id = $2;

-- name: GetSpaceEveryonePrincipal :one
-- そのスペースの「全員」を表す主体。
SELECT * FROM principals
WHERE workspace_id = $1 AND kind = 'space_all' AND space_id = $2;

-- name: DeletePrincipal :execrows
-- 主体の削除。grant / restriction / principal_members は FK の CASCADE で一緒に消える
-- （権限だけが残って別人に引き継がれることが無いように、行を残さず消す）。
DELETE FROM principals
WHERE workspace_id = $1 AND id = $2;

-- name: IsWorkspaceMember :one
-- ワークスペース所属の判定。principals が所属の唯一の表現なので、この 1 行の有無がすべて。
SELECT EXISTS (
    SELECT 1 FROM principals
    WHERE workspace_id = $1 AND kind = 'user' AND user_id = $2
) AS is_member;

-- name: InsertPrincipalMember :exec
-- グループへの所属追加。既にあれば何もしない（冪等）。
-- group / member の kind は DB 側の生成列 + 複合 FK が固定するため、ここでは渡さない。
INSERT INTO principal_members (workspace_id, group_principal_id, member_principal_id)
VALUES ($1, $2, $3)
ON CONFLICT (group_principal_id, member_principal_id) DO NOTHING;

-- name: DeletePrincipalMember :execrows
-- グループからの所属削除。
DELETE FROM principal_members
WHERE workspace_id = $1 AND group_principal_id = $2 AND member_principal_id = $3;

-- name: UpsertWorkspaceGrant :one
-- ワークスペース全体での既定の役割の付与（同じ主体には 1 行だけ）。
INSERT INTO workspace_grants (workspace_id, principal_id, "role")
VALUES ($1, $2, $3)
ON CONFLICT (workspace_id, principal_id)
DO UPDATE SET "role" = EXCLUDED."role", updated_at = now()
RETURNING *;

-- name: DeleteWorkspaceGrant :execrows
-- ワークスペース全体での既定の役割の剥奪。
DELETE FROM workspace_grants
WHERE workspace_id = $1 AND principal_id = $2;

-- name: ListWorkspaceGrants :many
-- ワークスペースの grant 一覧（権限設定画面用）。
SELECT * FROM workspace_grants
WHERE workspace_id = $1
ORDER BY principal_id;

-- name: UpsertSpaceGrant :one
-- スペースでの既定の役割の付与（同じ主体には 1 行だけ）。
INSERT INTO space_grants (workspace_id, space_id, principal_id, "role")
VALUES ($1, $2, $3, $4)
ON CONFLICT (workspace_id, space_id, principal_id)
DO UPDATE SET "role" = EXCLUDED."role", updated_at = now()
RETURNING *;

-- name: DeleteSpaceGrant :execrows
-- スペースでの既定の役割の剥奪。
DELETE FROM space_grants
WHERE workspace_id = $1 AND space_id = $2 AND principal_id = $3;

-- name: ListSpaceGrants :many
-- スペースの grant 一覧（権限設定画面用）。
SELECT * FROM space_grants
WHERE workspace_id = $1 AND space_id = $2
ORDER BY principal_id;

-- name: UpsertPageRestriction :one
-- ページの例外の設定。同じ (ページ, 主体, ケイパビリティ) の行は 1 つだけなので、
-- allow と deny を入れ替えるときも行は増えない。
INSERT INTO page_restrictions (workspace_id, page_id, principal_id, capability, mode)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (workspace_id, page_id, principal_id, capability)
DO UPDATE SET mode = EXCLUDED.mode, updated_at = now()
RETURNING *;

-- name: DeletePageRestriction :execrows
-- ページの例外の解除。最後の 1 行が消えるとその段には制限が無くなり、
-- 解決は「より遠い祖先の制限」→「grant の既定」の順に戻る。
DELETE FROM page_restrictions
WHERE workspace_id = $1 AND page_id = $2 AND principal_id = $3 AND capability = $4;

-- name: ListPageRestrictions :many
-- そのページ自身に張られた例外の一覧（祖先から継承したものは含まない）。
SELECT * FROM page_restrictions
WHERE workspace_id = $1 AND page_id = $2
ORDER BY principal_id, capability;

-- name: InsertShareLink :one
-- 共有リンクの発行。principal（kind='share_link'）は同じトランザクションで先に作る。
INSERT INTO share_links (
    id, workspace_id, page_id, principal_id, capability,
    token_hash, password_hash, expires_at, created_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: RevokeShareLink :execrows
-- 共有リンクの失効。行は消さず revoked_at を立てる（誰がいつ止めたかを追えるように）。
-- 既に失効しているものは触らない（最初に止めた時刻を保つ）。
UPDATE share_links
SET revoked_at = now(), updated_at = now()
WHERE workspace_id = $1 AND id = $2 AND revoked_at IS NULL;

-- name: GetShareLinkByTokenHash :one
-- トークン（の SHA-256）から共有リンクを引く。期限・失効・パスワードの判定は呼び出し側で行う
-- （「トークンが違う」と「期限切れ」を同じ経路で扱うと、どちらなのかを利用者に返せない）。
SELECT * FROM share_links
WHERE token_hash = $1;

-- name: GetShareLink :one
-- 共有リンクを 1 件取得（失効操作の対象確認用）。
SELECT * FROM share_links
WHERE workspace_id = $1 AND id = $2;

-- name: ListPageShareLinks :many
-- そのページに発行された共有リンクの一覧（失効済みも含む）。
SELECT * FROM share_links
WHERE workspace_id = $1 AND page_id = $2
ORDER BY created_at DESC;

-- name: SubtreeHasForeignSpaceAllRestriction :one
-- 移動するサブツリー（自分自身 + 子孫）に「移動先スペース以外のスペース全員」宛ての例外が
-- あるか。KnowledgeBaseRepository.MovePage が同じトランザクションで使う。
--
-- space_all の主体は「そのスペースの全員」を表すため、スペースをまたぐ移動で行だけが残り
-- 評価されなくなる（権限解決は対象ページが今いるスペースの space_all しか主体に取らない）。
-- 行は権限設定画面に見えているのに実効は違う、という追跡困難なずれになり、
-- しかも allow は締め出す側・deny は開く側へ倒れるという非対称ができる。
-- 移動そのものを止めることで、緩む向きにも締まる向きにも黙って変わらないようにする。
SELECT EXISTS (
    SELECT 1
    FROM page_paths pp
    JOIN page_restrictions r
      ON r.workspace_id = pp.workspace_id AND r.page_id = pp.page_id
    JOIN principals pr
      ON pr.workspace_id = r.workspace_id AND pr.id = r.principal_id
    WHERE pp.workspace_id = sqlc.arg(workspace_id)
      AND pp.ancestor_id = sqlc.arg(page_id)
      AND pr.kind = 'space_all'
      AND pr.space_id IS DISTINCT FROM sqlc.arg(new_space_id)::uuid
) AS found;

-- name: ResolvePagePermissionFacts :one
-- 1 ページの実効権限を決めるのに必要な「事実」を 1 回のクエリで集める。
-- 判定そのものは domain.ResolvePagePermission が行う（ここには規則を書かない）。
--
-- user_id / principal_id はどちらか一方だけを渡す。前者はログイン済みユーザーとして、
-- 後者は共有リンクの来訪者（kind='share_link'）として解決する。
--
-- CTE の役割:
--   target      … 対象ページの所属スペース（space_grants を引くのに要る）
--   me          … 自分自身の主体
--   mine        … 自分に効く主体すべて（自分 + 所属グループ + スペース全員）。
--                  グループの入れ子は DB 側で禁じているので 1 段の JOIN で足りる。
--                  スペース全員はメンバーにだけ効かせる（共有リンクの来訪者には効かせない）。
--   onpath      … 対象ページ自身と祖先に張られた制限を depth 付きで集める（0 が自分自身）
--   allow_level … ケイパビリティごとの「allow 行を持つ最も近い段」の depth
--   exception   … ケイパビリティごとに 3 つの事実へ畳む:
--                  (a) 経路上のどこかに自分宛ての deny があるか
--                  (b) 経路上に allow 行を持つ段があるか
--                  (c) その最も近い段に自分宛ての allow があるか
--
-- deny を経路全体で見るのが肝。最も近い段だけを見ると、deny 行しか無い段が最近段になった
-- 瞬間に「deny だけの段は既定に戻す」規則が働き、より遠い祖先の許可リストが
-- 無関係な deny 1 行で解除されてしまう（規則の適用は domain 側だが、
-- 事実として最近段しか返さない限り domain からは直しようがない）。
WITH target AS (
    SELECT p.space_id
    FROM pages p
    WHERE p.workspace_id = sqlc.arg(workspace_id) AND p.id = sqlc.arg(page_id)
),
me AS (
    SELECT p.id, p.kind
    FROM principals p
    WHERE p.workspace_id = sqlc.arg(workspace_id)
      AND (
            (p.kind = 'user' AND p.user_id = sqlc.narg(user_id)::bigint)
         OR (p.kind = 'share_link' AND p.id = sqlc.narg(principal_id)::uuid)
      )
),
mine AS (
    SELECT id FROM me
    UNION
    SELECT pm.group_principal_id
    FROM principal_members pm
    JOIN me ON me.id = pm.member_principal_id
    WHERE pm.workspace_id = sqlc.arg(workspace_id)
    UNION
    SELECT sp.id
    FROM principals sp
    CROSS JOIN target t
    WHERE sp.workspace_id = sqlc.arg(workspace_id)
      AND sp.kind = 'space_all'
      AND sp.space_id = t.space_id
      AND EXISTS (SELECT 1 FROM me WHERE me.kind = 'user')
),
onpath AS (
    SELECT r.capability, r.mode, r.principal_id, pp.depth
    FROM page_paths pp
    JOIN page_restrictions r
      ON r.workspace_id = pp.workspace_id AND r.page_id = pp.ancestor_id
    WHERE pp.workspace_id = sqlc.arg(workspace_id) AND pp.page_id = sqlc.arg(page_id)
),
allow_level AS (
    SELECT o.capability, MIN(o.depth) AS nearest_allow_depth
    FROM onpath o
    WHERE o.mode = 'allow'
    GROUP BY o.capability
),
exception AS (
    SELECT o.capability,
           bool_or(o.mode = 'deny' AND o.principal_id IN (SELECT id FROM mine)) AS denied_anywhere,
           bool_or(o.mode = 'allow') AS has_allow_list,
           -- 「そのケイパビリティの」最も近い allow の段を相関副問い合わせで引く。
           -- JOIN + bool_or にすると、ケイパビリティの突き合わせを落としても
           -- 正しい組が論理和に残るせいで誤りが表に出ない。ここは 1 行に絞れることが
           -- 意味なので、絞り込みを外したら副問い合わせが複数行でエラーになる形にする。
           bool_or(o.mode = 'allow'
                   AND o.depth = (SELECT al.nearest_allow_depth FROM allow_level al
                                   WHERE al.capability = o.capability)
                   AND o.principal_id IN (SELECT id FROM mine)) AS allowed_at_nearest
    FROM onpath o
    GROUP BY o.capability
)
SELECT
    EXISTS (SELECT 1 FROM target) AS page_exists,
    EXISTS (SELECT 1 FROM me WHERE me.kind = 'user') AS is_member,
    -- ワークスペースの grant とスペースの grant を合わせ、最も強い役割の強さを返す。
    -- 弱い方を採るとスペースに viewer を張るだけでワークスペース管理者を締め出せてしまう。
    --
    -- 役割そのもの（text）ではなく強さ（整数）を返すのは、役割が 1 つも無いときに
    -- NULL ではなく 0 で返すため。sqlc はスカラ副問い合わせの NULL 可能性を推論できず
    -- string 型を生成してしまい、grant が無い行の Scan がそこで落ちる。
    -- 0 は「grant が無い」を表し、persistence が domain.GrantRoleByRank で nil に直す
    -- （この値がそのまま上の層へ出ることはない）。
    -- CASE の並びは domain.GrantRole.Rank と一対一に対応させること。
    COALESCE((
        SELECT max(CASE g."role"
                     WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                     WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END)
        FROM (
            SELECT wg."role" FROM workspace_grants wg
             WHERE wg.workspace_id = sqlc.arg(workspace_id)
               AND wg.principal_id IN (SELECT id FROM mine)
            UNION ALL
            SELECT sg."role" FROM space_grants sg CROSS JOIN target t
             WHERE sg.workspace_id = sqlc.arg(workspace_id) AND sg.space_id = t.space_id
               AND sg.principal_id IN (SELECT id FROM mine)
        ) g
    ), 0)::integer AS grant_rank,
    EXISTS (SELECT 1 FROM exception e WHERE e.capability = 'view') AS view_restricted,
    COALESCE((SELECT e.denied_anywhere FROM exception e WHERE e.capability = 'view'), false)::boolean AS view_denied_anywhere,
    COALESCE((SELECT e.has_allow_list FROM exception e WHERE e.capability = 'view'), false)::boolean AS view_has_allow_list,
    COALESCE((SELECT e.allowed_at_nearest FROM exception e WHERE e.capability = 'view'), false)::boolean AS view_allowed_at_nearest,
    EXISTS (SELECT 1 FROM exception e WHERE e.capability = 'edit') AS edit_restricted,
    COALESCE((SELECT e.denied_anywhere FROM exception e WHERE e.capability = 'edit'), false)::boolean AS edit_denied_anywhere,
    COALESCE((SELECT e.has_allow_list FROM exception e WHERE e.capability = 'edit'), false)::boolean AS edit_has_allow_list,
    COALESCE((SELECT e.allowed_at_nearest FROM exception e WHERE e.capability = 'edit'), false)::boolean AS edit_allowed_at_nearest;

-- name: ListSpacePageViewFacts :many
-- スペース配下の現役ページ全件と、それぞれの「閲覧の事実」を 1 回のクエリで返す。
-- 判定は domain.ResolvePagePermission が行い、呼び出し側がふるい落とす。
--
-- ページごとに権限クエリを投げる（N+1）ことは避ける。ツリー表示は 1 スペースで
-- 数百〜数千ページを一度に扱うため、1 ページ 1 往復では表示のたびにその回数だけ往復する。
-- 制限を持つページはごく少数なので、closure との JOIN で拾えるのは実際には数行だけになる。
--
-- 集計は ResolvePagePermissionFacts と同じ 3 つの事実（deny は経路全体・許可リストは
-- 最も近い段）。ケイパビリティは 'view' に絞ってあるので、分けるのはページ単位だけで足りる。
-- 1 ページの解決と一覧で違う畳み方をすると「開くと見えるのに一覧に出ない」ずれになる。
WITH me AS (
    SELECT p.id
    FROM principals p
    WHERE p.workspace_id = sqlc.arg(workspace_id)
      AND p.kind = 'user' AND p.user_id = sqlc.arg(user_id)
),
mine AS (
    SELECT id FROM me
    UNION
    SELECT pm.group_principal_id
    FROM principal_members pm
    JOIN me ON me.id = pm.member_principal_id
    WHERE pm.workspace_id = sqlc.arg(workspace_id)
    UNION
    SELECT sp.id
    FROM principals sp
    WHERE sp.workspace_id = sqlc.arg(workspace_id)
      AND sp.kind = 'space_all' AND sp.space_id = sqlc.arg(space_id)
      AND EXISTS (SELECT 1 FROM me)
),
onpath AS (
    SELECT pp.page_id, pp.depth, r.mode, r.principal_id
    FROM page_paths pp
    JOIN pages tp ON tp.workspace_id = pp.workspace_id AND tp.id = pp.page_id
    JOIN page_restrictions r
      ON r.workspace_id = pp.workspace_id AND r.page_id = pp.ancestor_id AND r.capability = 'view'
    WHERE pp.workspace_id = sqlc.arg(workspace_id)
      AND tp.space_id = sqlc.arg(space_id) AND tp.archived_at IS NULL
),
allow_level AS (
    SELECT o.page_id, MIN(o.depth) AS nearest_allow_depth
    FROM onpath o
    WHERE o.mode = 'allow'
    GROUP BY o.page_id
),
exception AS (
    SELECT o.page_id,
           bool_or(o.mode = 'deny' AND o.principal_id IN (SELECT id FROM mine)) AS denied_anywhere,
           bool_or(o.mode = 'allow') AS has_allow_list,
           -- 相関副問い合わせにする理由は ResolvePagePermissionFacts と同じ。
           bool_or(o.mode = 'allow'
                   AND o.depth = (SELECT al.nearest_allow_depth FROM allow_level al
                                   WHERE al.page_id = o.page_id)
                   AND o.principal_id IN (SELECT id FROM mine)) AS allowed_at_nearest
    FROM onpath o
    GROUP BY o.page_id
)
SELECT
    p.*,
    EXISTS (SELECT 1 FROM me) AS is_member,
    -- 既定の役割の強さ。意味と 0 の扱いは ResolvePagePermissionFacts と同じ。
    COALESCE((
        SELECT max(CASE g."role"
                     WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                     WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END)
        FROM (
            SELECT wg."role" FROM workspace_grants wg
             WHERE wg.workspace_id = sqlc.arg(workspace_id)
               AND wg.principal_id IN (SELECT id FROM mine)
            UNION ALL
            SELECT sg."role" FROM space_grants sg
             WHERE sg.workspace_id = sqlc.arg(workspace_id) AND sg.space_id = sqlc.arg(space_id)
               AND sg.principal_id IN (SELECT id FROM mine)
        ) g
    ), 0)::integer AS grant_rank,
    (e.page_id IS NOT NULL)::boolean AS view_restricted,
    COALESCE(e.denied_anywhere, false)::boolean AS view_denied_anywhere,
    COALESCE(e.has_allow_list, false)::boolean AS view_has_allow_list,
    COALESCE(e.allowed_at_nearest, false)::boolean AS view_allowed_at_nearest
FROM pages p
LEFT JOIN exception e ON e.page_id = p.id
WHERE p.workspace_id = sqlc.arg(workspace_id)
  AND p.space_id = sqlc.arg(space_id)
  AND p.archived_at IS NULL
ORDER BY p."position";
