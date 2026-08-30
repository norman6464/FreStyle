package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// statsCompanyRepo は ListAll の応答だけを設定した CompanyRepository の mock を返す。
func statsCompanyRepo(rows []domain.Company, err error) *mockCompanyRepo {
	repo := &mockCompanyRepo{}
	repo.On("ListAll", mock.Anything).Return(rows, err)
	return repo
}

// statsWsA / statsWsB は会社に 1:1 で紐づくワークスペース ID。
const (
	statsWsA = "0198a000-0000-7000-8000-0000000000e1"
	statsWsB = "0198a000-0000-7000-8000-0000000000e2"
)

func statsStrPtr(v string) *string { return &v }

// memberCounter は CompanyMemberCounter の mock を返す。
func memberCounter(rows []repository.WorkspaceMemberCount, err error) *mockMemberCounter {
	repo := &mockMemberCounter{}
	repo.On("CountMembersByWorkspace", mock.Anything).Return(rows, err).Maybe()
	return repo
}

func Test_会社横断ビュー_会社にメンバー集計をマージして返す(t *testing.T) {
	companies := statsCompanyRepo([]domain.Company{
		{ID: 1, Name: "アクメ社", IsActive: true, WorkspaceID: statsStrPtr(statsWsA)},
		{ID: 2, Name: "ベータ社", IsActive: false, WorkspaceID: statsStrPtr(statsWsB)},
		{ID: 3, Name: "メンバー無し社", IsActive: true},
	}, nil)
	counter := memberCounter([]repository.WorkspaceMemberCount{
		{WorkspaceID: statsWsA, Total: 5, Active: 4, Trainees: 3},
		{WorkspaceID: statsWsB, Total: 2, Active: 0, Trainees: 1},
		// 会社 3 は集計に出てこない（メンバー 0・ワークスペース未紐付け）→ zero value で埋まることを検証
	}, nil)
	uc := usecase.NewListCompanyStatsUseCase(companies, counter)

	stats, err := uc.Execute(context.Background())
	require.NoError(t, err)
	require.Len(t, stats, 3)

	assert.Equal(t, uint64(1), stats[0].ID)
	assert.Equal(t, 5, stats[0].MemberTotal)
	assert.Equal(t, 4, stats[0].ActiveMembers)
	assert.Equal(t, 3, stats[0].TraineeCount)
	assert.True(t, stats[0].IsActive)

	assert.Equal(t, 2, stats[1].MemberTotal)
	assert.Equal(t, 0, stats[1].ActiveMembers)
	assert.False(t, stats[1].IsActive)

	// メンバーがいない会社は 0 件で返る
	assert.Equal(t, "メンバー無し社", stats[2].Name)
	assert.Equal(t, 0, stats[2].MemberTotal)
	assert.Equal(t, 0, stats[2].TraineeCount)
}

func Test_会社横断ビュー_会社一覧の取得失敗を伝播(t *testing.T) {
	wantErr := errors.New("db down")
	uc := usecase.NewListCompanyStatsUseCase(
		statsCompanyRepo(nil, wantErr),
		memberCounter(nil, nil),
	)
	_, err := uc.Execute(context.Background())
	require.ErrorIs(t, err, wantErr)
}

func Test_会社横断ビュー_集計の取得失敗を伝播(t *testing.T) {
	wantErr := errors.New("count failed")
	uc := usecase.NewListCompanyStatsUseCase(
		statsCompanyRepo([]domain.Company{{ID: 1, Name: "X"}}, nil),
		memberCounter(nil, wantErr),
	)
	_, err := uc.Execute(context.Background())
	require.ErrorIs(t, err, wantErr)
}
