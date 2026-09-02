-- ノートの権限モデル（principals / principal_members / workspace_grants /
-- space_grants / page_grants / share_links）のクエリ。
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
-- 主体の削除。grant / principal_members は FK の CASCADE で一緒に消える
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

-- name: InsertWorkspaceGrantIfAbsent :exec
-- ワークスペース全体での既定の役割を**無いときだけ**与える（メンバー追加の既定 editor 用）。
-- Upsert（上書き）を使わないのが要点: 追加は冪等な操作で、既に admin の人へもう一度
-- 実行され得る。上書きだと admin が editor に落ち、しかも最後の admin なら
-- 保護の検査に当たって追加そのものが失敗する。既にある行は一切触らない。
INSERT INTO workspace_grants (workspace_id, principal_id, "role")
VALUES ($1, $2, $3)
ON CONFLICT (workspace_id, principal_id) DO NOTHING;

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

-- name: LockWorkspaceAdminGrantsForRemoval :one
-- 「この主体から admin を外しても、ユーザーの admin が 1 人以上残るか」を答える。
-- **答えるだけでなく、この文が admin の行をロックする。** ロックは呼び出し側の
-- トランザクションが終わるまで続く（＝ 判定と書き換えのあいだに割り込ませない）。
--
-- なぜロックまでするのか（検査を単一文にするだけでは足りなかった）:
--   検査と削除が別トランザクションだと、2 人の admin をほぼ同時に外す要求が
--   両方とも検査を通り抜け、ワークスペースの admin が 0 人になる。0 人になると
--   ノートには super_admin の抜け道が無いので、元 admin を含む誰も API から
--   権限を張り直せない（復旧は DB を直接触るしかない）。
--
--   検査を DELETE の EXISTS へ畳んで単一文にしても、これは塞がらない。
--   PostgreSQL の既定は READ COMMITTED で、EXISTS の副問い合わせは行をロックしない。
--   2 つのトランザクションが互いの admin 行を「まだ在る」と見たまま、それぞれ自分の
--   相手を消せてしまう（実測: 明示トランザクションを重ねると再現し、admin が 0 人になる）。
--   FOR UPDATE を付けて初めて、後から来た側が先の削除を待ち、待ったあとに
--   「その行はもう無い」と読み直して断るようになる。
--
-- 作りの要点:
--   - ORDER BY principal_id … 同時に走る 2 つの要求が同じ順でロックを取るので、
--     互いに待ち合って動けなくなる（デッドロック）ことがない。
--   - AS MATERIALIZED + 集約 … CTE を必ず最後まで読み切らせ、経路上の admin 行を
--     取りこぼさずロックする。EXISTS で書くと最初の 1 行で読むのをやめる可能性があり、
--     ロックする行が実行計画次第で変わってしまう。
--   - kind の扱い … 残る admin として数えるのは kind='user' だけ。グループ宛ての admin を
--     数に入れると、メンバーが 1 人も居ないグループが「最後の admin」として残り、
--     結局誰も権限を変えられないワークスペースが同じようにできる。
--     一方 target_is_admin は kind を問わない（外そうとしている行が admin かどうかの事実）。
WITH admin_grants AS MATERIALIZED (
    SELECT wg.principal_id, p.kind
    FROM workspace_grants wg
    JOIN principals p ON p.workspace_id = wg.workspace_id AND p.id = wg.principal_id
    WHERE wg.workspace_id = $1 AND wg."role" = 'admin'
    ORDER BY wg.principal_id
    FOR UPDATE OF wg
)
SELECT
    count(*) FILTER (WHERE principal_id = $2) > 0 AS target_is_admin,
    count(*) FILTER (WHERE principal_id <> $2 AND kind = 'user') > 0 AS other_user_admin_remains
FROM admin_grants;

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

-- name: ListGrantablePrincipals :many
-- 権限を張れる相手の一覧（画面の相手選びに使う）。
--
-- 表示名の正本はそれぞれ別の表にある。principals.name が埋まるのは group だけで、
-- ユーザー名は users、スペース名は spaces が持つ（principals へ写すと二重管理になる）。
-- 画面には名前が要るので、ここで 1 回だけ突き合わせる。
--
-- share_link は除く。あれは「リンクを踏んだ来訪者」を表す主体で、リンクを発行したときに
-- 自動で作られる。人が選んで役割を与える相手ではない（与えても意味を持たない）。
--
-- 並びは kind → 名前 → id。名前が空でも順序が決まるように id まで入れる
-- （ユーザーが消えた直後など、名前が引けない行が混ざり得る）。
--
-- 名前を組み立ててから CTE の外で並べ替える。JOIN したままの ORDER BY name は
-- principals.name と users.name のどちらを指すのか決まらず、sqlc が曖昧だと断る。
WITH grantable AS (
    SELECT p.id, p.kind,
           CASE p.kind
               WHEN 'group' THEN p.name
               WHEN 'user' THEN COALESCE(u.name, '')
               WHEN 'space_all' THEN COALESCE(s.name, '')
               ELSE ''
           END AS name
    FROM principals p
    LEFT JOIN users u
           ON p.kind = 'user' AND u.id = p.user_id
    LEFT JOIN spaces s
           ON p.kind = 'space_all' AND s.workspace_id = p.workspace_id AND s.id = p.space_id
    WHERE p.workspace_id = sqlc.arg(workspace_id)
      AND p.kind <> 'share_link'
)
SELECT id, kind, name FROM grantable
ORDER BY kind, name, id;

-- name: UpsertPageGrant :one
-- ページでの既定の役割の付与（同じ主体には 1 行だけ）。
--
-- 3 段目の既定で、このページとその子孫に効く。合成は他の 2 段と同じ「最も強いものを採る」
-- なので、ここに弱い役割を張っても上位で得た強い役割は下がらない。**弱める手段はどの層にも
-- 無い**ので、狭めたい内容は private のスペースへ置く。
INSERT INTO page_grants (workspace_id, page_id, principal_id, "role")
VALUES ($1, $2, $3, $4)
ON CONFLICT (workspace_id, page_id, principal_id)
DO UPDATE SET "role" = EXCLUDED."role", updated_at = now()
RETURNING *;

-- name: DeletePageGrant :execrows
-- ページでの既定の役割の剥奪。
DELETE FROM page_grants
WHERE workspace_id = $1 AND page_id = $2 AND principal_id = $3;

-- name: ListPageGrants :many
-- そのページ自身に張られた grant の一覧（祖先から降りてくる分は含まない）。
-- ListPageRestrictions と同じ見方で、返るのは「この段で足したもの」だけ。
SELECT * FROM page_grants
WHERE workspace_id = $1 AND page_id = $2
ORDER BY principal_id;

-- name: SubtreeHasForeignSpaceAllGrant :one
-- 移動するサブツリー（自分自身 + 子孫）に「移動先スペース以外のスペース全員」宛ての
-- ページ付与があるか。KnowledgeBaseRepository.MovePage が同じトランザクションで使う。
--
-- space_all の主体は「そのスペースの全員」を表すため、スペースをまたぐ移動で行だけが残り
-- 評価されなくなる（権限解決は対象ページが今いるスペースの space_all しか主体に取らない）。
-- 行は権限設定画面に見えているのに実効は違う、という追跡困難なずれになる。
-- 倒れる向きは常に「狭まる側」だが、見えている行が効かないこと自体が説明できないので、
-- 移動そのものを止める。
SELECT EXISTS (
    SELECT 1
    FROM page_paths pp
    JOIN page_grants g
      ON g.workspace_id = pp.workspace_id AND g.page_id = pp.page_id
    JOIN principals pr
      ON pr.workspace_id = g.workspace_id AND pr.id = g.principal_id
    WHERE pp.workspace_id = sqlc.arg(workspace_id)
      AND pp.ancestor_id = sqlc.arg(page_id)
      AND pr.kind = 'space_all'
      AND pr.space_id IS DISTINCT FROM sqlc.arg(new_space_id)::uuid
) AS found;

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

-- name: ResolvePagePermissionFacts :one
-- 1 ページの実効権限を決めるのに必要な「事実」を 1 回のクエリで集める。
-- 判定そのものは domain.ResolvePagePermission が行う（ここには規則を書かない）。
--
-- user_id / principal_id はどちらか一方だけを渡す。前者はログイン済みユーザーとして、
-- 後者は共有リンクの来訪者（kind='share_link'）として解決する。
--
-- CTE の役割:
--   target … 対象ページの所属スペース（space_grants を引くのに要る）
--   me     … 自分自身の主体
--   mine   … 自分に効く主体すべて（自分 + 所属グループ + スペース全員）。
--            グループの入れ子は DB 側で禁じているので 1 段の JOIN で足りる。
--            スペース全員はメンバーにだけ効かせる（共有リンクの来訪者には効かせない）。
--
-- **打ち消す層は無い。** 権限は 3 段の付与を足し合わせ、届いた中で最も強い役割で決まる。
-- 下の段が上の段を弱めることはないので、経路をさかのぼって拾うのは「最も強い役割」だけでよく、
-- どの段にあったかを覚えておく必要がない（最近段の depth も要らない）。
WITH target AS (
    -- スペースの visibility も一緒に引く。'private' のスペースには
    -- ワークスペース全体の grant と space_all（そのスペースの全員）を届かせない
    -- （届かせ方の規則は domain のまま。ここで変えるのは「事実の集め方」だけ）。
    SELECT p.space_id, s.visibility AS space_visibility
    FROM pages p
    JOIN spaces s ON s.workspace_id = p.workspace_id AND s.id = p.space_id
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
      AND t.space_visibility = 'workspace'
      AND EXISTS (SELECT 1 FROM me WHERE me.kind = 'user')
),
-- 経路上のページ付与（自分自身と祖先）のうち最も強いもの。祖先に editor を張れば
-- 子孫の既定が editor 以上になる、という降り方は grant の他の 2 段と同じ。
page_grant_rank AS (
    SELECT max(CASE pg."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS rank
    FROM page_paths pp
    JOIN page_grants pg
      ON pg.workspace_id = pp.workspace_id AND pg.page_id = pp.ancestor_id
    WHERE pp.workspace_id = sqlc.arg(workspace_id) AND pp.page_id = sqlc.arg(page_id)
      AND pg.principal_id IN (SELECT id FROM mine)
)
SELECT
    EXISTS (SELECT 1 FROM target) AS page_exists,
    EXISTS (SELECT 1 FROM me WHERE me.kind = 'user') AS is_member,
    -- 3 段の grant を合わせ、最も強い役割の強さを返す。
    -- 弱い方を採るとスペースに viewer を張るだけでワークスペース管理者を締め出せてしまう。
    --
    -- 役割そのもの（text）ではなく強さ（整数）を返すのは、役割が 1 つも無いときに
    -- NULL ではなく 0 で返すため。sqlc はスカラ副問い合わせの NULL 可能性を推論できず
    -- string 型を生成してしまい、grant が無い行の Scan がそこで落ちる。
    -- 0 は「grant が無い」を表し、persistence が domain.GrantRoleByRank で nil に直す
    -- （この値がそのまま上の層へ出ることはない）。
    -- CASE の並びは domain.GrantRole.Rank と一対一に対応させること。
    GREATEST(COALESCE((
        SELECT max(CASE g."role"
                     WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                     WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END)
        FROM (
            SELECT wg."role" FROM workspace_grants wg CROSS JOIN target t
             WHERE wg.workspace_id = sqlc.arg(workspace_id)
               AND t.space_visibility = 'workspace'
               AND wg.principal_id IN (SELECT id FROM mine)
            UNION ALL
            SELECT sg."role" FROM space_grants sg CROSS JOIN target t
             WHERE sg.workspace_id = sqlc.arg(workspace_id) AND sg.space_id = t.space_id
               AND sg.principal_id IN (SELECT id FROM mine)
        ) g
    ), 0), COALESCE((SELECT rank FROM page_grant_rank), 0))::integer AS grant_rank;

-- name: ListSpacePageViewFacts :many
-- スペース配下のページ全件と、それぞれの「閲覧の事実」を 1 回のクエリで返す。
-- archived で現役／アーカイブ済みを切り替える（既定の一覧は現役）。
--
-- # なぜアーカイブ用に別のクエリを作らないのか
--
-- 権限の事実を組み立てる部分を写経することになるため。**同じ判断を 2 箇所に置くと必ずずれる。**
-- 違うのは対象の絞り込みだけなので、そこだけを引数にする。
-- 判定は domain.ResolvePageView が行い、呼び出し側がふるい落とす。
--
-- ページごとに権限クエリを投げる（N+1）ことは避ける。ツリー表示は 1 スペースで
-- 数百〜数千ページを一度に扱うため、1 ページ 1 往復では表示のたびにその回数だけ往復する。
--
-- 集めるのは ResolvePagePermissionFacts と同じ事実（届いた中で最も強い役割）。
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
      -- private のスペースには space_all を届かせない（1 枚解決と同じ規則）。
      AND EXISTS (
        SELECT 1 FROM spaces sv1
        WHERE sv1.workspace_id = sqlc.arg(workspace_id) AND sv1.id = sqlc.arg(space_id)
          AND sv1.visibility = 'workspace'
      )
      AND EXISTS (SELECT 1 FROM me)
),
-- 各ページについて、経路上（自分と祖先）のページ付与の最大値。ページごとに値が変わるので
-- スペース単位の既定のように 1 行へ畳めない。1 回のクエリで集めて LEFT JOIN する
-- （ページごとに引き直すと行数ぶんの集約になる）。
-- 「最も近い段」は見ない — 付与に降格は無く、近い付与が遠い付与を弱めることはないため。
page_grant_rank AS (
    SELECT pp.page_id,
           max(CASE pg."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS rank
    FROM page_paths pp
    JOIN pages tp ON tp.workspace_id = pp.workspace_id AND tp.id = pp.page_id
    JOIN page_grants pg
      ON pg.workspace_id = pp.workspace_id AND pg.page_id = pp.ancestor_id
    WHERE pp.workspace_id = sqlc.arg(workspace_id)
      AND tp.space_id = sqlc.arg(space_id)
      AND (sqlc.arg(archived)::boolean) = (tp.archived_at IS NOT NULL)
      AND pg.principal_id IN (SELECT id FROM mine)
    GROUP BY pp.page_id
)
SELECT
    p.*,
    -- 既定の役割の強さ。意味と 0 の扱いは ResolvePagePermissionFacts と同じ。
    -- 所属（is_member）は返さない。役割が 1 つも無ければ強さ 0 で「何もできない」に
    -- なるため閲覧の判定には要らず、使われない事実を返すと編集可否にも答えられる顔をする。
    GREATEST(COALESCE((
        SELECT max(CASE g."role"
                     WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                     WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END)
        FROM (
            -- private のスペースにはワークスペース全体の grant を届かせない。
            SELECT wg."role" FROM workspace_grants wg
             WHERE wg.workspace_id = sqlc.arg(workspace_id)
               AND wg.principal_id IN (SELECT id FROM mine)
               AND EXISTS (
                 SELECT 1 FROM spaces sv2
                 WHERE sv2.workspace_id = sqlc.arg(workspace_id) AND sv2.id = sqlc.arg(space_id)
                   AND sv2.visibility = 'workspace'
               )
            UNION ALL
            SELECT sg."role" FROM space_grants sg
             WHERE sg.workspace_id = sqlc.arg(workspace_id) AND sg.space_id = sqlc.arg(space_id)
               AND sg.principal_id IN (SELECT id FROM mine)
        ) g
    ), 0), COALESCE(pgr.rank, 0))::integer AS grant_rank,
    -- 親がアーカイブ済みか（親を持たない行は false）。
    (par.archived_at IS NOT NULL)::boolean AS parent_archived
FROM pages p
LEFT JOIN page_grant_rank pgr ON pgr.page_id = p.id
-- 親がアーカイブ済みかは**事実**として返すだけで、ここでは何の判断にも使わない。
-- 「復帰できるか」の規則は UnarchivePageUseCase が持つ（親がアーカイブ中なら断る）。
LEFT JOIN pages par ON par.workspace_id = p.workspace_id AND par.id = p.parent_id
WHERE p.workspace_id = sqlc.arg(workspace_id)
  AND p.space_id = sqlc.arg(space_id)
  AND (sqlc.arg(archived)::boolean) = (p.archived_at IS NOT NULL)
ORDER BY p."position";

-- name: ListSubtreePagePermissionFacts :many
-- サブツリー（対象ページ自身 + 全子孫）の各ページについて、実効権限を決める「事実」を
-- 1 回のクエリで集める。判定は domain.ResolvePagePermission が行う（ここには規則を書かない）。
--
-- 用途はアーカイブ / 復帰のように「1 枚を名指しして子孫ごと書き換える」操作の入口検査。
-- 根 1 枚の編集権限だけで通すと、同じページを直接 rename すると 403 になるのに
-- 祖先のアーカイブ経由なら書き換えられる、という経路依存の食い違いになる。
--
-- ページごとに ResolvePagePermissionFacts を投げない（N+1 にしない）。集めるのは
-- ListSpacePageViewFacts と同じ形で、違いは対象の絞り方（スペース全体 → closure の
-- サブツリー）だけ。
--
-- アーカイブ済みのページも外さない。操作の影響が及ぶ範囲はアーカイブ状態と関係なく
-- サブツリー全体で、外すと「先に子だけアーカイブしておけば検査を迂回できる」経路ができる。
--
-- サブツリーは必ず 1 つのスペースに収まる（スペースをまたぐ移動はサブツリーの space_id を
-- まとめて付け替える）ので、space_grants と space_all の主体は根のスペースで引けば足りる。
--
-- ページごとに値が変わるのは経路上のページ付与だけなので、closure を辿るのは
-- page_grant_rank の 1 本で済む。呼ぶのはアーカイブ / 復帰の 1 回だけで、閲覧経路には足さない。
WITH target AS (
    -- visibility の意味は ResolvePagePermissionFacts の target と同じ。
    SELECT p.space_id, s.visibility AS space_visibility
    FROM pages p
    JOIN spaces s ON s.workspace_id = p.workspace_id AND s.id = p.space_id
    WHERE p.workspace_id = sqlc.arg(workspace_id) AND p.id = sqlc.arg(page_id)
),
subtree AS (
    -- closure なので自分自身（depth 0）も含む。
    SELECT pp.page_id
    FROM page_paths pp
    WHERE pp.workspace_id = sqlc.arg(workspace_id) AND pp.ancestor_id = sqlc.arg(page_id)
),
me AS (
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
    CROSS JOIN target t
    WHERE sp.workspace_id = sqlc.arg(workspace_id)
      AND sp.kind = 'space_all' AND sp.space_id = t.space_id
      AND t.space_visibility = 'workspace'
      AND EXISTS (SELECT 1 FROM me)
),
grants AS (
    -- ワークスペースとスペースの既定はサブツリー全体で同じ値なので 1 行に畳んでから配る
    -- （ページごとに引き直すと行数ぶんの集約になる）。意味と 0 の扱いは
    -- ResolvePagePermissionFacts と同じ。
    SELECT max(CASE g."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS grant_rank
    FROM (
        SELECT wg."role" FROM workspace_grants wg CROSS JOIN target t
         WHERE wg.workspace_id = sqlc.arg(workspace_id)
           AND t.space_visibility = 'workspace'
           AND wg.principal_id IN (SELECT id FROM mine)
        UNION ALL
        SELECT sg."role" FROM space_grants sg CROSS JOIN target t
         WHERE sg.workspace_id = sqlc.arg(workspace_id) AND sg.space_id = t.space_id
           AND sg.principal_id IN (SELECT id FROM mine)
    ) g
),
-- ページ付与だけは畳めない。サブツリーの中でも「祖先のどこに張られているか」で
-- ページごとに値が変わるため、page_id ごとに集めて下の SELECT へ LEFT JOIN する。
-- 「最も近い段」は見ない — 付与に降格は無く、近い付与が遠い付与を弱めることはないため。
page_grant_rank AS (
    SELECT pp.page_id,
           max(CASE pg."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS rank
    FROM page_paths pp
    JOIN subtree st ON st.page_id = pp.page_id
    JOIN page_grants pg
      ON pg.workspace_id = pp.workspace_id AND pg.page_id = pp.ancestor_id
    WHERE pp.workspace_id = sqlc.arg(workspace_id)
      AND pg.principal_id IN (SELECT id FROM mine)
    GROUP BY pp.page_id
)
SELECT
    s.page_id,
    EXISTS (SELECT 1 FROM me) AS is_member,
    GREATEST(COALESCE((SELECT grants.grant_rank FROM grants), 0), COALESCE(pgr.rank, 0))::integer AS grant_rank
FROM subtree s
LEFT JOIN page_grant_rank pgr ON pgr.page_id = s.page_id
ORDER BY s.page_id;

-- name: ListMemberWorkspaces :many
-- そのユーザーが所属するワークスペース一覧と、自分がそこの admin かどうか。
--
-- 所属の正本は principals（kind='user'）の行なので、JOIN の結果がそのまま答えになる。
-- このファイルの作法（WHERE に workspace_id を必ず含める）に対する唯一の例外で、
-- テナントを絞る手前の「どのテナントに入れるか」を答えるクエリだから workspace_id を取らない。
-- 代わりに principals 側で user_id を必ず縛る（ここが緩むと全テナントが漏れる）。
--
-- is_admin は workspace_grants を自分の principal で LEFT JOIN するだけで求まる
-- （admin だけが consequential なので role = 'admin' の 1 行があるかどうかだけを見る）。
-- DeleteWorkspace が要求する権限（CanManage）と同じ判定を、一覧の段階で添えて返す。
--
-- grant が 1 行も無い所属では wg.role が NULL になり、(wg.role = 'admin') も NULL になる
-- （SQL の三値論理）。COALESCE(wg.role, '') で先に text を NULL 抜きにしてから比較すると、
-- 比較結果そのものは NULL になり得ない。COALESCE(bool式, false) だと sqlc の型推論が
-- interface{} に落ちてしまうため、text 側で COALESCE する形にしている
-- （driver が NULL を bool へ Scan できずに落ちることをローカル PostgreSQL で確認済み）。
SELECT w.*, (COALESCE(wg.role, '') = 'admin') AS is_admin FROM workspaces w
JOIN principals p
  ON p.workspace_id = w.id AND p.kind = 'user' AND p.user_id = sqlc.arg(user_id)
LEFT JOIN workspace_grants wg
  ON wg.workspace_id = w.id AND wg.principal_id = p.id
ORDER BY w.slug;

-- name: ListSpaceScopeGrantRoles :many
-- そのスペースの「既定の役割」として自分に届いている役割をすべて返す（事実だけ）。
--
-- ページを介さずスペース単位で権限を判定する経路（スペース直下へのページ作成など）の土台。
-- 返すのは役割の集合であって、どれを採るかの規則は domain.StrongestGrantRole が持つ。
-- max(rank) を SQL 側で計算すると「最も強いものを採る」という規則が DB へ写り、
-- ページ 1 枚の解決と食い違ったときにどちらが正か決められなくなる。
--
-- ページ付与（page_grants）はここでは一切見ない。この結果を「そのスペースのあるページを
-- 編集してよいか」に使ってはいけない（祖先のページ付与を見ていないため必ず狭い側へ倒れる）。
-- 呼び出し側は対象がまだ存在しない操作にだけ使う。
--
-- mine（自分に効く主体）の作り方は ResolvePagePermissionFacts と同じ:
-- 自分自身 + 所属グループ + そのスペースの「全員」。グループの入れ子は DB 側で
-- 禁じてあるので 1 段の JOIN で足りる。
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
      -- private のスペースには space_all を届かせない。
      AND EXISTS (
        SELECT 1 FROM spaces sv3
        WHERE sv3.workspace_id = sqlc.arg(workspace_id) AND sv3.id = sqlc.arg(space_id)
          AND sv3.visibility = 'workspace'
      )
      AND EXISTS (SELECT 1 FROM me)
)
-- ワークスペースの grant は visibility='workspace' のスペースにだけ届く。
-- スペースの grant と合わせて返す。
SELECT wg."role" FROM workspace_grants wg
 WHERE wg.workspace_id = sqlc.arg(workspace_id)
   AND wg.principal_id IN (SELECT id FROM mine)
   AND EXISTS (
     SELECT 1 FROM spaces sv4
     WHERE sv4.workspace_id = sqlc.arg(workspace_id) AND sv4.id = sqlc.arg(space_id)
       AND sv4.visibility = 'workspace'
   )
UNION
SELECT sg."role" FROM space_grants sg
 WHERE sg.workspace_id = sqlc.arg(workspace_id) AND sg.space_id = sqlc.arg(space_id)
   AND sg.principal_id IN (SELECT id FROM mine);

-- name: ListWorkspaceScopeGrantRoles :many
-- ワークスペースそのものに対して自分に届いている役割をすべて返す（事実だけ）。
-- スペースを作る操作のように、どのスペースにも属さない判定に使う。
--
-- kind='space_all' の主体はここでは数えない。あれは「そのスペースの全員」という
-- スペースの中でだけ意味を持つ主体で、入れ物が決まっていないワークスペース単位の判定に
-- 混ぜると、どこか 1 つのスペースの space_all に張られた grant がテナント全体の権限に化ける。
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
)
SELECT wg."role" FROM workspace_grants wg
 WHERE wg.workspace_id = sqlc.arg(workspace_id)
   AND wg.principal_id IN (SELECT id FROM mine);

-- name: ListWorkspaceSpaceScopeFacts :many
-- ワークスペース配下のスペース全件と、それぞれで呼び出し元に届いている「既定の役割」を
-- 1 回のクエリで返す（サイドバーがスペースを列挙するための土台）。
--
-- 返すのは事実だけ。「その役割で中身を見てよいか」は domain.ResolveScopePermission が決め、
-- 呼び出し側（ListViewableSpacesUseCase）が見えないスペースをふるい落とす。
-- ここで役割を畳んだり WHERE で絞ったりしないのは、ページ 1 枚の解決・スペース 1 つの解決と
-- 同じ規則を 1 箇所（domain）だけに置くため。SQL 側にも規則を書くと、片方だけ直したときに
-- 「スペースを開けるのに一覧に出ない」というずれ方をする。
--
-- スペースを 1 件も落とさずに返す（LEFT JOIN）のが要点。役割の届いていないスペースを
-- SQL 側で消してしまうと、ふるいが SQL と domain の 2 箇所に散る。事実として
-- 「役割が 1 つも無い」（role が NULL の行）まで返し、判定は 1 箇所に集める。
-- 同じ作りの先例が ListSpacePageViewFacts（スペース配下の全ページを返して domain がふるう）。
--
-- N+1 を作らない。スペースごとに ListSpaceScopeGrantRoles を投げると、
-- サイドバーを開くたびにスペース数だけ往復する。
--
-- mine（自分に効く主体）の作り方は ListSpaceScopeGrantRoles と同じだが、
-- 「そのスペースの全員（kind='space_all'）」だけは mine に混ぜられない。あれはスペース 1 つに
-- 紐づく主体で、対象スペースが 1 つに決まっているときしか「自分」に畳めないため
-- （混ぜると、どこか 1 つのスペースの space_all 宛て grant が全スペースへ効く）。
-- ここでは grant の側でスペースを突き合わせる（下の (c)）。
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
),
roles AS (
    -- (a) ワークスペース全体の grant は visibility='workspace' のスペースへだけ届く。
    -- private のスペースは (b) のスペース単位の付与だけで見える。
    SELECT s.id AS space_id, wg."role"
    FROM spaces s
    JOIN workspace_grants wg
      ON wg.workspace_id = sqlc.arg(workspace_id)
     AND wg.principal_id IN (SELECT id FROM mine)
    WHERE s.workspace_id = sqlc.arg(workspace_id)
      AND s.visibility = 'workspace'
    UNION
    -- (b) スペース単位の grant のうち、自分 / 所属グループ宛てのもの。
    SELECT sg.space_id, sg."role"
    FROM space_grants sg
    WHERE sg.workspace_id = sqlc.arg(workspace_id)
      AND sg.principal_id IN (SELECT id FROM mine)
    UNION
    -- (c) スペース単位の grant のうち、そのスペース自身の「全員」宛てのもの。
    -- sa.space_id = sg.space_id で結ぶので、別スペースの「全員」宛て grant は混ざらない。
    -- EXISTS (me) は非メンバーを弾く（所属していない相手は「全員」に含まれない）。
    SELECT sg.space_id, sg."role"
    FROM space_grants sg
    JOIN principals sa
      ON sa.workspace_id = sqlc.arg(workspace_id) AND sa.id = sg.principal_id
     AND sa.kind = 'space_all' AND sa.space_id = sg.space_id
    -- private のスペースには space_all を届かせない。
    JOIN spaces sv5
      ON sv5.workspace_id = sqlc.arg(workspace_id) AND sv5.id = sg.space_id
     AND sv5.visibility = 'workspace'
    WHERE sg.workspace_id = sqlc.arg(workspace_id)
      AND EXISTS (SELECT 1 FROM me)
)
SELECT s.*, r."role"
FROM spaces s
LEFT JOIN roles r ON r.space_id = s.id
WHERE s.workspace_id = sqlc.arg(workspace_id)
ORDER BY s."key", r."role";

-- name: SearchWorkspacePageViewFacts :many
-- ワークスペース全体から、題名が部分一致する**現役**ページを候補にして、
-- それぞれの「閲覧の事実」を 1 回のクエリで返す（サイドバーの題名検索用）。
--
-- 事実の組み立ては ListSpacePageViewFacts と同じ見方（届いた中で最も強い役割）で、
-- 判定は domain.ResolvePageView が行う。違いは 2 つだけ:
--   1. 対象がスペース 1 つではなくワークスペース全体（題名の一致で先に候補を絞る）
--   2. スペース単位の主体（space_all）と space_grants は**そのページのスペース**のもので
--      突き合わせる。1 スペース版は引数のスペースに固定できたが、こちらは行ごとに違うので、
--      space_allp（スペースごとの space_all 主体。自分が所属するときだけ行がある）を
--      JOIN で当てる。集計の中に相関副問い合わせを書かない流儀は他のクエリと同じ。
--
-- 候補の LIMIT 200 は「事実を計算する行数の天井」であって応答の件数ではない
-- （呼び出し側が可視でふるった後にさらに件数を切る）。
--
-- needle は呼び出し側（Go）が % _ とバックスラッシュをエスケープして渡す
-- （LIKE の既定のエスケープ文字はバックスラッシュ）。生で渡すと「%」1 文字で全件一致になり、
-- 候補の天井まで無関係な行が埋まる。
--
-- 索引について: 部分一致（中間一致）は B-tree では引けないため、この絞り込みは
-- workspace_id の索引で範囲を狭めたうえでの逐次比較になる。現状の規模（1 ワークスペース
-- 数百〜数千ページ）では十分速い。伸びたら pg_trgm の GIN を検討する（拡張が要るので
-- そのときに判断する）。
--
-- 表の別名はクエリ全体で一意にしてある（pr / pg / spx / c …）。CTE ごとに同じ
-- 別名（p 等）を使い回すと sqlc の列解決が別の CTE の表に混線して
-- 「column ... does not exist」で生成が落ちる（実測）。
WITH me AS (
    SELECT pr.id
    FROM principals pr
    WHERE pr.workspace_id = sqlc.arg(workspace_id)
      AND pr.kind = 'user' AND pr.user_id = sqlc.arg(user_id)
),
mine AS (
    -- 自分と、自分が入っているグループ。space_all はスペースごとに違うので space_allp で持つ。
    SELECT id FROM me
    UNION
    SELECT pmb.group_principal_id
    FROM principal_members pmb
    JOIN me ON me.id = pmb.member_principal_id
    WHERE pmb.workspace_id = sqlc.arg(workspace_id)
),
space_allp AS (
    -- スペースごとの「全員」主体。自分がワークスペースの所属者のときだけ意味を持つ。
    -- private のスペースには space_all を届かせない（スペース単位の付与だけが届く）。
    SELECT spx.space_id, spx.id
    FROM principals spx
    JOIN spaces svx ON svx.workspace_id = sqlc.arg(workspace_id) AND svx.id = spx.space_id
     AND svx.visibility = 'workspace'
    WHERE spx.workspace_id = sqlc.arg(workspace_id)
      AND spx.kind = 'space_all'
      AND EXISTS (SELECT 1 FROM me)
),
cand AS (
    SELECT pg.*
    FROM pages pg
    WHERE pg.workspace_id = sqlc.arg(workspace_id)
      AND pg.archived_at IS NULL
      AND pg.title ILIKE ('%' || sqlc.arg(needle)::text || '%')
    ORDER BY pg.title, pg.id
    LIMIT 200
),
wsrank AS (
    SELECT COALESCE(max(CASE wg."role"
                          WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                          WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END), 0) AS v
    FROM workspace_grants wg
    WHERE wg.workspace_id = sqlc.arg(workspace_id)
      AND wg.principal_id IN (SELECT id FROM mine)
),
sgrank AS (
    SELECT sg.space_id,
           max(CASE sg."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS v
    FROM space_grants sg
    LEFT JOIN space_allp sap2 ON sap2.space_id = sg.space_id
    WHERE sg.workspace_id = sqlc.arg(workspace_id)
      AND (sg.principal_id IN (SELECT id FROM mine) OR sg.principal_id = sap2.id)
    GROUP BY sg.space_id
),
-- ページ付与は経路（自分と祖先）を辿るので page_id ごとに値が変わる。
-- 「最も近い段」は見ない — 付与に降格は無く、近い付与が遠い付与を弱めることはないため。
--
-- この経路の mine は「自分と所属グループ」だけで、スペース全員（space_all）は space_allp が
-- 別に持つ。両方を見ないと、全員宛ての付与が 1 ページの解決では効くのにここでは効かず、
-- 「開けるのに検索に出ない」ずれになる。
--
-- 候補（cand）に絞ってから集計する。ワークスペース全体の経路を集めると、候補が数件でも
-- 全ページ分の JOIN を回すことになる。
pgrank AS (
    SELECT pp.page_id,
           max(CASE pgt."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS v
    FROM page_paths pp
    JOIN cand c ON c.id = pp.page_id
    JOIN page_grants pgt
      ON pgt.workspace_id = pp.workspace_id AND pgt.page_id = pp.ancestor_id
    LEFT JOIN space_allp sap3 ON sap3.space_id = c.space_id
    WHERE pp.workspace_id = sqlc.arg(workspace_id)
      AND (pgt.principal_id IN (SELECT id FROM mine) OR pgt.principal_id = sap3.id)
    GROUP BY pp.page_id
)
SELECT
    cnd.*,
    -- ワークスペース全体の強さ（wsrank）は visibility='workspace' のスペースの行にだけ効かせる。
    -- private のスペースはスペース単位の強さ（sgrank）だけで決まる。
    GREATEST(
      CASE WHEN spvis.visibility = 'workspace' THEN (SELECT v FROM wsrank) ELSE 0 END,
      COALESCE(sr.v, 0),
      COALESCE(pgr.v, 0)
    )::integer AS grant_rank
FROM cand cnd
-- pages → spaces は複合 FK があるので必ず 1 行に当たる。
JOIN spaces spvis ON spvis.workspace_id = sqlc.arg(workspace_id) AND spvis.id = cnd.space_id
LEFT JOIN sgrank sr ON sr.space_id = cnd.space_id
LEFT JOIN pgrank pgr ON pgr.page_id = cnd.id
ORDER BY cnd.title, cnd.id;

-- name: ListWorkspacePageViewFactsByIDs :many
-- 指定した ID 群の**現役**ページについて「閲覧の事実」を 1 回のクエリで返す
-- （本文中のページ参照の題名解決用）。事実の組み立ては SearchWorkspacePageViewFacts と
-- 同一で、違いは候補の絞り方だけ（題名の部分一致 → ID の一致）。判定は
-- domain.ResolvePageView が行う。
--
-- page_ids は json 配列（文字列の UUID）。IN 句のスライス展開を使わないのは
-- ListMasterExerciseExamplesByExerciseIDs と同じ理由（database/sql モードでは lib/pq 依存が
-- 増えるため。json_array_elements_text で展開して uuid へ落とす）。呼び出し側（Go）が
-- UUID として読めない値を先に落として渡す — ここで ::uuid が失敗するとクエリ全体が落ちる。
--
-- アーカイブ済みも行として返す（Page.ArchivedAt に載る）。除外するかは用途で違う —
-- 題名解決は除外し（隠したページの題名を本文へ映さない）、パンくずは含める
-- （アーカイブ済みのページは開けるので、経路から抜くと場所を偽る）。その判断は
-- 呼び出し側（usecase）が Page.ArchivedAt を見て行う。
-- 他ワークスペースの ID は workspace_id の条件で自然に 0 行になる。
--
-- 表の別名はクエリ全体で一意（CTE をまたぐ使い回しは sqlc の列解決が混線する — 検索の
-- クエリのコメントを参照）。
WITH me AS (
    SELECT pr.id
    FROM principals pr
    WHERE pr.workspace_id = sqlc.arg(workspace_id)
      AND pr.kind = 'user' AND pr.user_id = sqlc.arg(user_id)
),
mine AS (
    SELECT id FROM me
    UNION
    SELECT pmb.group_principal_id
    FROM principal_members pmb
    JOIN me ON me.id = pmb.member_principal_id
    WHERE pmb.workspace_id = sqlc.arg(workspace_id)
),
space_allp AS (
    -- private のスペースには space_all を届かせない（Search 側と同じ規則）。
    SELECT spx.space_id, spx.id
    FROM principals spx
    JOIN spaces svy ON svy.workspace_id = sqlc.arg(workspace_id) AND svy.id = spx.space_id
     AND svy.visibility = 'workspace'
    WHERE spx.workspace_id = sqlc.arg(workspace_id)
      AND spx.kind = 'space_all'
      AND EXISTS (SELECT 1 FROM me)
),
cand AS (
    SELECT pg.*
    FROM pages pg
    WHERE pg.workspace_id = sqlc.arg(workspace_id)
      AND pg.id IN (
        SELECT value::uuid FROM json_array_elements_text(sqlc.arg(page_ids)::json) AS t(value)
      )
    ORDER BY pg.title, pg.id
),
wsrank AS (
    SELECT COALESCE(max(CASE wg."role"
                          WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                          WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END), 0) AS v
    FROM workspace_grants wg
    WHERE wg.workspace_id = sqlc.arg(workspace_id)
      AND wg.principal_id IN (SELECT id FROM mine)
),
sgrank AS (
    SELECT sg.space_id,
           max(CASE sg."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS v
    FROM space_grants sg
    LEFT JOIN space_allp sap2 ON sap2.space_id = sg.space_id
    WHERE sg.workspace_id = sqlc.arg(workspace_id)
      AND (sg.principal_id IN (SELECT id FROM mine) OR sg.principal_id = sap2.id)
    GROUP BY sg.space_id
),
-- ページ付与。候補に絞ってから経路を辿る（意味は検索側と同じ）。
pgrank AS (
    SELECT pp.page_id,
           max(CASE pgt."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS v
    FROM page_paths pp
    JOIN cand c ON c.id = pp.page_id
    JOIN page_grants pgt
      ON pgt.workspace_id = pp.workspace_id AND pgt.page_id = pp.ancestor_id
    LEFT JOIN space_allp sap3 ON sap3.space_id = c.space_id
    WHERE pp.workspace_id = sqlc.arg(workspace_id)
      AND (pgt.principal_id IN (SELECT id FROM mine) OR pgt.principal_id = sap3.id)
    GROUP BY pp.page_id
)
SELECT
    cnd.*,
    GREATEST(
      CASE WHEN spvis.visibility = 'workspace' THEN (SELECT v FROM wsrank) ELSE 0 END,
      COALESCE(sr.v, 0),
      COALESCE(pgr.v, 0)
    )::integer AS grant_rank
FROM cand cnd
JOIN spaces spvis ON spvis.workspace_id = sqlc.arg(workspace_id) AND spvis.id = cnd.space_id
LEFT JOIN sgrank sr ON sr.space_id = cnd.space_id
LEFT JOIN pgrank pgr ON pgr.page_id = cnd.id
ORDER BY cnd.title, cnd.id;
