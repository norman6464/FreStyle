//go:build integration

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// platformAdminRouter は本番と同じ順序（JWTAuth → CurrentUser → SyncPlatformAdmin → RequireAdmin → handler）で
// ルートを張る。claims は access_token の検証結果として JWTAuth に返させるもので、
// "cognito:groups" キーを入れる / 入れないで claim の欠落と空を撃ち分けられる。
func platformAdminRouter(db *gorm.DB, claims map[string]any) *gin.Engine {
	gin.SetMode(gin.TestMode)
	userRepo := persistence.NewUserRepository(db)
	companyRepo := persistence.NewCompanyRepository(db)

	authHandler := NewAuthHandler(
		usecase.NewGetCurrentUserUseCase(userRepo),
		usecase.NewUpsertUserFromIDTokenUseCase(userRepo, persistence.NewAdminInvitationRepository(db), ""),
		usecase.NewPromoteCognitoAdminRoleUseCase(userRepo),
		usecase.NewSyncPlatformAdminUseCase(userRepo),
		&config.CognitoConfig{},
		nil,
		nil,
	)

	r := gin.New()
	g := r.Group("")
	// 横断処理は本番と同じ組み立て（authedMiddlewares）をそのまま使う。ここで並べ直すと、
	// 本番の並びから 1 つ抜け落ちてもこのテストが緑のままになる。
	g.Use(authedMiddlewares(
		func(context.Context, string) (map[string]any, error) { return claims, nil },
		userRepo, companyRepo,
	)...)
	registerAuthAuthedRoutes(g, authHandler)

	admin := g.Group("", middleware.RequireAdmin())
	admin.GET("/admin/companies", NewAdminCompanyHandler(
		usecase.NewListCompaniesUseCase(companyRepo),
		usecase.NewListCompanyStatsUseCase(companyRepo, persistence.NewCompanyStatsRepository(db)),
		usecase.NewSetCompanyActiveUseCase(companyRepo),
	).List)
	return r
}

func platformAdminGet(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.AddCookie(&http.Cookie{Name: middleware.CookieAccessToken, Value: "dummy-access-token"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// platformAdminFlag は DB 上の users.is_platform_admin を直接読む（表示ではなく事実を見る）。
func platformAdminFlag(t *testing.T, db *gorm.DB, userID uint64) bool {
	t.Helper()
	var flag bool
	require.NoError(t, db.Raw(`SELECT is_platform_admin FROM users WHERE id = ?`, userID).Scan(&flag).Error)
	return flag
}

// TestPlatformAdminOffboarding_Integration は運営権限のオフボーディングを実 PostgreSQL で固定する。
//
// 見るのは 3 つ:
//   - バックフィル直後の状態（is_platform_admin = true）では挙動が何も変わらないこと
//   - false にした瞬間にプラットフォーム権限を要する操作が通らなくなること
//   - cognito:groups claim の有無・内容が、付与 / 剥奪 / 「何もしない」を正しく撃ち分けること
func TestPlatformAdminOffboarding_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "user_oidc_identities", "users", "companies")

	userRepo := persistence.NewUserRepository(db)

	// 運営管理者は会社に属さない（本番の super_admin と同じく company_id は NULL）。
	mkSuperAdmin := func(sub, email string, isPlatformAdmin bool) *domain.User {
		u := &domain.User{
			Email: email, Name: email, Role: domain.RoleSuperAdmin,
			IsActive: true, IsPlatformAdmin: isPlatformAdmin,
		}
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		return u
	}
	claimsFor := func(sub string, groups []string, withGroupsKey bool) map[string]any {
		claims := map[string]any{"sub": sub}
		if withGroupsKey {
			raw := make([]any, 0, len(groups))
			for _, g := range groups {
				raw = append(raw, g)
			}
			claims["cognito:groups"] = raw
		}
		return claims
	}
	meRole := func(t *testing.T, w *httptest.ResponseRecorder) (string, bool) {
		t.Helper()
		var body struct {
			Role    string `json:"role"`
			IsAdmin bool   `json:"isAdmin"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		return body.Role, body.IsAdmin
	}

	t.Run("バックフィル直後(true)の運営管理者は従来どおり操作できる", func(t *testing.T) {
		sub := "platform-admin-kept"
		u := mkSuperAdmin(sub, "kept@example.com", true)
		// groups claim があり admin を含む、本番の正常な状態。
		r := platformAdminRouter(db, claimsFor(sub, []string{"admin"}, true))

		w := platformAdminGet(t, r, "/admin/companies")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		w = platformAdminGet(t, r, "/auth/me")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		role, isAdmin := meRole(t, w)
		require.Equal(t, string(domain.RoleSuperAdmin), role)
		require.True(t, isAdmin)
		require.True(t, platformAdminFlag(t, db, u.ID))
	})

	t.Run("false にすると role が super_admin のままでも通らない", func(t *testing.T) {
		sub := "platform-admin-revoked"
		u := mkSuperAdmin(sub, "revoked@example.com", false)
		// claim は欠落させる（失効の原因が「今のリクエストの claim」ではなく
		// 「DB の列」であることを見るため）。
		r := platformAdminRouter(db, claimsFor(sub, nil, false))

		w := platformAdminGet(t, r, "/admin/companies")
		require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())

		w = platformAdminGet(t, r, "/auth/me")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		role, isAdmin := meRole(t, w)
		require.Equal(t, string(domain.RoleTrainee), role, "実効役割は最小権限へ倒れる")
		require.False(t, isAdmin)

		// 保存された役割自体は下げない（下げ先が決まらないため）。
		var roleID uint16
		require.NoError(t, db.Raw(`SELECT role_id FROM users WHERE id = ?`, u.ID).Scan(&roleID).Error)
		require.Equal(t, domain.RoleIDSuperAdmin, roleID)
	})

	t.Run("claim が存在しないときは降格しない", func(t *testing.T) {
		sub := "platform-admin-no-claim"
		u := mkSuperAdmin(sub, "no-claim@example.com", true)
		// federated（Google 連携）ユーザーのトークンには groups claim が載らないことがある。
		// これを「グループに居ない」と読むと、正当な運営管理者を締め出す。
		r := platformAdminRouter(db, claimsFor(sub, nil, false))

		w := platformAdminGet(t, r, "/auth/me")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		role, isAdmin := meRole(t, w)
		require.Equal(t, string(domain.RoleSuperAdmin), role)
		require.True(t, isAdmin)
		require.True(t, platformAdminFlag(t, db, u.ID), "claim 欠落で剥奪してはならない")

		// 次のリクエストでも通り続ける。
		w = platformAdminGet(t, r, "/admin/companies")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.True(t, platformAdminFlag(t, db, u.ID))
	})

	t.Run("claim が空配列でも欠落と混同せず剥奪する", func(t *testing.T) {
		sub := "platform-admin-empty-claim"
		u := mkSuperAdmin(sub, "empty-claim@example.com", true)
		r := platformAdminRouter(db, claimsFor(sub, nil, true))

		w := platformAdminGet(t, r, "/auth/me")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.False(t, platformAdminFlag(t, db, u.ID))
	})

	// 失効は /auth/me ではなく認可の手前で起きなければならない。管理 API は
	// JWTAuth → CurrentUser → RequireAdmin と進むだけで /auth/me を通らないので、
	// /auth/me でしか同期しないと退任者は直接 /admin/* を叩けてしまう。
	t.Run("グループから外れたら_authmeを通らなくても管理APIが通らない", func(t *testing.T) {
		sub := "platform-admin-offboarded"
		u := mkSuperAdmin(sub, "offboarded@example.com", true)

		// 1) まだ admin グループに居るあいだは通る。
		inGroup := platformAdminRouter(db, claimsFor(sub, []string{"admin"}, true))
		w := platformAdminGet(t, inGroup, "/admin/companies")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		// 2) Cognito 側で admin グループから外す（claim はあるが admin を含まない）。
		//    /auth/me を一度も呼ばずに管理 API を叩く。
		outOfGroup := platformAdminRouter(db, claimsFor(sub, []string{"users"}, true))
		w = platformAdminGet(t, outOfGroup, "/admin/companies")
		require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		require.False(t, platformAdminFlag(t, db, u.ID), "認可の手前で失効していること")

		// 3) /auth/me の応答も管理者ではなくなる。
		w = platformAdminGet(t, outOfGroup, "/auth/me")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		role, isAdmin := meRole(t, w)
		require.Equal(t, string(domain.RoleTrainee), role)
		require.False(t, isAdmin)

		// 4) 以後も全顧客企業の一覧に触れない。
		w = platformAdminGet(t, outOfGroup, "/admin/companies")
		require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	})

	t.Run("グループに戻れば付与される", func(t *testing.T) {
		sub := "platform-admin-rejoined"
		u := mkSuperAdmin(sub, "rejoined@example.com", false)
		r := platformAdminRouter(db, claimsFor(sub, []string{"admin"}, true))

		// 付与も認可の手前で起きるので、/auth/me を先に呼ぶ必要はない。
		w := platformAdminGet(t, r, "/admin/companies")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.True(t, platformAdminFlag(t, db, u.ID))

		w = platformAdminGet(t, r, "/auth/me")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		role, isAdmin := meRole(t, w)
		require.Equal(t, string(domain.RoleSuperAdmin), role)
		require.True(t, isAdmin)
	})

	// 読めない形の claim（配列に非 string が混ざる）は「グループに居ない」ではない。
	// これを失効と読むと、正当な運営管理者を締め出す。
	t.Run("読めない形のclaimでは降格しない", func(t *testing.T) {
		sub := "platform-admin-broken-claim"
		u := mkSuperAdmin(sub, "broken-claim@example.com", true)
		r := platformAdminRouter(db, map[string]any{
			"sub":            sub,
			"cognito:groups": []any{"admin", 42},
		})

		w := platformAdminGet(t, r, "/admin/companies")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.True(t, platformAdminFlag(t, db, u.ID), "読めない claim で剥奪してはならない")
	})
}
