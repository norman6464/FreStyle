package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTeachingMaterialRepo は usecase テスト用フェイク。
type fakeTeachingMaterialRepo struct {
	rows    []domain.TeachingMaterial
	getResp *domain.TeachingMaterial
	getErr  error
	created *domain.TeachingMaterial
	updated *domain.TeachingMaterial
	deleted uint64
	// list 呼び出し時の引数記録
	listCompany      uint64
	listIncludeDraft bool
}

func (r *fakeTeachingMaterialRepo) ListByCompany(_ context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	r.listCompany, r.listIncludeDraft = companyID, includeUnpublished
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

func TestTeachingMaterialUseCase_List_TraineeOnlyPublished(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	_, err := uc.List(context.Background(), 10, domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(10), repo.listCompany)
	assert.False(t, repo.listIncludeDraft, "trainee は draft を含まない")
}

func TestTeachingMaterialUseCase_List_CompanyAdminIncludesDraft(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	_, err := uc.List(context.Background(), 10, domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.True(t, repo.listIncludeDraft)
}

func TestTeachingMaterialUseCase_List_NoCompanyReturnsEmpty(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{rows: []domain.TeachingMaterial{{ID: 1}}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	out, err := uc.List(context.Background(), 0, domain.RoleSuperAdmin)
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestTeachingMaterialUseCase_Get_TraineeCannotReadDraft(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, IsPublished: false,
	}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	_, err := uc.Get(context.Background(), 1, 10, domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeachingMaterialUseCase_Get_TraineeCanReadPublishedSameCompany(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, IsPublished: true,
	}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	got, err := uc.Get(context.Background(), 1, 10, domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func TestTeachingMaterialUseCase_Get_CrossCompanyForbidden(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, IsPublished: true,
	}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	_, err := uc.Get(context.Background(), 1, 99, domain.RoleCompanyAdmin)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestTeachingMaterialUseCase_Get_SuperAdminCrossCompanyAllowed(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, IsPublished: false,
	}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	got, err := uc.Get(context.Background(), 1, 99, domain.RoleSuperAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)
}

func TestTeachingMaterialUseCase_Create_TraineeForbidden(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
		Title: "X", Content: "Y", IsPublished: true,
	})
	require.Error(t, err)
}

func TestTeachingMaterialUseCase_Create_CompanyAdminSucceeds(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	got, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "Spring 入門", Content: "# Spring", IsPublished: true,
	})
	require.NoError(t, err)
	require.NotNil(t, repo.created)
	assert.Equal(t, uint64(7), repo.created.CreatedByUserID)
	assert.Equal(t, uint64(10), repo.created.CompanyID)
	assert.Equal(t, "Spring 入門", got.Title)
}

func TestTeachingMaterialUseCase_Create_NoCompanyForbidden(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	_, err := uc.Create(context.Background(), usecase.CreateTeachingMaterialInput{
		ActorUserID: 7, ActorCompanyID: 0, ActorRole: domain.RoleCompanyAdmin,
		Title: "X",
	})
	require.Error(t, err)
}

func TestTeachingMaterialUseCase_Update_CrossCompanyForbidden(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, Title: "old",
	}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	_, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
		ID: 1, ActorCompanyID: 99, ActorRole: domain.RoleCompanyAdmin, Title: "new",
	})
	require.Error(t, err)
	assert.Nil(t, repo.updated)
}

func TestTeachingMaterialUseCase_Update_SameCompanyAdminSucceeds(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10, Title: "old",
	}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	got, err := uc.Update(context.Background(), usecase.UpdateTeachingMaterialInput{
		ID: 1, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "new", Content: "X", IsPublished: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "new", got.Title)
	assert.NotNil(t, repo.updated)
}

func TestTeachingMaterialUseCase_Delete_TraineeForbidden(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10,
	}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleTrainee)
	require.Error(t, err)
}

func TestTeachingMaterialUseCase_Delete_SameCompanyAdminSucceeds(t *testing.T) {
	repo := &fakeTeachingMaterialRepo{getResp: &domain.TeachingMaterial{
		ID: 1, CompanyID: 10,
	}}
	uc := usecase.NewTeachingMaterialUseCase(repo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), repo.deleted)
}
