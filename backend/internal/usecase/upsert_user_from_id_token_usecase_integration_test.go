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

type failingInvitationRepository struct {
	repository.AdminInvitationRepository
	err error
}

func (r *failingInvitationRepository) UpdateStatus(
	_ context.Context,
	_ uint64,
	_ string,
) error {
	return r.err
}

type failingInvitationTransactionRunner struct {
	inner repository.UserInvitationTransactionRunner
	err   error
}

type synchronizedInvitationRepository struct {
	repository.AdminInvitationRepository
	ready   chan<- struct{}
	release <-chan struct{}
}

func (r *synchronizedInvitationRepository) UpdateStatus(
	ctx context.Context,
	id uint64,
	status string,
) error {
	r.ready <- struct{}{}
	<-r.release

	return r.AdminInvitationRepository.UpdateStatus(
		ctx,
		id,
		status,
	)
}

type synchronizedInvitationTransactionRunner struct {
	inner   repository.UserInvitationTransactionRunner
	ready   chan<- struct{}
	release <-chan struct{}
}

func (r *synchronizedInvitationTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(
		users repository.UserRepository,
		invitations repository.AdminInvitationRepository,
	) error,
) error {
	return r.inner.WithinTransaction(
		ctx,
		func(
			users repository.UserRepository,
			invitations repository.AdminInvitationRepository,
		) error {
			return fn(
				users,
				&synchronizedInvitationRepository{
					AdminInvitationRepository: invitations,
					ready:                     r.ready,
					release:                   r.release,
				},
			)
		},
	)
}

func (r *failingInvitationTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(
		users repository.UserRepository,
		invitations repository.AdminInvitationRepository,
	) error,
) error {
	return r.inner.WithinTransaction(
		ctx,
		func(
			users repository.UserRepository,
			invitations repository.AdminInvitationRepository,
		) error {
			return fn(
				users,
				&failingInvitationRepository{
					AdminInvitationRepository: invitations,
					err:                       r.err,
				},
			)
		},
	)
}

func TestUpsertUserFromIDToken_Transaction_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()

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

	t.Run("同じ招待tokenの並行利用ではユーザーを1件だけ作成する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "invitations")

		users := persistence.NewUserRepository(db)
		invitations := persistence.NewAdminInvitationRepository(db)

		token := "concurrent-invitation-token"
		invitation := &domain.AdminInvitation{
			CompanyID: 42,
			Email:     "concurrent@example.com",
			Name:      "Concurrent User",
			Role:      domain.RoleTrainee,
			Status:    domain.InvitationStatusPending,
			Token:     &token,
			ExpiresAt: time.Now().UTC().Add(time.Hour),
		}
		require.NoError(t, invitations.Create(ctx, invitation))

		ready := make(chan struct{}, 2)
		release := make(chan struct{})

		// 途中でテストが失敗しても、待機中のgoroutineを解放する。
		defer func() {
			select {
			case <-release:
			default:
				close(release)
			}
		}()

		runner := &synchronizedInvitationTransactionRunner{
			inner:   persistence.NewUserInvitationTransactionRunner(db),
			ready:   ready,
			release: release,
		}
		uc := usecase.NewUpsertUserFromIDTokenUseCase(
			users,
			invitations,
			runner,
		)

		type executeResult struct {
			allowed bool
			err     error
		}

		results := make(chan executeResult, 2)
		execute := func(cognitoSub string) {
			allowed, err := uc.Execute(
				ctx,
				usecase.UpsertUserFromIDTokenInput{
					CognitoSub:      cognitoSub,
					Email:           invitation.Email,
					InvitationToken: token,
				},
			)
			results <- executeResult{
				allowed: allowed,
				err:     err,
			}
		}

		go execute("concurrent-sub-1")
		go execute("concurrent-sub-2")

		for i := 0; i < 2; i++ {
			select {
			case <-ready:
			case <-time.After(5 * time.Second):
				t.Fatal("2つの処理が招待更新直前まで到達しなかった")
			}
		}

		close(release)

		first := <-results
		second := <-results

		successCount := 0
		failureCount := 0
		for _, result := range []executeResult{first, second} {
			switch {
			case result.allowed && result.err == nil:
				successCount++
			case !result.allowed && result.err != nil:
				failureCount++
			}
		}

		require.Equal(t, 1, successCount)
		require.Equal(t, 1, failureCount)

		var userCount int64
		require.NoError(
			t,
			db.Model(&domain.User{}).
				Where(
					"cognito_sub IN ?",
					[]string{"concurrent-sub-1", "concurrent-sub-2"},
				).
				Count(&userCount).
				Error,
		)
		require.Equal(t, int64(1), userCount)

		accepted, err := invitations.FindByID(ctx, invitation.ID)
		require.NoError(t, err)
		require.NotNil(t, accepted)
		require.Equal(
			t,
			domain.InvitationStatusAccepted,
			accepted.Status,
		)
	})
}
