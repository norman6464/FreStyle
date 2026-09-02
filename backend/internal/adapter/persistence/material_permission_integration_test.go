//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 教材（コース / 章）の権限を実 PostgreSQL で固定する。
//
// 事実を集めるのは SQL、そこから何ができるかを決めるのは domain。ここで見るのは
// **どの事実を集めるか**で、規則そのものは domain の単体テストが持つ。
//
// いちばん大事なのは「ワークスペースの grant は admin だけが届く」。ノートの editor が
// 教材の編集権まで一度に得ないための境目で、ここが緩むと本番のワークスペース editor 3 名が
// 黙って全教材を編集できるようになる。

type materialFixture struct {
	db      *sql.DB
	perm    repository.MaterialPermissionRepository
	kb      repository.KnowledgeBasePermissionRepository
	ws      string
	otherWS string
	course  uint64
	// draftCourse は未公開のコース。公開判定がコース側でも効くことを見るために持つ。
	draftCourse uint64
	chapter     uint64
	// draft は同じコースの未公開の章。
	draft uint64
	alice uint64
	bob   uint64
}

func setupMaterialPermission(t *testing.T, db *sql.DB) materialFixture {
	t.Helper()
	testsupport.TruncateAll(t, db, append([]string{"course_chapters", "courses"}, kbTables...)...)
	f := materialFixture{
		db:   db,
		perm: persistence.NewMaterialPermissionRepository(db),
		kb:   persistence.NewKnowledgeBasePermissionRepository(db),
	}
	f.ws = createWorkspace(t, db, "mat-main")
	f.otherWS = createWorkspace(t, db, "mat-other")
	f.alice = createUser(t, db, "mat-alice")
	f.bob = createUser(t, db, "mat-bob")

	f.course = 7001
	f.draftCourse = 7002
	_, err := db.Exec(
		`INSERT INTO courses (id, workspace_id, created_by_user_id, title, is_published, created_at, updated_at)
		 VALUES ($1, $3, $4, '公開のコース', true, now(), now()),
		        ($2, $3, $4, '下書きのコース', false, now(), now())`,
		f.course, f.draftCourse, f.ws, f.alice,
	)
	require.NoError(t, err)
	f.chapter = 7101
	f.draft = 7102
	_, err = db.Exec(
		`INSERT INTO course_chapters (id, workspace_id, course_id, created_by_user_id, title, is_published, created_at, updated_at)
		 VALUES ($1, $3, $4, $5, '公開の章', true, now(), now()),
		        ($2, $3, $4, $5, '下書きの章', false, now(), now())`,
		f.chapter, f.draft, f.ws, f.course, f.alice,
	)
	require.NoError(t, err)
	return f
}

// member はユーザーをワークスペースの一員にして主体を返す。
func (f materialFixture) member(ctx context.Context, t *testing.T, userID uint64) *domain.Principal {
	t.Helper()
	p, err := f.kb.EnsureUserPrincipal(ctx, f.ws, userID)
	require.NoError(t, err)
	return p
}

func (f materialFixture) chapterPerm(ctx context.Context, t *testing.T, userID uint64) domain.MaterialPermission {
	t.Helper()
	facts, err := f.perm.ChapterFactsForUser(ctx, f.ws, f.chapter, userID)
	require.NoError(t, err)
	return domain.ResolveMaterialPermission(*facts)
}

func (f materialFixture) draftPerm(ctx context.Context, t *testing.T, userID uint64) domain.MaterialPermission {
	t.Helper()
	facts, err := f.perm.ChapterFactsForUser(ctx, f.ws, f.draft, userID)
	require.NoError(t, err)
	return domain.ResolveMaterialPermission(*facts)
}

func (f materialFixture) coursePerm(ctx context.Context, t *testing.T, userID uint64) domain.MaterialPermission {
	t.Helper()
	return f.coursePermOf(ctx, t, f.course, userID)
}

func (f materialFixture) coursePermOf(
	ctx context.Context, t *testing.T, courseID, userID uint64,
) domain.MaterialPermission {
	t.Helper()
	facts, err := f.perm.CourseFactsForUser(ctx, f.ws, courseID, userID)
	require.NoError(t, err)
	return domain.ResolveMaterialPermission(*facts)
}

func TestMaterialPermission_ワークスペースのgrantはadminだけが届く_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("ノートのeditorは教材を編集できない", func(t *testing.T) {
		// **この検査がこの PR の核心。** 本番のワークスペース editor 3 名は
		// いま教材を編集できない。ここが緩むと、その 3 名が黙って全教材を編集できるようになる。
		f := setupMaterialPermission(t, sqlDB)
		alice := f.member(ctx, t, f.alice)
		_, err := f.kb.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		got := f.chapterPerm(ctx, t, f.alice)
		assert.True(t, got.CanView, "公開済みなので読める")
		assert.False(t, got.CanEdit, "ノートの editor が教材まで編集できてはいけない")
		assert.False(t, got.CanManage)
		assert.False(t, f.draftPerm(ctx, t, f.alice).CanView, "下書きも見えない")
	})

	t.Run("ワークスペースのadminは配下すべてを扱える", func(t *testing.T) {
		f := setupMaterialPermission(t, sqlDB)
		alice := f.member(ctx, t, f.alice)
		_, err := f.kb.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)

		got := f.chapterPerm(ctx, t, f.alice)
		assert.True(t, got.CanEdit)
		assert.True(t, got.CanManage)
		assert.True(t, f.draftPerm(ctx, t, f.alice).CanView, "下書きも見える")
		assert.True(t, f.coursePerm(ctx, t, f.alice).CanManage, "コースも同じ")
	})
}

func TestMaterialPermission_付与の届き方_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("コースの付与は配下の章へ降りる", func(t *testing.T) {
		f := setupMaterialPermission(t, sqlDB)
		alice := f.member(ctx, t, f.alice)
		_, err := f.perm.UpsertCourseGrant(ctx, f.ws, f.course, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		assert.True(t, f.coursePerm(ctx, t, f.alice).CanEdit, "コース自身")
		assert.True(t, f.chapterPerm(ctx, t, f.alice).CanEdit, "配下の章へ降りる")
		assert.True(t, f.draftPerm(ctx, t, f.alice).CanView, "下書きも見える")
	})

	t.Run("章の付与はその章だけに効く", func(t *testing.T) {
		f := setupMaterialPermission(t, sqlDB)
		alice := f.member(ctx, t, f.alice)
		_, err := f.perm.UpsertChapterGrant(ctx, f.ws, f.chapter, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		assert.True(t, f.chapterPerm(ctx, t, f.alice).CanEdit, "張った章")
		assert.False(t, f.draftPerm(ctx, t, f.alice).CanEdit, "同じコースの別の章へは広がらない")
		assert.False(t, f.coursePerm(ctx, t, f.alice).CanEdit, "コースへは上がらない")
	})

	t.Run("弱い付与を重ねても下がらない", func(t *testing.T) {
		// 弱める操作は既定の層では表せない（合成は最も強いものを採る）。
		f := setupMaterialPermission(t, sqlDB)
		alice := f.member(ctx, t, f.alice)
		_, err := f.perm.UpsertCourseGrant(ctx, f.ws, f.course, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		_, err = f.perm.UpsertChapterGrant(ctx, f.ws, f.chapter, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, err)

		assert.True(t, f.chapterPerm(ctx, t, f.alice).CanEdit,
			"章に viewer を張ってもコースの editor は下がらない")
	})

	t.Run("付与は張った相手だけに効く", func(t *testing.T) {
		f := setupMaterialPermission(t, sqlDB)
		alice := f.member(ctx, t, f.alice)
		f.member(ctx, t, f.bob)
		_, err := f.perm.UpsertCourseGrant(ctx, f.ws, f.course, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		assert.True(t, f.chapterPerm(ctx, t, f.alice).CanEdit)
		assert.False(t, f.chapterPerm(ctx, t, f.bob).CanEdit, "他人の権限は変わらない")
	})

	t.Run("所属していない相手には何も見えない", func(t *testing.T) {
		f := setupMaterialPermission(t, sqlDB)
		// bob はワークスペースの一員ではない。公開済みでも読めない。
		got := f.chapterPerm(ctx, t, f.bob)
		assert.False(t, got.CanView)
		assert.False(t, got.CanEdit)
	})
}

func TestMaterialPermission_読むことに付与を要求しない_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupMaterialPermission(t, sqlDB)
	f.member(ctx, t, f.alice)

	// 付与は 1 つも無い。研修を受ける人が教材を開くたびに権限を配らずに済むこと。
	got := f.chapterPerm(ctx, t, f.alice)
	assert.True(t, got.CanView, "公開済みなら一員は誰でも読める")
	assert.False(t, got.CanEdit)
	assert.False(t, f.draftPerm(ctx, t, f.alice).CanView, "下書きは編集できる人にしか見せない")

	// コース側でも同じ規則が効くこと。章だけを見ていると、コースの is_published を
	// 取り違えても気付けない（章は公開・コースは下書き、という食い違いが起こり得る）。
	assert.True(t, f.coursePerm(ctx, t, f.alice).CanView, "公開のコースは読める")
	assert.False(t, f.coursePermOf(ctx, t, f.draftCourse, f.alice).CanView,
		"付与の無い一員に下書きのコースは見えない")

	// 付与を張れば下書きのコースも扱える。
	alice, err := f.kb.EnsureUserPrincipal(ctx, f.ws, f.alice)
	require.NoError(t, err)
	_, err = f.perm.UpsertCourseGrant(ctx, f.ws, f.draftCourse, alice.ID, domain.GrantRoleEditor)
	require.NoError(t, err)
	got = f.coursePermOf(ctx, t, f.draftCourse, f.alice)
	assert.True(t, got.CanView, "editor には下書きのコースも見える")
	assert.True(t, got.CanEdit)
}

func TestMaterialPermission_テナントを跨がない_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupMaterialPermission(t, sqlDB)
	alice := f.member(ctx, t, f.alice)
	_, err := f.kb.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleAdmin)
	require.NoError(t, err)

	t.Run("他ワークスペースから引くと見つからない", func(t *testing.T) {
		// 自分のワークスペースの admin でも、別テナントの箱から同じコースは引けない。
		_, err := f.perm.CourseFactsForUser(ctx, f.otherWS, f.course, f.alice)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		_, err = f.perm.ChapterFactsForUser(ctx, f.otherWS, f.chapter, f.alice)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("存在しない対象は見つからない", func(t *testing.T) {
		_, err := f.perm.CourseFactsForUser(ctx, f.ws, 999999, f.alice)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		_, err = f.perm.ChapterFactsForUser(ctx, f.ws, 999999, f.alice)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("別テナントの主体には張れない", func(t *testing.T) {
		other, err := f.kb.EnsureUserPrincipal(ctx, f.otherWS, f.bob)
		require.NoError(t, err)
		_, err = f.perm.UpsertCourseGrant(ctx, f.ws, f.course, other.ID, domain.GrantRoleEditor)
		assert.Error(t, err, "複合 FK がテナント跨ぎを塞ぐ")
	})
}

func TestMaterialPermission_書き経路_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupMaterialPermission(t, sqlDB)
	alice := f.member(ctx, t, f.alice)

	t.Run("2度張っても行は増えず役割だけ変わる", func(t *testing.T) {
		before, err := f.perm.UpsertCourseGrant(ctx, f.ws, f.course, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		after, err := f.perm.UpsertCourseGrant(ctx, f.ws, f.course, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		assert.Equal(t, domain.GrantRoleViewer, after.Role)
		assert.Equal(t, before.CreatedAt, after.CreatedAt, "作成時刻は動かない")

		rows, err := f.perm.ListCourseGrants(ctx, f.ws, f.course)
		require.NoError(t, err)
		assert.Len(t, rows, 1)
	})

	t.Run("取り消しは冪等", func(t *testing.T) {
		require.NoError(t, f.perm.DeleteCourseGrant(ctx, f.ws, f.course, alice.ID))
		require.NoError(t, f.perm.DeleteCourseGrant(ctx, f.ws, f.course, alice.ID))
		rows, err := f.perm.ListCourseGrants(ctx, f.ws, f.course)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})

	t.Run("一覧はその段に張った行だけを返す", func(t *testing.T) {
		_, err := f.perm.UpsertCourseGrant(ctx, f.ws, f.course, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		// コースの付与は章へ降りるが、章の一覧には出さない
		//（どの段で足したかが分からないと、取り消すべき行を人が選べない）。
		require.True(t, f.chapterPerm(ctx, t, f.alice).CanEdit, "前提: 降りている")
		onChapter, err := f.perm.ListChapterGrants(ctx, f.ws, f.chapter)
		require.NoError(t, err)
		assert.Empty(t, onChapter, "コースの行は章の一覧に含めない")
	})

	t.Run("コースを消すと付与も消える", func(t *testing.T) {
		_, err := f.perm.UpsertCourseGrant(ctx, f.ws, f.course, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		_, err = f.perm.UpsertChapterGrant(ctx, f.ws, f.chapter, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		_, err = f.db.Exec(`DELETE FROM courses WHERE id = $1`, f.course)
		require.NoError(t, err)

		rows, err := f.perm.ListCourseGrants(ctx, f.ws, f.course)
		require.NoError(t, err)
		assert.Empty(t, rows, "コースと一緒に消える")
		chapterRows, err := f.perm.ListChapterGrants(ctx, f.ws, f.chapter)
		require.NoError(t, err)
		assert.Empty(t, chapterRows, "章ごと消えるので、章の付与も残らない")
	})
}
