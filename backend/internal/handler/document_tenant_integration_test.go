//go:build integration

package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

const tenantDocBody = `{"type":"doc","content":[{"type":"paragraph"}]}`

// docTenantRouter は本番と同じ registerDocumentRoutes でルートを張り、user を current user として注入する。
// 認証済みであること以外は本番と同じ経路を通るので、「ログインしていれば他ワークスペースの公開文書が読めるか」を
// エンドポイント単位で確かめられる。
func docTenantRouter(db *sql.DB, user *domain.User) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("")
	g.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUserID, user.ID)
		c.Set(middleware.ContextKeyCurrentUser, user)
		c.Next()
	})
	registerDocumentRoutes(g, &routeDeps{db: db})
	return r
}

func docTenantGet(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

// TestDocumentTenantIsolation_Integration は「公開文書は同一ワークスペースの中でだけ読める」を実 PostgreSQL で固定する。
// ワークスペースをまたいだ閲覧は、読み取りエンドポイントのどれからも成立しないこと（取得・一覧の両方）を見る。
func TestDocumentTenantIsolation_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "rich_documents", "user_oidc_identities", "users", "workspaces")

	userRepo := persistence.NewUserRepository(sqlDB)
	docRepo := persistence.NewRichDocumentRepository(sqlDB)

	// テナントの正本は workspaces なので、下ごしらえもそこへ直接 1 行入れる。
	mkWorkspace := func(slug, name string) *string {
		id := uuid.NewString()
		_, err := sqlDB.ExecContext(
			ctx,
			`INSERT INTO workspaces (id, slug, name) VALUES ($1, $2, $3)`, id, slug, name,
		)
		require.NoError(t, err)
		return &id
	}
	mkUser := func(sub, email string, workspaceID *string) *domain.User {
		u := &domain.User{Email: email, Role: domain.RoleTrainee, WorkspaceID: workspaceID, IsActive: true}
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		// 本番の current user 解決（毎回 DB から引く）と同じ状態にするため、作成後に読み直す。
		got, err := userRepo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		return got
	}
	mkDoc := func(owner uint64, workspaceID *string, title string, isPublic bool) *domain.RichDocument {
		// workspace_id は呼び出し側（usecase）が解決して渡す設計のため、ここでも明示的に渡す。
		d := &domain.RichDocument{
			OwnerID: owner, WorkspaceID: workspaceID, Kind: domain.DocumentKindNote,
			Title: title, IsPublic: isPublic, Doc: tenantDocBody, Revision: 1, SchemaVersion: 1,
		}
		require.NoError(t, docRepo.Create(ctx, d))
		return d
	}

	workspaceA := mkWorkspace("tenant-doc-a", "ワークスペース A")
	workspaceB := mkWorkspace("tenant-doc-b", "ワークスペース B")

	ownerA := mkUser("tenant-owner-a", "tenant-owner-a@example.com", workspaceA)
	peerA := mkUser("tenant-peer-a", "tenant-peer-a@example.com", workspaceA)
	userB := mkUser("tenant-user-b", "tenant-user-b@example.com", workspaceB)
	// ワークスペース未所属（運営管理者相当）。workspace_id が NULL の文書の所有者になる。
	loneOwner := mkUser("tenant-lone", "tenant-lone@example.com", nil)

	publicA := mkDoc(ownerA.ID, workspaceA, "ワークスペース A の公開メモ", true)
	privateA := mkDoc(ownerA.ID, workspaceA, "ワークスペース A の非公開メモ", false)
	publicNullWorkspace := mkDoc(loneOwner.ID, nil, "ワークスペース不明の公開メモ", true)
	// 作成後に異動した所有者を模す（workspace_id は作成時の写しのままで更新されない）。
	staleWorkspaceDoc := mkDoc(userB.ID, workspaceA, "ワークスペース B のユーザーが A 在籍時に作った公開メモ", true)

	t.Run("取得: 別ワークスペースのユーザーは他ワークスペースの公開文書を読めない(404)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, userB), "/documents/"+publicA.ID)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("取得: 同一ワークスペースの別ユーザーは公開文書を読める(200)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, peerA), "/documents/"+publicA.ID)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("取得: 同一ワークスペースの別ユーザーでも非公開は読めない(404)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, peerA), "/documents/"+privateA.ID)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("取得: ワークスペース不明(NULL)の公開文書は所有者以外から読めない(404)", func(t *testing.T) {
		for _, viewer := range []*domain.User{peerA, userB} {
			w := docTenantGet(t, docTenantRouter(sqlDB, viewer), "/documents/"+publicNullWorkspace.ID)
			require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		}
	})

	t.Run("取得: 所有者はワークスペース不明(NULL)の自分の文書を読める(200)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, loneOwner), "/documents/"+publicNullWorkspace.ID)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("取得: 所有者は文書のワークスペースが自分のワークスペースと食い違っても読める(200)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, userB), "/documents/"+staleWorkspaceDoc.ID)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("取得: 別ワークスペースのユーザーは他ワークスペースの非公開文書も読めない(404)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, userB), "/documents/"+privateA.ID)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("一覧: 他ワークスペースユーザーの一覧に他ワークスペースの文書は出ない", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, userB), "/documents")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var rows []struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rows))
		ids := make(map[string]bool, len(rows))
		for _, row := range rows {
			ids[row.ID] = true
		}
		require.False(t, ids[publicA.ID], "他ワークスペースの公開文書が一覧に出ている")
		require.False(t, ids[privateA.ID], "他ワークスペースの非公開文書が一覧に出ている")
		require.False(t, ids[publicNullWorkspace.ID], "他人の文書が一覧に出ている")
		require.True(t, ids[staleWorkspaceDoc.ID], "自分が所有する文書は一覧に出るべき")
	})

	t.Run("一覧: 同一ワークスペースの別ユーザーの公開文書も一覧には出ない(owner スコープ)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, peerA), "/documents")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.Equal(t, "[]", w.Body.String())
	})

	t.Run("一覧: 所有者は自分の文書をワークスペースの値によらず取れる", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, loneOwner), "/documents")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var rows []struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rows))
		require.Len(t, rows, 1)
		require.Equal(t, publicNullWorkspace.ID, rows[0].ID)
	})
}
