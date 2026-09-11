-- name: InsertActiveWorkspaceMember :exec
-- 自分でワークスペースを作った／個人ワークスペースを自動作成した本人を、招待の手順を踏まず
-- 直接 active な所属として記録する。principal の作成（InsertPrincipal）と同じ
-- トランザクションで呼ぶこと（workspace_members のコメントにある procedural invariant）。
INSERT INTO workspace_members (workspace_id, user_id, status, joined_at, created_at, updated_at)
VALUES ($1, $2, 'active', now(), now(), now());

-- name: UpsertInvitedWorkspaceMember :execrows
-- 招待中の所属を作る。既存行が無ければ invited で新規作成し、left/suspended だった相手には
-- invited へ戻して再招待できるようにする（invited_by_user_id も招いた人へ更新）。
-- 既に active/invited の行には触らない（0 件で返る）。呼び出し側はこれを「今回新しく
-- 招待状態にしたか」の判定には使わず、常に成功として扱ってよい（同じ意味の状態へ収束する）。
INSERT INTO workspace_members (workspace_id, user_id, status, invited_by_user_id, created_at, updated_at)
VALUES ($1, $2, 'invited', $3, now(), now())
ON CONFLICT (workspace_id, user_id) DO UPDATE SET
  status              = 'invited',
  invited_by_user_id  = EXCLUDED.invited_by_user_id,
  left_at             = NULL,
  updated_at          = now()
WHERE workspace_members.status IN ('left', 'suspended');

-- name: ActivateWorkspaceMembership :execrows
-- invited → active。principal の作成（EnsureUserPrincipal 相当）と同じトランザクションで
-- 呼ぶこと。0 件なら invited の行が無い（招待されていない・既に受諾済み）。
UPDATE workspace_members SET status = 'active', joined_at = now(), updated_at = now()
WHERE workspace_id = $1 AND user_id = $2 AND status = 'invited';

-- name: DeclineWorkspaceInvitation :execrows
-- invited → left（辞退）。0 件なら invited の行が無い。
UPDATE workspace_members SET status = 'left', left_at = now(), updated_at = now()
WHERE workspace_id = $1 AND user_id = $2 AND status = 'invited';

-- name: LeaveWorkspaceMembership :execrows
-- active/invited → left（退出・削除・招待の取り消し）。0 件なら既に left/suspended か、
-- そもそも所属したことが無い。呼び出し側はどちらも「非メンバーになった」として
-- 冪等に成功扱いする（行の有無を見ない）。
UPDATE workspace_members SET status = 'left', left_at = now(), updated_at = now()
WHERE workspace_id = $1 AND user_id = $2 AND status IN ('active', 'invited');

-- name: ListMyWorkspaceInvitations :many
-- 自分宛の未受諾の招待を新しい順で返す。停止中のワークスペースからの招待は出さない
-- （受諾しても入れないものを見せない。ResolveWorkspaceUseCase が停止中を無いものとして
-- 扱うのと同じ判断）。
SELECT w.slug AS workspace_slug, w.name AS workspace_name,
       wm.invited_by_user_id, wm.created_at AS invited_at
FROM workspace_members wm
JOIN workspaces w ON w.id = wm.workspace_id
WHERE wm.user_id = $1 AND wm.status = 'invited' AND w.is_active
ORDER BY wm.created_at DESC;
