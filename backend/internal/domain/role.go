package domain

import "time"

// Role はユーザーロールのマスタ。users.role_id から参照される。
// 名前（super_admin 等）はアプリ全体で使うビジネス定数（本ファイル下部の Role* 定数）と一致させる。
type Role struct {
	// ID は固定採番（1: super_admin / 2: company_admin / 3: trainee）。migrate の seedRoles が投入する。
	ID          int32     `json:"id"`
	Name        RoleName  `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ロールマスタの固定 ID。seedRoles と結合テストで使う。
const (
	RoleIDSuperAdmin   int32 = 1
	RoleIDCompanyAdmin int32 = 2
	RoleIDTrainee      int32 = 3
)
