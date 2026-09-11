package domain

import "time"

// UserStatus はユーザーアカウントの状態。
type UserStatus string

const (
	// UserStatusActive は通常どおり利用できる状態。
	UserStatusActive UserStatus = "active"
	// UserStatusSuspended は運営判断で一時的に止めた状態。本人の操作では戻せない。
	UserStatusSuspended UserStatus = "suspended"
	// UserStatusDeactivated は退会済み。実際の退会は匿名化で扱い、物理削除はしない
	// （users.id を指す記録列の FK が RESTRICT のため、DB 側でも禁じられている）。
	UserStatusDeactivated UserStatus = "deactivated"
)

// User はアプリケーション利用者のドメインモデル。
//
// 所属ワークスペースへの参照はここには無い（段 2 で撤去済みの旧 workspace_id）。
// 1 人が複数のワークスペースに所属できるため、所属は workspace_members が正本で持つ
// （usecase/kb.ListMemberWorkspacesUseCase 等を参照）。
type User struct {
	ID    uint64 `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	// Status はアカウントの状態。deactivated のときだけ DeletedAt が非 nil になる
	// （DB の ck_users_status_deleted_at が両者の整合を縛る）。
	Status    UserStatus `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// IsActive はアカウントが通常どおり利用できる状態かを返す。false なら
// ログイン/利用不可になる（middleware で弾く）。
func (u User) IsActive() bool { return u.Status == UserStatusActive }
