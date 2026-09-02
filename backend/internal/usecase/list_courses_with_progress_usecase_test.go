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
	return newListUC(mrepo, completed, rows)
}

// listUCSplitCounts は「公開だけの数」と「下書き込みの数」を撃ち分ける repo で組み立てる。
// 権限が混ざったときに、コースごとに正しい方を選べているかを見るために使う。
func listUCSplitCounts(
	rows []repository.CourseWithFacts, published, all map[uint64]int,
) (*usecase.ListCoursesWithProgressUseCase, *mockMaterialRepo) {
	mrepo := &mockMaterialRepo{}
	mrepo.On("CountByCourseForWorkspace", mock.Anything, mock.Anything, false).Return(published, nil).Maybe()
	mrepo.On("CountByCourseForWorkspace", mock.Anything, mock.Anything, true).Return(all, nil).Maybe()
	return newListUC(mrepo, nil, rows)
}

func newListUC(
	mrepo *mockMaterialRepo, completed map[uint64]int, rows []repository.CourseWithFacts,
) (*usecase.ListCoursesWithProgressUseCase, *mockMaterialRepo) {
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

func Test_コース一覧進捗付き_下書きの章を数えるのはコースごとに決まる(t *testing.T) {
	// **権限が混ざるときが本番。** 1 つでも編集できれば全部を下書き込みで数える、と
	// してはいけない。閲覧しかできないコースの下書き章数まで数に出てしまい、
	// そのコースに何本の下書きがあるかが漏れる。
	uc, _ := listUCSplitCounts(
		[]repository.CourseWithFacts{
			courseFacts(1, "読むだけ", materialFactsConfig{member: true, published: true}),
			courseFacts(2, "編集できる", materialFactsConfig{
				member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
			}),
		},
		map[uint64]int{1: 2, 2: 3}, // 公開だけ
		map[uint64]int{1: 9, 2: 7}, // 下書き込み
	)

	out, err := uc.Execute(context.Background(), listInput())
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, 2, out[0].MaterialCount, "読むだけのコースに下書きの数を出さない")
	assert.Equal(t, 7, out[1].MaterialCount, "編集できるコースは下書きも数える")
}

func Test_コース一覧進捗付き_編集できるコースがあっても進捗は返る(t *testing.T) {
	// 以前は「編集できるコースが 1 つでもあれば完了記録を引かない」としていたため、
	// 受講もしている人の進捗が全コースで 0 に見えていた。
	uc, _ := listUC([]repository.CourseWithFacts{
		courseFacts(1, "受講している", materialFactsConfig{member: true, published: true}),
		courseFacts(2, "編集できる", materialFactsConfig{
			member: true, published: true, role: grantRoleOf(domain.GrantRoleEditor),
		}),
	}, map[uint64]int{1: 5}, map[uint64]int{1: 3})

	out, err := uc.Execute(context.Background(), listInput())
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, 3, out[0].CompletedCount, "自分の進捗が消えている")
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
