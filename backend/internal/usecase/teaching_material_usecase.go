package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// TeachingMaterialUseCase は教材の list / get / create / update / delete を 1 構造体で扱う。
// 教材は必ず Course に所属するため list は Course 単位、create は course_id 必須。
// trainee は published コース内の published 教材のみ閲覧、編集系は company_admin / super_admin。
//
//naminglint:allow 複数 CRUD を束ねる集約 usecase のため Execute 単一メソッドではなく List/Get/Create 等で公開する
type TeachingMaterialUseCase struct {
	repo    repository.TeachingMaterialRepository
	courses repository.CourseRepository
}

func NewTeachingMaterialUseCase(repo repository.TeachingMaterialRepository, courses repository.CourseRepository) *TeachingMaterialUseCase {
	return &TeachingMaterialUseCase{repo: repo, courses: courses}
}

// canManage は教材を作成 / 編集 / 削除できる role 判定。
func canManage(role domain.RoleName) bool {
	return role == domain.RoleCompanyAdmin || role == domain.RoleSuperAdmin
}

// List はワークスペース内の全教材を返す backward-compat 用（コース対応への移行後に削除予定）。
func (uc *TeachingMaterialUseCase) List(ctx context.Context, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) ([]domain.TeachingMaterial, error) {
	// workspace 単位の一覧なので、未所属の actor には（super_admin であっても）空を返す。
	workspaceID, affiliated := actorWorkspace.WorkspaceID()
	if !affiliated {
		return []domain.TeachingMaterial{}, nil
	}
	includeUnpublished := canManage(actorRole)
	return uc.repo.ListByWorkspace(ctx, workspaceID, includeUnpublished)
}

// ListByCourse は指定コース配下の教材を返す（role / workspace を検証してから）。
func (uc *TeachingMaterialUseCase) ListByCourse(ctx context.Context, courseID uint64, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) ([]domain.TeachingMaterial, error) {
	course, err := uc.courses.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if !canReadCourse(course, actorWorkspace, actorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	includeUnpublished := canManage(actorRole)
	return uc.repo.ListByCourse(ctx, courseID, includeUnpublished)
}

// Get は ID 指定で 1 件取得する（workspace 不一致 / 非公開コースは閲覧不可）。
func (uc *TeachingMaterialUseCase) Get(ctx context.Context, id uint64, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) (*domain.TeachingMaterial, error) {
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	course, err := uc.courses.GetByID(ctx, m.CourseID)
	if err != nil {
		return nil, err
	}
	if !canRead(m, course, actorWorkspace, actorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	return m, nil
}

// canRead は対象教材を actorWorkspace が閲覧できるかを判定する。
// courses.workspace_id / course_chapters.workspace_id は起動時バックフィルと
// 作成時の書き込みにより、リクエストを捌く時点で必ず埋まっている。
func canRead(m *domain.TeachingMaterial, course *domain.Course, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) bool {
	if !materialBelongsToWorkspace(m, actorWorkspace, actorRole) {
		return false
	}
	// 所属コースが閲覧可能でなければ教材も見せない。
	if !canReadCourse(course, actorWorkspace, actorRole) {
		return false
	}
	if !m.IsPublished && !canManage(actorRole) {
		return false
	}
	return true
}

// materialBelongsToWorkspace は super_admin か、対象教材が actorWorkspace に属するかを返す。
// Get（canRead 経由）/ Update / UpdateDoc / Delete で同じ形の所属チェックが個別に書かれていた
// 重複を、courseBelongsToWorkspace と対になる形でここに集約した（判定結果は変えない）。
func materialBelongsToWorkspace(m *domain.TeachingMaterial, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) bool {
	if actorRole == domain.RoleSuperAdmin {
		return true
	}
	wid, ok := m.WorkspaceRef().WorkspaceID()
	// 未所属の actor・対象教材の workspace_id 未設定はどちらも一致し得ない。
	return ok && actorWorkspace.Matches(wid)
}

// CreateTeachingMaterialInput は POST 入力。CourseID は必須。
type CreateTeachingMaterialInput struct {
	ActorUserID    uint64
	ActorWorkspace domain.WorkspaceRef
	ActorRole      domain.RoleName
	CourseID       uint64
	Title          string
	OrderInCourse  int
	IsPublished    bool
}

func (uc *TeachingMaterialUseCase) Create(ctx context.Context, in CreateTeachingMaterialInput) (*domain.TeachingMaterial, error) {
	if !canManage(in.ActorRole) {
		return nil, fmt.Errorf("forbidden: only company_admin or super_admin can create materials")
	}
	// 教材はコース（= ワークスペース）配下に作るため、未所属の actor は super_admin でも作成できない。
	if _, affiliated := in.ActorWorkspace.WorkspaceID(); !affiliated {
		return nil, fmt.Errorf("actor must belong to a workspace")
	}
	if in.CourseID == 0 {
		return nil, fmt.Errorf("course_id is required")
	}
	course, err := uc.courses.GetByID(ctx, in.CourseID)
	if err != nil {
		return nil, err
	}
	if !courseBelongsToWorkspace(course, in.ActorWorkspace, in.ActorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	m := &domain.TeachingMaterial{
		CourseID:        in.CourseID,
		CreatedByUserID: in.ActorUserID,
		Title:           in.Title,
		OrderInCourse:   in.OrderInCourse,
		IsPublished:     in.IsPublished,
		// WorkspaceID はコースから継承する（InsertChapter が company_id からも同じ値を
		// dual-write するため、DB に書かれる値と一致する）。POST 応答に workspaceId が
		// 欠けないようにするため、DB 往復を待たずここで埋める。
		WorkspaceID: course.WorkspaceID,
	}
	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

type UpdateTeachingMaterialInput struct {
	ID             uint64
	ActorWorkspace domain.WorkspaceRef
	ActorRole      domain.RoleName
	Title          string
	OrderInCourse  int
	IsPublished    bool
}

func (uc *TeachingMaterialUseCase) Update(ctx context.Context, in UpdateTeachingMaterialInput) (*domain.TeachingMaterial, error) {
	existing, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !canManage(in.ActorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	if !materialBelongsToWorkspace(existing, in.ActorWorkspace, in.ActorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	existing.Title = in.Title
	existing.OrderInCourse = in.OrderInCourse
	existing.IsPublished = in.IsPublished
	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// UpdateChapterDocInput は章のリッチ本文（tiptap JSON）更新の入力。
type UpdateChapterDocInput struct {
	ID               uint64
	ActorWorkspace   domain.WorkspaceRef
	ActorRole        domain.RoleName
	Doc              string
	ExpectedRevision int
}

// ErrChapterDocInvalid は doc がリッチ本文として不正（型不一致・サイズ超過・NUL 含み等）。
var ErrChapterDocInvalid = errors.New("chapter doc is invalid")

// UpdateDoc は章のリッチ本文を楽観ロックで保存する。canManage（company_admin / super_admin）のみ。
// doc の検証（object かつ type='doc'・1MiB 上限・NUL 拒否）は rich_documents と同じ基準で行う。
func (uc *TeachingMaterialUseCase) UpdateDoc(ctx context.Context, in UpdateChapterDocInput) (*domain.TeachingMaterial, error) {
	existing, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !canManage(in.ActorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	if !materialBelongsToWorkspace(existing, in.ActorWorkspace, in.ActorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	if in.ExpectedRevision < 1 {
		return nil, fmt.Errorf("%w: expectedRevision must be >= 1", ErrChapterDocInvalid)
	}
	if err := validateDoc(in.Doc); err != nil {
		// rich_documents 用のエラー種別を章用に読み替える（HTTP 400 へ落とすため）。
		return nil, fmt.Errorf("%w: %w", ErrChapterDocInvalid, err)
	}
	return uc.repo.UpdateDocWithRevision(ctx, in.ID, in.Doc, in.ExpectedRevision)
}

func (uc *TeachingMaterialUseCase) Delete(ctx context.Context, id uint64, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !canManage(actorRole) {
		return fmt.Errorf("forbidden")
	}
	if !materialBelongsToWorkspace(existing, actorWorkspace, actorRole) {
		return fmt.Errorf("forbidden")
	}
	return uc.repo.Delete(ctx, id)
}
