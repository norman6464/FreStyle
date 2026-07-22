package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeCompanyLearningActivityRepoH は CompanyLearningActivitySummarizer の最小 fake。
type fakeCompanyLearningActivityRepoH struct {
	rows []repository.MemberLearningActivity
	err  error
}

func (f *fakeCompanyLearningActivityRepoH) ListMemberActivities(context.Context, uint64, time.Time) ([]repository.MemberLearningActivity, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.rows, nil
}

func newLearningSummaryHandler(repo repository.CompanyLearningActivitySummarizer) *AdminMemberHandler {
	return NewAdminMemberHandler(nil, nil, nil, nil, usecase.NewGetCompanyLearningSummaryUseCase(repo))
}

func companyAdminUser() *domain.User {
	cid := uint64(10)
	return &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: &cid}
}

func Test_学習サマリーハンドラ(t *testing.T) {
	t.Run("未認証 → 403", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, nil)
		newLearningSummaryHandler(&fakeCompanyLearningActivityRepoH{}).LearningSummary(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", w.Code)
		}
	})
	t.Run("trainee → 403", func(t *testing.T) {
		cid := uint64(10)
		trainee := &domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: &cid}
		w, c := ctxJSON(http.MethodGet, "", nil, trainee)
		newLearningSummaryHandler(&fakeCompanyLearningActivityRepoH{}).LearningSummary(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", w.Code)
		}
	})
	t.Run("company_admin → 200 と集計", func(t *testing.T) {
		now := time.Now().UTC()
		repo := &fakeCompanyLearningActivityRepoH{rows: []repository.MemberLearningActivity{
			{UserID: 11, Name: "member-a", LastActiveDate: &now, RecentActivityCount: 2},
		}}
		w, c := ctxJSON(http.MethodGet, "", nil, companyAdminUser())
		newLearningSummaryHandler(repo).LearningSummary(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		var out usecase.CompanyLearningSummaryOutput
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("json parse: %v", err)
		}
		if out.TraineeCount != 1 || out.ActiveToday != 1 || len(out.RecentMembers) != 1 {
			t.Fatalf("unexpected summary: %+v", out)
		}
	})
	t.Run("集計エラー → 500", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, companyAdminUser())
		newLearningSummaryHandler(&fakeCompanyLearningActivityRepoH{err: context.DeadlineExceeded}).LearningSummary(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("want 500, got %d", w.Code)
		}
	})
}
