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
	// WorkspaceID は所属ワークスペースへの参照。未所属（運営管理者等）は NULL。
	WorkspaceID *string `json:"workspaceId,omitempty"`
	// Role はロール名。users.role 列そのもので、名前と ID を突き合わせる変換は挟まらない
	// （かつては roles マスタ表への FK だったが撤去した。理由は schema/schema.sql の users を参照）。
	Role RoleName `json:"role"`
	// IsActive はユーザーアカウントの有効/無効。false（無効）にすると、このユーザーは
	// ログイン/利用不可になる（middleware で弾く）。super_admin / company_admin が個別に停止できる。
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

// RoleName はユーザーロール名（roles.name）の型。生の string の写経とタイポを
// コンパイル時に弾くための named type（JSON へは underlying string としてそのまま出る）。
type RoleName string

const (
	RoleSuperAdmin   RoleName = "super_admin"
	RoleCompanyAdmin RoleName = "company_admin"
	RoleTrainee      RoleName = "trainee"
)

// Valid は既知のロール名かを返す。外部入力（招待作成リクエスト等）の検証に使う。
func (r RoleName) Valid() bool {
	switch r {
	case RoleSuperAdmin, RoleCompanyAdmin, RoleTrainee:
		return true
	}
	return false
}
