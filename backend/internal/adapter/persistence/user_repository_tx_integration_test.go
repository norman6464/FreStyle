//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// bootstrapSuperAdminLockKeyForTest は repository が「最初の運営管理者を作る」経路の直列化に
// 使う advisory lock のキー（persistence.bootstrapSuperAdminLockKey と同じ値）。
// 非公開の定数をテストから覗く代わりに同じ値を書き写している。ここがずれると
// 「ロックを外から握れば repository が待つ」ことを確かめられなくなるので、
// キーを変えるときはこちらも合わせる。
const bootstrapSuperAdminLockKeyForTest int64 = 7_419_063

// userTxTables は users まわりの書き込み経路が触るテーブル。
var userTxTables = []string{"users", "user_oidc_identities", "workspaces"}

// userUpdatedAt はユーザーの updated_at を DB から直接読む。
func userUpdatedAt(t *testing.T, db *sql.DB, id uint64) time.Time {
	t.Helper()
	var ts time.Time
	require.NoError(t, db.QueryRow(`SELECT updated_at FROM users WHERE id = $1`, id).Scan(&ts))
	return ts
}

// countActiveSuperAdmins は論理削除されていない super_admin の人数を返す。
func countActiveSuperAdmins(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM users u JOIN roles r ON r.id = u.role_id
		 WHERE r.name = $1 AND u.deleted_at IS NULL`, string(domain.RoleSuperAdmin),
	).Scan(&n))
	return n
}

// TestUserRepositoryBootstrapSuperAdmin_Integration は「最初の運営管理者」を作る経路の
// 直列化を実 PostgreSQL で固定する。
//
// この経路は招待を経ずにユーザーを作れる唯一の免除で、1 人目ができた瞬間に閉じなければ
// ならない。READ COMMITTED では「0 人であること」を確かめてから作るまでの間に別の
// トランザクションが 1 人目を作っても互いの未コミット行が見えないため、判定と INSERT を
// 同じトランザクションに入れるだけでは足りず、経路自体を advisory lock で直列化している。
// ロックがトランザクションの外（別接続）へ出ると直列化は成立しないので、ここでは
// 「同時に来ても 1 人しか作られない」ことを実際に走らせて確かめる。
func TestUserRepositoryBootstrapSuperAdmin_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()

	t.Run("同時に来ても作られる運営管理者は 1 人だけ", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)

		const workers = 8
		start := make(chan struct{})
		created := make([]bool, workers)
		errs := make([]error, workers)
		var wg sync.WaitGroup
		for i := range workers {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-start // 全員を同じ瞬間に走らせる
				u := &domain.User{
					Email: fmt.Sprintf("boot%d@example.com", i),
					Name:  fmt.Sprintf("運営%d", i),
					Role:  domain.RoleSuperAdmin,
				}
				created[i], errs[i] = repo.CreateFirstSuperAdminWithOidcIdentity(
					ctx, u, domain.OidcProviderCognito, fmt.Sprintf("boot-%d", i),
				)
			}(i)
		}
		close(start)
		wg.Wait()

		wins := 0
		for i := range workers {
			require.NoErrorf(t, errs[i], "worker %d がエラーになった", i)
			if created[i] {
				wins++
			}
		}
		require.Equal(t, 1, wins, "created=true を返したのは 1 本だけであるべき")
		require.Equal(t, int64(1), countActiveSuperAdmins(t, sqlDB), "運営管理者は 1 人だけ作られる")
	})

	t.Run("ロックを外から握っているあいだ作成は待たされる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)

		blocker, err := sqlDB.BeginTx(ctx, nil)
		require.NoError(t, err)
		_, err = blocker.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, bootstrapSuperAdminLockKeyForTest)
		require.NoError(t, err)

		done := make(chan error, 1)
		go func() {
			_, createErr := repo.CreateFirstSuperAdminWithOidcIdentity(ctx, &domain.User{
				Email: "blocked@example.com", Name: "待たされる運営", Role: domain.RoleSuperAdmin,
			}, domain.OidcProviderCognito, "boot-blocked")
			done <- createErr
		}()

		select {
		case err := <-done:
			_ = blocker.Rollback()
			t.Fatalf("ロックを握っているあいだに作成が進んでしまった: %v", err)
		case <-time.After(500 * time.Millisecond):
			// 期待どおり待たされている。
		}

		require.NoError(t, blocker.Rollback()) // ロック解放
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("ロック解放後も作成が終わらない")
		}
		require.Equal(t, int64(1), countActiveSuperAdmins(t, sqlDB))
	})

	t.Run("既に運営管理者が居れば作らず created=false", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "first@example.com", Name: "先客", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "first-sub"))

		created, err := repo.CreateFirstSuperAdminWithOidcIdentity(ctx, &domain.User{
			Email: "second@example.com", Name: "2 人目", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "second-sub")
		require.NoError(t, err)
		require.False(t, created)
		require.Equal(t, int64(1), countActiveSuperAdmins(t, sqlDB))

		// 作られていないので identity も残らない。
		var n int64
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(*) FROM user_oidc_identities WHERE subject = $1`, "second-sub",
		).Scan(&n))
		require.Equal(t, int64(0), n)
	})

	t.Run("論理削除された運営管理者は人数に数えない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := &domain.User{Email: "gone@example.com", Name: "退任", Role: domain.RoleSuperAdmin}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "gone-sub"))
		require.NoError(t, repo.SoftDelete(ctx, u.ID))

		created, err := repo.CreateFirstSuperAdminWithOidcIdentity(ctx, &domain.User{
			Email: "again@example.com", Name: "再任", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "again-sub")
		require.NoError(t, err)
		require.True(t, created)
	})

	t.Run("super_admin 以外の role はエラーで何も作らない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		created, err := repo.CreateFirstSuperAdminWithOidcIdentity(ctx, &domain.User{
			Email: "trainee@example.com", Name: "研修生", Role: domain.RoleTrainee,
		}, domain.OidcProviderCognito, "trainee-sub")
		require.Error(t, err)
		require.False(t, created)
		var n int64
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM users`).Scan(&n))
		require.Equal(t, int64(0), n)
	})

	t.Run("identity の失敗で users 行ごと巻き戻る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		// 空 subject は CHECK 違反。users 行だけが残る（＝ログイン不能な孤児）状態を作らない。
		created, err := repo.CreateFirstSuperAdminWithOidcIdentity(ctx, &domain.User{
			Email: "orphan@example.com", Name: "孤児", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "")
		require.Error(t, err)
		require.False(t, created)
		var n int64
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM users`).Scan(&n))
		require.Equal(t, int64(0), n)
	})

	t.Run("所属ワークスペースを渡せばそのまま書かれる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		ws1 := uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)

		ws1Str := ws1.String()
		u := &domain.User{Email: "boot-ws@example.com", Name: "運営", Role: domain.RoleSuperAdmin, WorkspaceID: &ws1Str}
		created, err := repo.CreateFirstSuperAdminWithOidcIdentity(ctx, u, domain.OidcProviderCognito, "boot-ws")
		require.NoError(t, err)
		require.True(t, created)

		require.Equal(t, uuid.NullUUID{UUID: ws1, Valid: true}, tableWorkspaceID(t, sqlDB, "users", u.ID))
	})
}

// TestUserRepositoryWrites_Integration は users の書き込み経路（オフボーディングの芯である
// UpdateActive / SoftDelete を含む）の契約を実 PostgreSQL で固定する。
func TestUserRepositoryWrites_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()

	// newTrainee は研修生を 1 人作って返す。
	newTrainee := func(t *testing.T, email, sub string) *domain.User {
		t.Helper()
		u := &domain.User{Email: email, Name: email, Role: domain.RoleTrainee}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		require.NotZero(t, u.ID, "作成後は採番された id が書き戻る")
		require.False(t, u.CreatedAt.IsZero(), "作成後は created_at が書き戻る")
		require.False(t, u.UpdatedAt.IsZero(), "作成後は updated_at が書き戻る")
		return u
	}

	t.Run("作成直後は有効（is_active=true）", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "active@example.com", "active-1")
		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.True(t, got.IsActive)
	})

	t.Run("UpdateActive(false) は即時に効き FindActiveByEmail から消える", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "off@example.com", "off-1")
		before := userUpdatedAt(t, sqlDB, u.ID)

		require.NoError(t, repo.UpdateActive(ctx, u.ID, false))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.False(t, got.IsActive)
		byEmail, err := repo.FindActiveByEmail(ctx, "off@example.com")
		require.NoError(t, err)
		require.Nil(t, byEmail, "無効化されたユーザーはログイン経路から引けない")
		require.NotEqual(t, before, userUpdatedAt(t, sqlDB, u.ID), "updated_at が進む")

		// 戻せる。
		require.NoError(t, repo.UpdateActive(ctx, u.ID, true))
		byEmail, err = repo.FindActiveByEmail(ctx, "off@example.com")
		require.NoError(t, err)
		require.NotNil(t, byEmail)
	})

	t.Run("UpdateActive は存在しないユーザーで domain.ErrNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		require.ErrorIs(t, repo.UpdateActive(ctx, 999999, false), domain.ErrNotFound)
	})

	t.Run("SoftDelete は identity を解放し、二度目は domain.ErrNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "bye@example.com", "bye-1")

		require.NoError(t, repo.SoftDelete(ctx, u.ID))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Nil(t, got, "論理削除後は引けない")
		bySub, err := repo.FindByCognitoSub(ctx, "bye-1")
		require.NoError(t, err)
		require.Nil(t, bySub)
		var n int64
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1`, u.ID).Scan(&n))
		require.Equal(t, int64(0), n, "identity は解放される")

		require.ErrorIs(t, repo.SoftDelete(ctx, u.ID), domain.ErrNotFound)
		require.ErrorIs(t, repo.SoftDelete(ctx, 999999), domain.ErrNotFound)
	})

	t.Run("UpdateName / UpdateRole は指定列だけを更新する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "rename@example.com", "rename-1")
		before := userUpdatedAt(t, sqlDB, u.ID)

		require.NoError(t, repo.UpdateName(ctx, u.ID, "新しい名前"))
		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, "新しい名前", got.Name)
		require.Equal(t, "rename@example.com", got.Email, "email は触らない")
		require.True(t, got.IsActive, "is_active は触らない")
		require.NotEqual(t, before, userUpdatedAt(t, sqlDB, u.ID))

		require.NoError(t, repo.UpdateRole(ctx, u.ID, domain.RoleCompanyAdmin))
		got, err = repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, domain.RoleCompanyAdmin, got.Role)
		require.Equal(t, domain.RoleIDCompanyAdmin, got.RoleID)
		require.Equal(t, "新しい名前", got.Name, "name は触らない")
		require.True(t, got.IsActive, "is_active は触らない")
	})

	t.Run("ListByWorkspaceID はワークスペースで絞り id 昇順・論理削除を除く", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		ws1 := createWorkspace(t, sqlDB, "list-by-workspace-a")
		ws2 := createWorkspace(t, sqlDB, "list-by-workspace-b")
		a := &domain.User{Email: "m1@example.com", Name: "m1", Role: domain.RoleTrainee, WorkspaceID: &ws1}
		b := &domain.User{Email: "m2@example.com", Name: "m2", Role: domain.RoleTrainee, WorkspaceID: &ws1}
		other := &domain.User{Email: "m3@example.com", Name: "m3", Role: domain.RoleTrainee, WorkspaceID: &ws2}
		for _, u := range []*domain.User{a, b, other} {
			require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, u.Name))
		}

		rows, err := repo.ListByWorkspaceID(ctx, ws1)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.Equal(t, "m1", rows[0].Name)
		require.Equal(t, "m2", rows[1].Name)

		require.NoError(t, repo.SoftDelete(ctx, b.ID))
		rows, err = repo.ListByWorkspaceID(ctx, ws1)
		require.NoError(t, err)
		require.Len(t, rows, 1)

		// 該当なしでも nil ではなく空スライス（JSON が null にならない）。
		empty, err := repo.ListByWorkspaceID(ctx, "0198a000-0000-7000-8000-0000000000ff")
		require.NoError(t, err)
		require.NotNil(t, empty)
		require.Empty(t, empty)

		// 不正な文字列（uuid として解釈できない）も該当なしと同じ扱い。
		invalid, err := repo.ListByWorkspaceID(ctx, "not-a-uuid")
		require.NoError(t, err)
		require.NotNil(t, invalid)
		require.Empty(t, invalid)
	})

	t.Run("CognitoSubjectByUserID は subject を返し、無ければ空文字", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "sub@example.com", "sub-1")

		got, err := repo.CognitoSubjectByUserID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, "sub-1", got)

		got, err = repo.CognitoSubjectByUserID(ctx, 999999)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("EnsureOidcIdentity は他人が持つ subject を奪わない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		owner := newTrainee(t, "owner@example.com", "shared-subject")
		other := newTrainee(t, "other@example.com", "other-subject")

		// 自分の subject を張り直しても冪等に成功する。
		require.NoError(t, repo.EnsureOidcIdentity(ctx, owner.ID, domain.OidcProviderCognito, "shared-subject"))

		// 他人が持つ subject を要求したら黙って成功にせずエラー。
		err := repo.EnsureOidcIdentity(ctx, other.ID, domain.OidcProviderCognito, "shared-subject")
		require.ErrorContains(t, err, "oidc identity conflict")

		// 持ち主は変わっていない。
		got, err := repo.FindByCognitoSub(ctx, "shared-subject")
		require.NoError(t, err)
		require.Equal(t, owner.ID, got.ID)
	})
}
