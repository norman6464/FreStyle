package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStatsCompanyRepo は CompanyRepository の最小 fake。
type fakeStatsCompanyRepo struct {
	rows []domain.Company
	err  error
}

func (f *fakeStatsCompanyRepo) ListAll(context.Context) ([]domain.Company, error) {
	return f.rows, f.err
}

func (f *fakeStatsCompanyRepo) FindByID(context.Context, uint64) (*domain.Company, error) {
	return nil, nil
}

func (f *fakeStatsCompanyRepo) UpdateAiChatEnabled(context.Context, uint64, bool) error { return nil }

func (f *fakeStatsCompanyRepo) UpdateActive(context.Context, uint64, bool) error { return nil }

// fakeCounter は CompanyMemberCounter の fake。
type fakeCounter struct {
	rows []repository.CompanyMemberCount
	err  error
}

func (f *fakeCounter) CountMembersByCompany(context.Context) ([]repository.CompanyMemberCount, error) {
	return f.rows, f.err
}

func Test_会社横断ビュー_会社にメンバー集計をマージして返す(t *testing.T) {
	companies := &fakeStatsCompanyRepo{rows: []domain.Company{
		{ID: 1, Name: "アクメ社", IsActive: true},
		{ID: 2, Name: "ベータ社", IsActive: false},
		{ID: 3, Name: "メンバー無し社", IsActive: true},
	}}
	counter := &fakeCounter{rows: []repository.CompanyMemberCount{
		{CompanyID: 1, Total: 5, Active: 4, Trainees: 3},
		{CompanyID: 2, Total: 2, Active: 0, Trainees: 1},
		// 会社 3 は集計に出てこない（メンバー 0）→ zero value で埋まることを検証
	}}
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
	uc := usecase.NewListCompanyStatsUseCase(
		&fakeStatsCompanyRepo{err: errors.New("db down")},
		&fakeCounter{},
	)
	_, err := uc.Execute(context.Background())
	assert.Error(t, err)
}

func Test_会社横断ビュー_集計の取得失敗を伝播(t *testing.T) {
	uc := usecase.NewListCompanyStatsUseCase(
		&fakeStatsCompanyRepo{rows: []domain.Company{{ID: 1, Name: "X"}}},
		&fakeCounter{err: errors.New("count failed")},
	)
	_, err := uc.Execute(context.Background())
	assert.Error(t, err)
}
