package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wsA / wsB はコースの workspace_id 比較を固定するための 2 つのワークスペース ID。
// wsA が「自社」、wsB が「別会社」を表す。
const (
	wsA = "0198a000-0000-7000-8000-0000000000c1"
	wsB = "0198a000-0000-7000-8000-0000000000c2"
)

func Test_コース_取得_traineeは下書き不可(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, WorkspaceID: strPtr(wsA), IsPublished: false}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Get(context.Background(), 5, domain.WorkspaceRefOf(wsA), domain.RoleTrainee)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func Test_コース_取得_traineeは自社の公開を読める(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, WorkspaceID: strPtr(wsA), IsPublished: true}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Get(context.Background(), 5, domain.WorkspaceRefOf(wsA), domain.RoleTrainee)
	require.NoError(t, err)
	assert.Equal(t, uint64(5), got.ID)
}

func Test_コース_取得_別会社は禁止(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, WorkspaceID: strPtr(wsA), IsPublished: true}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Get(context.Background(), 5, domain.WorkspaceRefOf(wsB), domain.RoleCompanyAdmin)
	require.Error(t, err)
}

func Test_コース_作成_traineeは禁止(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 1, ActorRole: domain.RoleTrainee,
		Title: "Web 基礎",
	})
	require.Error(t, err)
}

func Test_コース_作成_会社管理者は成功(t *testing.T) {
	crepo, cstore := courseRepo(courseFakeConfig{})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "Web 基礎", Description: "HTTP / REST", SortOrder: 10, IsPublished: true,
	})
	require.NoError(t, err)
	require.NotNil(t, cstore.created)
	assert.Equal(t, uint64(7), cstore.created.CreatedByUserID)
	require.NotNil(t, cstore.created.WorkspaceID)
	assert.Equal(t, wsA, *cstore.created.WorkspaceID)
	assert.Equal(t, "Web 基礎", cstore.created.Title)
	assert.Equal(t, "HTTP / REST", cstore.created.Description)
	assert.Equal(t, 10, cstore.created.SortOrder)
	assert.True(t, cstore.created.IsPublished)
	assert.Equal(t, "Web 基礎", got.Title)
	assert.Equal(t, 10, got.SortOrder)
}

func Test_コース_更新_別会社は禁止(t *testing.T) {
	crepo, cstore := courseRepo(courseFakeConfig{get: &domain.Course{ID: 1, WorkspaceID: strPtr(wsA), Title: "old"}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsB), ActorRole: domain.RoleCompanyAdmin, Title: "new",
	})
	require.Error(t, err)
	assert.Nil(t, cstore.updated)
}

func Test_コース_更新_自社管理者は成功(t *testing.T) {
	crepo, cstore := courseRepo(courseFakeConfig{get: &domain.Course{ID: 1, WorkspaceID: strPtr(wsA), Title: "old"}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "new", Description: "X", SortOrder: 200, IsPublished: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "new", got.Title)
	require.NotNil(t, cstore.updated)
	assert.Equal(t, "new", cstore.updated.Title)
	assert.Equal(t, "X", cstore.updated.Description)
	assert.Equal(t, 200, cstore.updated.SortOrder)
	assert.True(t, cstore.updated.IsPublished)
}

func Test_コース_削除_traineeは禁止(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 1, WorkspaceID: strPtr(wsA)}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	err := uc.Delete(context.Background(), 1, domain.WorkspaceRefOf(wsA), domain.RoleTrainee)
	require.Error(t, err)
}

func Test_コース_削除_自社管理者は教材も連鎖削除(t *testing.T) {
	crepo, cstore := courseRepo(courseFakeConfig{get: &domain.Course{ID: 1, WorkspaceID: strPtr(wsA)}})
	mrepo, mstore := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	err := uc.Delete(context.Background(), 1, domain.WorkspaceRefOf(wsA), domain.RoleCompanyAdmin)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), cstore.deleted)
	assert.Equal(t, uint64(1), mstore.deletedByCourse, "コース配下の教材も cascade で削除される")
}

func Test_コース_作成_カテゴリ付きで成功(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "PostgreSQL 徹底入門", Category: domain.CourseCategoryDatabase,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.CourseCategoryDatabase, got.Category)
}

func Test_コース_作成_不正なカテゴリは拒否(t *testing.T) {
	crepo, cstore := courseRepo(courseFakeConfig{})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "X", Category: "unknown-category",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid course category")
	assert.Nil(t, cstore.created)
}

func Test_コース_作成_カテゴリ未分類は許可(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "X", Category: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "", got.Category)
}

func Test_コース_作成_ワークスペース未所属は禁止(t *testing.T) {
	crepo, cstore := courseRepo(courseFakeConfig{})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorWorkspace: domain.NoWorkspace(), ActorRole: domain.RoleCompanyAdmin,
		Title: "X",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace")
	assert.Nil(t, cstore.created)
}

func Test_コース_更新_カテゴリを変更できる(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 1, WorkspaceID: strPtr(wsA), Category: domain.CourseCategoryDevBasics}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "Terraform 入門", Category: domain.CourseCategoryInfra,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.CourseCategoryInfra, got.Category)
}

func Test_コース_作成_言語付きで成功(t *testing.T) {
	crepo, cstore := courseRepo(courseFakeConfig{})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		ActorUserID: 7, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "Go 言語徹底攻略", Category: domain.CourseCategoryBackend, Language: "go",
	})
	require.NoError(t, err)
	assert.Equal(t, "go", got.Language)
	assert.Equal(t, "go", cstore.created.Language)
}

func Test_コース_更新_言語を変更できる(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 1, WorkspaceID: strPtr(wsA), Language: "go"}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "Terraform 入門", Language: "terraform",
	})
	require.NoError(t, err)
	assert.Equal(t, "terraform", got.Language)
}

func Test_コース_更新_言語は空にもできる(t *testing.T) {
	crepo, _ := courseRepo(courseFakeConfig{get: &domain.Course{ID: 1, WorkspaceID: strPtr(wsA), Language: "go"}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	got, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "Design Doc 入門", Language: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "", got.Language)
}

func Test_コース_更新_不正なカテゴリは拒否(t *testing.T) {
	crepo, cstore := courseRepo(courseFakeConfig{get: &domain.Course{ID: 1, WorkspaceID: strPtr(wsA)}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo)
	_, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
		ID: 1, ActorWorkspace: domain.WorkspaceRefOf(wsA), ActorRole: domain.RoleCompanyAdmin,
		Title: "X", Category: "nope",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid course category")
	assert.Nil(t, cstore.updated)
}
