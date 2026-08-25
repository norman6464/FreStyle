package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// syncPlatformAdmin は 1 ケース分の実行をまとめる。repository は共有の testify/mock
// （repository_mocks_test.go の mockUserRepo）を使い、呼び出し条件も mock 側で固定する。
func syncPlatformAdmin(
	t *testing.T, repo *mockUserRepo, claim domain.PlatformAdminClaim,
) (bool, error) {
	t.Helper()
	changed, err := usecase.NewSyncPlatformAdminUseCase(repo).Execute(context.Background(),
		usecase.SyncPlatformAdminInput{CognitoSub: "sub-1", Claim: claim})
	repo.AssertExpectations(t)
	return changed, err
}

func Test_運営権限の同期_claimが無いときは触らない(t *testing.T) {
	repo := &mockUserRepo{}

	changed, err := syncPlatformAdmin(t, repo, domain.PlatformAdminClaimAbsent)

	require.NoError(t, err)
	require.False(t, changed)
	// claim 欠落は「グループに居ない」ではない。DB を読むことすらしない。
	repo.AssertNotCalled(t, "FindByCognitoSub", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "UpdatePlatformAdmin", mock.Anything, mock.Anything, mock.Anything)
}

func Test_運営権限の同期_グループから外れたら剥奪する(t *testing.T) {
	repo := &mockUserRepo{}
	repo.On("FindByCognitoSub", mock.Anything, "sub-1").
		Return(&domain.User{ID: 7, IsPlatformAdmin: true}, nil).Once()
	repo.On("UpdatePlatformAdmin", mock.Anything, uint64(7), false).Return(nil).Once()

	changed, err := syncPlatformAdmin(t, repo, domain.PlatformAdminClaimRevoked)

	require.NoError(t, err)
	require.True(t, changed)
}

func Test_運営権限の同期_グループに居れば付与する(t *testing.T) {
	repo := &mockUserRepo{}
	repo.On("FindByCognitoSub", mock.Anything, "sub-1").
		Return(&domain.User{ID: 9, IsPlatformAdmin: false}, nil).Once()
	repo.On("UpdatePlatformAdmin", mock.Anything, uint64(9), true).Return(nil).Once()

	changed, err := syncPlatformAdmin(t, repo, domain.PlatformAdminClaimGranted)

	require.NoError(t, err)
	require.True(t, changed)
}

func Test_運営権限の同期_既に同じ値なら書かない(t *testing.T) {
	repo := &mockUserRepo{}
	repo.On("FindByCognitoSub", mock.Anything, "sub-1").
		Return(&domain.User{ID: 9, IsPlatformAdmin: true}, nil).Once()

	changed, err := syncPlatformAdmin(t, repo, domain.PlatformAdminClaimGranted)

	require.NoError(t, err)
	require.False(t, changed)
	repo.AssertNotCalled(t, "UpdatePlatformAdmin", mock.Anything, mock.Anything, mock.Anything)
}

func Test_運営権限の同期_DB障害は握り潰さない(t *testing.T) {
	repo := &mockUserRepo{}
	repo.On("FindByCognitoSub", mock.Anything, "sub-1").
		Return((*domain.User)(nil), errors.New("db down")).Once()

	changed, err := syncPlatformAdmin(t, repo, domain.PlatformAdminClaimRevoked)

	require.Error(t, err)
	require.False(t, changed)
}

// 剥奪の書き込みが落ちたのを成功扱いにすると、退任者が管理者のまま残ったことに誰も気付けない。
func Test_運営権限の同期_更新の失敗を成功扱いにしない(t *testing.T) {
	repo := &mockUserRepo{}
	repo.On("FindByCognitoSub", mock.Anything, "sub-1").
		Return(&domain.User{ID: 7, IsPlatformAdmin: true}, nil).Once()
	repo.On("UpdatePlatformAdmin", mock.Anything, uint64(7), false).
		Return(errors.New("db down")).Once()

	changed, err := syncPlatformAdmin(t, repo, domain.PlatformAdminClaimRevoked)

	require.Error(t, err)
	require.False(t, changed)
}

func Test_運営権限の同期_ユーザーが居なければ何もしない(t *testing.T) {
	repo := &mockUserRepo{}
	repo.On("FindByCognitoSub", mock.Anything, "sub-1").
		Return((*domain.User)(nil), nil).Once()

	changed, err := syncPlatformAdmin(t, repo, domain.PlatformAdminClaimRevoked)

	require.NoError(t, err)
	require.False(t, changed)
	repo.AssertNotCalled(t, "UpdatePlatformAdmin", mock.Anything, mock.Anything, mock.Anything)
}

func Test_運営権限の同期_subが空ならエラー(t *testing.T) {
	repo := &mockUserRepo{}

	changed, err := usecase.NewSyncPlatformAdminUseCase(repo).Execute(context.Background(),
		usecase.SyncPlatformAdminInput{CognitoSub: "", Claim: domain.PlatformAdminClaimRevoked})

	require.Error(t, err)
	require.False(t, changed)
	repo.AssertExpectations(t)
}
