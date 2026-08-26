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
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

const tenantDocBody = `{"type":"doc","content":[{"type":"paragraph"}]}`

// docTenantRouter は本番と同じ registerDocumentRoutes でルートを張り、user を current user として注入する。
// 認証済みであること以外は本番と同じ経路を通るので、「ログインしていれば他社の公開文書が読めるか」を
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

// TestDocumentTenantIsolation_Integration は「公開文書は同一会社の中でだけ読める」を実 PostgreSQL で固定する。
// 会社をまたいだ閲覧は、読み取りエンドポイントのどれからも成立しないこと（取得・一覧の両方）を見る。
func TestDocumentTenantIsolation_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "rich_documents", "user_oidc_identities", "users", "companies")

	userRepo := persistence.NewUserRepository(sqlDB)
	docRepo := persistence.NewRichDocumentRepository(sqlDB)

	mkCompany := func(name string) uint64 {
		c := &domain.Company{Name: name, IsActive: true}
		require.NoError(t, db.Create(c).Error)
		return c.ID
	}
	mkUser := func(sub, email string, companyID *uint64) *domain.User {
		u := &domain.User{Email: email, Role: domain.RoleTrainee, CompanyID: companyID, IsActive: true}
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		return u
	}
	mkDoc := func(owner uint64, companyID *uint64, title string, isPublic bool) *domain.RichDocument {
		d := &domain.RichDocument{
			OwnerID: owner, CompanyID: companyID, Kind: domain.DocumentKindNote,
			Title: title, IsPublic: isPublic, Doc: tenantDocBody, Revision: 1, SchemaVersion: 1,
		}
		require.NoError(t, docRepo.Create(ctx, d))
		return d
	}

	companyA := mkCompany("A 社")
	companyB := mkCompany("B 社")

	ownerA := mkUser("tenant-owner-a", "tenant-owner-a@example.com", &companyA)
	peerA := mkUser("tenant-peer-a", "tenant-peer-a@example.com", &companyA)
	userB := mkUser("tenant-user-b", "tenant-user-b@example.com", &companyB)
	// 会社未所属（運営管理者相当）。company_id が NULL の文書の所有者になる。
	loneOwner := mkUser("tenant-lone", "tenant-lone@example.com", nil)

	publicA := mkDoc(ownerA.ID, &companyA, "A 社の公開メモ", true)
	privateA := mkDoc(ownerA.ID, &companyA, "A 社の非公開メモ", false)
	publicNullCompany := mkDoc(loneOwner.ID, nil, "会社不明の公開メモ", true)
	// 作成後に異動した所有者を模す（company_id は作成時の写しのままで更新されない）。
	staleCompanyDoc := mkDoc(userB.ID, &companyA, "B 社ユーザーが A 社在籍時に作った公開メモ", true)

	t.Run("取得: 別会社のユーザーは他社の公開文書を読めない(404)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, userB), "/documents/"+publicA.ID)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("取得: 同一会社の別ユーザーは公開文書を読める(200)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, peerA), "/documents/"+publicA.ID)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("取得: 同一会社の別ユーザーでも非公開は読めない(404)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, peerA), "/documents/"+privateA.ID)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("取得: 会社不明(NULL)の公開文書は所有者以外から読めない(404)", func(t *testing.T) {
		for _, viewer := range []*domain.User{peerA, userB} {
			w := docTenantGet(t, docTenantRouter(sqlDB, viewer), "/documents/"+publicNullCompany.ID)
			require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		}
	})

	t.Run("取得: 所有者は会社不明(NULL)の自分の文書を読める(200)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, loneOwner), "/documents/"+publicNullCompany.ID)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("取得: 所有者は文書の会社が自分の会社と食い違っても読める(200)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, userB), "/documents/"+staleCompanyDoc.ID)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("取得: 別会社のユーザーは他社の非公開文書も読めない(404)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, userB), "/documents/"+privateA.ID)
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("一覧: 他社ユーザーの一覧に他社の文書は出ない", func(t *testing.T) {
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
		require.False(t, ids[publicA.ID], "他社の公開文書が一覧に出ている")
		require.False(t, ids[privateA.ID], "他社の非公開文書が一覧に出ている")
		require.False(t, ids[publicNullCompany.ID], "他人の文書が一覧に出ている")
		require.True(t, ids[staleCompanyDoc.ID], "自分が所有する文書は一覧に出るべき")
	})

	t.Run("一覧: 同一会社の別ユーザーの公開文書も一覧には出ない(owner スコープ)", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, peerA), "/documents")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.Equal(t, "[]", w.Body.String())
	})

	t.Run("一覧: 所有者は自分の文書を会社の値によらず取れる", func(t *testing.T) {
		w := docTenantGet(t, docTenantRouter(sqlDB, loneOwner), "/documents")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var rows []struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &rows))
		require.Len(t, rows, 1)
		require.Equal(t, publicNullCompany.ID, rows[0].ID)
	})
}
