package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

type userInvitationTransactionRunner struct {
	db *sql.DB
}

// NewUserInvitationTransactionRunner はユーザー作成と招待更新の
// トランザクション実装を生成する。
func NewUserInvitationTransactionRunner(
	db *sql.DB,
) repository.UserInvitationTransactionRunner {
	return &userInvitationTransactionRunner{db: db}
}

type txUserWithOidcIdentityCreator struct {
	q *sqlcgen.Queries
}

func (c *txUserWithOidcIdentityCreator) CreateWithOidcIdentity(
	ctx context.Context,
	user *domain.User,
	provider string,
	subject string,
) error {
	return createWithOidcIdentity(ctx, c.q, user, provider, subject)
}

type txInvitationStatusUpdater struct {
	q *sqlcgen.Queries
}

func (u *txInvitationStatusUpdater) UpdateStatus(
	ctx context.Context,
	id uint64,
	status string,
) error {
	return updateInvitationStatus(ctx, u.q, id, status)
}

func (r *userInvitationTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(
		users repository.UserWithOidcIdentityCreator,
		invitations repository.InvitationStatusUpdater,
	) error,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := sqlcgen.New(tx)
	if err := fn(
		&txUserWithOidcIdentityCreator{q: qtx},
		&txInvitationStatusUpdater{q: qtx},
	); err != nil {
		return err
	}

	return tx.Commit()
}
