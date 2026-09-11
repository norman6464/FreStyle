-- 所属・権限の変更履歴（段 6・監査）。追記のみで UPDATE / DELETE はしない。
-- schema.hcl の table "membership_events" のコメントに設計の理由がある。

-- name: InsertMembershipEvent :exec
-- 必ず、記録対象の書き込み（workspace_members / workspace_grants への UPDATE や principals の
-- 作成・削除）と同じトランザクションで呼ぶこと。片方だけ書けると履歴が実際の状態とずれる。
INSERT INTO membership_events (
  id, workspace_id, target_user_id, actor_user_id, action, old_label, new_label, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, now());

-- name: ListMembershipEvents :many
-- ワークスペース 1 つ分の履歴を新しい順で返す。
SELECT * FROM membership_events
WHERE workspace_id = $1
ORDER BY created_at DESC, id DESC;

-- name: GetWorkspaceGrant :one
-- 主体 1 件のワークスペース既定の役割を 1 行だけ引く。UpsertWorkspaceGrant /
-- DeleteWorkspaceGrant が「変更前の役割」を履歴に残すために、書き換える前に読む
-- （無ければ ErrNoRows = まだ役割が無い）。
SELECT "role" FROM workspace_grants
WHERE workspace_id = $1 AND principal_id = $2;
