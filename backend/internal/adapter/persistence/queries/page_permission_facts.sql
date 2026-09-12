-- ResolvePagePermissionFacts / ListSubtreePagePermissionFacts は、もともと
-- knowledge_base_permission.sql に置いていたが、そちらへ編集を加えるたびに sqlc が
-- 「column ... does not exist / is ambiguous」を同じファイル内の無関係な既存クエリの行へ
-- 誤帰属させる現象が起きたため、この 1 ファイルへ切り出した（space_member.sql と同じ理由・
-- 同じ対処）。表の別名はこのファイル内で完全に一意にしてある — CTE をまたいで同じ別名
-- （p / s / t 等）を使い回すと sqlc の列解決が混線するため（実測）。

-- name: ResolvePagePermissionFacts :one
-- 1 ページの実効権限を決めるのに必要な「事実」を 1 回のクエリで集める。
-- 判定そのものは domain.ResolvePagePermission が行う（ここには規則を書かない）。
--
-- user_id / principal_id はどちらか一方だけを渡す。前者はログイン済みユーザーとして、
-- 後者は共有リンクの来訪者（kind='share_link'）として解決する。
--
-- CTE の役割:
--   rpf_target … 対象ページの所属スペース・ページ自身の visibility / 作成者
--   rpf_me     … 自分自身の主体
--   rpf_mine   … 自分に効く主体すべて（自分 + 所属グループ + スペース全員）。
--            グループの入れ子は DB 側で禁じているので 1 段の JOIN で足りる。
--            スペース全員はメンバーにだけ効かせる（共有リンクの来訪者には効かせない）。
--
-- **打ち消す層は無い（ページの visibility='private' だけが唯一の例外）。** 権限は 3 段の
-- 付与を足し合わせ、届いた中で最も強い役割で決まる。下の段が上の段を弱めることはないので、
-- 経路をさかのぼって拾うのは「最も強い役割」だけでよく、どの段にあったかを覚えておく
-- 必要がない（最近段の depth も要らない）。'private' の判定（作成者以外には一切見せない）は
-- domain.ResolvePagePermission が持ち、ここでは page_visibility / is_owner を事実として
-- 渡すだけ。
WITH rpf_target AS (
    -- スペースの visibility も一緒に引く。'private' のスペースには
    -- ワークスペース全体の grant と space_all（そのスペースの全員）を届かせない
    -- （届かせ方の規則は domain のまま。ここで変えるのは「事実の集め方」だけ）。
    --
    -- ページ自身の visibility / created_by_user_id も一緒に引く。
    SELECT rpf_pg.space_id, rpf_sp.visibility AS space_visibility,
           rpf_pg.visibility AS page_visibility, rpf_pg.created_by_user_id
    FROM pages rpf_pg
    JOIN spaces rpf_sp ON rpf_sp.workspace_id = rpf_pg.workspace_id AND rpf_sp.id = rpf_pg.space_id
    WHERE rpf_pg.workspace_id = sqlc.arg(workspace_id) AND rpf_pg.id = sqlc.arg(page_id)
),
rpf_me AS (
    SELECT rpf_pr.id, rpf_pr.kind
    FROM principals rpf_pr
    WHERE rpf_pr.workspace_id = sqlc.arg(workspace_id)
      AND (
            (rpf_pr.kind = 'user' AND rpf_pr.user_id = sqlc.narg(user_id)::bigint)
         OR (rpf_pr.kind = 'share_link' AND rpf_pr.id = sqlc.narg(principal_id)::uuid)
      )
),
rpf_mine AS (
    SELECT id FROM rpf_me
    UNION
    SELECT rpf_pm.group_principal_id
    FROM principal_members rpf_pm
    JOIN rpf_me ON rpf_me.id = rpf_pm.member_principal_id
    WHERE rpf_pm.workspace_id = sqlc.arg(workspace_id)
    UNION
    SELECT rpf_spall.id
    FROM principals rpf_spall
    CROSS JOIN rpf_target rpf_t1
    WHERE rpf_spall.workspace_id = sqlc.arg(workspace_id)
      AND rpf_spall.kind = 'space_all'
      AND rpf_spall.space_id = rpf_t1.space_id
      AND rpf_t1.space_visibility = 'workspace'
      AND EXISTS (SELECT 1 FROM rpf_me WHERE rpf_me.kind = 'user')
),
-- 経路上のページ付与（自分自身と祖先）のうち最も強いもの。祖先に editor を張れば
-- 子孫の既定が editor 以上になる、という降り方は grant の他の 2 段と同じ。
rpf_page_grant_rank AS (
    SELECT max(CASE rpf_pg2."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS rank
    FROM page_paths rpf_pp
    JOIN page_grants rpf_pg2
      ON rpf_pg2.workspace_id = rpf_pp.workspace_id AND rpf_pg2.page_id = rpf_pp.ancestor_id
    WHERE rpf_pp.workspace_id = sqlc.arg(workspace_id) AND rpf_pp.page_id = sqlc.arg(page_id)
      AND rpf_pg2.principal_id IN (SELECT id FROM rpf_mine)
)
SELECT
    EXISTS (SELECT 1 FROM rpf_target) AS page_exists,
    EXISTS (SELECT 1 FROM rpf_me WHERE rpf_me.kind = 'user') AS is_member,
    -- ページが無いときの既定値は 'space'（何も特別扱いしない値）。page_exists が false の
    -- ときは呼び出し側（persistence）がその場で ErrPageNotFound を返すので、実際に使われる
    -- ことはない。
    COALESCE((SELECT page_visibility FROM rpf_target), 'space')::text AS page_visibility,
    -- 共有リンク経由（user_id が NULL）では常に false。ログインしていない来訪者を
    -- 作成者と同一だと判定しようがないため。
    COALESCE((SELECT created_by_user_id FROM rpf_target) = sqlc.narg(user_id)::bigint, false)::boolean AS is_owner,
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
        SELECT max(CASE rpf_g."role"
                     WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                     WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END)
        FROM (
            SELECT rpf_wg."role" FROM workspace_grants rpf_wg CROSS JOIN rpf_target rpf_t2
             WHERE rpf_wg.workspace_id = sqlc.arg(workspace_id)
               AND rpf_t2.space_visibility = 'workspace'
               AND rpf_wg.principal_id IN (SELECT id FROM rpf_mine)
            UNION ALL
            SELECT rpf_sg."role" FROM space_grants rpf_sg CROSS JOIN rpf_target rpf_t3
             WHERE rpf_sg.workspace_id = sqlc.arg(workspace_id) AND rpf_sg.space_id = rpf_t3.space_id
               AND rpf_sg.principal_id IN (SELECT id FROM rpf_mine)
        ) rpf_g
    ), 0), COALESCE((SELECT rank FROM rpf_page_grant_rank), 0))::integer AS grant_rank;

