package repository

import "context"

// UserInvitationTransactionRunner はユーザーと招待の更新を
// 単一トランザクション内で実行する。
type UserInvitationTransactionRunner interface {
	WithinTransaction(
		ctx context.Context,
		fn func(
			users UserRepository,
			invitations AdminInvitationRepository,
		) error,
	) error
}
