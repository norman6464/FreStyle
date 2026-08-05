package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository/repofakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// materialStore は教材 fake が記録する呼び出し内容(course クラスタのテストで共有)。
type materialStore struct {
	listCourseID                uint64
	listIncludeAll              bool
	created                     *domain.TeachingMaterial
	updated                     *domain.TeachingMaterial
	deleted                     uint64
	deletedByCourse             uint64
	lastCountIncludeUnpublished *bool
}

// materialFakeConfig は教材 fake の応答設定。ゼロ値はすべて「空を返す」。
type materialFakeConfig struct {
	get      *domain.TeachingMaterial
	getErr   error
	counts   map[uint64]int
	countErr error
}

// materialRepo は TeachingMaterialRepository の生成 fake に、このクラスタが使う
// メソッドだけを差し込んで返す。残りは生成 fake がゼロ値を返すので no-op が要らない。
func materialRepo(cfg materialFakeConfig) (*repofakes.FakeTeachingMaterialRepository, *materialStore) {
	st := &materialStore{}
	repo := &repofakes.FakeTeachingMaterialRepository{
		GetByIDFunc: func(context.Context, uint64) (*domain.TeachingMaterial, error) {
			if cfg.getErr != nil {
				return nil, cfg.getErr
			}
			return cfg.get, nil
		},
		ListByCourseFunc: func(_ context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
			st.listCourseID, st.listIncludeAll = courseID, includeUnpublished
			return nil, nil
		},
		CountByCourseForCompanyFunc: func(_ context.Context, _ uint64, includeUnpublished bool) (map[uint64]int, error) {
			st.lastCountIncludeUnpublished = &includeUnpublished
			if cfg.countErr != nil {
				return nil, cfg.countErr
			}
			return cfg.counts, nil
		},
		CreateFunc: func(_ context.Context, m *domain.TeachingMaterial) error {
			m.ID = 99
			st.created = m
			return nil
		},
		UpdateFunc: func(_ context.Context, m *domain.TeachingMaterial) error {
			st.updated = m
			return nil
		},
		DeleteFunc: func(_ context.Context, id uint64) error {
			st.deleted = id
			return nil
		},
		DeleteByCourseFunc: func(_ context.Context, courseID uint64) error {
			st.deletedByCourse = courseID
			return nil
		},
	}
	return repo, st
}

// courseStore はコース fake が記録する更新内容(course クラスタのテストで共有)。
type courseStore struct {
	created *domain.Course
	updated *domain.Course
	deleted uint64
}

// courseFakeConfig はコース fake の応答設定。ゼロ値はすべて「空を返す」。
type courseFakeConfig struct {
	rows   []domain.Course
	get    *domain.Course
	getErr error
}

// courseRepo は CourseRepository の生成 fake に、このクラスタが使うメソッドだけを
// 差し込んで返す。
func courseRepo(cfg courseFakeConfig) (*repofakes.FakeCourseRepository, *courseStore) {
	st := &courseStore{}
	repo := &repofakes.FakeCourseRepository{
		ListByCompanyFunc: func(context.Context, uint64, bool) ([]domain.Course, error) {
			return cfg.rows, nil
		},
		GetByIDFunc: func(context.Context, uint64) (*domain.Course, error) {
			if cfg.getErr != nil {
				return nil, cfg.getErr
			}
			return cfg.get, nil
		},
		CreateFunc: func(_ context.Context, c *domain.Course) error {
			c.ID = 88
			st.created = c
			return nil
		},
		UpdateFunc: func(_ context.Context, c *domain.Course) error {
			st.updated = c
			return nil
		},
		DeleteFunc: func(_ context.Context, id uint64) error {
			st.deleted = id
			return nil
		},
	}
	return repo, st
}

func Test_教材_コース別一覧_traineeは公開のみ(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, 10, domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(5), mstore.listCourseID)
	assert.False(t, mstore.listIncludeAll, "trainee は draft を含まない")
}

func Test_教材_コース別一覧_traineeは非公開コースを見られない(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, 10, domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_教材_コース別一覧_会社管理者は下書きも含む(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, 10, domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.True(t, mstore.listIncludeAll)
}

func Test_教材_取得_traineeは下書き不可(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, IsPublished: false,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Get(context.Background(), 1, 10, domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_教材_取得_traineeは自社の公開を読める(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, IsPublished: true,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Get(context.Background(), 1, 10, domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func Test_教材_取得_別会社は禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, IsPublished: true,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Get(context.Background(), 1, 99, domain.RoleCompanyAdmin)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_教材_取得_運営は別会社も許可(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, IsPublished: false,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Get(context.Background(), 1, 99, domain.RoleSuperAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func Test_教材_作成_traineeは禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
		CourseID: 5, Title: "X", Content: "Y", IsPublished: true,
	})
	require.Error(t, err)
}

func Test_教材_作成_会社管理者は成功(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "Spring 入門", Content: "# Spring", IsPublished: true,
	})
	require.NoError(t, err)
	require.NotNil(t, mstore.created)
	assert.Equal(t, uint64(7), mstore.created.CreatedByUserID)
	assert.Equal(t, uint64(10), mstore.created.CompanyID)
	assert.Equal(t, uint64(5), mstore.created.CourseID)
	assert.Equal(t, "Spring 入門", got.Title)
}

func Test_教材_作成_コースID欠落は禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "X",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "course_id")
}

func Test_教材_作成_別会社コースは禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 99}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "X",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_教材_作成_会社未所属は禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{})
	crepo, _ := courseRepo(courseFakeConfig{})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 0, ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "X",
	})
	require.Error(t, err)
}

func Test_教材_更新_別会社は禁止(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, Title: "old",
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
		ID: 1, ActorCompanyID: 99, ActorRole: domain.RoleCompanyAdmin, Title: "new",
	})
	require.Error(t, err)
	assert.Nil(t, mstore.updated)
}

func Test_教材_更新_自社管理者は成功(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, Title: "old",
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
		ID: 1, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "new", Content: "X", OrderInCourse: 200, IsPublished: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "new", got.Title)
	assert.Equal(t, 200, got.OrderInCourse)
	assert.NotNil(t, mstore.updated)
}

func Test_教材_削除_traineeは禁止(t *testing.T) {
	mrepo, _ := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleTrainee)
	require.Error(t, err)
}

func Test_教材_削除_自社管理者は成功(t *testing.T) {
	mrepo, mstore := materialRepo(materialFakeConfig{get: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5,
	}})
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, CompanyID: 10}})
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), mstore.deleted)
}
