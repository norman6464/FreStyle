-- name: ListSpaceMemberGrantFacts :many
-- スペースに届いている権限を、人ごとの複数経路のまま返す（段 9）。集約（最も強い役割を選び、
-- どの経路が勝ったかを決める）は Go 側（ListSpaceMembers repository メソッド）が行う —
-- ResolvePagePermissionFacts と同じ「SQL は事実の集合を返すだけ、合成は 1 箇所の Go 関数に
-- 集める」方針（domain.GrantRole.Rank の doc 参照）。
--
-- space_reachable CTE は 1 人分の継承規則を返す ListSpaceScopeGrantRoles と同じ形（本人の主体 /
-- 所属グループの主体 / そのスペースが visibility='workspace' のときだけ加わる space_all の
-- 主体）を、ワークスペース内の全員に対して一度に展開したもの。列はすべてテーブル別名で
-- 完全修飾する（sqlc の型推論が裸の列参照でうまく解決できない場合があるため）。
WITH target_space AS (
  SELECT sv.visibility AS visibility FROM spaces sv
  WHERE sv.workspace_id = $1 AND sv.id = $2
),
users_in_ws AS (
  SELECT pr2.id AS principal_id, pr2.user_id AS user_id FROM principals pr2
  WHERE pr2.workspace_id = $1 AND pr2.kind = 'user'
),
space_reachable AS (
  SELECT uw1.principal_id AS user_principal_id, uw1.principal_id AS mine_id FROM users_in_ws uw1
  UNION ALL
  SELECT uw2.principal_id AS user_principal_id, pm.group_principal_id AS mine_id
  FROM users_in_ws uw2
  JOIN principal_members pm ON pm.workspace_id = $1
    AND pm.member_principal_id = uw2.principal_id
  UNION ALL
  -- space_all はそのスペースの「全員」を表す主体。ワークスペース全体の grant と同じく、
  -- 対象スペースが visibility='workspace'（プライベートでない）のときだけ全員に届く。
  SELECT uw3.principal_id AS user_principal_id, sp.id AS mine_id
  FROM users_in_ws uw3
  CROSS JOIN principals sp
  WHERE sp.workspace_id = $1 AND sp.kind = 'space_all'
    AND sp.space_id = $2
    AND EXISTS (SELECT 1 FROM target_space ts1 WHERE ts1.visibility = 'workspace')
),
space_role_facts AS (
  SELECT m1.user_principal_id AS user_principal_id, wg.role AS role, 'workspace'::text AS source
  FROM workspace_grants wg
  JOIN space_reachable m1 ON m1.mine_id = wg.principal_id
  WHERE wg.workspace_id = $1
    AND EXISTS (SELECT 1 FROM target_space ts2 WHERE ts2.visibility = 'workspace')
  UNION ALL
  SELECT m2.user_principal_id AS user_principal_id, sg.role AS role,
    CASE WHEN sg.principal_id = m2.user_principal_id THEN 'direct' ELSE 'group' END AS source
  FROM space_grants sg
  JOIN space_reachable m2 ON m2.mine_id = sg.principal_id
  WHERE sg.workspace_id = $1 AND sg.space_id = $2
)
SELECT
  u.id AS user_id,
  u.name AS name,
  COALESCE(pr.avatar_url, '') AS avatar_url,
  f.role AS role,
  f.source AS source
FROM space_role_facts f
JOIN users_in_ws uw ON uw.principal_id = f.user_principal_id
-- 停止中・退会済みのアカウント（users.status <> 'active'）はこの一覧には出さない
-- （ListWorkspaceMembers と同じ判断基準。復帰の入口が要る管理画面
-- 〈ListWorkspaceMembersForAdmin〉とは違い、こちらは「今スペースにアクセスできる人」の
-- 一覧のため。principal 行はユーザーの退会・停止だけでは自動では消えないので、
-- ここで明示的に絞る必要がある）。
JOIN users u ON u.id = uw.user_id AND u.status = 'active'
JOIN workspace_members wm ON wm.workspace_id = $1 AND wm.user_id = u.id
  AND wm.status = 'active'
LEFT JOIN profiles pr ON pr.user_id = u.id;
