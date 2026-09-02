package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// materialGrantDeps は教材の権限を書き換える usecase が共通で使う依存。
//
// 3 つとも要るのは、1 回の操作で 3 つのことを確かめるため:
// 呼び出し元がその対象を管理してよいか（check）、相手が同じワークスペースの主体か
// （principals）、そして実際の書き換え（perm）。
type materialGrantDeps struct {
	perm       repository.MaterialPermissionRepository
	check      *CheckMaterialPermissionUseCase
	principals repository.KnowledgeBasePermissionRepository
}

// requireCourseManage はコースの権限を変えてよいかを確かめる。
//
// **見えない相手には実在を教えない**（domain.ErrNotFound）。見えているが管理できない
// 相手には理由を返す（ErrMaterialForbidden）。教材の読み書きと同じ分け方。
func (d materialGrantDeps) requireCourseManage(
	ctx context.Context, actor MaterialActor, courseID uint64,
) (string, error) {
	workspaceID, affiliated := actor.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return "", domain.ErrNotFound
	}
	perm, err := d.check.Course(ctx, workspaceID, courseID, actor.ActorUserID)
	if err != nil {
		return "", err
	}
	if !perm.CanView {
		return "", domain.ErrNotFound
	}
	if !perm.CanManage {
		return "", ErrMaterialForbidden
	}
	return workspaceID, nil
}

// requireChapterManage は章の権限を変えてよいかを確かめる。
// コースに admin が届いていれば章も管理できる（付与は配下へ降りる）。
func (d materialGrantDeps) requireChapterManage(
	ctx context.Context, actor MaterialActor, chapterID uint64,
) (string, error) {
	workspaceID, affiliated := actor.ActorWorkspace.WorkspaceID()
	if !affiliated {
		return "", domain.ErrNotFound
	}
	perm, err := d.check.Chapter(ctx, workspaceID, chapterID, actor.ActorUserID)
	if err != nil {
		return "", err
	}
	if !perm.CanView {
		return "", domain.ErrNotFound
	}
	if !perm.CanManage {
		return "", ErrMaterialForbidden
	}
	return workspaceID, nil
}

// requirePrincipal は相手が同じワークスペースの主体かを確かめる。
//
// DB の複合 FK でも守られるが、先に引いて「別ワークスペースの ID を渡した」を
// FK 違反（500）ではなく not found として返す（ノート側と同じ扱い）。
func (d materialGrantDeps) requirePrincipal(ctx context.Context, workspaceID, principalID string) error {
	_, err := d.principals.FindPrincipal(ctx, workspaceID, principalID)
	return err
}

// GrantCourseRoleUseCase はコースでの既定の役割を主体に与える。
//
// 配下の章にも効く。合成は「最も強いものを採る」なので、**ここに弱い役割を張っても
// 上位で得ている役割は下がらない**（弱める手段はこの層には無い）。
type GrantCourseRoleUseCase struct {
	deps materialGrantDeps
}

func NewGrantCourseRoleUseCase(
	perm repository.MaterialPermissionRepository,
	check *CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *GrantCourseRoleUseCase {
	return &GrantCourseRoleUseCase{deps: materialGrantDeps{perm: perm, check: check, principals: principals}}
}

type GrantCourseRoleInput struct {
	MaterialActor
	CourseID    uint64
	PrincipalID string
	Role        domain.GrantRole
}

func (u *GrantCourseRoleUseCase) Execute(ctx context.Context, in GrantCourseRoleInput) (*domain.CourseGrant, error) {
	// 認可が先。落ちた要求は相手の ID にも役割にも触れない（相手の実在が漏れない）。
	workspaceID, err := u.deps.requireCourseManage(ctx, in.MaterialActor, in.CourseID)
	if err != nil {
		return nil, err
	}
	if !in.Role.Valid() {
		return nil, ErrInvalidGrantRole
	}
	if err := u.deps.requirePrincipal(ctx, workspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.deps.perm.UpsertCourseGrant(ctx, workspaceID, in.CourseID, in.PrincipalID, in.Role)
}

// RevokeCourseRoleUseCase はコースでの既定の役割を剥がす（冪等）。
//
// 「最後の管理者」の検査は置かない。ワークスペースの admin は配下すべてに届くので、
// コースの付与を全部消しても管理できる人が居なくなるわけではない。
type RevokeCourseRoleUseCase struct {
	deps materialGrantDeps
}

func NewRevokeCourseRoleUseCase(
	perm repository.MaterialPermissionRepository,
	check *CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *RevokeCourseRoleUseCase {
	return &RevokeCourseRoleUseCase{deps: materialGrantDeps{perm: perm, check: check, principals: principals}}
}

type RevokeCourseRoleInput struct {
	MaterialActor
	CourseID    uint64
	PrincipalID string
}

func (u *RevokeCourseRoleUseCase) Execute(ctx context.Context, in RevokeCourseRoleInput) error {
	workspaceID, err := u.deps.requireCourseManage(ctx, in.MaterialActor, in.CourseID)
	if err != nil {
		return err
	}
	return u.deps.perm.DeleteCourseGrant(ctx, workspaceID, in.CourseID, in.PrincipalID)
}

// ListCourseGrantsUseCase はそのコース自身に張られた付与を返す。
//
// **「このコースを編集できる人の一覧」ではない。** ワークスペースの admin は含まれず、
// 空でも「誰も編集できない」の意味にならない。画面はそれが分かる見せ方をすること。
type ListCourseGrantsUseCase struct {
	deps materialGrantDeps
}

func NewListCourseGrantsUseCase(
	perm repository.MaterialPermissionRepository,
	check *CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *ListCourseGrantsUseCase {
	return &ListCourseGrantsUseCase{deps: materialGrantDeps{perm: perm, check: check, principals: principals}}
}

type ListCourseGrantsInput struct {
	MaterialActor
	CourseID uint64
}

func (u *ListCourseGrantsUseCase) Execute(ctx context.Context, in ListCourseGrantsInput) ([]domain.CourseGrant, error) {
	workspaceID, err := u.deps.requireCourseManage(ctx, in.MaterialActor, in.CourseID)
	if err != nil {
		return nil, err
	}
	return u.deps.perm.ListCourseGrants(ctx, workspaceID, in.CourseID)
}

// GrantChapterRoleUseCase は章 1 つでの既定の役割を主体に与える（「この教材だけ」）。
type GrantChapterRoleUseCase struct {
	deps materialGrantDeps
}

func NewGrantChapterRoleUseCase(
	perm repository.MaterialPermissionRepository,
	check *CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *GrantChapterRoleUseCase {
	return &GrantChapterRoleUseCase{deps: materialGrantDeps{perm: perm, check: check, principals: principals}}
}

type GrantChapterRoleInput struct {
	MaterialActor
	ChapterID   uint64
	PrincipalID string
	Role        domain.GrantRole
}

func (u *GrantChapterRoleUseCase) Execute(ctx context.Context, in GrantChapterRoleInput) (*domain.ChapterGrant, error) {
	workspaceID, err := u.deps.requireChapterManage(ctx, in.MaterialActor, in.ChapterID)
	if err != nil {
		return nil, err
	}
	if !in.Role.Valid() {
		return nil, ErrInvalidGrantRole
	}
	if err := u.deps.requirePrincipal(ctx, workspaceID, in.PrincipalID); err != nil {
		return nil, err
	}
	return u.deps.perm.UpsertChapterGrant(ctx, workspaceID, in.ChapterID, in.PrincipalID, in.Role)
}

// RevokeChapterRoleUseCase は章での既定の役割を剥がす（冪等）。
// 消えるのはこの段で足した分だけで、コースから降りている役割はそのまま残る。
type RevokeChapterRoleUseCase struct {
	deps materialGrantDeps
}

func NewRevokeChapterRoleUseCase(
	perm repository.MaterialPermissionRepository,
	check *CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *RevokeChapterRoleUseCase {
	return &RevokeChapterRoleUseCase{deps: materialGrantDeps{perm: perm, check: check, principals: principals}}
}

type RevokeChapterRoleInput struct {
	MaterialActor
	ChapterID   uint64
	PrincipalID string
}

func (u *RevokeChapterRoleUseCase) Execute(ctx context.Context, in RevokeChapterRoleInput) error {
	workspaceID, err := u.deps.requireChapterManage(ctx, in.MaterialActor, in.ChapterID)
	if err != nil {
		return err
	}
	return u.deps.perm.DeleteChapterGrant(ctx, workspaceID, in.ChapterID, in.PrincipalID)
}

// ListChapterGrantsUseCase はその章自身に張られた付与を返す（コースから降りる分は含まない）。
type ListChapterGrantsUseCase struct {
	deps materialGrantDeps
}

func NewListChapterGrantsUseCase(
	perm repository.MaterialPermissionRepository,
	check *CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *ListChapterGrantsUseCase {
	return &ListChapterGrantsUseCase{deps: materialGrantDeps{perm: perm, check: check, principals: principals}}
}

type ListChapterGrantsInput struct {
	MaterialActor
	ChapterID uint64
}

func (u *ListChapterGrantsUseCase) Execute(ctx context.Context, in ListChapterGrantsInput) ([]domain.ChapterGrant, error) {
	workspaceID, err := u.deps.requireChapterManage(ctx, in.MaterialActor, in.ChapterID)
	if err != nil {
		return nil, err
	}
	return u.deps.perm.ListChapterGrants(ctx, workspaceID, in.ChapterID)
}

// ListGrantableMaterialPrincipalsUseCase は教材に権限を張れる相手を表示名つきで返す。
//
// 中身はワークスペース全体だが、**呼べるかはコース単位で決める**。ワークスペースの
// admin に絞ると、コースに admin を張られた人が相手を選べなくなる（権限はあるのに
// 画面が使えない）。ノート側の同じ口と同じ考え方。
type ListGrantableMaterialPrincipalsUseCase struct {
	deps materialGrantDeps
}

func NewListGrantableMaterialPrincipalsUseCase(
	perm repository.MaterialPermissionRepository,
	check *CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *ListGrantableMaterialPrincipalsUseCase {
	return &ListGrantableMaterialPrincipalsUseCase{
		deps: materialGrantDeps{perm: perm, check: check, principals: principals},
	}
}

type ListGrantableMaterialPrincipalsInput struct {
	MaterialActor
	CourseID uint64
}

func (u *ListGrantableMaterialPrincipalsUseCase) Execute(
	ctx context.Context, in ListGrantableMaterialPrincipalsInput,
) ([]domain.GrantablePrincipal, error) {
	workspaceID, err := u.deps.requireCourseManage(ctx, in.MaterialActor, in.CourseID)
	if err != nil {
		return nil, err
	}
	return u.deps.principals.ListGrantablePrincipals(ctx, workspaceID)
}
