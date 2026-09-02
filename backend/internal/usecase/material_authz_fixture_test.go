package usecase_test

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/mock"
)

// materialFactsConfig は「その人にとって対象がどう見えるか」をテストが指定するための設定。
//
// **できること（編集してよいか）は直接指定できない。** 事実だけを置き、そこから何が
// できるかは domain.ResolveMaterialPermission が決める。ここで「編集できる」を直に
// 差し込めるようにすると、規則を通らない状態をテストが作れてしまう。
type materialFactsConfig struct {
	// member はワークスペースの一員か。既定 false（＝ 何も見えない）。
	member bool
	// workspaceAdmin はワークスペース全体の admin か。
	workspaceAdmin bool
	// role はコース / 章から届いている役割。nil なら付与なし。
	role *domain.GrantRole
	// published は対象が公開済みか。
	published bool
	// notFound を立てると、対象そのものが引けない（別テナント・存在しない）。
	notFound bool
}

func (c materialFactsConfig) facts() *domain.MaterialFacts {
	return &domain.MaterialFacts{
		Member:         c.member,
		WorkspaceAdmin: c.workspaceAdmin,
		Role:           c.role,
		Published:      c.published,
	}
}

// materialPerm は教材の権限判定を組み立てる。コースと章に同じ事実を返す
// （階層をまたぐ違いを見たいテストは perm.On(...) で個別に上書きする）。
func materialPerm(cfg materialFactsConfig) (*mockMaterialPermRepo, *usecase.CheckMaterialPermissionUseCase) {
	repo := &mockMaterialPermRepo{}
	var facts *domain.MaterialFacts
	var err error
	if cfg.notFound {
		err = domain.ErrNotFound
	} else {
		facts = cfg.facts()
	}
	repo.On("CourseFactsForUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(facts, err).Maybe()
	repo.On("ChapterFactsForUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(facts, err).Maybe()
	repo.On("ListCourseFactsForUser", mock.Anything, mock.Anything, mock.Anything).
		Return([]repository.CourseWithFacts{}, nil).Maybe()
	return repo, usecase.NewCheckMaterialPermissionUseCase(repo)
}

// principalsFor は「その人がワークスペースの一員か」だけを答える主体のモック。
// コース作成が所属を確かめるのに使う。
func principalsFor(member bool) *mockKBPermissionRepo {
	repo := &mockKBPermissionRepo{}
	if member {
		repo.On("FindUserPrincipal", mock.Anything, mock.Anything, mock.Anything).
			Return(&domain.Principal{ID: "0198a000-0000-7000-8000-0000000000a1", Kind: domain.PrincipalKindUser}, nil).Maybe()
	} else {
		repo.On("FindUserPrincipal", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, repository.ErrPrincipalNotFound).Maybe()
	}
	return repo
}

// grantRoleOf は付与の役割をポインタで返す小道具。
func grantRoleOf(r domain.GrantRole) *domain.GrantRole { return &r }

// actorIn は所属つきの呼び出し元。
func actorIn(workspaceID string) usecase.MaterialActor {
	return usecase.MaterialActor{ActorUserID: 1, ActorWorkspace: domain.WorkspaceRefOf(workspaceID)}
}
