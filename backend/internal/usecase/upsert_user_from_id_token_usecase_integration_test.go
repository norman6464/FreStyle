//go:build integration

package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

type updateThenFailInvitationStatusUpdater struct {
	inner repository.InvitationStatusUpdater
	err   error
}

func (u *updateThenFailInvitationStatusUpdater) UpdateStatus(
	ctx context.Context,
	id uint64,
	status string,
) error {
	if err := u.inner.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	return u.err
}

type failingInvitationTransactionRunner struct {
	inner repository.UserInvitationTransactionRunner
	err   error
}

func (r *failingInvitationTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(
		users repository.UserWithOidcIdentityCreator,
		invitations repository.InvitationStatusUpdater,
	) error,
) error {
	return r.inner.WithinTransaction(
		ctx,
		func(
			users repository.UserWithOidcIdentityCreator,
			invitations repository.InvitationStatusUpdater,
		) error {
			return fn(
				users,
				&updateThenFailInvitationStatusUpdater{
					inner: invitations,
					err:   r.err,
				},
			)
		},
	)
}

func TestUpsertUserFromIDToken_Transaction_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Cleanup(func() {
		testsupport.TruncateAll(t, db, "users", "invitations")
	})

	t.Run("ユーザー作成と招待受諾をまとめてコミットする", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "invitations")

		users := persistence.NewUserRepository(db)
		invitations := persistence.NewAdminInvitationRepository(db)
		runner := persistence.NewUserInvitationTransactionRunner(db)

		invitation := &domain.AdminInvitation{
			CompanyID: 42,
			Email:     "commit@example.com",
			Name:      "Commit User",
			Role:      domain.RoleTrainee,
			Status:    domain.InvitationStatusPending,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		require.NoError(t, invitations.Create(ctx, invitation))

		uc := usecase.NewUpsertUserFromIDTokenUseCase(
			users,
			invitations,
			"",
			runner,
		)

		allowed, err := uc.Execute(
			ctx,
			usecase.UpsertUserFromIDTokenInput{
				CognitoSub: "commit-sub",
				Email:      invitation.Email,
			},
		)

		require.NoError(t, err)
		require.True(t, allowed)

		created, err := users.FindByCognitoSub(ctx, "commit-sub")
		require.NoError(t, err)
		require.NotNil(t, created)

		accepted, err := invitations.FindByID(ctx, invitation.ID)
		require.NoError(t, err)
		require.NotNil(t, accepted)
		require.Equal(
			t,
			domain.InvitationStatusAccepted,
			accepted.Status,
		)
	})

	t.Run("招待更新失敗時にユーザー作成もロールバックする", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "invitations")

		users := persistence.NewUserRepository(db)
		invitations := persistence.NewAdminInvitationRepository(db)
		updateErr := errors.New("forced invitation update failure")
		runner := &failingInvitationTransactionRunner{
			inner: persistence.NewUserInvitationTransactionRunner(db),
			err:   updateErr,
		}

		invitation := &domain.AdminInvitation{
			CompanyID: 42,
			Email:     "rollback@example.com",
			Name:      "Rollback User",
			Role:      domain.RoleTrainee,
			Status:    domain.InvitationStatusPending,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		require.NoError(t, invitations.Create(ctx, invitation))

		uc := usecase.NewUpsertUserFromIDTokenUseCase(
			users,
			invitations,
			"",
			runner,
		)

		allowed, err := uc.Execute(
			ctx,
			usecase.UpsertUserFromIDTokenInput{
				CognitoSub: "rollback-sub",
				Email:      invitation.Email,
			},
		)

		require.False(t, allowed)
		require.ErrorIs(t, err, updateErr)

		created, err := users.FindByCognitoSub(ctx, "rollback-sub")
		require.NoError(t, err)
		require.Nil(t, created)

		pending, err := invitations.FindByID(ctx, invitation.ID)
		require.NoError(t, err)
		require.NotNil(t, pending)
		require.Equal(
			t,
			domain.InvitationStatusPending,
			pending.Status,
		)
	})
}
