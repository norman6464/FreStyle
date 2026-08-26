package domain

import "time"

// AuditEvent は運営/管理者の重要操作（会社の有効/無効・従業員の停止/削除・招待など）の監査記録。
// actor_email / actor_role は、actor が後で削除・変更されても「誰の操作か」を追えるよう非正規化して残す。
type AuditEvent struct {
	ID         uint64 `json:"id"`
	ActorID    uint64 `json:"actorId"`
	ActorEmail string `json:"actorEmail"`
	ActorRole  string `json:"actorRole"`
	// Action は「METHOD ルートパターン」（例: "PATCH /api/v2/admin/companies/:id/active"）。
	Action string `json:"action"`
	// TargetID は操作対象の ID（会社 ID / ユーザー ID など。取得できないときは 0）。
	TargetID  uint64    `json:"targetId"`
	CreatedAt time.Time `json:"createdAt"`
}
