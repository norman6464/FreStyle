package domain

import "time"

// User はアプリケーション利用者のドメインモデル。
type User struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	// PasswordHash はパスワードログイン用の bcrypt ハッシュ。NULL = パスワードログイン不可
	// （OIDC のみ）。API へは絶対に出さない。検証はローカル専用 authenticator（infra/localauth）
	// が行い、本番のログイン経路（Cognito）はこの列を参照しない。
	PasswordHash *string `json:"-"`
	Name         string  `json:"name"`
	// WorkspaceID は所属ワークスペースへの参照。未所属は NULL。
	WorkspaceID *string `json:"workspaceId,omitempty"`
	// IsActive はユーザーアカウントの有効/無効。false（無効）にすると、このユーザーは
	// ログイン/利用不可になる（middleware で弾く）。
	IsActive  bool       `json:"isActive"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// WorkspaceRef は所属ワークスペースへの参照を返す。未所属(workspace_id = NULL)は NoWorkspace。
func (u User) WorkspaceRef() WorkspaceRef {
	if u.WorkspaceID == nil {
		return NoWorkspace()
	}
	return WorkspaceRefOf(*u.WorkspaceID)
}
