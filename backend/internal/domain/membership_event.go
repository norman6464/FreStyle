package domain

import "time"

// MembershipEventAction は membership_events.action に入る操作の種類。
type MembershipEventAction string

const (
	// MembershipEventMemberAdded は招待の手順を踏まず所属が直接 active になった
	// （自分でワークスペースを作った本人）。
	MembershipEventMemberAdded MembershipEventAction = "member_added"
	// MembershipEventInvited は招待を発行した（まだ principal も権限も無い）。
	MembershipEventInvited MembershipEventAction = "invited"
	// MembershipEventInvitationAccepted は招待を受諾し、principal と既定の役割が付いた。
	MembershipEventInvitationAccepted MembershipEventAction = "invitation_accepted"
	// MembershipEventInvitationDeclined は招待を辞退した。
	MembershipEventInvitationDeclined MembershipEventAction = "invitation_declined"
	// MembershipEventRoleChanged はワークスペース全体での既定の役割が変わった
	// （付与・変更・剥奪のいずれも同じ action。NewLabel が nil なら剥奪）。
	MembershipEventRoleChanged MembershipEventAction = "role_changed"
	// MembershipEventMemberRemoved は本人以外（admin）がその人の所属を終えた。
	MembershipEventMemberRemoved MembershipEventAction = "member_removed"
	// MembershipEventLeft は本人が自分の所属を終えた。
	MembershipEventLeft MembershipEventAction = "left"
	// MembershipEventSuspended は所属を止めた（段 7 の停止 API 用。書き込むコードは
	// まだ無いが、schema.hcl の CHECK を後から DROP + ADD し直さないよう先に列挙してある）。
	MembershipEventSuspended MembershipEventAction = "suspended"
)

// MembershipEvent は所属・権限が変わった 1 件の記録（段 6・監査）。
//
// OldLabel / NewLabel は当時の表示名の写し（domain.TicketChangeItem と同じ設計。
// schema.hcl の table "membership_events" のコメント参照）。役割名・所属状態名は
// 後から改名されない固定の小さな値なので、ticket の OldValue/NewValue に相当する
// 別列は持たない。
type MembershipEvent struct {
	ID           string                `json:"id"`
	TargetUserID uint64                `json:"targetUserId"`
	ActorUserID  uint64                `json:"actorUserId"`
	Action       MembershipEventAction `json:"action"`
	OldLabel     *string               `json:"oldLabel,omitempty"`
	NewLabel     *string               `json:"newLabel,omitempty"`
	CreatedAt    time.Time             `json:"createdAt"`
}
