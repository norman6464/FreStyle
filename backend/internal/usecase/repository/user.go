// Package repository は usecase 層が依存する永続化境界（port）を定義する。
// 実装は adapter/persistence 配下で提供される（依存方向: usecase ← persistence、DIP）。
package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// UserRepository は users テーブルへのアクセスを提供する。
type UserRepository interface {
	FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error)
	// FindActiveByEmail は email で有効ユーザー（未削除・is_active）を引く。
	// ローカルのパスワードログインが使う。見つからなければ (nil, nil)。
	FindActiveByEmail(ctx context.Context, email string) (*domain.User, error)
	// CognitoSubjectByUserID はユーザーの cognito provider の OIDC subject を返す。
	// 無ければ ("", nil)。ローカルのパスワードログインのトークン発行に使う。
	CognitoSubjectByUserID(ctx context.Context, userID uint64) (string, error)
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
	// ListByRole は指定 role のユーザー一覧を返す（super_admin への一斉通知などに使う）。
	ListByRole(ctx context.Context, role domain.RoleName) ([]domain.User, error)
	// ListByCompanyID は会社単位の従業員一覧を返す（company_admin の従業員管理画面用）。
	ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.User, error)
	// CreateWithOidcIdentity は users 行と OIDC identity（provider + subject）を
	// 単一トランザクションで作成する。正規化後は「識別子を持たないユーザー」は存在し得ない
	// ため、ユーザー作成は必ず identity 作成と不可分に行う（片方だけ成功する状態を作らない）。
	CreateWithOidcIdentity(ctx context.Context, user *domain.User, provider, subject string) error
	// CreateFirstSuperAdminWithOidcIdentity は super_admin が 1 人も居ないときに限り、
	// CreateWithOidcIdentity と同じ内容（users 行 + OIDC identity）を作る。
	// 「居ないこと」の確認と作成は同一トランザクションで不可分に行い、同時に来た 2 本が
	// どちらも「0 人」を見て 2 人目を作ることを防ぐ。既に居た場合は作成せず (false, nil)。
	// user.Role が super_admin でない場合はエラー（この経路は最初の運営管理者専用）。
	CreateFirstSuperAdminWithOidcIdentity(ctx context.Context, user *domain.User, provider, subject string) (bool, error)
	// EnsureOidcIdentity は OIDC identity（provider + subject）を無ければ作る（冪等）。
	// 既存ユーザーのセルフヒール（provider 追加・張り直し）で呼ばれる。
	EnsureOidcIdentity(ctx context.Context, userID uint64, provider, subject string) error
	// UpdateAiChatEnabled は AI チャットの個別上書きを更新する（nil で会社設定に従う）。
	UpdateAiChatEnabled(ctx context.Context, userID uint64, enabled *bool) error
	// UpdateActive はユーザーアカウントの有効/無効を更新する（false で無効化 → 利用不可）。
	UpdateActive(ctx context.Context, userID uint64, active bool) error
	// SoftDelete はユーザーを論理削除する（deleted_at = NOW()）。認証時にも除外される。
	SoftDelete(ctx context.Context, userID uint64) error
	// UpdateName は氏名変更、および OIDC ログイン時の name 自動補正で呼ばれる。
	UpdateName(ctx context.Context, userID uint64, name string) error
	// UpdateRole は Cognito group → DB role 同期、または招待受諾時に呼ばれる。
	UpdateRole(ctx context.Context, userID uint64, role domain.RoleName) error
	// UpdateCompanyID は既存ユーザーが招待を受けて company に紐付くときに呼ばれる。
	UpdateCompanyID(ctx context.Context, userID uint64, companyID uint64) error
}
