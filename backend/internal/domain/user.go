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
	CompanyID    *uint64 `json:"companyId,omitempty"`
	// WorkspaceID は所属ワークスペースへの参照。CompanyID と同じ理由で NULL を許容する。
	WorkspaceID *string `json:"workspaceId,omitempty"`
	// Role はロール名（roles.name）。DB には列を持たない導出値で、読み出し時は repository が
	// roles を JOIN して role_id から解決し、書き込み時は RoleID へ変換して保存する。
	Role RoleName `json:"role"`
	// RoleID は roles マスタへの参照（正規化後の正）。repository が Role 名から解決して設定する。
	// 列の NOT NULL / DEFAULT 3（= RoleIDTrainee）は schema/core.sql が持つ。
	// この既定値は、ローリングデプロイ中の旧コード（role_id を書かない INSERT）を
	// NOT NULL 違反で壊さないための安全弁で、起動時バックフィルが role 文字列と同期する。
	RoleID int32 `json:"-"`
	// IsActive はユーザーアカウントの有効/無効。false（無効）にすると、このユーザーは
	// ログイン/利用不可になる（middleware で弾く）。super_admin / company_admin が個別に停止できる。
	IsActive  bool       `json:"isActive"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// CompanyRef は所属会社への参照を返す。未所属(company_id = NULL)は NoCompany。
// handler / usecase へは常にこの形で渡し、「未所属」を 0 に潰さない。
func (u User) CompanyRef() CompanyRef {
	if u.CompanyID == nil {
		return NoCompany()
	}
	return CompanyRefOf(*u.CompanyID)
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
