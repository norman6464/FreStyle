package repository

import "context"

// CompanyMemberCount は会社単位のメンバー集計（運営の横断ビュー用）。
type CompanyMemberCount struct {
	CompanyID uint64
	Total     int
	Active    int
	Trainees  int
}

// WorkspaceMemberCount はワークスペース単位のメンバー集計（CompanyMemberCount のワークスペース版）。
type WorkspaceMemberCount struct {
	WorkspaceID string
	Total       int
	Active      int
	Trainees    int
}

// CompanyMemberCounter は所属単位でメンバー数を集計する単一責務 port（Effective Go 流の -er 命名）。
// UserRepository を肥大化させないため独立 port として切り出す。
type CompanyMemberCounter interface {
	// CountMembersByCompany は会社ごとの在籍メンバー数（総数 / 有効 / trainee）を返す。
	// 論理削除済み（deleted_at IS NOT NULL）や会社未所属（company_id IS NULL）は除外する。
	CountMembersByCompany(ctx context.Context) ([]CompanyMemberCount, error)
	// CountMembersByWorkspace は CountMembersByCompany のワークスペース版。
	// 論理削除済みやワークスペース未所属（workspace_id IS NULL）は除外する。
	CountMembersByWorkspace(ctx context.Context) ([]WorkspaceMemberCount, error)
}
