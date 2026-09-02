package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// wsA / wsB はコースの所属比較を固定するための 2 つのワークスペース ID。
const (
	wsA = "0198a000-0000-7000-8000-0000000000c1"
	wsB = "0198a000-0000-7000-8000-0000000000c2"
)

// newCourseUC はコースの usecase を、指定した「見え方」で組み立てる。
func newCourseUC(cfg materialFactsConfig, member bool) (*usecase.CourseUseCase, *courseStore, *mockCourseRepo) {
	crepo, cstore := courseRepo(courseFakeConfig{get: &domain.Course{ID: 5, WorkspaceID: strPtr(wsA)}})
	mrepo, _ := materialRepo(materialFakeConfig{})
	_, perm := materialPerm(cfg)
	return usecase.NewCourseUseCase(crepo, mrepo, perm, principalsFor(member)), cstore, crepo
}

func Test_コース_見えない相手には実在を教えない(t *testing.T) {
	// 「無い」と「見えない」を撃ち分けると、ID を総当たりするだけで隠したコースの
	// 実在が分かる。どちらも同じ ErrNotFound に落ちること。
	for _, c := range []struct {
		name string
		cfg  materialFactsConfig
	}{
		{"下書きに付与が無い", materialFactsConfig{member: true, published: false}},
		{"別テナントのコース", materialFactsConfig{notFound: true}},
		{"ワークスペースに所属していない", materialFactsConfig{member: false, published: true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			uc, _, _ := newCourseUC(c.cfg, true)
			_, err := uc.Get(context.Background(), 5, actorIn(wsA))
			assert.ErrorIs(t, err, domain.ErrNotFound)
		})
	}
}

func Test_コース_公開済みは一員なら誰でも読める(t *testing.T) {
	// 読むことに付与を要求しない（研修を受ける人が教材を開くたびに権限を配らない）。
	uc, _, _ := newCourseUC(materialFactsConfig{member: true, published: true}, true)
	got, err := uc.Get(context.Background(), 5, actorIn(wsA))
	require.NoError(t, err)
	assert.Equal(t, uint64(5), got.ID)
}

func Test_コース_未所属は作れない(t *testing.T) {
	uc, _, _ := newCourseUC(materialFactsConfig{}, false)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		MaterialActor: usecase.MaterialActor{ActorUserID: 1},
		Title:         "Web 基礎", Category: domain.ValidCourseCategories[0],
	})
	assert.ErrorIs(t, err, usecase.ErrMaterialForbidden)
}

func Test_コース_一員でなければ作れない(t *testing.T) {
	// 所属は principals の行が唯一の表現。行が無ければ一員ではない。
	uc, _, _ := newCourseUC(materialFactsConfig{}, false)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		MaterialActor: actorIn(wsA),
		Title:         "Web 基礎", Category: domain.ValidCourseCategories[0],
	})
	assert.ErrorIs(t, err, usecase.ErrMaterialForbidden)
}

func Test_コース_一員なら作れて作成者がadminになる(t *testing.T) {
	// 誰でも作れるのに誰も扱えない、という状態を作らない。コースと付与は 1 つの
	// トランザクションで書くので、repository も専用のメソッドを通ること。
	uc, _, crepo := newCourseUC(materialFactsConfig{}, true)
	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		MaterialActor: actorIn(wsA),
		Title:         "Web 基礎", Category: domain.ValidCourseCategories[0],
	})
	require.NoError(t, err)
	assert.Equal(t, "Web 基礎", got.Title)
	assert.Equal(t, wsA, *got.WorkspaceID)
	crepo.AssertCalled(t, "CreateWithOwnerGrant", mock.Anything, mock.Anything,
		"0198a000-0000-7000-8000-0000000000a1")
	// 付与を伴わない Create は使わない（使うと権限の無いコースが残る）。
	crepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func Test_コース_付与に失敗したら作成も失敗する(t *testing.T) {
	// コースだけ出来て誰も扱えない、という状態を返さないこと。
	crepo, _ := courseRepo(courseFakeConfig{writeErr: errors.New("db")})
	mrepo, _ := materialRepo(materialFakeConfig{})
	_, perm := materialPerm(materialFactsConfig{})
	uc := usecase.NewCourseUseCase(crepo, mrepo, perm, principalsFor(true))

	got, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		MaterialActor: actorIn(wsA),
		Title:         "Web 基礎", Category: domain.ValidCourseCategories[0],
	})
	require.Error(t, err)
	assert.Nil(t, got, "失敗したのにコースを返している")
}

func Test_コース_分類が既知でなければ作れない(t *testing.T) {
	uc, _, _ := newCourseUC(materialFactsConfig{}, true)
	_, err := uc.Create(context.Background(), usecase.CreateCourseInput{
		MaterialActor: actorIn(wsA),
		Title:         "Web 基礎", Category: "unknown",
	})
	require.Error(t, err)
	assert.NotErrorIs(t, err, usecase.ErrMaterialForbidden)
}

func Test_コース_編集は付与で決まる(t *testing.T) {
	update := func(uc *usecase.CourseUseCase) error {
		_, err := uc.Update(context.Background(), usecase.UpdateCourseInput{
			MaterialActor: actorIn(wsA), ID: 5,
			Title: "改題", Category: domain.ValidCourseCategories[0],
		})
		return err
	}

	t.Run("読めるが付与が無ければ 403", func(t *testing.T) {
		// 見えている相手には理由を返してよい（実在は既に知っている）。
		uc, _, _ := newCourseUC(materialFactsConfig{member: true, published: true}, true)
		assert.ErrorIs(t, update(uc), usecase.ErrMaterialForbidden)
	})

	t.Run("editor の付与があれば編集できる", func(t *testing.T) {
		uc, cstore, _ := newCourseUC(materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
		}, true)
		require.NoError(t, update(uc))
		require.NotNil(t, cstore.updated)
		assert.Equal(t, "改題", cstore.updated.Title)
	})

	t.Run("ワークスペースの admin も編集できる", func(t *testing.T) {
		uc, _, _ := newCourseUC(materialFactsConfig{member: true, workspaceAdmin: true}, true)
		assert.NoError(t, update(uc))
	})

	t.Run("viewer の付与では編集できない", func(t *testing.T) {
		uc, _, _ := newCourseUC(materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleViewer),
		}, true)
		assert.ErrorIs(t, update(uc), usecase.ErrMaterialForbidden)
	})
}

func Test_コース_削除も編集と同じ条件(t *testing.T) {
	t.Run("付与が無ければ消せない", func(t *testing.T) {
		uc, cstore, _ := newCourseUC(materialFactsConfig{member: true, published: true}, true)
		assert.ErrorIs(t, uc.Delete(context.Background(), 5, actorIn(wsA)), usecase.ErrMaterialForbidden)
		assert.Zero(t, cstore.deleted, "断ったのに消しに行っている")
	})

	t.Run("editor なら配下ごと消せる", func(t *testing.T) {
		uc, cstore, _ := newCourseUC(materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
		}, true)
		require.NoError(t, uc.Delete(context.Background(), 5, actorIn(wsA)))
		assert.Equal(t, uint64(5), cstore.deleted)
	})
}
