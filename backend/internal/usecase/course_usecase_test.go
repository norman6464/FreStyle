package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCourseUseCase_List_TraineeOnlyPublished(t *testing.T) {
	crepo := &fakeCourseRepo{rows: []domain.Course{{ID: 1}}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.List(context.Background(), 10, domain.RoleTrainee)
	require.NoError(t, err)
}

func TestCourseUseCase_List_NoCompanyReturnsEmpty(t *testing.T) {
	crepo := &fakeCourseRepo{rows: []domain.Course{{ID: 1}}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	out, err := uc.List(context.Background(), 0, domain.RoleSuperAdmin)
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestCourseUseCase_Get_TraineeCannotReadDraft(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Get(context.Background(), 5, 10, domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestCourseUseCase_Get_TraineeCanReadPublishedSameCompany(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Get(context.Background(), 5, 10, domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(5), got.ID)
}

func TestCourseUseCase_Get_CrossCompanyForbidden(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Get(context.Background(), 5, 99, domain.RoleCompanyAdmin)
	require.Error(t, err)
}

func TestCourseUseCase_Create_TraineeForbidden(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
		Title: "Web 基礎",
	})
	require.Error(t, err)
}

func TestCourseUseCase_Create_CompanyAdminSucceeds(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "Web 基礎", Description: "HTTP / REST", SortOrder: 10, IsPublished: true,
	})
	require.NoError(t, err)
	require.NotNil(t, crepo.created)
	assert.Equal(t, uint64(7), crepo.created.CreatedByUserID)
	assert.Equal(t, uint64(10), crepo.created.CompanyID)
	assert.Equal(t, "Web 基礎", got.Title)
	assert.Equal(t, 10, got.SortOrder)
}

func TestCourseUseCase_Create_NoCompanyForbidden(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorCompanyID: 0, ActorRole: domain.RoleCompanyAdmin,
		Title: "X",
	})
	require.Error(t, err)
}

func TestCourseUseCase_Update_CrossCompanyForbidden(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10, Title: "old"}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorCompanyID: 99, ActorRole: domain.RoleCompanyAdmin, Title: "new",
	})
	require.Error(t, err)
	assert.Nil(t, crepo.updated)
}

func TestCourseUseCase_Update_SameCompanyAdminSucceeds(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10, Title: "old"}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "new", Description: "X", SortOrder: 200, IsPublished: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "new", got.Title)
	assert.NotNil(t, crepo.updated)
}

func TestCourseUseCase_Delete_TraineeForbidden(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleTrainee)
	require.Error(t, err)
}

func TestCourseUseCase_Delete_SameCompanyAdminCascadesMaterials(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), crepo.deleted)
	assert.Equal(t, uint64(1), mrepo.deletedByCo, "コース配下の教材も cascade で削除される")
}
