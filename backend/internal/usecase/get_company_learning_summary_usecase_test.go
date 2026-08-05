package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// summarizerArgs は mock が受け取った引数の控え。usecase が「何を渡したか」を検証するために使う。
type summarizerArgs struct {
	companyID uint64
	fromDate  time.Time
}

// learningActivityRepo は指定の結果を返す mock と、渡された引数の控えを返す。
func learningActivityRepo(rows []repository.MemberLearningActivity, err error) (*mockLearningSummarizer, *summarizerArgs) {
	got := &summarizerArgs{}
	repo := &mockLearningSummarizer{}
	repo.On("ListMemberActivities", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			got.companyID = args.Get(1).(uint64)
			got.fromDate = args.Get(2).(time.Time)
		}).Return(rows, err).Maybe()
	return repo, got
}

func companyAdminActor(companyID uint64) *domain.User {
	return &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &companyID}
}

func datePtr(t time.Time) *time.Time { return &t }

func Test_メンバー学習サマリー_集計される(t *testing.T) {
	today := time.Now().UTC()
	yesterday := today.AddDate(0, 0, -1)
	tenDaysAgo := today.AddDate(0, 0, -10)
	repo, got := learningActivityRepo([]repository.MemberLearningActivity{
		{UserID: 11, Name: "今日学習した人", LastActiveDate: datePtr(today), RecentActivityCount: 3},
		{UserID: 12, Name: "昨日学習した人", LastActiveDate: datePtr(yesterday), RecentActivityCount: 1},
		{UserID: 13, Name: "先週以前の人", LastActiveDate: datePtr(tenDaysAgo), RecentActivityCount: 0},
		{UserID: 14, Name: "未学習の人", LastActiveDate: nil, RecentActivityCount: 0},
	}, nil)
	uc := usecase.NewGetCompanyLearningSummaryUseCase(repo)

	out, err := uc.Execute(context.Background(), companyAdminActor(10))
	require.NoError(t, err)
	assert.Equal(t, uint64(10), got.companyID)
	assert.Equal(t, 4, out.TraineeCount)
	assert.Equal(t, 1, out.ActiveToday)
	assert.Equal(t, 2, out.ActiveThisWeek, "直近 7 日の活動回数 > 0 の 2 名")
	require.Len(t, out.RecentMembers, 2, "直近 7 日に活動が無い trainee は直近アクティブに含めない")
	assert.Equal(t, "今日学習した人", out.RecentMembers[0].Name)
	assert.Equal(t, today.Format("2006-01-02"), out.RecentMembers[0].LastActiveDate)
	assert.Equal(t, "昨日学習した人", out.RecentMembers[1].Name)
}

func Test_メンバー学習サマリー_直近リストは5名まで(t *testing.T) {
	today := time.Now().UTC()
	rows := make([]repository.MemberLearningActivity, 0, 7)
	for i := 0; i < 7; i++ {
		rows = append(rows, repository.MemberLearningActivity{
			UserID: uint64(100 + i), Name: fmt.Sprintf("member-%d", i),
			LastActiveDate: datePtr(today.AddDate(0, 0, -i)), RecentActivityCount: 1,
		})
	}
	repo, _ := learningActivityRepo(rows, nil)
	uc := usecase.NewGetCompanyLearningSummaryUseCase(repo)

	out, err := uc.Execute(context.Background(), companyAdminActor(10))
	require.NoError(t, err)
	assert.Equal(t, 7, out.TraineeCount)
	assert.Len(t, out.RecentMembers, 5)
}

func Test_メンバー学習サマリー_会社未所属は空サマリー(t *testing.T) {
	repo, _ := learningActivityRepo(nil, nil)
	uc := usecase.NewGetCompanyLearningSummaryUseCase(repo)

	out, err := uc.Execute(context.Background(), &domain.User{ID: 1, Role: domain.RoleSuperAdmin})
	require.NoError(t, err)
	assert.Equal(t, 0, out.TraineeCount)
	assert.NotNil(t, out.RecentMembers)
	assert.Empty(t, out.RecentMembers)
	repo.AssertNotCalled(t, "ListMemberActivities", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー学習サマリー_集計ウィンドウは今日を含む7日間(t *testing.T) {
	repo, got := learningActivityRepo(nil, nil)
	uc := usecase.NewGetCompanyLearningSummaryUseCase(repo)

	_, err := uc.Execute(context.Background(), companyAdminActor(10))
	require.NoError(t, err)
	wantFrom := time.Now().UTC().AddDate(0, 0, -6).Format("2006-01-02")
	assert.Equal(t, wantFrom, got.fromDate.Format("2006-01-02"))
}

func Test_メンバー学習サマリー_集計エラーはそのまま返す(t *testing.T) {
	repo, _ := learningActivityRepo(nil, context.DeadlineExceeded)
	uc := usecase.NewGetCompanyLearningSummaryUseCase(repo)
	_, err := uc.Execute(context.Background(), companyAdminActor(10))
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
