package domain

import "time"

// User はアプリケーション利用者のドメインモデル。
type User struct {
	ID uint64 `gorm:"primaryKey" json:"id"`
	// CognitoSub は旧カラム（正は user_oidc_identities.subject）。移行期間中のロールバック保全のため
	// 書き込みは継続する。旧カラム撤去（FRESTYLE-311 PR3）でフィールドごと削除する。
	CognitoSub string  `gorm:"column:cognito_sub;uniqueIndex" json:"cognitoSub"`
	Email      string  `gorm:"column:email;not null" json:"email"`
	Name       string  `gorm:"column:name" json:"name"`
	CompanyID  *uint64 `gorm:"column:company_id" json:"companyId,omitempty"`
	// Role はロール名（roles.name）。読み出し時は repository が roles を JOIN して解決する。
	// 旧 users.role カラムへの書き込みは移行期間中のロールバック保全のため継続し、
	// 旧カラム撤去（FRESTYLE-311 PR3）で本フィールドの column 書き込みも止める。
	Role RoleName `gorm:"column:role" json:"role"`
	// RoleID は roles マスタへの参照（正規化後の正）。repository が Role 名から解決して設定する。
	// not null / default は AutoMigrate が管理する（別で ALTER すると毎起動剥がされる）。
	// default:3 は RoleIDTrainee。ローリングデプロイ中の旧コード（role_id を書かない INSERT）を
	// NOT NULL 違反で壊さないための安全弁で、起動時バックフィルが role 文字列と同期する。
	RoleID uint16 `gorm:"column:role_id;not null;default:3" json:"-"`
	// AiChatEnabled は AI チャット利用可否の個別上書き。nil = 会社設定に従う、
	// true/false = この user 個別に強制 ON/OFF（company_admin が従業員ごとに設定）。
	AiChatEnabled *bool `gorm:"column:ai_chat_enabled" json:"aiChatEnabled,omitempty"`
	// IsActive はユーザーアカウントの有効/無効。false（無効）にすると、このユーザーは
	// ログイン/利用不可になる（middleware で弾く）。super_admin / company_admin が個別に停止できる。
	IsActive  bool       `gorm:"column:is_active;not null;default:true" json:"isActive"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deletedAt,omitempty"`
}

func (User) TableName() string { return "users" }

// CompanyIDValue は CompanyID を非ポインタで返す。未所属(nil)なら 0。
// handler/usecase で「nil なら 0」の展開を繰り返さないための小道具。
func (u User) CompanyIDValue() uint64 {
	if u.CompanyID == nil {
		return 0
	}
	return *u.CompanyID
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
