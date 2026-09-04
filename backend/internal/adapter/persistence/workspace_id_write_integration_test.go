//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// workspaceWriteTables はこの節のテストが TRUNCATE する対象。
var workspaceWriteTables = []string{"users", "user_oidc_identities", "workspaces"}

// insertWorkspaceWithActive はワークスペースを 1 件作る（id を明示して固定する）。
func insertWorkspaceWithActive(t *testing.T, db *sql.DB, id uuid.UUID, name string, active bool) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO workspaces (id, slug, name, is_active) VALUES ($1, $2, $3, $4)`,
		id, "ws-"+id.String(), name, active,
	)
	require.NoError(t, err)
}

// userWorkspaceID はユーザーの workspace_id を返す（未設定なら Valid=false）。
func userWorkspaceID(t *testing.T, db *sql.DB, userID uint64) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.QueryRow(`SELECT workspace_id FROM users WHERE id = $1`, userID).Scan(&got))
	return got
}

// TestUserWorkspaceWrite_Integration は users.workspace_id が「渡された値そのまま」で
// 書かれることを固定する。所属参照の解決は呼び出し側（usecase）の責務で、
// リポジトリ側が引き直したりはしない。
func TestUserWorkspaceWrite_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("渡した所属ワークスペースがそのまま書かれる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, workspaceWriteTables...)
		ws1, ws2 := uuid.New(), uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)
		insertWorkspaceWithActive(t, sqlDB, ws2, "ワークスペース B", true)
		ws1Str, ws2Str := ws1.String(), ws2.String()

		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "a@example.com", Name: "A", WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-a"))
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "b@example.com", Name: "B", WorkspaceID: &ws2Str,
		}, domain.OidcProviderCognito, "sub-b"))

		userA, err := repo.FindByCognitoSub(ctx, "sub-a")
		require.NoError(t, err)
		userB, err := repo.FindByCognitoSub(ctx, "sub-b")
		require.NoError(t, err)

		require.Equal(t, uuid.NullUUID{UUID: ws1, Valid: true}, userWorkspaceID(t, sqlDB, userA.ID))
		require.Equal(t, uuid.NullUUID{UUID: ws2, Valid: true}, userWorkspaceID(t, sqlDB, userB.ID))
		// 読み出しも同じ所属を返す。
		require.Equal(t, ws1Str, *userA.WorkspaceID)
		require.Equal(t, ws2Str, *userB.WorkspaceID)
	})

	t.Run("未所属のまま作られたユーザーは所属先ワークスペースが無い", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, workspaceWriteTables...)
		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "運営",
		}, domain.OidcProviderCognito, "sub-root"))

		got, err := repo.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)
		require.Nil(t, got.WorkspaceID)
		require.False(t, tableWorkspaceID(t, sqlDB, "users", got.ID).Valid)
	})

	t.Run("所属の付け替えでワークスペースが入れ替わる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, workspaceWriteTables...)
		ws1, ws2 := uuid.New(), uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)
		insertWorkspaceWithActive(t, sqlDB, ws2, "ワークスペース B", true)
		ws1Str, ws2Str := ws1.String(), ws2.String()

		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "move@example.com", Name: "異動", WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-move"))
		got, err := repo.FindByCognitoSub(ctx, "sub-move")
		require.NoError(t, err)
		require.Equal(t, uuid.NullUUID{UUID: ws1, Valid: true}, tableWorkspaceID(t, sqlDB, "users", got.ID))

		require.NoError(t, repo.UpdateWorkspaceID(ctx, got.ID, &ws2Str))

		moved, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.Equal(t, ws2Str, *moved.WorkspaceID)
		require.Equal(t, uuid.NullUUID{UUID: ws2, Valid: true}, tableWorkspaceID(t, sqlDB, "users", got.ID))
	})

	t.Run("nil を渡すと未所属へ戻る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, workspaceWriteTables...)
		ws1 := uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)
		ws1Str := ws1.String()

		repo := persistence.NewUserRepository(sqlDB)
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "leave@example.com", Name: "退所", WorkspaceID: &ws1Str,
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
var businessTablesWithWorkspace = []string{
	"courses", "course_chapters", "rich_documents",
}

// businessTableTruncateTables はこの節のテストが TRUNCATE する対象。
var businessTableTruncateTables = append(
	append([]string{}, workspaceWriteTables...),
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
// 作成者の所属ワークスペースと異なるワークスペースを渡して見分ける。
func TestBusinessTableWorkspaceWrite_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("新規作成の直後に渡した workspace_id が入る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, businessTableTruncateTables...)
		ws1, ws2 := uuid.New(), uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)
		insertWorkspaceWithActive(t, sqlDB, ws2, "ワークスペース B", true)
		// 作成者はワークスペース A（ws1）に所属させたうえで、行の所属先には B（ws2）を渡す。
		// 作成者の所属から引き直していれば ws1 になってしまうので、これで見分けられる。
		ws1Str, ws2Str := ws1.String(), ws2.String()

		users := persistence.NewUserRepository(sqlDB)
		require.NoError(t, users.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "author@example.com", Name: "作成者", WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-author"))
		author, err := users.FindByCognitoSub(ctx, "sub-author")
		require.NoError(t, err)

		courses := persistence.NewCourseRepository(sqlDB)
		course := &domain.Course{WorkspaceID: &ws2Str, CreatedByUserID: author.ID, Title: "コース", Language: "go"}
		require.NoError(t, courses.Create(ctx, course))
		require.Equal(t, uuid.NullUUID{UUID: ws2, Valid: true}, tableWorkspaceID(t, sqlDB, "courses", course.ID), "InsertCourse は渡された workspace_id を書く")

		materials := persistence.NewTeachingMaterialRepository(sqlDB)
		chapter := &domain.TeachingMaterial{WorkspaceID: &ws2Str, CourseID: course.ID, CreatedByUserID: author.ID, Title: "第1章"}
		require.NoError(t, materials.Create(ctx, chapter))
		require.Equal(t, uuid.NullUUID{UUID: ws2, Valid: true}, tableWorkspaceID(t, sqlDB, "course_chapters", chapter.ID), "InsertChapter は渡された workspace_id を書く")

		richDocs := persistence.NewRichDocumentRepository(sqlDB)
		doc := &domain.RichDocument{
			OwnerID: author.ID, WorkspaceID: &ws2Str, Kind: domain.DocumentKindNote,
			Title: "メモ", Doc: `{"type":"doc","content":[]}`,
		}
		require.NoError(t, richDocs.Create(ctx, doc))
		require.Equal(t, uuid.NullUUID{UUID: ws2, Valid: true}, tableWorkspaceID(t, sqlDB, "rich_documents", doc.ID), "InsertRichDocument は渡された workspace_id を書く")
	})

	t.Run("所属を渡さない rich_documents は workspace_id も NULL のまま", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, businessTableTruncateTables...)
		users := persistence.NewUserRepository(sqlDB)
		require.NoError(t, users.CreateWithOidcIdentity(ctx, &domain.User{
			Email: "root@example.com", Name: "運営",
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
