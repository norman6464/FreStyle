package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// TeachingMaterialUseCase は教材の get / create / update / delete を 1 構造体で扱う。
// 教材は必ず Course に所属するため一覧は Course 単位、create は course_id 必須。
//
// **可否は対象ごとの付与だけが決める。** アプリのロール（company_admin など）は見ない。
// 判定は CheckMaterialPermissionUseCase を通し、規則は domain が持つ。
type TeachingMaterialUseCase struct {
	repo    repository.TeachingMaterialRepository
	courses repository.CourseRepository
	perm    *CheckMaterialPermissionUseCase
}

func NewTeachingMaterialUseCase(
	repo repository.TeachingMaterialRepository,
	courses repository.CourseRepository,
	perm *CheckMaterialPermissionUseCase,
) *TeachingMaterialUseCase {
	return &TeachingMaterialUseCase{repo: repo, courses: courses, perm: perm}
}

// requireChapter は章 1 つの実効権限を引き、求める条件を満たさなければ断る。
//
// **見えない相手には実在を教えない**（domain.ErrNotFound）。見えている相手には
// 理由を返してよい（ErrMaterialForbidden）。前者を撃ち分けると、ID を総当たりする
// だけで隠した教材の実在が分かる。後者は既に実在を知っているので隠す意味が無い。
func (uc *TeachingMaterialUseCase) requireChapter(
	ctx context.Context, in MaterialActor, chapterID uint64, want func(domain.MaterialPermission) bool,
) error {
	workspaceID, affiliated := in.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return domain.ErrNotFound
	}
	perm, err := uc.chapterPermission(ctx, workspaceID, chapterID, in.ActorUserID)
	if err != nil {
		return err
	}
	if !want(*perm) {
		return ErrMaterialForbidden
	}
	return nil
}

// coursePermission はコースの実効権限を返す。見えない相手には domain.ErrNotFound。
//
// 一覧（読む）と作成（書く）の両方から通す。手順が 2 つに分かれていると、片方だけを
// 直したときに同じコースの可否が経路で食い違う — 認可で最も起こりやすい穴。
// 未所属の扱いだけは呼び出し側に残す（作る操作には隠す対象がまだ無いため）。
func (uc *TeachingMaterialUseCase) coursePermission(
	ctx context.Context, workspaceID string, courseID, userID uint64,
) (*domain.MaterialPermission, error) {
	perm, err := uc.perm.Course(ctx, workspaceID, courseID, userID)
	if err != nil {
		return nil, err
	}
	if !perm.CanView {
		return nil, domain.ErrNotFound
	}
	return perm, nil
}

// chapterPermission は章の実効権限を返す。見えない相手には domain.ErrNotFound。
func (uc *TeachingMaterialUseCase) chapterPermission(
	ctx context.Context, workspaceID string, chapterID, userID uint64,
) (*domain.MaterialPermission, error) {
	perm, err := uc.perm.Chapter(ctx, workspaceID, chapterID, userID)
	if err != nil {
		return nil, err
	}
	if !perm.CanView {
		return nil, domain.ErrNotFound
	}
	return perm, nil
}

// MaterialActor は「誰が」を表す共通部分。教材の入力はどれもこの 2 つを要るので、
// 埋め込みで持たせる。
//
// **アプリのロール（company_admin など）は持たない。** 教材の可否は対象ごとの付与だけで
// 決まるので、ロールを受け取る口を残すと「役割で通す」判定が戻ってくる余地になる。
type MaterialActor struct {
	// ActorUserID は呼び出し元。付与は主体（principals）経由でこの人へ届く。
	ActorUserID uint64
	// ActorWorkspace は呼び出し元の所属。未所属なら教材には一切触れない。
	ActorWorkspace domain.WorkspaceRef
}

// ListByCourse は指定コース配下の教材を返す。
//
// 下書きを混ぜるかはコースの実効権限で決める。編集できる人にだけ見せる、という規則は
// domain.ResolveMaterialPermission が持っていて、ここはその答えを使うだけ。
func (uc *TeachingMaterialUseCase) ListByCourse(
	ctx context.Context, courseID uint64, actor MaterialActor,
) ([]domain.TeachingMaterial, error) {
	workspaceID, affiliated := actor.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return nil, domain.ErrNotFound
	}
	perm, err := uc.coursePermission(ctx, workspaceID, courseID, actor.ActorUserID)
	if err != nil {
		return nil, err
	}
	return uc.repo.ListByCourse(ctx, courseID, perm.CanEdit)
}

// Get は ID 指定で 1 件取得する。認可が先で、通らなければ中身を一度も読まない。
func (uc *TeachingMaterialUseCase) Get(
	ctx context.Context, id uint64, actor MaterialActor,
) (*domain.TeachingMaterial, error) {
	// 見えることだけが条件なので chapterPermission を直に通す。
	// requireChapter に「常に true」の述語を渡すと、そこで何かを見ているように読める。
	workspaceID, affiliated := actor.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return nil, domain.ErrNotFound
	}
	if _, err := uc.chapterPermission(ctx, workspaceID, id, actor.ActorUserID); err != nil {
		return nil, err
	}
	return uc.repo.GetByID(ctx, id)
}

// CreateTeachingMaterialInput は POST 入力。CourseID は必須。
type CreateTeachingMaterialInput struct {
	MaterialActor
	CourseID      uint64
	Title         string
	OrderInCourse int
	IsPublished   bool
}

func (uc *TeachingMaterialUseCase) Create(ctx context.Context, in CreateTeachingMaterialInput) (*domain.TeachingMaterial, error) {
	if in.CourseID == 0 {
		return nil, fmt.Errorf("course_id is required")
	}
	// 章を足すのはコースを書き換えること。コースを編集できる人だけが足せる。
	//
	// 未所属はここで断る。作る操作には隠すべき対象がまだ無いので、実在を伏せる
	// 必要も無い（読み取りの経路とは扱いが違う）。
	workspaceID, affiliated := in.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return nil, ErrMaterialForbidden
	}
	perm, err := uc.coursePermission(ctx, workspaceID, in.CourseID, in.ActorUserID)
	if err != nil {
		return nil, err
	}
	if !perm.CanEdit {
		return nil, ErrMaterialForbidden
	}
	course, err := uc.courses.GetByID(ctx, in.CourseID)
	if err != nil {
		return nil, err
	}
	m := &domain.TeachingMaterial{
		CourseID:        in.CourseID,
		CreatedByUserID: in.ActorUserID,
		Title:           in.Title,
		OrderInCourse:   in.OrderInCourse,
		IsPublished:     in.IsPublished,
		// WorkspaceID はコースから継承する（InsertChapter がそのまま書くので、DB に
		// 書かれる値と一致する）。POST 応答に workspaceId が欠けないようにするため、
		// DB 往復を待たずここで埋める。
		WorkspaceID: course.WorkspaceID,
	}
	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

type UpdateTeachingMaterialInput struct {
	MaterialActor
	ID            uint64
	Title         string
	OrderInCourse int
	IsPublished   bool
}

func (uc *TeachingMaterialUseCase) Update(ctx context.Context, in UpdateTeachingMaterialInput) (*domain.TeachingMaterial, error) {
	if err := uc.requireChapter(ctx, in.MaterialActor, in.ID, func(p domain.MaterialPermission) bool {
		return p.CanEdit
	}); err != nil {
		return nil, err
	}
	existing, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
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
	MaterialActor
	ID               uint64
	Doc              string
	ExpectedRevision int
}

// ErrChapterDocInvalid は doc がリッチ本文として不正（型不一致・サイズ超過・NUL 含み等）。
var ErrChapterDocInvalid = errors.New("chapter doc is invalid")

// UpdateDoc は章のリッチ本文を楽観ロックで保存する。その章を編集できる人だけ。
// doc の検証（object かつ type='doc'・1MiB 上限・NUL 拒否）は rich_documents と同じ基準で行う。
func (uc *TeachingMaterialUseCase) UpdateDoc(ctx context.Context, in UpdateChapterDocInput) (*domain.TeachingMaterial, error) {
	if err := uc.requireChapter(ctx, in.MaterialActor, in.ID, func(p domain.MaterialPermission) bool {
		return p.CanEdit
	}); err != nil {
		return nil, err
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

func (uc *TeachingMaterialUseCase) Delete(ctx context.Context, id uint64, actor MaterialActor) error {
	if err := uc.requireChapter(ctx, actor, id, func(p domain.MaterialPermission) bool {
		return p.CanEdit
	}); err != nil {
		return err
	}
	return uc.repo.Delete(ctx, id)
}
