package usecase

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// learningSummaryDateLayout はサマリー内の日付表現(YYYY-MM-DD)。
const learningSummaryDateLayout = "2006-01-02"

// recentMembersLimit はホームのサイドバーに出す直近アクティブメンバーの最大件数。
const recentMembersLimit = 5

// MemberLearningSummaryItem は直近アクティブメンバー 1 人分の表示用サマリー。
type MemberLearningSummaryItem struct {
	UserID uint64 `json:"userId"`
	Name   string `json:"name"`
	// LastActiveDate は最後に学習活動があった日(YYYY-MM-DD、UTC 基準)。
	LastActiveDate string `json:"lastActiveDate"`
	// RecentActivityCount は直近 7 日間の活動回数合計。
	RecentActivityCount int `json:"recentActivityCount"`
}

// CompanyLearningSummaryOutput は company_admin のホームに出す自社メンバーの学習状況サマリー。
type CompanyLearningSummaryOutput struct {
	// TraineeCount は在籍 trainee 数(論理削除済みを除く)。
	TraineeCount int `json:"traineeCount"`
	// ActiveToday は今日(UTC)学習活動があった trainee 数。
	ActiveToday int `json:"activeToday"`
	// ActiveThisWeek は直近 7 日間(今日を含む)に学習活動があった trainee 数。
	ActiveThisWeek int `json:"activeThisWeek"`
	// RecentMembers は最終活動日の新しい順の直近アクティブメンバー
	// (直近 7 日間に活動があるメンバーのみ、最大 5 名)。
	RecentMembers []MemberLearningSummaryItem `json:"recentMembers"`
}

// GetCompanyLearningSummaryUseCase は company_admin のホーム向けに自社 trainee の学習状況を集計する。
// 日付境界は既存ダッシュボード(GetUserDashboardUseCase)と同じ UTC。
// ロールの認可は handler 層(isAdminActor)で行い、ここでは actor の所属会社だけを見る。
type GetCompanyLearningSummaryUseCase struct {
	activities repository.CompanyLearningActivitySummarizer
}

// NewGetCompanyLearningSummaryUseCase は GetCompanyLearningSummaryUseCase を組み立てる。
func NewGetCompanyLearningSummaryUseCase(activities repository.CompanyLearningActivitySummarizer) *GetCompanyLearningSummaryUseCase {
	return &GetCompanyLearningSummaryUseCase{activities: activities}
}

// Execute は自社 trainee の学習状況サマリーを返す。会社未所属(super_admin 等)は空サマリー。
func (u *GetCompanyLearningSummaryUseCase) Execute(ctx context.Context, actor *domain.User) (*CompanyLearningSummaryOutput, error) {
	out := &CompanyLearningSummaryOutput{RecentMembers: []MemberLearningSummaryItem{}}
	if actor == nil || actor.CompanyID == nil {
		return out, nil
	}

	today := time.Now().UTC()
	weekStart := today.AddDate(0, 0, -6)
	rows, err := u.activities.ListMemberActivities(ctx, *actor.CompanyID, weekStart)
	if err != nil {
		return nil, err
	}

	todayStr := today.Format(learningSummaryDateLayout)
	out.TraineeCount = len(rows)
	for _, r := range rows {
		if r.LastActiveDate != nil && r.LastActiveDate.Format(learningSummaryDateLayout) == todayStr {
			out.ActiveToday++
		}
		if r.RecentActivityCount > 0 {
			out.ActiveThisWeek++
		}
		// repository は最終活動日の新しい順(未活動は末尾)で返す。
		// 「直近アクティブ」の見出しと矛盾しないよう、直近 7 日間に活動があるメンバーだけを載せる
		// (7 日超前の活動しか無い人を 0 回で並べない)。
		if r.LastActiveDate != nil && r.RecentActivityCount > 0 && len(out.RecentMembers) < recentMembersLimit {
			out.RecentMembers = append(out.RecentMembers, MemberLearningSummaryItem{
				UserID:              r.UserID,
				Name:                r.Name,
				LastActiveDate:      r.LastActiveDate.Format(learningSummaryDateLayout),
				RecentActivityCount: r.RecentActivityCount,
			})
		}
	}
	return out, nil
}
