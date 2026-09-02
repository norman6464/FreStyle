package handler

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeMaterialPerm は教材の権限モデルの最小 fake。
//
// **事実だけを返す。** 何ができるかは domain.ResolveMaterialPermission が決めるので、
// 「編集できる」を直に差し込む口は持たない（持つと、規則を通らない状態をテストが作れる）。
type fakeMaterialPerm struct {
	facts *domain.MaterialFacts
	// notFound を立てると、対象そのものが引けない（別テナント・存在しない）。
	notFound bool
	// courses は一覧が返す行。
	courses []repository.CourseWithFacts
	// listErr を立てると一覧が失敗する（500 経路の確認用）。
	listErr error
}

var _ repository.MaterialPermissionRepository = (*fakeMaterialPerm)(nil)

// materialPermOf は「一員で、公開済みを読めて、編集もできる」既定の fake を返す。
// 既存の handler テストの多くは認可そのものではなく配線を見ているので、通る側を既定にする。
func materialPermOf(role domain.GrantRole) *fakeMaterialPerm {
	return &fakeMaterialPerm{facts: &domain.MaterialFacts{
		Member: true, Published: true, Role: &role,
	}}
}

func (f *fakeMaterialPerm) result() (*domain.MaterialFacts, error) {
	if f.notFound {
		return nil, domain.ErrNotFound
	}
	if f.facts == nil {
		return &domain.MaterialFacts{}, nil
	}
	return f.facts, nil
}

func (f *fakeMaterialPerm) CourseFactsForUser(context.Context, string, uint64, uint64) (*domain.MaterialFacts, error) {
	return f.result()
}

func (f *fakeMaterialPerm) ChapterFactsForUser(context.Context, string, uint64, uint64) (*domain.MaterialFacts, error) {
	return f.result()
}

func (f *fakeMaterialPerm) ListCourseFactsForUser(context.Context, string, uint64) ([]repository.CourseWithFacts, error) {
	return f.courses, f.listErr
}

func (f *fakeMaterialPerm) UpsertCourseGrant(context.Context, string, uint64, string, domain.GrantRole) (*domain.CourseGrant, error) {
	return &domain.CourseGrant{}, nil
}

func (f *fakeMaterialPerm) DeleteCourseGrant(context.Context, string, uint64, string) error {
	return nil
}

func (f *fakeMaterialPerm) ListCourseGrants(context.Context, string, uint64) ([]domain.CourseGrant, error) {
	return nil, nil
}

func (f *fakeMaterialPerm) UpsertChapterGrant(context.Context, string, uint64, string, domain.GrantRole) (*domain.ChapterGrant, error) {
	return &domain.ChapterGrant{}, nil
}

func (f *fakeMaterialPerm) DeleteChapterGrant(context.Context, string, uint64, string) error {
	return nil
}

func (f *fakeMaterialPerm) ListChapterGrants(context.Context, string, uint64) ([]domain.ChapterGrant, error) {
	return nil, nil
}
