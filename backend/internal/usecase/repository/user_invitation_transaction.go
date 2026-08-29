package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// UserWithOidcIdentityCreator はユーザーとOIDC identityを作成する。
type UserWithOidcIdentityCreator interface {
	CreateWithOidcIdentity(
		ctx context.Context,
		user *domain.User,
		provider string,
		subject string,
	) error
}

// InvitationStatusUpdater は招待ステータスを更新する。
type InvitationStatusUpdater interface {
	UpdateStatus(ctx context.Context, id uint64, status string) error
}

// UserInvitationTransactionRunner はユーザー・OIDC identityの作成と
// 招待更新を単一のトランザクション内で実行する。
type UserInvitationTransactionRunner interface {
	WithinTransaction(
		ctx context.Context,
		fn func(
			users UserWithOidcIdentityCreator,
			invitations InvitationStatusUpdater,
		) error,
	) error
}
