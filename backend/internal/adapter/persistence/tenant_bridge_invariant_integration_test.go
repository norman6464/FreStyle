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

// このファイルは「テナントの正本を companies から workspaces へ移す」移行のうち、
// 両方の表現が同時に生きているあいだ守り続けなければならない不変条件を固定する。
//
// 移行はローリングデプロイを跨ぐため、旧タスク（companies を読む）と新タスク（workspaces を読む）が
// 同時に走る瞬間がある。その間どちらを読んでも同じ答えが返らないと、会社をまたいだ取り違えや
// 「所属しているのに見えない」が起きる。つまり守るべきは個々の書き込み経路ではなく、
// **どの時点で DB を覗いても両側が同じ事実を語っている**という状態そのもの。
//
// 経路ごとの振る舞い（招待からの作成・付け替え・設定更新）は tenant_bridge_integration_test.go が
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

	// (2) 所属は両側で同じことを言っている。会社に属するユーザーの workspace_id は
	//     必ずその会社のワークスペース。未所属なら両側とも空。
	require.Zero(t, tenantCount(t, db,
		`SELECT count(*) FROM users u JOIN companies c ON u.company_id = c.id
		 WHERE u.workspace_id IS DISTINCT FROM c.workspace_id`),
		"所属会社のワークスペースを指していないユーザーがいる")
	require.Zero(t, tenantCount(t, db,
		`SELECT count(*) FROM users WHERE company_id IS NULL AND workspace_id IS NOT NULL`),
		"未所属なのにワークスペースを指しているユーザーがいる")

	// (3) テナント設定は移行中 companies が正本。写し先がずれたまま残っていない。
	require.Zero(t, tenantCount(t, db,
		`SELECT count(*) FROM workspaces w JOIN companies c ON c.workspace_id = w.id
		 WHERE w.is_active IS DISTINCT FROM c.is_active`),
		"会社の設定がワークスペースへ写っていない")
}

// insertTraineeIn は会社に属する研修生を 1 人作る（作成経路の二重書きもここで通る）。
func insertTraineeIn(ctx context.Context, t *testing.T, db *sql.DB, companyID uint64, sub string) uint64 {
	t.Helper()
	repo := persistence.NewUserRepository(db)
	require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
		Email: sub + "@example.com", Name: sub, Role: domain.RoleTrainee, CompanyID: &companyID,
	}, domain.OidcProviderCognito, sub))
	got, err := repo.FindByCognitoSub(ctx, sub)
	require.NoError(t, err)
	return got.ID
}

// TestTenantBridgeInvariants_Integration は段 1（companies と workspaces の両方が
// 常に正しい値を持つ）の不変条件を実 PostgreSQL で固定する。
func TestTenantBridgeInvariants_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("再実行しても DB の状態が 1 列も変わらない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", false)
		// バックフィル前に作られたユーザー（workspace_id がまだ空の状態）を混ぜる。
		insertTraineeIn(ctx, t, sqlDB, 1, "sub-a")
		insertTraineeIn(ctx, t, sqlDB, 2, "sub-b")
		_, err := sqlDB.Exec(`UPDATE users SET workspace_id = NULL`)
		require.NoError(t, err)

		runStartupBackfill(ctx, t, sqlDB)
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

		// 起動後に新しい会社と、その所属ユーザーが増える。
		insertCompany(t, sqlDB, 2, "後から増えた会社", true)
		newUser := insertTraineeIn(ctx, t, sqlDB, 2, "sub-late")

		runStartupBackfill(ctx, t, sqlDB)

		require.Equal(t, ws1, companyWorkspaceID(t, sqlDB, 1), "既存会社の紐付け先は動かない")
		require.Equal(t, before.workspaces[0], captureTenantSnapshot(t, sqlDB).workspaces[0],
			"既存ワークスペースの行は 1 列も変わらない")
		ws2 := companyWorkspaceID(t, sqlDB, 2)
		require.True(t, ws2.Valid, "後から増えた会社にもワークスペースができる")
		require.NotEqual(t, ws1.UUID, ws2.UUID)
		require.Equal(t, ws2, userWorkspaceID(t, sqlDB, newUser))
		requireTenantInvariants(t, sqlDB)
	})

	t.Run("写し先がずれても次の起動で companies に合わせて直る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", true)
		userA := insertTraineeIn(ctx, t, sqlDB, 1, "sub-drift-a")
		userB := insertTraineeIn(ctx, t, sqlDB, 2, "sub-drift-b")
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)
		ws2 := companyWorkspaceID(t, sqlDB, 2)

		// 旧タスクの書き込みや手作業で写し側だけがずれた状態を作る。
		// 正本は companies なので、次の起動で companies の値に戻らなければならない。
		_, err := sqlDB.Exec(
			`UPDATE workspaces SET is_active = false WHERE id = $1`,
			ws1.UUID,
		)
		require.NoError(t, err)
		_, err = sqlDB.Exec(
			`UPDATE workspaces SET is_active = NULL WHERE id = $1`,
			ws2.UUID,
		)
		require.NoError(t, err)
		// 他社のワークスペースを指してしまったユーザーと、写し漏れたユーザー。
		_, err = sqlDB.Exec(`UPDATE users SET workspace_id = $1 WHERE id = $2`, ws2.UUID, userA)
		require.NoError(t, err)
		_, err = sqlDB.Exec(`UPDATE users SET workspace_id = NULL WHERE id = $1`, userB)
		require.NoError(t, err)

		runStartupBackfill(ctx, t, sqlDB)

		require.Equal(t, ws1, userWorkspaceID(t, sqlDB, userA), "他社を指していたユーザーが正しい所属へ戻る")
		require.Equal(t, ws2, userWorkspaceID(t, sqlDB, userB), "写し漏れたユーザーが埋まる")
		requireTenantInvariants(t, sqlDB)
	})

	t.Run("未所属ユーザーはワークスペースへ流し込まれない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "唯一の会社", true)
		root := insertTraineeIn(ctx, t, sqlDB, 1, "sub-root")
		// 会社から外れた（未所属になった）ユーザー。既定のテナントへ寄せてはいけない。
		_, err := sqlDB.Exec(
			`UPDATE users SET company_id = NULL, workspace_id = NULL WHERE id = $1`, root,
		)
		require.NoError(t, err)

		runStartupBackfill(ctx, t, sqlDB)

		require.False(t, userWorkspaceID(t, sqlDB, root).Valid, "未所属は NULL のまま")
		requireTenantInvariants(t, sqlDB)
	})
}
