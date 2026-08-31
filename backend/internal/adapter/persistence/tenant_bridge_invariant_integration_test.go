//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// このファイルは会社（companies）とワークスペース（workspaces）の 1:1 紐付けについて、
// 起動時バックフィルが何度走っても守り続けなければならない不変条件を固定する。
//
// 会社は「契約している組織」、ワークスペースは「テナントとしての入れ物」で、片方だけが
// 増える・相乗りする・余るといった崩れ方をすると、テナント境界そのものが曖昧になる。
// つまり守るべきは個々の書き込み経路ではなく、**どの時点で DB を覗いても紐付けが
// 1:1 のまま**という状態そのもの。
//
// 経路ごとの振る舞い（所属の書き込み・付け替え・設定更新）は tenant_bridge_integration_test.go が
// 見ている。ここは経路を問わず「起動時のバックフィルを流し終えた後の DB の形」だけを見る。

// tenantSnapshot はバックフィル後の DB の状態を丸ごと写し取ったもの。
// 冪等性は「行が増えない」だけでは足りない（値の書き換えや updated_at の空打ちを見逃す）。
// 会社・ワークスペース・ユーザーの全列を突き合わせられるよう文字列に畳んで持つ。
type tenantSnapshot struct {
	workspaces []string
	companies  []string
	users      []string
}

// captureTenantSnapshot は会社↔ワークスペースに関わる全行を決定的な順序で読み出す。
//
// 列を手で並べず to_jsonb で行ごと写すのは、列挙から漏れた列を書き換える実装を
// 見逃さないため。手書きの列リストは書いた時点の列しか守らないので、あとから
// 足された列は誰にも見られないまま素通りする。updated_at も当然含まれる —
// 値が同じでも UPDATE を空打ちしていれば時刻が動くので、「一致していれば 0 件更新」
// という実装の約束をここで検証できる。
//
// jsonb はキーを正規化した順で持つため、列の追加や順序変更があっても比較は決定的。
func captureTenantSnapshot(t *testing.T, db *sql.DB) tenantSnapshot {
	t.Helper()
	return tenantSnapshot{
		workspaces: tenantRowStrings(t, db,
			`SELECT to_jsonb(w)::text FROM workspaces AS w ORDER BY id`),
		companies: tenantRowStrings(t, db,
			`SELECT to_jsonb(c)::text FROM companies AS c ORDER BY id`),
		users: tenantRowStrings(t, db,
			`SELECT to_jsonb(u)::text FROM users AS u ORDER BY id`),
	}
}

// tenantRowStrings は 1 列の文字列を全行読み出す。
func tenantRowStrings(t *testing.T, db *sql.DB, query string) []string {
	t.Helper()
	rows, err := db.Query(query)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()
	out := make([]string, 0)
	for rows.Next() {
		var v string
		require.NoError(t, rows.Scan(&v))
		out = append(out, v)
	}
	require.NoError(t, rows.Err())
	return out
}

// tenantCount は 1 行 1 列の件数クエリを読む。
func tenantCount(t *testing.T, db *sql.DB, query string, args ...any) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.QueryRow(query, args...).Scan(&n))
	return n
}

// requireTenantInvariants は段 1 のあいだ常に成り立っていなければならない条件を検査する。
// 個別のシナリオではなく「DB 全体の形」を見るので、どのテストの末尾に置いても意味がある。
func requireTenantInvariants(t *testing.T, db *sql.DB) {
	t.Helper()

	// (1) 会社とワークスペースは 1:1。取りこぼし（会社にワークスペースが無い）も、
	//     相乗り（2 社が同じワークスペースを指す）も、余り（誰も指していないワークスペース）も無い。
	require.Zero(t, tenantCount(t, db, `SELECT count(*) FROM companies WHERE workspace_id IS NULL`),
		"ワークスペースが割り当たっていない会社が残っている")
	require.Zero(t, tenantCount(t, db,
		`SELECT count(*) FROM (
		   SELECT workspace_id FROM companies GROUP BY workspace_id HAVING count(*) > 1
		 ) dup`),
		"複数の会社が同じワークスペースを指している")
	require.Zero(t, tenantCount(t, db,
		`SELECT count(*) FROM workspaces w
		 WHERE NOT EXISTS (SELECT 1 FROM companies c WHERE c.workspace_id = w.id)`),
		"どの会社からも指されていないワークスペースが残っている")
	require.Equal(t,
		tenantCount(t, db, `SELECT count(*) FROM companies`),
		tenantCount(t, db, `SELECT count(*) FROM workspaces`),
		"会社の数とワークスペースの数が一致しない")

	// (2) 所属参照は宙に浮かない。workspace_id が入っているユーザーは必ず実在する
	//     ワークスペースを指す（未所属は NULL で表し、既定のテナントへは寄せない）。
	require.Zero(t, tenantCount(t, db,
		`SELECT count(*) FROM users u
		 WHERE u.workspace_id IS NOT NULL
		   AND NOT EXISTS (SELECT 1 FROM workspaces w WHERE w.id = u.workspace_id)`),
		"実在しないワークスペースを指しているユーザーがいる")
}

// insertTraineeIn はワークスペースに属する研修生を 1 人作る。
func insertTraineeIn(ctx context.Context, t *testing.T, db *sql.DB, workspaceID string, sub string) uint64 {
	t.Helper()
	repo := persistence.NewUserRepository(db)
	require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
		Email: sub + "@example.com", Name: sub, Role: domain.RoleTrainee, WorkspaceID: &workspaceID,
	}, domain.OidcProviderCognito, sub))
	got, err := repo.FindByCognitoSub(ctx, sub)
	require.NoError(t, err)
	return got.ID
}

// TestTenantBridgeInvariants_Integration は会社↔ワークスペースの 1:1 紐付けが
// 常に保たれることを実 PostgreSQL で固定する。
func TestTenantBridgeInvariants_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("再実行しても DB の状態が 1 列も変わらない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", false)

		runStartupBackfill(ctx, t, sqlDB)
		// 所属済みのユーザーと未所属のユーザーを混ぜる（再実行で書き換えられないこと）。
		insertTraineeIn(ctx, t, sqlDB, companyWorkspaceID(t, sqlDB, 1).UUID.String(), "sub-a")
		insertTraineeIn(ctx, t, sqlDB, companyWorkspaceID(t, sqlDB, 2).UUID.String(), "sub-b")
		require.NoError(t, persistence.NewUserRepository(sqlDB).CreateWithOidcIdentity(ctx, &domain.User{
			Email: "sub-root@example.com", Name: "sub-root", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "sub-root"))

		first := captureTenantSnapshot(t, sqlDB)
		requireTenantInvariants(t, sqlDB)

		// 起動 2 回目・3 回目。行数だけでなく全列（updated_at 含む）が同一であること。
		// 値が合っていても UPDATE を空打ちしていれば updated_at が動いて落ちる。
		for i := 2; i <= 3; i++ {
			runStartupBackfill(ctx, t, sqlDB)
			require.Equal(t, first, captureTenantSnapshot(t, sqlDB), "%d 回目の実行で状態が変わった", i)
			requireTenantInvariants(t, sqlDB)
		}
	})

	t.Run("既にワークスペースが紐づいた会社は作り直しも指し替えもしない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "手で紐付けた会社", true)
		insertCompany(t, sqlDB, 2, "まだ紐付いていない会社", true)

		// 運用や先行移行で既に用意されたワークスペース。自動採番とは違う slug / name を持つ。
		manual := uuid.New()
		_, err := sqlDB.Exec(
			`INSERT INTO workspaces (id, slug, name, is_active)
			 VALUES ($1, 'handmade-workspace', '手で付けた名前', true)`, manual,
		)
		require.NoError(t, err)
		_, err = sqlDB.Exec(`UPDATE companies SET workspace_id = $1 WHERE id = 1`, manual)
		require.NoError(t, err)

		runStartupBackfill(ctx, t, sqlDB)

		require.Equal(t, manual, companyWorkspaceID(t, sqlDB, 1).UUID, "紐付け先を書き換えてはいけない")
		var slug, name string
		require.NoError(t, sqlDB.QueryRow(`SELECT slug, name FROM workspaces WHERE id = $1`, manual).
			Scan(&slug, &name))
		require.Equal(t, "handmade-workspace", slug, "既存ワークスペースの slug を採番し直してはいけない")
		require.Equal(t, "手で付けた名前", name, "既存ワークスペースの表示名を会社名で上書きしてはいけない")

		// 紐付いていなかった会社の分だけが増える（手持ちの分を作り直して 3 行にならない）。
		require.Equal(t, int64(2), tenantCount(t, sqlDB, `SELECT count(*) FROM workspaces`))
		requireTenantInvariants(t, sqlDB)
	})

	t.Run("バックフィル後に会社が増えても次の起動で追随する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "先にあった会社", true)
		runStartupBackfill(ctx, t, sqlDB)
		before := captureTenantSnapshot(t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)

		// 起動後に新しい会社が増える。
		insertCompany(t, sqlDB, 2, "後から増えた会社", true)

		runStartupBackfill(ctx, t, sqlDB)

		require.Equal(t, ws1, companyWorkspaceID(t, sqlDB, 1), "既存会社の紐付け先は動かない")
		require.Equal(t, before.workspaces[0], captureTenantSnapshot(t, sqlDB).workspaces[0],
			"既存ワークスペースの行は 1 列も変わらない")
		ws2 := companyWorkspaceID(t, sqlDB, 2)
		require.True(t, ws2.Valid, "後から増えた会社にもワークスペースができる")
		require.NotEqual(t, ws1.UUID, ws2.UUID)

		// 新しいワークスペースへ所属させたユーザーは、次の起動でも書き換えられない。
		newUser := insertTraineeIn(ctx, t, sqlDB, ws2.UUID.String(), "sub-late")
		runStartupBackfill(ctx, t, sqlDB)
		require.Equal(t, ws2, userWorkspaceID(t, sqlDB, newUser))
		requireTenantInvariants(t, sqlDB)
	})

	t.Run("停止したワークスペースは次の起動でも停止したまま", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)

		// 停止の正本は workspaces.is_active へ移った。会社は有効のまま片方だけ止める。
		// バックフィルが会社から写し直すと、止めたはずのワークスペースが起動のたびに
		// 有効へ戻り、停止そのものが効かなくなる。
		_, err := sqlDB.Exec(
			`UPDATE workspaces SET is_active = false WHERE id = $1`,
			ws1.UUID,
		)
		require.NoError(t, err)

		runStartupBackfill(ctx, t, sqlDB)

		var active sql.NullBool
		require.NoError(t, sqlDB.QueryRow(
			`SELECT is_active FROM workspaces WHERE id = $1`, ws1.UUID,
		).Scan(&active))
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, active, "会社の値で上書きされない")
		requireTenantInvariants(t, sqlDB)
	})

	t.Run("未所属ユーザーはワークスペースへ流し込まれない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "唯一の会社", true)
		runStartupBackfill(ctx, t, sqlDB)
		root := insertTraineeIn(ctx, t, sqlDB, companyWorkspaceID(t, sqlDB, 1).UUID.String(), "sub-root")
		// ワークスペースから外れた（未所属になった）ユーザー。既定のテナントへ寄せてはいけない。
		_, err := sqlDB.Exec(
			`UPDATE users SET workspace_id = NULL WHERE id = $1`, root,
		)
		require.NoError(t, err)

		runStartupBackfill(ctx, t, sqlDB)

		require.False(t, userWorkspaceID(t, sqlDB, root).Valid, "未所属は NULL のまま")
		requireTenantInvariants(t, sqlDB)
	})
}
