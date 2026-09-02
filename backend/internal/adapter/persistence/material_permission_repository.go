package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// materialPermissionRepository は [repository.MaterialPermissionRepository] の実装。
type materialPermissionRepository struct {
	q *sqlcgen.Queries
}

func NewMaterialPermissionRepository(db *sql.DB) repository.MaterialPermissionRepository {
	return &materialPermissionRepository{q: sqlcgen.New(db)}
}

func toDomainCourseGrant(row sqlcgen.CourseGrant) domain.CourseGrant {
	return domain.CourseGrant{
		WorkspaceID: row.WorkspaceID.String(),
		CourseID:    uint64(row.CourseID),
		PrincipalID: row.PrincipalID.String(),
		Role:        domain.GrantRole(row.Role),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toDomainChapterGrant(row sqlcgen.ChapterGrant) domain.ChapterGrant {
	return domain.ChapterGrant{
		WorkspaceID: row.WorkspaceID.String(),
		ChapterID:   uint64(row.ChapterID),
		PrincipalID: row.PrincipalID.String(),
		Role:        domain.GrantRole(row.Role),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func (r *materialPermissionRepository) CourseFactsForUser(
	ctx context.Context, workspaceID string, courseID, userID uint64,
) (*domain.MaterialFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	cID, ok2 := toInt64ID(courseID)
	uID, ok3 := toInt64ID(userID)
	if !ok || !ok2 || !ok3 {
		// bigint に収まらない ID はどの行にも一致しない。無いのと同じ扱いにする。
		return nil, domain.ErrNotFound
	}
	row, err := r.q.ResolveCoursePermissionFacts(ctx, sqlcgen.ResolveCoursePermissionFactsParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uID, Valid: true},
		CourseID:    cID,
	})
	if err != nil {
		return nil, err
	}
	if !row.TargetExists {
		return nil, domain.ErrNotFound
	}
	return &domain.MaterialFacts{
		Member:         row.IsMember,
		WorkspaceAdmin: row.IsWorkspaceAdmin,
		Role:           domain.GrantRoleByRank(int(row.GrantRank)),
		Published:      row.IsPublished,
	}, nil
}

func (r *materialPermissionRepository) ChapterFactsForUser(
	ctx context.Context, workspaceID string, chapterID, userID uint64,
) (*domain.MaterialFacts, error) {
	wsID, ok := kbParseID(workspaceID)
	chID, ok2 := toInt64ID(chapterID)
	uID, ok3 := toInt64ID(userID)
	if !ok || !ok2 || !ok3 {
		return nil, domain.ErrNotFound
	}
	row, err := r.q.ResolveChapterPermissionFacts(ctx, sqlcgen.ResolveChapterPermissionFactsParams{
		WorkspaceID: wsID,
		UserID:      sql.NullInt64{Int64: uID, Valid: true},
		ChapterID:   chID,
	})
	if err != nil {
		return nil, err
	}
	if !row.TargetExists {
		return nil, domain.ErrNotFound
	}
	return &domain.MaterialFacts{
		Member:         row.IsMember,
		WorkspaceAdmin: row.IsWorkspaceAdmin,
		Role:           domain.GrantRoleByRank(int(row.GrantRank)),
		Published:      row.IsPublished,
	}, nil
}

func (r *materialPermissionRepository) UpsertCourseGrant(
	ctx context.Context, workspaceID string, courseID uint64, principalID string, role domain.GrantRole,
) (*domain.CourseGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	cID, ok3 := toInt64ID(courseID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrPrincipalNotFound
	}
	row, err := r.q.UpsertCourseGrant(ctx, sqlcgen.UpsertCourseGrantParams{
		WorkspaceID: wsID, CourseID: cID, PrincipalID: prID, Role: string(role),
	})
	if err != nil {
		return nil, err
	}
	g := toDomainCourseGrant(row)
	return &g, nil
}

// DeleteCourseGrant はコース権限を 1 件取り消す。
// ノート側と同じ理由で 0 行削除は成功のまま（取り消しは冪等）。
func (r *materialPermissionRepository) DeleteCourseGrant(
	ctx context.Context, workspaceID string, courseID uint64, principalID string,
) error {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	cID, ok3 := toInt64ID(courseID)
	if !ok || !ok2 || !ok3 {
		return nil
	}
	_, err := r.q.DeleteCourseGrant(ctx, sqlcgen.DeleteCourseGrantParams{
		WorkspaceID: wsID, CourseID: cID, PrincipalID: prID,
	})
	return err
}

func (r *materialPermissionRepository) ListCourseGrants(
	ctx context.Context, workspaceID string, courseID uint64,
) ([]domain.CourseGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	cID, ok2 := toInt64ID(courseID)
	if !ok || !ok2 {
		return []domain.CourseGrant{}, nil
	}
	rows, err := r.q.ListCourseGrants(ctx, sqlcgen.ListCourseGrantsParams{WorkspaceID: wsID, CourseID: cID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.CourseGrant, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainCourseGrant(row))
	}
	return out, nil
}

func (r *materialPermissionRepository) UpsertChapterGrant(
	ctx context.Context, workspaceID string, chapterID uint64, principalID string, role domain.GrantRole,
) (*domain.ChapterGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	chID, ok3 := toInt64ID(chapterID)
	if !ok || !ok2 || !ok3 {
		return nil, repository.ErrPrincipalNotFound
	}
	row, err := r.q.UpsertChapterGrant(ctx, sqlcgen.UpsertChapterGrantParams{
		WorkspaceID: wsID, ChapterID: chID, PrincipalID: prID, Role: string(role),
	})
	if err != nil {
		return nil, err
	}
	g := toDomainChapterGrant(row)
	return &g, nil
}

func (r *materialPermissionRepository) DeleteChapterGrant(
	ctx context.Context, workspaceID string, chapterID uint64, principalID string,
) error {
	wsID, ok := kbParseID(workspaceID)
	prID, ok2 := kbParseID(principalID)
	chID, ok3 := toInt64ID(chapterID)
	if !ok || !ok2 || !ok3 {
		return nil
	}
	_, err := r.q.DeleteChapterGrant(ctx, sqlcgen.DeleteChapterGrantParams{
		WorkspaceID: wsID, ChapterID: chID, PrincipalID: prID,
	})
	return err
}

func (r *materialPermissionRepository) ListChapterGrants(
	ctx context.Context, workspaceID string, chapterID uint64,
) ([]domain.ChapterGrant, error) {
	wsID, ok := kbParseID(workspaceID)
	chID, ok2 := toInt64ID(chapterID)
	if !ok || !ok2 {
		return []domain.ChapterGrant{}, nil
	}
	rows, err := r.q.ListChapterGrants(ctx, sqlcgen.ListChapterGrantsParams{WorkspaceID: wsID, ChapterID: chID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ChapterGrant, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainChapterGrant(row))
	}
	return out, nil
}
