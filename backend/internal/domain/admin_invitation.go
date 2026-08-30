package domain

import "time"

type AdminInvitation struct {
	ID        uint64   `json:"id"`
	CompanyID uint64   `json:"companyId"`
	Email     string   `json:"email"`
	Role      RoleName `json:"role"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	// Token はマジックリンク用の不透明 UUID（秘匿値なので json では返さない）。
	// 未設定値を NULL にして UNIQUE 制約に引っかけないため *string にしている。
	Token     *string   `json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	// WorkspaceID は所属ワークスペースへの参照。CompanyID から dual-write されるため
	// 通常は必ず埋まっているが、Course と同じ理由で NULL を許容する型にする（FRESTYLE-297）。
	WorkspaceID *string `json:"workspaceId,omitempty"`
}

// WorkspaceRef は所属ワークスペースへの参照を返す。未設定(workspace_id = NULL)は NoWorkspace。
func (i AdminInvitation) WorkspaceRef() WorkspaceRef {
	if i.WorkspaceID == nil {
		return NoWorkspace()
	}
	return WorkspaceRefOf(*i.WorkspaceID)
}

const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusCanceled = "canceled"
)
