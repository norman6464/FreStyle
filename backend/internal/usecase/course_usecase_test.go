package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_コース_取得_traineeは下書き不可(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: false}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Get(context.Background(), 5, 10, domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_コース_取得_traineeは自社の公開を読める(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Get(context.Background(), 5, 10, domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(5), got.ID)
}

func Test_コース_取得_別会社は禁止(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 5, CompanyID: 10, IsPublished: true}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Get(context.Background(), 5, 99, domain.RoleCompanyAdmin)
	require.Error(t, err)
}

func Test_コース_作成_traineeは禁止(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 1, ActorCompanyID: 10, ActorRole: domain.RoleTrainee,
		Title: "Web 基礎",
	})
	require.Error(t, err)
}

func Test_コース_作成_会社管理者は成功(t *testing.T) {
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

func Test_コース_作成_会社未所属は禁止(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorCompanyID: 0, ActorRole: domain.RoleCompanyAdmin,
		Title: "X",
	})
	require.Error(t, err)
}

func Test_コース_更新_別会社は禁止(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10, Title: "old"}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorCompanyID: 99, ActorRole: domain.RoleCompanyAdmin, Title: "new",
	})
	require.Error(t, err)
	assert.Nil(t, crepo.updated)
}

func Test_コース_更新_自社管理者は成功(t *testing.T) {
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

func Test_コース_削除_traineeは禁止(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleTrainee)
	require.Error(t, err)
}

func Test_コース_削除_自社管理者は教材も連鎖削除(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	err := uc.Delete(context.Background(), 1, 10, domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), crepo.deleted)
	assert.Equal(t, uint64(1), mrepo.deletedByCo, "コース配下の教材も cascade で削除される")
}

func Test_コース_作成_カテゴリ付きで成功(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "PostgreSQL 徹底入門", Category: domain.CourseCategoryDatabase,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.CourseCategoryDatabase, got.Category)
}

func Test_コース_作成_不正なカテゴリは拒否(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "X", Category: "unknown-category",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid course category")
	assert.Nil(t, crepo.created)
}

func Test_コース_作成_カテゴリ未分類は許可(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "X", Category: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "", got.Category)
}

func Test_コース_更新_カテゴリを変更できる(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10, Category: domain.CourseCategoryDevBasics}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "Terraform 入門", Category: domain.CourseCategoryInfra,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.CourseCategoryInfra, got.Category)
}

func Test_コース_作成_言語付きで成功(t *testing.T) {
	crepo := &fakeCourseRepo{}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "Go 言語徹底攻略", Category: domain.CourseCategoryBackend, Language: "go",
	})
	require.NoError(t, err)
	assert.Equal(t, "go", got.Language)
	assert.Equal(t, "go", crepo.created.Language)
}

func Test_コース_更新_言語を変更できる(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10, Language: "go"}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "Terraform 入門", Language: "terraform",
	})
	require.NoError(t, err)
	assert.Equal(t, "terraform", got.Language)
}

func Test_コース_更新_言語は空にもできる(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10, Language: "go"}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "Design Doc 入門", Language: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "", got.Language)
}

func Test_コース_更新_不正なカテゴリは拒否(t *testing.T) {
	crepo := &fakeCourseRepo{getResp: &domain.Course{ID: 1, CompanyID: 10}}
	mrepo := &fakeTeachingMaterialRepo{}
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorCompanyID: 10, ActorRole: domain.RoleCompanyAdmin,
		Title: "X", Category: "nope",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid course category")
	assert.Nil(t, crepo.updated)
}
