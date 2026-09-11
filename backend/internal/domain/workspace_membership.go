package domain

import "time"

// MembershipStatus はワークスペース所属のライフサイクル状態（workspace_members.status）。
//
// スキーマの正本は schema.hcl の table "workspace_members" のコメント（procedural invariant
// を含む）を参照。
type MembershipStatus string

const (
	// MembershipStatusInvited は管理者が招いたが本人がまだ受諾していない状態。
	// 権限はまだ何も届かない（principals(kind='user') の行がまだ無い）。
	MembershipStatusInvited MembershipStatus = "invited"
	// MembershipStatusActive は実際のメンバー。principals(kind='user') の対応する行がある。
	MembershipStatusActive MembershipStatus = "active"
	// MembershipStatusSuspended は運営判断で一時的に外した状態（今の usecase はまだ
	// 書き込まない。将来の管理操作のための予約）。
	MembershipStatusSuspended MembershipStatus = "suspended"
	// MembershipStatusLeft は離脱・招待の辞退・削除。行は消さず記録として残す。
	MembershipStatusLeft MembershipStatus = "left"
)

// WorkspaceInvitation は「自分宛の未受諾の招待」1 件（ListMyWorkspaceInvitationsUseCase が返す）。
type WorkspaceInvitation struct {
	// WorkspaceSlug は受諾・辞退の URL に使う識別子。
	WorkspaceSlug string `json:"workspaceSlug"`
	WorkspaceName string `json:"workspaceName"`
	// InvitedByUserID は招いた人。招待の起点が常にある（自分でワークスペースを作る経路は
	// invited の状態を経ないため）ので nil にはならない。
	InvitedByUserID uint64    `json:"invitedByUserId"`
	InvitedAt       time.Time `json:"invitedAt"`
}
