package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// fakeTeachingMaterialRepo は usecase テスト用フェイク。
type fakeTeachingMaterialRepo struct {
	rows           []domain.TeachingMaterial
	getResp        *domain.TeachingMaterial
	getErr         error
	created        *domain.TeachingMaterial
	updated        *domain.TeachingMaterial
	deleted        uint64
	deletedByCo    uint64
	listCourseID   uint64
	listIncludeAll bool
}

func (r *fakeTeachingMaterialRepo) ListByCompany(_ context.Context, _ uint64, _ bool) ([]domain.TeachingMaterial, error) {
	return r.rows, nil
}

func (r *fakeTeachingMaterialRepo) ListByCourse(_ context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	r.listCourseID, r.listIncludeAll = courseID, includeUnpublished
	return r.rows, nil
}
func (r *fakeTeachingMaterialRepo) GetByID(_ context.Context, _ uint64) (*domain.TeachingMaterial, error) {
	return r.getResp, r.getErr
}
func (r *fakeTeachingMaterialRepo) Create(_ context.Context, m *domain.TeachingMaterial) error {
	m.ID = 99
	r.created = m
	return nil
}
func (r *fakeTeachingMaterialRepo) Update(_ context.Context, m *domain.TeachingMaterial) error {
	r.updated = m
	return nil
}
func (r *fakeTeachingMaterialRepo) Delete(_ context.Context, id uint64) error {
	r.deleted = id
	return nil
}
func (r *fakeTeachingMaterialRepo) DeleteByCourse(_ context.Context, courseID uint64) error {
	r.deletedByCo = courseID
	return nil
}

// fakeCourseRepo は course 依存をスタブ化する。
type fakeCourseRepo struct {
	rows    []domain.Course
	getResp *domain.Course
	getErr  error
	created *domain.Course
	updated *domain.Course
	deleted uint64
}

func (r *fakeCourseRepo) ListByCompany(_ context.Context, _ uint64, _ bool) ([]domain.Course, error) {
	return r.rows, nil
}
func (r *fakeCourseRepo) GetByID(_ context.Context, _ uint64) (*domain.Course, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.getResp, nil
}
func (r *fakeCourseRepo) Create(_ context.Context, c *domain.Course) error {
	c.ID = 88
	r.created = c
	return nil
}
func (r *fakeCourseRepo) Update(_ context.Context, c *domain.Course) error {
	r.updated = c
	return nil
}
func (r *fakeCourseRepo) Delete(_ context.Context, id uint64) error {
	r.deleted = id
	return nil
}

func TestTeachingMaterialUseCase_ListByCourse_TraineeOnlyPublished(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, 10, domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(5), mrepo.listCourseID)
	assert.False(t, mrepo.listIncludeAll, "trainee は draft を含まない")
}

func TestTeachingMaterialUseCase_ListByCourse_TraineeCannotSeeUnpublishedCourse(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, 10, domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeachingMaterialUseCase_ListByCourse_CompanyAdminIncludesDraft(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.ListByCourse(context.Background(), 5, 10, domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.True(t, mrepo.listIncludeAll)
}

func TestTeachingMaterialUseCase_Get_TraineeCannotReadDraft(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, IsPublished: false,
	}}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Get(context.Background(), 1, 10, domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeachingMaterialUseCase_Get_TraineeCanReadPublishedSameCompany(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, IsPublished: true,
	}}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Get(context.Background(), 1, 10, domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func TestTeachingMaterialUseCase_Get_CrossCompanyForbidden(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, IsPublished: true,
	}}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Get(context.Background(), 1, 99, domain.RoleCompanyAdmin)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeachingMaterialUseCase_Get_SuperAdminCrossCompanyAllowed(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, IsPublished: false,
	}}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Get(context.Background(), 1, 99, domain.RoleSuperAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func TestTeachingMaterialUseCase_Create_TraineeForbidden(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
		CourseID: 5, Title: "X", Content: "Y", IsPublished: true,
	})
	require.Error(t, err)
}

func TestTeachingMaterialUseCase_Create_CompanyAdminSucceeds(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "Spring 入門", Content: "# Spring", IsPublished: true,
	})
	require.NoError(t, err)
	require.NotNil(t, mrepo.created)
	assert.Equal(t, uint64(7), mrepo.created.CreatedByUserID)
	assert.Equal(t, uint64(10), mrepo.created.CompanyID)
	assert.Equal(t, uint64(5), mrepo.created.CourseID)
	assert.Equal(t, "Spring 入門", got.Title)
}

func TestTeachingMaterialUseCase_Create_MissingCourseIDForbidden(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{}
	crepo := &fakeCourseRepo{}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "X",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "course_id")
}

func TestTeachingMaterialUseCase_Create_CrossCompanyCourseForbidden(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 99}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "X",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeachingMaterialUseCase_Create_NoCompanyForbidden(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{}
	crepo := &fakeCourseRepo{}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 0, ActorRole: domain.RoleCompanyAdmin,
		CourseID: 5, Title: "X",
	})
	require.Error(t, err)
}

func TestTeachingMaterialUseCase_Update_CrossCompanyForbidden(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, Title: "old",
	}}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	_, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
		ID: 1, ActorCompanyID: 99, ActorRole: domain.RoleCompanyAdmin, Title: "new",
	})
	require.Error(t, err)
	assert.Nil(t, mrepo.updated)
}

func TestTeachingMaterialUseCase_Update_SameCompanyAdminSucceeds(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5, Title: "old",
	}}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	got, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
		ID: 1, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "new", Content: "X", OrderInCourse: 200, IsPublished: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "new", got.Title)
	assert.Equal(t, 200, got.OrderInCourse)
	assert.NotNil(t, mrepo.updated)
}

func TestTeachingMaterialUseCase_Delete_TraineeForbidden(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5,
	}}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleTrainee)
	require.Error(t, err)
}

func TestTeachingMaterialUseCase_Delete_SameCompanyAdminSucceeds(t *testing.T) {
	mrepo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, CourseID: 5,
	}}
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10}}
	uc := usecase.NewTeachingMaterialUseCase(mrepo, crepo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), mrepo.deleted)
}

// 念のため: gorm 依存を import で残すための no-op（一部 import path 揃え用）。
var _ = gorm.ErrRecordNotFound
