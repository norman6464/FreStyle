// Package repository は usecase 層が依存する永続化境界（port）を定義する。
// 実装は adapter/persistence 配下で提供される（依存方向: usecase ← persistence、DIP）。
package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrEmailTaken は作成しようとした email が既に別のアクティブユーザーに使われているときに返す
// （uq_users_email_active）。同時サインアップ・招待の二重受諾で起き得る。
var ErrEmailTaken = errors.New("email is already used by another active user")

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
	// ListByWorkspaceID はワークスペース単位のユーザー一覧を返す。
	ListByWorkspaceID(ctx context.Context, workspaceID string) ([]domain.User, error)
	// Create は users 行を 1 件作る。OIDC identity と不可分に作りたい場合は、
	// 呼び出し側（usecase）が TxManager.DoInTx の中で UserOidcIdentityRepository.EnsureIdentity と
	// 併せて呼ぶ（repository 層はまたがるテーブルのトランザクションを自前で持たない）。
	Create(ctx context.Context, user *domain.User) error
	// UpdateActive はユーザーアカウントの有効/無効を更新する（false で無効化 → 利用不可）。
	UpdateActive(ctx context.Context, userID uint64, active bool) error
	// SoftDelete はユーザーを論理削除する（deleted_at = NOW()）。認証時にも除外される。
	SoftDelete(ctx context.Context, userID uint64) error
	// UpdateName は氏名変更、および OIDC ログイン時の name 自動補正で呼ばれる。
	UpdateName(ctx context.Context, userID uint64, name string) error
	// UpdateWorkspaceID は既存ユーザーが workspace に紐付くときに呼ばれる。
	UpdateWorkspaceID(ctx context.Context, userID uint64, workspaceID *string) error
}
