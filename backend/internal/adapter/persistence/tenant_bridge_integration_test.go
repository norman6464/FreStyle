//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// tenantBridgeTables は会社↔ワークスペースの橋渡しが触るテーブル（TRUNCATE 対象）。
// workspaces を参照する FK があるため CASCADE で companies / users も一緒に消えるが、
// 依存を読み手に見せるために全部並べる。
var tenantBridgeTables = []string{"users", "user_oidc_identities", "companies", "workspaces"}

// workspaceSlugPattern は自動採番した slug の形（ws- + UUID の 32 桁 hex）。
var workspaceSlugPattern = regexp.MustCompile(`^ws-[0-9a-f]{32}$`)

// insertCompany は会社を 1 件作る（id を明示して固定する）。
func insertCompany(t *testing.T, db *sql.DB, id int64, name string, active bool) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO companies (id, name, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())`, id, name, active,
	)
	require.NoError(t, err)
}

// companyWorkspaceID は会社に紐づいたワークスペース ID を返す（未紐付けなら Valid=false）。
func companyWorkspaceID(t *testing.T, db *sql.DB, companyID int64) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.QueryRow(`SELECT workspace_id FROM companies WHERE id = $1`, companyID).Scan(&got))
	return got
}

// userWorkspaceID はユーザーの workspace_id を返す（未設定なら Valid=false）。
func userWorkspaceID(t *testing.T, db *sql.DB, userID uint64) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.QueryRow(`SELECT workspace_id FROM users WHERE id = $1`, userID).Scan(&got))
	return got
}

// runStartupBackfill は起動時に走る会社→ワークスペースのバックフィルを 1 回分流す。
// スキーマ（列 / FK）は testsupport.OpenTestDB が既に適用済み。
func runStartupBackfill(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()
	require.NoError(t, database.BackfillWorkspacesFromCompanies(ctx, db))
}

// TestTenantBridgeBackfill_Integration は「会社を workspaces へ映す」バックフィルを実 Postgres で固定する。
// バックフィルが受け持つのは companies↔workspaces の 1:1 紐付けと設定の反映だけで、
// 利用者や業務テーブルの所属参照（workspace_id）は書き込み経路が直接決めて渡す。
func TestTenantBridgeBackfill_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("会社ごとにワークスペースが 1 つできる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "株式会社ふれすたいる", true)
		insertCompany(t, sqlDB, 2, "Second Corp", false)

		runStartupBackfill(ctx, t, sqlDB)

		var count int64
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM workspaces`).Scan(&count))
		require.Equal(t, int64(2), count, "会社 1 件につきワークスペースは 1 つ")

		ws1 := companyWorkspaceID(t, sqlDB, 1)
		ws2 := companyWorkspaceID(t, sqlDB, 2)
		require.True(t, ws1.Valid)
		require.True(t, ws2.Valid)
		require.NotEqual(t, ws1.UUID, ws2.UUID, "会社ごとに別のワークスペース")

		// 会社名と設定が写っていること（設定は移行期間中 companies が正本）。
		var name string
		var active sql.NullBool
		require.NoError(t, sqlDB.QueryRow(
			`SELECT name, is_active FROM workspaces WHERE id = $1`, ws1.UUID,
		).Scan(&name, &active))
		require.Equal(t, "株式会社ふれすたいる", name)
		require.Equal(t, sql.NullBool{Bool: true, Valid: true}, active)
		require.NoError(t, sqlDB.QueryRow(
			`SELECT is_active FROM workspaces WHERE id = $1`, ws2.UUID,
		).Scan(&active))
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, active)
	})

	t.Run("何度流してもワークスペースが増えず ID もぶれない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", true)

		runStartupBackfill(ctx, t, sqlDB)
		first1 := companyWorkspaceID(t, sqlDB, 1)
		first2 := companyWorkspaceID(t, sqlDB, 2)

		for i := 0; i < 2; i++ { // 起動 2 回目・3 回目
			runStartupBackfill(ctx, t, sqlDB)
			var count int64
			require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM workspaces`).Scan(&count))
			require.Equal(t, int64(2), count, "再実行でワークスペースが増えない")
			require.Equal(t, first1, companyWorkspaceID(t, sqlDB, 1), "再実行で紐付け先が変わらない")
			require.Equal(t, first2, companyWorkspaceID(t, sqlDB, 2))
		}
	})

	t.Run("slug はグローバル一意で長さ制約を満たす", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "同名の会社", true)
		insertCompany(t, sqlDB, 2, "同名の会社", true) // companies.name に一意制約は無い

		runStartupBackfill(ctx, t, sqlDB)

		rows, err := sqlDB.Query(`SELECT slug FROM workspaces ORDER BY slug`)
		require.NoError(t, err)
		defer func() { require.NoError(t, rows.Close()) }()
		slugs := make([]string, 0, 2)
		for rows.Next() {
			var slug string
			require.NoError(t, rows.Scan(&slug))
			require.Regexp(t, workspaceSlugPattern, slug)
			require.LessOrEqual(t, len(slug), 64)
			slugs = append(slugs, slug)
		}
		require.NoError(t, rows.Err())
		require.Len(t, slugs, 2)
		require.NotEqual(t, slugs[0], slugs[1], "会社名が同じでも slug は衝突しない")
	})

	t.Run("存在しないワークスペースは指せない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社", true)

		_, err := sqlDB.Exec(`UPDATE companies SET workspace_id = $1 WHERE id = 1`, uuid.New())
		require.Error(t, err, "実在しないワークスペースは FK が弾く")
	})
}

// TestUserWorkspaceWrite_Integration は users.workspace_id が「渡された値そのまま」で
// 書かれることを固定する。所属参照の解決は呼び出し側（usecase）の責務で、
// リポジトリ側が会社から引き直したりはしない。
func TestUserWorkspaceWrite_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("渡した所属ワークスペースがそのまま書かれる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)
		ws2 := companyWorkspaceID(t, sqlDB, 2)
		ws1Str, ws2Str := ws1.UUID.String(), ws2.UUID.String()

		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "a@example.com", Name: "A", Role: domain.RoleTrainee, WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-a"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "b@example.com", Name: "B", Role: domain.RoleTrainee, WorkspaceID: &ws2Str,
		}, domain.OidcProviderCognito, "sub-b"))

		userA, err := repo.FindByCognitoSub(ctx, "sub-a")
		require.NoError(t, err)
		userB, err := repo.FindByCognitoSub(ctx, "sub-b")
		require.NoError(t, err)

		require.Equal(t, ws1, userWorkspaceID(t, sqlDB, userA.ID))
		require.Equal(t, ws2, userWorkspaceID(t, sqlDB, userB.ID))
		// 読み出しも同じ所属を返す。
		require.Equal(t, ws1Str, *userA.WorkspaceID)
		require.Equal(t, ws2Str, *userB.WorkspaceID)
	})

	t.Run("未所属のまま作られたユーザーは所属先ワークスペースが無い", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "運営", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "sub-root"))

		got, err := repo.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)
		require.Nil(t, got.WorkspaceID)
		require.False(t, tableWorkspaceID(t, sqlDB, "users", got.ID).Valid)
	})

	t.Run("所属の付け替えでワークスペースが入れ替わる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)
		ws2 := companyWorkspaceID(t, sqlDB, 2)
		ws1Str, ws2Str := ws1.UUID.String(), ws2.UUID.String()

		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "move@example.com", Name: "異動", Role: domain.RoleTrainee, WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-move"))
		got, err := repo.FindByCognitoSub(ctx, "sub-move")
		require.NoError(t, err)
		require.Equal(t, ws1, tableWorkspaceID(t, sqlDB, "users", got.ID))

		require.NoError(t, repo.UpdateWorkspaceID(ctx, got.ID, &ws2Str))

		moved, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.Equal(t, ws2Str, *moved.WorkspaceID)
		require.Equal(t, ws2, tableWorkspaceID(t, sqlDB, "users", got.ID))
	})

	t.Run("nil を渡すと未所属へ戻る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1Str := companyWorkspaceID(t, sqlDB, 1).UUID.String()

		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "leave@example.com", Name: "退所", Role: domain.RoleTrainee, WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-leave"))
		got, err := repo.FindByCognitoSub(ctx, "sub-leave")
		require.NoError(t, err)

		require.NoError(t, repo.UpdateWorkspaceID(ctx, got.ID, nil))

		left, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.Nil(t, left.WorkspaceID)
		require.False(t, userWorkspaceID(t, sqlDB, got.ID).Valid)
	})
}

// businessTablesWithWorkspace は所属参照として workspace_id を持つ業務テーブル。
//
// company_exercises だけは下の書き込みテストで検証していない。アプリから書く経路が
// 無いため（repository も sqlc のクエリも無く、schema と移行の SQL にしか出てこない）。
// TRUNCATE の対象には要るのでリストには残す。書き込む repository を足したら、他の表と
// 同じく「渡した workspace_id がそのまま入る」ことをここで固定すること。
var businessTablesWithWorkspace = []string{
	"courses", "course_chapters", "company_exercises", "invitations", "rich_documents",
}

// businessTableTruncateTables はこの節のテストが TRUNCATE する対象。
var businessTableTruncateTables = append(
	append([]string{}, tenantBridgeTables...),
	businessTablesWithWorkspace...,
)

// tableWorkspaceID は業務テーブル 1 行の workspace_id を返す。table は本ファイル内の
// ハードコードされた定数のみを渡す（外部入力を SQL に組み込まない）。
func tableWorkspaceID(t *testing.T, db *sql.DB, table string, id any) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.QueryRow(fmt.Sprintf(`SELECT workspace_id FROM %s WHERE id = $1`, table), id).Scan(&got))
	return got
}

// TestBusinessTableWorkspaceWrite_Integration は業務テーブルの所属参照（workspace_id）が
// 呼び出し側から渡された値そのままで書かれることを実 PostgreSQL で固定する。
// リポジトリ側が会社から引き直していないことを、会社と対応しないワークスペースを
// 渡して見分ける。
func TestBusinessTableWorkspaceWrite_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("新規作成の直後に渡した workspace_id が入る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, businessTableTruncateTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)
		require.True(t, ws1.Valid)
		ws2 := companyWorkspaceID(t, sqlDB, 2)
		require.True(t, ws2.Valid)
		require.NotEqual(t, ws1.UUID, ws2.UUID)
		// 作成者は会社 A（ws1）に所属させたうえで、行の所属先には会社 B（ws2）を渡す。
		// 作成者の所属から引き直していれば ws1 になってしまうので、これで見分けられる。
		ws1Str, ws2Str := ws1.UUID.String(), ws2.UUID.String()

		users := persistence.NewUserRepository(sqlDB)
		require.NoError(t, users.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "author@example.com", Name: "作成者", Role: domain.RoleCompanyAdmin, WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-author"))
		author, err := users.FindByCognitoSub(ctx, "sub-author")
		require.NoError(t, err)

		courses := persistence.NewCourseRepository(sqlDB)
		course := &domain.Course{WorkspaceID: &ws2Str, CreatedByUserID: author.ID, Title: "コース", Language: "go"}
		require.NoError(t, courses.Create(ctx, course))
		require.Equal(t, ws2, tableWorkspaceID(t, sqlDB, "courses", course.ID), "InsertCourse は渡された workspace_id を書く")

		materials := persistence.NewTeachingMaterialRepository(sqlDB)
		chapter := &domain.TeachingMaterial{WorkspaceID: &ws2Str, CourseID: course.ID, CreatedByUserID: author.ID, Title: "第1章"}
		require.NoError(t, materials.Create(ctx, chapter))
		require.Equal(t, ws2, tableWorkspaceID(t, sqlDB, "course_chapters", chapter.ID), "InsertChapter は渡された workspace_id を書く")

		invitations := persistence.NewAdminInvitationRepository(sqlDB)
		inv := &domain.AdminInvitation{
			WorkspaceID: &ws2Str, Email: "invitee@example.com", Role: domain.RoleTrainee,
			Status: domain.InvitationStatusPending, ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, invitations.Create(ctx, inv))
		require.Equal(t, ws2, tableWorkspaceID(t, sqlDB, "invitations", inv.ID), "InsertInvitation は渡された workspace_id を書く")

		richDocs := persistence.NewRichDocumentRepository(sqlDB)
		doc := &domain.RichDocument{
			OwnerID: author.ID, WorkspaceID: &ws2Str, Kind: domain.DocumentKindNote,
			Title: "メモ", Doc: `{"type":"doc","content":[]}`,
		}
		require.NoError(t, richDocs.Create(ctx, doc))
		require.Equal(t, ws2, tableWorkspaceID(t, sqlDB, "rich_documents", doc.ID), "InsertRichDocument は渡された workspace_id を書く")
	})

	t.Run("所属を渡さない rich_documents は workspace_id も NULL のまま", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, businessTableTruncateTables...)
		users := persistence.NewUserRepository(sqlDB)
		require.NoError(t, users.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "運営", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "sub-root"))
		root, err := users.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)

		richDocs := persistence.NewRichDocumentRepository(sqlDB)
		doc := &domain.RichDocument{
			OwnerID: root.ID, WorkspaceID: nil, Kind: domain.DocumentKindNote,
			Title: "運営メモ", Doc: `{"type":"doc","content":[]}`,
		}
		require.NoError(t, richDocs.Create(ctx, doc))
		require.False(t, tableWorkspaceID(t, sqlDB, "rich_documents", doc.ID).Valid)
	})
}
