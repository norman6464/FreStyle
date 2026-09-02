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

const listWS = "0198a000-0000-7000-8000-0000000000c1"

// courseFacts は「このコースがその人にどう見えるか」を 1 件ぶん作る。
func courseFacts(id uint64, title string, cfg materialFactsConfig) repository.CourseWithFacts {
	return repository.CourseWithFacts{
		Course: domain.Course{ID: id, Title: title, IsPublished: cfg.published},
		Facts:  *cfg.facts(),
	}
}

// listUC は一覧の usecase を、指定したコースの見え方で組み立てる。
func listUC(
	rows []repository.CourseWithFacts, counts map[uint64]int, completed map[uint64]int,
) (*usecase.ListCoursesWithProgressUseCase, *mockMaterialRepo) {
	mrepo, _ := materialRepo(materialFakeConfig{counts: counts})
	prepo, _ := progressRepo(progressFakeConfig{counts: completed})
	perm := &mockMaterialPermRepo{}
	perm.On("ListCourseFactsForUser", mock.Anything, mock.Anything, mock.Anything).Return(rows, nil).Maybe()
	return usecase.NewListCoursesWithProgressUseCase(mrepo, prepo, perm), mrepo
}

func listInput() usecase.ListCoursesWithProgressInput {
	return usecase.ListCoursesWithProgressInput{MaterialActor: actorIn(listWS)}
}

func Test_コース一覧進捗付き_各コースに章数と完了章数が合成される(t *testing.T) {
	uc, _ := listUC([]repository.CourseWithFacts{
		courseFacts(1, "Git", materialFactsConfig{member: true, published: true}),
		courseFacts(2, "Docker", materialFactsConfig{member: true, published: true}),
	}, map[uint64]int{1: 3, 2: 12}, map[uint64]int{1: 2})

	out, err := uc.Execute(context.Background(), listInput())
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, uint64(1), out[0].ID)
	assert.Equal(t, 3, out[0].MaterialCount)
	assert.Equal(t, 2, out[0].CompletedCount)
	assert.Equal(t, 12, out[1].MaterialCount)
	assert.Equal(t, 0, out[1].CompletedCount, "完了記録が無いコースは 0")
}

func Test_コース一覧進捗付き_見せてよいコースだけを並べる(t *testing.T) {
	// **この一覧の核心。** 事実は全件返るので、ふるい落としはここが担う。
	// ここが緩むと、下書きのコースが受講者の一覧に並ぶ。
	uc, _ := listUC([]repository.CourseWithFacts{
		courseFacts(1, "公開", materialFactsConfig{member: true, published: true}),
		courseFacts(2, "下書き", materialFactsConfig{member: true, published: false}),
		courseFacts(3, "付与のある下書き", materialFactsConfig{
			member: true, published: false, role: grantRoleOf(domain.GrantRoleEditor),
		}),
	}, nil, nil)

	out, err := uc.Execute(context.Background(), listInput())
	require.NoError(t, err)
	ids := make([]uint64, 0, len(out))
	for _, c := range out {
		ids = append(ids, c.ID)
	}
	assert.Equal(t, []uint64{1, 3}, ids, "付与の無い下書きは出さない")
}

func Test_コース一覧進捗付き_下書きの章を数えるのは編集できる人だけ(t *testing.T) {
	t.Run("付与が無ければ公開の章だけ数える", func(t *testing.T) {
		uc, mrepo := listUC([]repository.CourseWithFacts{
			courseFacts(1, "公開", materialFactsConfig{member: true, published: true}),
		}, nil, nil)
		_, err := uc.Execute(context.Background(), listInput())
		require.NoError(t, err)
		mrepo.AssertCalled(t, "CountByCourseForWorkspace", mock.Anything, listWS, false)
	})

	t.Run("編集できるコースが 1 つでもあれば下書き込みで数える", func(t *testing.T) {
		uc, mrepo := listUC([]repository.CourseWithFacts{
			courseFacts(1, "編集できる", materialFactsConfig{
				member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
			}),
		}, nil, nil)
		_, err := uc.Execute(context.Background(), listInput())
		require.NoError(t, err)
		mrepo.AssertCalled(t, "CountByCourseForWorkspace", mock.Anything, listWS, true)
	})
}

func Test_コース一覧進捗付き_集計に無いコースは0章(t *testing.T) {
	uc, _ := listUC([]repository.CourseWithFacts{
		courseFacts(7, "空のコース", materialFactsConfig{member: true, published: true}),
	}, nil, nil)

	out, err := uc.Execute(context.Background(), listInput())
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, 0, out[0].MaterialCount)
	assert.Equal(t, 0, out[0].CompletedCount)
}

func Test_コース一覧進捗付き_未所属は空スライス(t *testing.T) {
	uc, _ := listUC(nil, nil, nil)
	out, err := uc.Execute(context.Background(), usecase.ListCoursesWithProgressInput{
		MaterialActor: usecase.MaterialActor{ActorUserID: 5, ActorWorkspace: domain.NoWorkspace()},
	})
	require.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out)
}

func Test_コース一覧進捗付き_0件でもnilではなく空スライス(t *testing.T) {
	// 0 件で nil を返すと JSON が null になり、画面が配列として扱えない。
	uc, _ := listUC([]repository.CourseWithFacts{}, nil, nil)
	out, err := uc.Execute(context.Background(), listInput())
	require.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out)
}
