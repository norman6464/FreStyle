package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CheckMaterialPermissionUseCase は教材（コース / 章）1 つに対する実効権限を返す。
//
// 教材の可否を問うところは必ずここを通す。role を見る古い判定（company_admin なら何でも
// できる）はもう無く、**誰が何をしてよいかは対象ごとの付与だけが決める**。
//
// 判定規則そのものは domain.ResolveMaterialPermission が持つ。ここがやるのは
// 事実を引いて規則へ渡すことだけで、規則をこちらへ写さない。
//
//naminglint:allow コースと章の 2 つを問うので Execute 単一メソッドではなく Course/Chapter で公開する
type CheckMaterialPermissionUseCase struct {
	repo repository.MaterialPermissionRepository
}

func NewCheckMaterialPermissionUseCase(r repository.MaterialPermissionRepository) *CheckMaterialPermissionUseCase {
	return &CheckMaterialPermissionUseCase{repo: r}
}

// Course はコース 1 つの実効権限を返す。コースが無い・別ワークスペースなら domain.ErrNotFound。
func (u *CheckMaterialPermissionUseCase) Course(
	ctx context.Context, workspaceID string, courseID, userID uint64,
) (*domain.MaterialPermission, error) {
	facts, err := u.repo.CourseFactsForUser(ctx, workspaceID, courseID, userID)
	if err != nil {
		return nil, err
	}
	perm := domain.ResolveMaterialPermission(*facts)
	return &perm, nil
}

// Chapter は章 1 つの実効権限を返す。コースに張られた付与も見る（章へ降りるため）。
func (u *CheckMaterialPermissionUseCase) Chapter(
	ctx context.Context, workspaceID string, chapterID, userID uint64,
) (*domain.MaterialPermission, error) {
	facts, err := u.repo.ChapterFactsForUser(ctx, workspaceID, chapterID, userID)
	if err != nil {
		return nil, err
	}
	perm := domain.ResolveMaterialPermission(*facts)
	return &perm, nil
}

// ErrMaterialForbidden は教材に対して権限が足りないときに返す。
//
// 「無い」と「見えない」を撃ち分けないため、呼び出し側（handler）はこれと
// domain.ErrNotFound を同じ応答へ落とす。撃ち分けると、ID を総当たりするだけで
// 隠した教材の実在が分かる。
var ErrMaterialForbidden = errors.New("material forbidden")
