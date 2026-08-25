package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/require"
)

// syncPlatformAdminRepo は SyncPlatformAdminUseCase が触る 2 メソッドだけを持つ最小の spy。
type syncPlatformAdminRepo struct {
	stubUserRepo
	existing    *domain.User
	findErr     error
	updateCalls int
	updatedID   uint64
	updatedVal  bool
	updateErr   error
}

func (s *syncPlatformAdminRepo) FindByCognitoSub(context.Context, string) (*domain.User, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}
	return s.existing, nil
}

func (s *syncPlatformAdminRepo) UpdatePlatformAdmin(_ context.Context, userID uint64, v bool) error {
	s.updateCalls++
	if s.updateErr != nil {
		return s.updateErr
	}
	s.updatedID, s.updatedVal = userID, v
	return nil
}

func Test_運営権限の同期_claimが無いときは触らない(t *testing.T) {
	repo := &syncPlatformAdminRepo{existing: &domain.User{ID: 7, IsPlatformAdmin: true}}
	changed, err := NewSyncPlatformAdminUseCase(repo).Execute(context.Background(),
		SyncPlatformAdminInput{CognitoSub: "sub-1", Claim: domain.PlatformAdminClaimAbsent})

	require.NoError(t, err)
	require.False(t, changed)
	require.Zero(t, repo.updateCalls, "claim 欠落は「グループに居ない」ではない")
}

func Test_運営権限の同期_グループから外れたら剥奪する(t *testing.T) {
	repo := &syncPlatformAdminRepo{existing: &domain.User{ID: 7, IsPlatformAdmin: true}}
	changed, err := NewSyncPlatformAdminUseCase(repo).Execute(context.Background(),
		SyncPlatformAdminInput{CognitoSub: "sub-1", Claim: domain.PlatformAdminClaimRevoked})

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, 1, repo.updateCalls)
	require.Equal(t, uint64(7), repo.updatedID)
	require.False(t, repo.updatedVal)
}

func Test_運営権限の同期_グループに居れば付与する(t *testing.T) {
	repo := &syncPlatformAdminRepo{existing: &domain.User{ID: 9, IsPlatformAdmin: false}}
	changed, err := NewSyncPlatformAdminUseCase(repo).Execute(context.Background(),
		SyncPlatformAdminInput{CognitoSub: "sub-1", Claim: domain.PlatformAdminClaimGranted})

	require.NoError(t, err)
	require.True(t, changed)
	require.True(t, repo.updatedVal)
}

func Test_運営権限の同期_既に同じ値なら書かない(t *testing.T) {
	repo := &syncPlatformAdminRepo{existing: &domain.User{ID: 9, IsPlatformAdmin: true}}
	changed, err := NewSyncPlatformAdminUseCase(repo).Execute(context.Background(),
		SyncPlatformAdminInput{CognitoSub: "sub-1", Claim: domain.PlatformAdminClaimGranted})

	require.NoError(t, err)
	require.False(t, changed)
	require.Zero(t, repo.updateCalls)
}

func Test_運営権限の同期_DB障害は握り潰さない(t *testing.T) {
	repo := &syncPlatformAdminRepo{findErr: errors.New("db down")}
	_, err := NewSyncPlatformAdminUseCase(repo).Execute(context.Background(),
		SyncPlatformAdminInput{CognitoSub: "sub-1", Claim: domain.PlatformAdminClaimRevoked})

	require.Error(t, err)
}

func Test_運営権限の同期_ユーザーが居なければ何もしない(t *testing.T) {
	repo := &syncPlatformAdminRepo{}
	changed, err := NewSyncPlatformAdminUseCase(repo).Execute(context.Background(),
		SyncPlatformAdminInput{CognitoSub: "sub-1", Claim: domain.PlatformAdminClaimRevoked})

	require.NoError(t, err)
	require.False(t, changed)
	require.Zero(t, repo.updateCalls)
}
