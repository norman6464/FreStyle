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
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
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
// 読み取りは何も変えていない段なので、ここで確かめるのは新しい列の中身だけ。
func TestTenantBridgeBackfill_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("会社ごとにワークスペースが 1 つでき、所属ユーザーへ写る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "株式会社ふれすたいる", true)
		insertCompany(t, sqlDB, 2, "Second Corp", false)

		repo := persistence.NewUserRepository(sqlDB)
		c1, c2 := uint64(1), uint64(2)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "a@example.com", Name: "A", Role: domain.RoleTrainee, CompanyID: &c1,
		}, domain.OidcProviderCognito, "sub-a"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "b@example.com", Name: "B", Role: domain.RoleTrainee, CompanyID: &c2,
		}, domain.OidcProviderCognito, "sub-b"))
		// 未所属（company_id IS NULL）。運営管理者のようにどの会社にも属さない利用者。
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "Root", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "sub-root"))

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

		userA, err := repo.FindByCognitoSub(ctx, "sub-a")
		require.NoError(t, err)
		userB, err := repo.FindByCognitoSub(ctx, "sub-b")
		require.NoError(t, err)
		userRoot, err := repo.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)

		require.Equal(t, ws1, userWorkspaceID(t, sqlDB, userA.ID))
		require.Equal(t, ws2, userWorkspaceID(t, sqlDB, userB.ID))
		require.False(t, userWorkspaceID(t, sqlDB, userRoot.ID).Valid, "未所属ユーザーは NULL のまま")

		// 読み取りは company_id のまま。バックフィルで見え方が変わっていないこと。
		require.Equal(t, uint64(1), *userA.CompanyID)
		require.Nil(t, userRoot.CompanyID)
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
		require.Error(t, err, "FK が無いまま company_id の轍を踏まない")
	})
}

// TestTenantBridgeDualWrite_Integration は所属の解決（FindUserCompanyWorkspaceID）が
// users.company_id → companies.workspace_id の JOIN で正しく求まることを固定する。
// users.workspace_id への書き込みはもう無い（都度 JOIN で求めるため）。
func TestTenantBridgeDualWrite_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	permissions := persistence.NewKnowledgeBasePermissionRepository(sqlDB)

	t.Run("招待からのユーザー作成で所属先ワークスペースが求まる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)

		repo := persistence.NewUserRepository(sqlDB)
		cid := uint64(1)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "new@example.com", Name: "新入社員", Role: domain.RoleTrainee, CompanyID: &cid,
		}, domain.OidcProviderCognito, "sub-new"))

		got, err := repo.FindByCognitoSub(ctx, "sub-new")
		require.NoError(t, err)
		require.Equal(t, uint64(1), *got.CompanyID)
		wsID, err := permissions.FindUserCompanyWorkspaceID(ctx, got.ID)
		require.NoError(t, err)
		require.Equal(t, ws1.UUID.String(), wsID)
	})

	t.Run("未所属のまま作られたユーザーは所属先ワークスペースが無い", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "運営", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "sub-root"))

		got, err := repo.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)
		require.Nil(t, got.CompanyID)
		_, err = permissions.FindUserCompanyWorkspaceID(ctx, got.ID)
		require.ErrorIs(t, err, repository.ErrWorkspaceNotFound)
	})

	t.Run("会社の付け替えで所属先ワークスペースも追随する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		insertCompany(t, sqlDB, 2, "会社 B", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)
		ws2 := companyWorkspaceID(t, sqlDB, 2)

		repo := persistence.NewUserRepository(sqlDB)
		cid := uint64(1)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "move@example.com", Name: "異動", Role: domain.RoleTrainee, CompanyID: &cid,
		}, domain.OidcProviderCognito, "sub-move"))
		got, err := repo.FindByCognitoSub(ctx, "sub-move")
		require.NoError(t, err)
		wsID, err := permissions.FindUserCompanyWorkspaceID(ctx, got.ID)
		require.NoError(t, err)
		require.Equal(t, ws1.UUID.String(), wsID)

		require.NoError(t, repo.UpdateCompanyID(ctx, got.ID, 2))

		moved, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.Equal(t, uint64(2), *moved.CompanyID)
		wsID, err = permissions.FindUserCompanyWorkspaceID(ctx, got.ID)
		require.NoError(t, err)
		require.Equal(t, ws2.UUID.String(), wsID)
	})

	t.Run("会社設定の更新がワークスペースへ写る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)

		companies := persistence.NewCompanyRepository(sqlDB)
		require.NoError(t, companies.UpdateActive(ctx, 1, false))

		var active sql.NullBool
		require.NoError(t, sqlDB.QueryRow(
			`SELECT is_active FROM workspaces WHERE id = $1`, ws1.UUID,
		).Scan(&active))
		require.Equal(t, sql.NullBool{Bool: false, Valid: true}, active)

		// 読み取り（companies）は従来どおり。
		company, err := companies.FindByID(ctx, 1)
		require.NoError(t, err)
		require.False(t, company.IsActive)
	})

	t.Run("存在しない会社の有効/無効更新は not found のまま", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		companies := persistence.NewCompanyRepository(sqlDB)
		require.ErrorIs(t, companies.UpdateActive(ctx, 999, false), domain.ErrNotFound)
	})

	t.Run("ワークスペース未紐付けの会社でも設定更新は成功する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)
		insertCompany(t, sqlDB, 1, "未紐付けの会社", true) // バックフィル前
		companies := persistence.NewCompanyRepository(sqlDB)
		require.NoError(t, companies.UpdateActive(ctx, 1, false))
		require.False(t, companyWorkspaceID(t, sqlDB, 1).Valid)
	})
}

// businessTablesWithWorkspaceMirror は FRESTYLE-399 で workspace_id を足した業務テーブルのうち、
// 書き込み経路を持つもの（company_exercises は repository / usecase が未実装のため対象外）。
var businessTablesWithWorkspaceMirror = []string{"courses", "course_chapters", "invitations", "rich_documents"}

// businessTableWorkspaceMirrorTruncateTables はこの節のテストが TRUNCATE する対象。
var businessTableWorkspaceMirrorTruncateTables = append(
	append([]string{}, tenantBridgeTables...),
	businessTablesWithWorkspaceMirror...,
)

// tableWorkspaceID は業務テーブル 1 行の workspace_id を返す。table は本ファイル内の
// ハードコードされた定数のみを渡す（外部入力を SQL に組み込まない）。
func tableWorkspaceID(t *testing.T, db *sql.DB, table string, id any) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.QueryRow(fmt.Sprintf(`SELECT workspace_id FROM %s WHERE id = $1`, table), id).Scan(&got))
	return got
}

// TestTenantBridgeBusinessTableBackfill_Integration は FRESTYLE-399（courses / course_chapters /
// invitations / rich_documents への workspace_id 列追加）のバックフィルと dual-write を
// 実 PostgreSQL で固定する。読み取りは何も変えていない段なので、ここで確かめるのは
// 新しい列の中身だけ（company_id 側の挙動は既存テストがそのまま守る）。
func TestTenantBridgeBusinessTableBackfill_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("新規作成の直後に workspace_id が入る（dual-write）", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, businessTableWorkspaceMirrorTruncateTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		runStartupBackfill(ctx, t, sqlDB)
		ws1 := companyWorkspaceID(t, sqlDB, 1)
		require.True(t, ws1.Valid)

		users := persistence.NewUserRepository(sqlDB)
		cid := uint64(1)
		require.NoError(t, users.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "author@example.com", Name: "作成者", Role: domain.RoleCompanyAdmin, CompanyID: &cid,
		}, domain.OidcProviderCognito, "sub-author"))
		author, err := users.FindByCognitoSub(ctx, "sub-author")
		require.NoError(t, err)

		courses := persistence.NewCourseRepository(sqlDB)
		course := &domain.Course{CompanyID: 1, CreatedByUserID: author.ID, Title: "コース", Language: "go"}
		require.NoError(t, courses.Create(ctx, course))
		require.Equal(t, ws1, tableWorkspaceID(t, sqlDB, "courses", course.ID), "InsertCourse が workspace_id を dual-write する")

		materials := persistence.NewTeachingMaterialRepository(sqlDB)
		chapter := &domain.TeachingMaterial{CompanyID: 1, CourseID: course.ID, CreatedByUserID: author.ID, Title: "第1章"}
		require.NoError(t, materials.Create(ctx, chapter))
		require.Equal(t, ws1, tableWorkspaceID(t, sqlDB, "course_chapters", chapter.ID), "InsertChapter が workspace_id を dual-write する")

		invitations := persistence.NewAdminInvitationRepository(sqlDB)
		inv := &domain.AdminInvitation{
			CompanyID: 1, Email: "invitee@example.com", Role: domain.RoleTrainee,
			Status: domain.InvitationStatusPending, ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, invitations.Create(ctx, inv))
		require.Equal(t, ws1, tableWorkspaceID(t, sqlDB, "invitations", inv.ID), "InsertInvitation が workspace_id を dual-write する")

		richDocs := persistence.NewRichDocumentRepository(sqlDB)
		doc := &domain.RichDocument{
			OwnerID: author.ID, CompanyID: &cid, Kind: domain.DocumentKindNote,
			Title: "メモ", Doc: `{"type":"doc","content":[]}`,
		}
		require.NoError(t, richDocs.Create(ctx, doc))
		require.Equal(t, ws1, tableWorkspaceID(t, sqlDB, "rich_documents", doc.ID), "InsertRichDocument が workspace_id を dual-write する")
	})

	t.Run("会社を持たない rich_documents は workspace_id も NULL のまま", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, businessTableWorkspaceMirrorTruncateTables...)
		users := persistence.NewUserRepository(sqlDB)
		require.NoError(t, users.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "運営", Role: domain.RoleSuperAdmin,
		}, domain.OidcProviderCognito, "sub-root"))
		root, err := users.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)

		richDocs := persistence.NewRichDocumentRepository(sqlDB)
		doc := &domain.RichDocument{
			OwnerID: root.ID, CompanyID: nil, Kind: domain.DocumentKindNote,
			Title: "運営メモ", Doc: `{"type":"doc","content":[]}`,
		}
		require.NoError(t, richDocs.Create(ctx, doc))
		require.False(t, tableWorkspaceID(t, sqlDB, "rich_documents", doc.ID).Valid)
	})

	t.Run("バックフィル前に作られた行は起動時バックフィルで写る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, businessTableWorkspaceMirrorTruncateTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		runStartupBackfill(ctx, t, sqlDB) // 会社→ワークスペースの対応だけ先に作る
		ws1 := companyWorkspaceID(t, sqlDB, 1)

		// dual-write を経由しない、直接 INSERT でバックフィル前の状態（workspace_id 未設定）を再現する。
		var courseID int64
		require.NoError(t, sqlDB.QueryRow(
			`INSERT INTO courses (company_id, created_by_user_id, title, created_at, updated_at)
			 VALUES (1, 1, 'legacy', NOW(), NOW()) RETURNING id`,
		).Scan(&courseID))
		require.False(t, tableWorkspaceID(t, sqlDB, "courses", courseID).Valid, "作成直後はまだ写っていない")

		runStartupBackfill(ctx, t, sqlDB)
		require.Equal(t, ws1, tableWorkspaceID(t, sqlDB, "courses", courseID))
	})

	t.Run("何度流しても workspace_id がぶれない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, businessTableWorkspaceMirrorTruncateTables...)
		insertCompany(t, sqlDB, 1, "会社 A", true)
		var courseID int64
		require.NoError(t, sqlDB.QueryRow(
			`INSERT INTO courses (company_id, created_by_user_id, title, created_at, updated_at)
			 VALUES (1, 1, 'legacy', NOW(), NOW()) RETURNING id`,
		).Scan(&courseID))

		runStartupBackfill(ctx, t, sqlDB)
		first := tableWorkspaceID(t, sqlDB, "courses", courseID)
		require.True(t, first.Valid)

		for i := 0; i < 2; i++ { // 起動 2 回目・3 回目
			runStartupBackfill(ctx, t, sqlDB)
			require.Equal(t, first, tableWorkspaceID(t, sqlDB, "courses", courseID))
		}
	})
}
