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
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
	// ListByRole は指定 role のユーザー一覧を返す（super_admin への一斉通知などに使う）。
	ListByRole(ctx context.Context, role string) ([]domain.User, error)
	// ListByCompanyID は会社単位の従業員一覧を返す（company_admin の従業員管理画面用）。
	ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	// UpdateAiChatEnabled は AI チャットの個別上書きを更新する（nil で会社設定に従う）。
	UpdateAiChatEnabled(ctx context.Context, userID uint64, enabled *bool) error
	// UpdateDisplayName は氏名変更、および OIDC ログイン時の displayName 自動補正で呼ばれる。
	UpdateDisplayName(ctx context.Context, userID uint64, displayName string) error
	// UpdateRole は Cognito group → DB role 同期、または招待受諾時に呼ばれる。
	UpdateRole(ctx context.Context, userID uint64, role string) error
	// UpdateCompanyID は既存ユーザーが招待を受けて company に紐付くときに呼ばれる。
	UpdateCompanyID(ctx context.Context, userID uint64, companyID uint64) error
	// MarkOnboarded は onboarded_at = NOW() に更新する（冪等、既存値は上書きしない）。
	MarkOnboarded(ctx context.Context, userID uint64) error
}
