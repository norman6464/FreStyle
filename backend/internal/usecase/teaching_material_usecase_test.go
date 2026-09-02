package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// materialStore は教材 mock が記録する呼び出し内容(course クラスタのテストで共有)。
type materialStore struct {
	listCourseID                uint64
	listIncludeAll              bool
	listWorkspaceID             string
	listWorkspaceIncludeAll     bool
	created                     *domain.TeachingMaterial
	updated                     *domain.TeachingMaterial
	deleted                     uint64
	deletedByCourse             uint64
	lastCountIncludeUnpublished *bool
}

// materialFakeConfig は教材 mock の応答設定。ゼロ値はすべて「空を返す」。
type materialFakeConfig struct {
	get      *domain.TeachingMaterial
	getErr   error
	counts   map[uint64]int
	countErr error
	listErr  error
}

// materialRepo は TeachingMaterialRepository の mock に、このクラスタが使う応答を
// 設定して返す(呼ばれないメソッドの期待は .Maybe で任意にする)。
func materialRepo(cfg materialFakeConfig) (*mockMaterialRepo, *materialStore) {
	st := &materialStore{}
	repo := &mockMaterialRepo{}
	repo.On("GetByID", mock.Anything, mock.Anything).Return(cfg.get, cfg.getErr).Maybe()
	repo.On("ListByCourse", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			st.listCourseID = args.Get(1).(uint64)
			st.listIncludeAll = args.Get(2).(bool)
		}).Return(nil, nil).Maybe()
	repo.On("ListByWorkspace", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			st.listWorkspaceID = args.Get(1).(string)
			st.listWorkspaceIncludeAll = args.Get(2).(bool)
		}).Return(nil, cfg.listErr).Maybe()
	repo.On("CountByCourseForWorkspace", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			include := args.Get(2).(bool)
			st.lastCountIncludeUnpublished = &include
		}).Return(cfg.counts, cfg.countErr).Maybe()
	repo.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			m := args.Get(1).(*domain.TeachingMaterial)
			m.ID = 99
			st.created = m
		}).Return(nil).Maybe()
	repo.On("Update", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			st.updated = args.Get(1).(*domain.TeachingMaterial)
		}).Return(nil).Maybe()
	repo.On("Delete", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			st.deleted = args.Get(1).(uint64)
		}).Return(nil).Maybe()
	repo.On("DeleteByCourse", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			st.deletedByCourse = args.Get(1).(uint64)
		}).Return(nil).Maybe()
	return repo, st
}

// courseStore はコース mock が記録する更新内容(course クラスタのテストで共有)。
type courseStore struct {
	created *domain.Course
	updated *domain.Course
	deleted uint64
	// ownerPrincipalID は CreateWithOwnerGrant が受け取った作成者の主体。
	ownerPrincipalID string
}

// courseFakeConfig はコース mock の応答設定。ゼロ値はすべて「空を返す」。
type courseFakeConfig struct {
	rows   []domain.Course
	get    *domain.Course
	getErr error
}

// courseRepo は CourseRepository の mock に、このクラスタが使う応答を設定して返す。
func courseRepo(cfg courseFakeConfig) (*mockCourseRepo, *courseStore) {
	st := &courseStore{}
	repo := &mockCourseRepo{}
	repo.On("ListByWorkspaceID", mock.Anything, mock.Anything, mock.Anything).Return(cfg.rows, nil).Maybe()
	repo.On("GetByID", mock.Anything, mock.Anything).Return(cfg.get, cfg.getErr).Maybe()
	repo.On("Create", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			c := args.Get(1).(*domain.Course)
			c.ID = 88
			st.created = c
		}).Return(nil).Maybe()
	repo.On("CreateWithOwnerGrant", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			c := args.Get(1).(*domain.Course)
			c.ID = 88
			st.created = c
			st.ownerPrincipalID = args.Get(2).(string)
		}).Return(nil).Maybe()
	repo.On("Update", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			st.updated = args.Get(1).(*domain.Course)
		}).Return(nil).Maybe()
	repo.On("Delete", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			st.deleted = args.Get(1).(uint64)
		}).Return(nil).Maybe()
	return repo, st
}

// newMaterialUC は教材の usecase を、指定した「見え方」で組み立てる。
func newMaterialUC(cfg materialFactsConfig, mcfg materialFakeConfig) (*usecase.TeachingMaterialUseCase, *materialStore, *mockMaterialRepo) {
	mrepo, mstore := materialRepo(mcfg)
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, WorkspaceID: strPtr(wsA)}})
	_, perm := materialPerm(cfg)
	return usecase.NewTeachingMaterialUseCase(mrepo, crepo, perm), mstore, mrepo
}

// existingChapter は下ごしらえの 1 件。
func existingChapter() materialFakeConfig {
	return materialFakeConfig{get: &domain.TeachingMaterial{ID: 1, CourseID: 5, WorkspaceID: strPtr(wsA)}}
}

func Test_教材_コース別一覧_見えないコースは実在を教えない(t *testing.T) {
	for _, c := range []struct {
		name string
		cfg  materialFactsConfig
	}{
		{"別テナント", materialFactsConfig{notFound: true}},
		{"付与の無い下書き", materialFactsConfig{member: true, published: false}},
		{"所属していない", materialFactsConfig{member: false, published: true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			uc, _, _ := newMaterialUC(c.cfg, materialFakeConfig{})
			_, err := uc.ListByCourse(context.Background(), 5, actorIn(wsA))
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func Test_教材_コース別一覧_下書きを混ぜるかは編集できるかで決まる(t *testing.T) {
	t.Run("付与が無ければ公開のみ", func(t *testing.T) {
		uc, mstore, _ := newMaterialUC(materialFactsConfig{member: true, published: true}, materialFakeConfig{})
		_, err := uc.ListByCourse(context.Background(), 5, actorIn(wsA))
		require.NoError(t, err)
		assert.False(t, mstore.listIncludeAll)
	})

	t.Run("editor なら下書きも含む", func(t *testing.T) {
		uc, mstore, _ := newMaterialUC(materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
		}, materialFakeConfig{})
		_, err := uc.ListByCourse(context.Background(), 5, actorIn(wsA))
		require.NoError(t, err)
		assert.True(t, mstore.listIncludeAll)
	})
}

func Test_教材_取得_見えなければ実在を教えない(t *testing.T) {
	uc, _, mrepo := newMaterialUC(materialFactsConfig{member: true, published: false}, existingChapter())
	_, err := uc.Get(context.Background(), 1, actorIn(wsA))
	assert.ErrorIs(t, err, domain.ErrNotFound)
	// 認可が先。断った要求は中身を一度も読まない。
	mrepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

func Test_教材_取得_公開済みは一員なら誰でも読める(t *testing.T) {
	uc, _, _ := newMaterialUC(materialFactsConfig{member: true, published: true}, existingChapter())
	got, err := uc.Get(context.Background(), 1, actorIn(wsA))
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func Test_教材_作成_コースを編集できる人だけが足せる(t *testing.T) {
	t.Run("コースID欠落は 400 相当", func(t *testing.T) {
		uc, _, _ := newMaterialUC(materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
		}, materialFakeConfig{})
		_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
			MaterialActor: actorIn(wsA), Title: "章",
		})
		require.Error(t, err)
		assert.NotErrorIs(t, err, usecase.ErrMaterialForbidden)
	})

	t.Run("読めるだけでは足せない", func(t *testing.T) {
		uc, mstore, _ := newMaterialUC(materialFactsConfig{member: true, published: true}, materialFakeConfig{})
		_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
			MaterialActor: actorIn(wsA), CourseID: 5, Title: "章",
		})
		assert.ErrorIs(t, err, usecase.ErrMaterialForbidden)
		assert.Nil(t, mstore.created)
	})

	t.Run("見えないコースへは実在を教えない", func(t *testing.T) {
		uc, _, _ := newMaterialUC(materialFactsConfig{notFound: true}, materialFakeConfig{})
		_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
			MaterialActor: actorIn(wsA), CourseID: 5, Title: "章",
		})
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("editor なら足せて、所属はコースから継ぐ", func(t *testing.T) {
		uc, mstore, _ := newMaterialUC(materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
		}, materialFakeConfig{})
		got, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
			MaterialActor: actorIn(wsA), CourseID: 5, Title: "章", OrderInCourse: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, mstore.created)
		assert.Equal(t, "章", mstore.created.Title)
		assert.Equal(t, wsA, *got.WorkspaceID)
	})
}

func Test_教材_更新は編集できる人だけ(t *testing.T) {
	update := func(uc *usecase.TeachingMaterialUseCase) error {
		_, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
			MaterialActor: actorIn(wsA), ID: 1, Title: "改題",
		})
		return err
	}

	t.Run("読めるが付与が無ければ 403", func(t *testing.T) {
		uc, mstore, _ := newMaterialUC(materialFactsConfig{member: true, published: true}, existingChapter())
		assert.ErrorIs(t, update(uc), usecase.ErrMaterialForbidden)
		assert.Nil(t, mstore.updated)
	})

	t.Run("editor なら書き換えられる", func(t *testing.T) {
		uc, mstore, _ := newMaterialUC(materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
		}, existingChapter())
		require.NoError(t, update(uc))
		require.NotNil(t, mstore.updated)
		assert.Equal(t, "改題", mstore.updated.Title)
	})

	t.Run("ワークスペースの admin も書き換えられる", func(t *testing.T) {
		uc, _, _ := newMaterialUC(materialFactsConfig{member: true, workspaceAdmin: true}, existingChapter())
		assert.NoError(t, update(uc))
	})
}

func Test_教材_削除は編集できる人だけ(t *testing.T) {
	t.Run("付与が無ければ消せない", func(t *testing.T) {
		uc, mstore, _ := newMaterialUC(materialFactsConfig{member: true, published: true}, existingChapter())
		assert.ErrorIs(t, uc.Delete(context.Background(), 1, actorIn(wsA)), usecase.ErrMaterialForbidden)
		assert.Zero(t, mstore.deleted)
	})

	t.Run("editor なら消せる", func(t *testing.T) {
		uc, mstore, _ := newMaterialUC(materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
		}, existingChapter())
		require.NoError(t, uc.Delete(context.Background(), 1, actorIn(wsA)))
		assert.Equal(t, uint64(1), mstore.deleted)
	})
}

// --- UpdateDoc（リッチ本文の楽観ロック保存） ---

func docUpdateRepo(existing *domain.TeachingMaterial, updated *domain.TeachingMaterial, updateErr error) *mockMaterialRepo {
	repo := &mockMaterialRepo{}
	repo.On("GetByID", mock.Anything, mock.Anything).Return(existing, nil).Maybe()
	repo.On("UpdateDocWithRevision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(updated, updateErr).Maybe()
	return repo
}

func newDocUC(cfg materialFactsConfig, repo *mockMaterialRepo) *usecase.TeachingMaterialUseCase {
	crepo, _ := courseRepo(courseFakeConfig{})
	_, perm := materialPerm(cfg)
	return usecase.NewTeachingMaterialUseCase(repo, crepo, perm)
}

func TestTeachingMaterialUseCase_UpdateDoc(t *testing.T) {
	validDoc := `{"type":"doc","content":[{"type":"paragraph"}]}`
	existing := &domain.TeachingMaterial{ID: 1, WorkspaceID: strPtr(wsA), Revision: 3}
	editable := materialFactsConfig{member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor)}

	t.Run("編集できる人は保存でき revision 付きで返る", func(t *testing.T) {
		updatedDoc := validDoc
		updated := &domain.TeachingMaterial{ID: 1, WorkspaceID: strPtr(wsA), Revision: 4, Doc: &updatedDoc}
		repo := docUpdateRepo(existing, updated, nil)
		got, err := newDocUC(editable, repo).UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			MaterialActor: actorIn(wsA), ID: 1, Doc: validDoc, ExpectedRevision: 3,
		})
		require.NoError(t, err)
		assert.Equal(t, 4, got.Revision)
		repo.AssertCalled(t, "UpdateDocWithRevision", mock.Anything, uint64(1), validDoc, 3)
	})

	t.Run("読めるだけでは保存できない", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, nil)
		_, err := newDocUC(materialFactsConfig{member: true, published: true}, repo).
			UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
				MaterialActor: actorIn(wsA), ID: 1, Doc: validDoc, ExpectedRevision: 3,
			})
		require.ErrorIs(t, err, usecase.ErrMaterialForbidden)
		repo.AssertNotCalled(t, "UpdateDocWithRevision", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("見えない章は実在を教えない", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, nil)
		_, err := newDocUC(materialFactsConfig{notFound: true}, repo).
			UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
				MaterialActor: actorIn(wsA), ID: 1, Doc: validDoc, ExpectedRevision: 3,
			})
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("revision が 1 未満は 400 相当", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, nil)
		_, err := newDocUC(editable, repo).UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			MaterialActor: actorIn(wsA), ID: 1, Doc: validDoc, ExpectedRevision: 0,
		})
		require.ErrorIs(t, err, usecase.ErrChapterDocInvalid)
	})

	t.Run("doc が不正なら 400 相当", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, nil)
		_, err := newDocUC(editable, repo).UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			MaterialActor: actorIn(wsA), ID: 1, Doc: `{"type":"paragraph"}`, ExpectedRevision: 3,
		})
		require.ErrorIs(t, err, usecase.ErrChapterDocInvalid)
	})

	t.Run("競合はそのまま伝える", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, repository.ErrChapterDocConflict)
		_, err := newDocUC(editable, repo).UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			MaterialActor: actorIn(wsA), ID: 1, Doc: validDoc, ExpectedRevision: 3,
		})
		require.ErrorIs(t, err, repository.ErrChapterDocConflict)
	})
}
