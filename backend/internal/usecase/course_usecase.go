package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CourseUseCase はコースの get / create / update / delete を 1 構造体で扱う。
//
// **可否は対象ごとの付与だけが決める。** アプリのロール（company_admin など）は見ない。
// Delete は配下教材も cascade 削除。一覧は進捗集計を伴うため
// ListCoursesWithProgressUseCase が担う。
//
//naminglint:allow 複数 CRUD を束ねる集約 usecase のため Execute 単一メソッドではなく Get/Create 等で公開する
type CourseUseCase struct {
	courses    repository.CourseRepository
	materials  repository.TeachingMaterialRepository
	perm       *CheckMaterialPermissionUseCase
	principals repository.KnowledgeBasePermissionRepository
}

func NewCourseUseCase(
	courses repository.CourseRepository,
	materials repository.TeachingMaterialRepository,
	perm *CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *CourseUseCase {
	return &CourseUseCase{courses: courses, materials: materials, perm: perm, principals: principals}
}

// requireCourse はコース 1 つの実効権限を引き、求める条件を満たさなければ断る。
// 見えない相手には実在を教えない（domain.ErrNotFound）。見えている相手には理由を返す。
func (uc *CourseUseCase) requireCourse(
	ctx context.Context, actor MaterialActor, courseID uint64, want func(domain.MaterialPermission) bool,
) error {
	workspaceID, affiliated := actor.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return domain.ErrNotFound
	}
	perm, err := uc.perm.Course(ctx, workspaceID, courseID, actor.ActorUserID)
	if err != nil {
		return err
	}
	if !perm.CanView {
		return domain.ErrNotFound
	}
	if !want(*perm) {
		return ErrMaterialForbidden
	}
	return nil
}

func (uc *CourseUseCase) Get(ctx context.Context, id uint64, actor MaterialActor) (*domain.Course, error) {
	if err := uc.requireCourse(ctx, actor, id, func(p domain.MaterialPermission) bool {
		return p.CanView
	}); err != nil {
		return nil, err
	}
	return uc.courses.GetByID(ctx, id)
}

type CreateCourseInput struct {
	MaterialActor
	Title       string
	Description string
	Category    string
	Language    string
	SortOrder   int
	IsPublished bool
}

// Create はコースを作る。**ワークスペースの一員なら誰でも作れる**（ユーザー指示）。
//
// 作った人はそのコースの admin になる。誰でも作れるのに誰も扱えない、という状態を
// 作らないためで、コースと付与は 1 つのトランザクションで書く。
func (uc *CourseUseCase) Create(ctx context.Context, in CreateCourseInput) (*domain.Course, error) {
	// 作成したコースの所属先が決まらないため、未所属の actor は作成できない。
	workspaceID, workspaceAffiliated := in.ActorWorkspace.WorkspaceID()
	if !workspaceAffiliated {
		return nil, ErrMaterialForbidden
	}
	// 所属は principals（kind='user'）の行が唯一の表現。無ければ一員ではない。
	owner, err := uc.principals.FindUserPrincipal(ctx, workspaceID, in.ActorUserID)
	if err != nil {
		if errors.Is(err, repository.ErrPrincipalNotFound) {
			return nil, ErrMaterialForbidden
		}
		return nil, err
	}
	if !domain.IsValidCourseCategory(in.Category) {
		return nil, fmt.Errorf("invalid course category: %s", in.Category)
	}
	c := &domain.Course{
		WorkspaceID:     &workspaceID,
		CreatedByUserID: in.ActorUserID,
		Title:           in.Title,
		Description:     in.Description,
		Category:        in.Category,
		Language:        in.Language,
		SortOrder:       in.SortOrder,
		IsPublished:     in.IsPublished,
	}
	if err := uc.courses.CreateWithOwnerGrant(ctx, c, owner.ID); err != nil {
		return nil, err
	}
	return c, nil
}

type UpdateCourseInput struct {
	MaterialActor
	ID          uint64
	Title       string
	Description string
	Category    string
	Language    string
	SortOrder   int
	IsPublished bool
}

func (uc *CourseUseCase) Update(ctx context.Context, in UpdateCourseInput) (*domain.Course, error) {
	if err := uc.requireCourse(ctx, in.MaterialActor, in.ID, func(p domain.MaterialPermission) bool {
		return p.CanEdit
	}); err != nil {
		return nil, err
	}
	existing, err := uc.courses.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !domain.IsValidCourseCategory(in.Category) {
		return nil, fmt.Errorf("invalid course category: %s", in.Category)
	}
	existing.Title = in.Title
	existing.Description = in.Description
	existing.Category = in.Category
	existing.Language = in.Language
	existing.SortOrder = in.SortOrder
	existing.IsPublished = in.IsPublished
	if err := uc.courses.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete はコースと配下教材を同時に削除する（cascade 相当）。
func (uc *CourseUseCase) Delete(ctx context.Context, id uint64, actor MaterialActor) error {
	if err := uc.requireCourse(ctx, actor, id, func(p domain.MaterialPermission) bool {
		return p.CanEdit
	}); err != nil {
		return err
	}
	if err := uc.materials.DeleteByCourse(ctx, id); err != nil {
		return err
	}
	return uc.courses.Delete(ctx, id)
}
