package usecase

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CompanyStat は会社横断ビューの 1 行（会社情報 + メンバー集計）。handler がそのまま JSON 返却する。
type CompanyStat struct {
	ID            uint64    `json:"id"`
	Name          string    `json:"name"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	MemberTotal   int       `json:"memberTotal"`
	ActiveMembers int       `json:"activeMembers"`
	TraineeCount  int       `json:"traineeCount"`
}

// ListCompanyStatsUseCase は会社一覧に各社のメンバー集計を付けて返す（super_admin 専用の横断ビュー）。
type ListCompanyStatsUseCase struct {
	companies repository.CompanyRepository
	counter   repository.CompanyMemberCounter
}

func NewListCompanyStatsUseCase(
	c repository.CompanyRepository,
	mc repository.CompanyMemberCounter,
) *ListCompanyStatsUseCase {
	return &ListCompanyStatsUseCase{companies: c, counter: mc}
}

func (u *ListCompanyStatsUseCase) Execute(ctx context.Context) ([]CompanyStat, error) {
	companies, err := u.companies.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	counts, err := u.counter.CountMembersByWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	// 集計はワークスペース単位で返るので、会社に紐づくワークスペース ID で引き当てる
	// （companies 1 : 1 workspaces）。
	byWorkspace := make(map[string]repository.WorkspaceMemberCount, len(counts))
	for _, cnt := range counts {
		byWorkspace[cnt.WorkspaceID] = cnt
	}
	stats := make([]CompanyStat, 0, len(companies))
	for _, co := range companies {
		// メンバーがいない会社・ワークスペース未紐付けの会社は zero value（0 件）で埋まる。
		var cnt repository.WorkspaceMemberCount
		if co.WorkspaceID != nil {
			cnt = byWorkspace[*co.WorkspaceID]
		}
		stats = append(stats, CompanyStat{
			ID:            co.ID,
			Name:          co.Name,
			IsActive:      co.IsActive,
			CreatedAt:     co.CreatedAt,
			MemberTotal:   cnt.Total,
			ActiveMembers: cnt.Active,
			TraineeCount:  cnt.Trainees,
		})
	}
	return stats, nil
}
