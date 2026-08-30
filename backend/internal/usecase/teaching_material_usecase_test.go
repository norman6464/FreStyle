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

func Test_教材_コース別一覧_traineeは公開のみ(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: true}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, domain.WorkspaceRefOf(wsA), domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(5), mstore.listCourseID)
	assert.False(t, mstore.listIncludeAll, "trainee は draft を含まない")
}

func Test_教材_コース別一覧_traineeは非公開コースを見られない(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: false}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, domain.WorkspaceRefOf(wsA), domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_教材_コース別一覧_会社管理者は下書きも含む(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: false}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, domain.WorkspaceRefOf(wsA), domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.True(t, mstore.listIncludeAll)
}

func Test_教材_取得_traineeは下書き不可(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, WorkspaceID: strPtr(wsA), IsPublished: false,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: true}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Get(context.Background(), 1, domain.WorkspaceRefOf(wsA), domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_教材_取得_traineeは自社の公開を読める(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, WorkspaceID: strPtr(wsA), IsPublished: true,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: true}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Get(context.Background(), 1, domain.WorkspaceRefOf(wsA), domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func Test_教材_取得_別会社は禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, WorkspaceID: strPtr(wsA), IsPublished: true,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: true}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Get(context.Background(), 1, domain.WorkspaceRefOf(wsB), domain.RoleCompanyAdmin)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_教材_取得_運営は別会社も許可(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, WorkspaceID: strPtr(wsA), IsPublished: false,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA), IsPublished: false}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Get(context.Background(), 1, domain.WorkspaceRefOf(wsB), domain.RoleSuperAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func Test_教材_作成_traineeは禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA)}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleTrainee,
		CourseID: 5, Title: "X", IsPublished: true,
	})
	require.Error(t, err)
	mrepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func Test_教材_作成_会社管理者は成功(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA)}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "Spring 入門", IsPublished: true,
	})
	require.NoError(t, err)
	require.NotNil(t, mstore.created)
	assert.Equal(t, uint64(7), mstore.created.CreatedByUserID)
	assert.Equal(t, uint64(10), mstore.created.CompanyID)
	assert.Equal(t, uint64(5), mstore.created.CourseID)
	assert.Equal(t, "Spring 入門", mstore.created.Title)
	assert.True(t, mstore.created.IsPublished)
	require.NotNil(t, mstore.created.WorkspaceID)
	assert.Equal(t, wsA, *mstore.created.WorkspaceID, "コースの workspace_id を継承する")
	assert.Equal(t, "Spring 入門", got.Title)
	require.NotNil(t, got.WorkspaceID)
	assert.Equal(t, wsA, *got.WorkspaceID)
}

func Test_教材_作成_コースID欠落は禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "X",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "course_id")
}

func Test_教材_作成_別会社コースは禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 99, WorkspaceID: strPtr(wsB)}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "X",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
	mrepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func Test_教材_作成_会社未所属は禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorWorkspace: domain.NoWorkspace(), ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "X",
	})
	require.Error(t, err)
}

func Test_教材_更新_別会社は禁止(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, WorkspaceID: strPtr(wsA), Title: "old",
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA)}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
		ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsB), ActorRole: domain.RoleCompanyAdmin, Title: "new",
	})
	require.Error(t, err)
	assert.Nil(t, mstore.updated)
}

func Test_教材_更新_自社管理者は成功(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, WorkspaceID: strPtr(wsA), Title: "old",
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA)}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
		ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "new", OrderInCourse: 200, IsPublished: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "new", got.Title)
	assert.Equal(t, 200, got.OrderInCourse)
	require.NotNil(t, mstore.updated)
	assert.Equal(t, "new", mstore.updated.Title)
	assert.Equal(t, 200, mstore.updated.OrderInCourse)
	assert.True(t, mstore.updated.IsPublished)
}

func Test_教材_削除_traineeは禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, WorkspaceID: strPtr(wsA),
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA)}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	err := uc.Delete(context.Background(), 1, domain.WorkspaceRefOf(wsA), domain.RoleTrainee)
	require.Error(t, err)
	mrepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func Test_教材_削除_自社管理者は成功(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, WorkspaceID: strPtr(wsA),
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, WorkspaceID: strPtr(wsA)}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	err := uc.Delete(context.Background(), 1, domain.WorkspaceRefOf(wsA), domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), mstore.deleted)
}

// newIdleCourseRepo は UpdateDoc 系テスト用の「呼ばれない」course repo mock。
func newIdleCourseRepo() *mockCourseRepo {
	repo := &mockCourseRepo{}
	repo.On("GetByID", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	return repo
}

// --- UpdateDoc（リッチ本文の楽観ロック保存） ---

func docUpdateRepo(existing *domain.TeachingMaterial, updated *domain.TeachingMaterial, updateErr error) *mockMaterialRepo {
	repo := &mockMaterialRepo{}
	repo.On("GetByID", mock.Anything, mock.Anything).Return(existing, nil).Maybe()
	repo.On("UpdateDocWithRevision", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(updated, updateErr).Maybe()
	return repo
}

func TestTeachingMaterialUseCase_UpdateDoc(t *testing.T) {
	validDoc := `{"type":"doc","content":[{"type":"paragraph"}]}`
	existing := &domain.TeachingMaterial{ID: 1, CompanyID: 10, WorkspaceID: strPtr(wsA), Revision: 3}

	t.Run("company_admin は自社の章を保存でき revision 付きで返る", func(t *testing.T) {
		updatedDoc := validDoc
		updated := &domain.TeachingMaterial{ID: 1, CompanyID: 10, WorkspaceID: strPtr(wsA), Revision: 4, Doc: &updatedDoc}
		repo := docUpdateRepo(existing, updated, nil)
		uc := usecase.NewTeachingMaterialUseCase(repo, newIdleCourseRepo())
		got, err := uc.UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
			Doc: validDoc, ExpectedRevision: 3,
		})
		require.NoError(t, err)
		assert.Equal(t, 4, got.Revision)
		repo.AssertCalled(t, "UpdateDocWithRevision", mock.Anything, uint64(1), validDoc, 3)
	})

	t.Run("trainee は forbidden", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, nil)
		uc := usecase.NewTeachingMaterialUseCase(repo, newIdleCourseRepo())
		_, err := uc.UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleTrainee,
			Doc: validDoc, ExpectedRevision: 3,
		})
		require.ErrorContains(t, err, "forbidden")
		repo.AssertNotCalled(t, "UpdateDocWithRevision", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("未所属の super_admin は他社の章も保存できる", func(t *testing.T) {
		updatedDoc := validDoc
		updated := &domain.TeachingMaterial{ID: 1, CompanyID: 10, WorkspaceID: strPtr(wsA), Revision: 4, Doc: &updatedDoc}
		repo := docUpdateRepo(existing, updated, nil)
		uc := usecase.NewTeachingMaterialUseCase(repo, newIdleCourseRepo())
		got, err := uc.UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			ID: 1, ActorWorkspace: domain.NoWorkspace(), ActorRole: domain.RoleSuperAdmin,
			Doc: validDoc, ExpectedRevision: 3,
		})
		require.NoError(t, err)
		assert.Equal(t, 4, got.Revision)
	})

	t.Run("他社の章は company_admin でも forbidden", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, nil)
		uc := usecase.NewTeachingMaterialUseCase(repo, newIdleCourseRepo())
		_, err := uc.UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsB), ActorRole: domain.RoleCompanyAdmin,
			Doc: validDoc, ExpectedRevision: 3,
		})
		require.ErrorContains(t, err, "forbidden")
	})

	t.Run("doc の型不正（type が doc でない）は ErrChapterDocInvalid", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, nil)
		uc := usecase.NewTeachingMaterialUseCase(repo, newIdleCourseRepo())
		_, err := uc.UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
			Doc: `{"type":"paragraph"}`, ExpectedRevision: 3,
		})
		require.ErrorIs(t, err, usecase.ErrChapterDocInvalid)
	})

	t.Run("expectedRevision が 0 以下は ErrChapterDocInvalid", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, nil)
		uc := usecase.NewTeachingMaterialUseCase(repo, newIdleCourseRepo())
		_, err := uc.UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
			Doc: validDoc, ExpectedRevision: 0,
		})
		require.ErrorIs(t, err, usecase.ErrChapterDocInvalid)
	})

	t.Run("repository の版不一致（ErrChapterDocConflict）は素通しする", func(t *testing.T) {
		repo := docUpdateRepo(existing, nil, repository.ErrChapterDocConflict)
		uc := usecase.NewTeachingMaterialUseCase(repo, newIdleCourseRepo())
		_, err := uc.UpdateDoc(context.Background(), usecase.UpdateChapterDocInput{
			ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
			Doc: validDoc, ExpectedRevision: 2,
		})
		require.ErrorIs(t, err, repository.ErrChapterDocConflict)
	})
}
