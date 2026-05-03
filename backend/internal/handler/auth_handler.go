package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AuthHandler は Cognito 関連の認証エンドポイントを提供する。
// Spring Boot の CognitoAuthController に相当。
//
// HTTP / OAuth2 通信の詳細は infra/cognito.TokenExchanger に切り出してあり、
// このハンドラは HTTP プロトコル境界とユーザー upsert ロジックだけを持つ。
type AuthHandler struct {
	getCurrentUser *usecase.GetCurrentUserUseCase
	users          repository.UserRepository
	cognitoCfg     *config.CognitoConfig
	tokens         *cognito.TokenExchanger
}

// NewAuthHandler は本番用に http.Client + 10s timeout の TokenExchanger を組み立てて DI する。
func NewAuthHandler(getCurrentUser *usecase.GetCurrentUserUseCase, users repository.UserRepository, cognitoCfg *config.CognitoConfig) *AuthHandler {
	return &AuthHandler{
		getCurrentUser: getCurrentUser,
		users:          users,
		cognitoCfg:     cognitoCfg,
		tokens: cognito.NewTokenExchanger(cognito.Config{
			ClientID:     cognitoCfg.ClientID,
			ClientSecret: cognitoCfg.ClientSecret,
			RedirectURI:  cognitoCfg.RedirectURI,
			TokenURI:     cognitoCfg.TokenURI,
		}),
	}
}

// Me は現在ログイン中のユーザー情報を返す。
// レスポンスは domain.User の各フィールド + 派生 `isAdmin` / `groups` を含める。
// isAdmin の判定:
//  1. Cognito の `cognito:groups` claim に "admin" が含まれている (Spring Boot 時代と同等)
//  2. または DB users.role が super_admin / company_admin
//
// 上記いずれかで true。フロントは `isAdmin` を見て管理画面の表示可否を決める。
func (h *AuthHandler) Me(c *gin.Context) {
	sub, ok := c.Get(middleware.ContextKeyCognitoSub)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.getCurrentUser.Execute(c.Request.Context(), sub.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found"})
		return
	}
	groups := middleware.CognitoGroupsFromContext(c)
	isAdmin := middleware.IsAdminFromGroups(groups) ||
		user.Role == domain.RoleSuperAdmin ||
		user.Role == domain.RoleCompanyAdmin
	// access_token に admin グループがあるが DB role が未昇格の場合はここで同期する。
	// Google federated ユーザーはログイン時の upsert で groups が取れないケースがある。
	if middleware.IsAdminFromGroups(groups) && user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleCompanyAdmin {
		if h.users != nil {
			_ = h.users.UpdateRole(c.Request.Context(), user.ID, domain.RoleSuperAdmin)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          user.ID,
		"cognitoSub":  user.CognitoSub,
		"email":       user.Email,
		"displayName": user.DisplayName,
		"companyId":   user.CompanyID,
		"role":        user.Role,
		"createdAt":   user.CreatedAt,
		"updatedAt":   user.UpdatedAt,
		"groups":      groups,
		"isAdmin":     isAdmin,
	})
}

// Logout はリフレッシュ・アクセストークンの Cookie を消去する。
func (h *AuthHandler) Logout(c *gin.Context) {
	middleware.ClearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "ログアウトしました。"})
}

type cognitoCallbackReq struct {
	Code string `json:"code" binding:"required"`
}

// Callback は Cognito Hosted UI から戻ってきた認可コードをアクセストークンに交換し、
// HttpOnly Cookie に格納する。Spring Boot の CognitoAuthController#callback 相当。
func (h *AuthHandler) Callback(c *gin.Context) {
	var req cognitoCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tok, err := h.tokens.ExchangeAuthorizationCode(c.Request.Context(), req.Code)
	if status, body, ok := h.handleTokenError(c, "callback", err); ok {
		c.JSON(status, body)
		return
	}

	middleware.SetAccessTokenCookie(c, tok.AccessToken, tok.ExpiresIn)
	middleware.SetRefreshTokenCookie(c, tok.RefreshToken)

	// 初回ログインで users 行が無いと /auth/me が 404 になるため自動で upsert する。
	// id_token の payload から sub / email / cognito:groups を取り出して同期する:
	//   - 未登録なら role = (group に admin なら super_admin / 無ければ trainee) で create
	//   - 既存ユーザーは Cognito group が変わっていれば role を update する
	// これで Spring Boot 時代と同じ「Cognito group が真のソース」で運用できる。
	h.upsertUserFromIDToken(c, tok.IDToken)

	c.JSON(http.StatusOK, gin.H{"message": "ログインしました。"})
}

// Refresh は HttpOnly Cookie の refresh_token を使ってアクセストークンを再発行する。
func (h *AuthHandler) Refresh(c *gin.Context) {
	rt, err := c.Cookie(middleware.CookieRefreshToken)
	if err != nil || rt == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_token_missing"})
		return
	}

	tok, err := h.tokens.RefreshAccessToken(c.Request.Context(), rt)
	if err != nil {
		// refresh_token が無効と判明した時はログイン状態をクリアして 401。
		// それ以外（502: Cognito 不到達など）は Cookie を残してリトライ余地を残す。
		var exErr *cognito.TokenExchangeError
		if errors.As(err, &exErr) {
			log.Printf("cognito refresh: status=%d body=%s", exErr.HTTPStatus, exErr.Body)
			middleware.ClearAuthCookies(c)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_failed"})
			return
		}
		status, body, _ := h.handleTokenError(c, "refresh", err)
		c.JSON(status, body)
		return
	}

	middleware.SetAccessTokenCookie(c, tok.AccessToken, tok.ExpiresIn)
	// refresh_token grant でも id_token が返る場合は DB role を同期する。
	// Google federated ユーザーは ID token に cognito:groups が含まれないことがあるため
	// access_token の claims からも昇格を試みる。
	if tok.IDToken != "" {
		h.upsertUserFromIDToken(c, tok.IDToken)
	} else {
		h.syncRoleFromAccessToken(c, tok.AccessToken)
	}
	c.JSON(http.StatusOK, gin.H{"message": "refreshed"})
}

// syncRoleFromAccessToken は access_token の cognito:groups を見て DB role を super_admin に昇格する。
// ID token に groups が含まれない Google federated ユーザー向けのフォールバック。
func (h *AuthHandler) syncRoleFromAccessToken(c *gin.Context, accessToken string) {
	if h.users == nil {
		return
	}
	claims, err := middleware.DecodeClaims(accessToken)
	if err != nil {
		return
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return
	}
	groups := middleware.ToStringSliceFromClaim(claims["cognito:groups"])
	if !middleware.IsAdminFromGroups(groups) {
		return
	}
	existing, _ := h.users.FindByCognitoSub(c.Request.Context(), sub)
	if existing != nil && existing.Role != domain.RoleSuperAdmin && existing.Role != domain.RoleCompanyAdmin {
		_ = h.users.UpdateRole(c.Request.Context(), existing.ID, domain.RoleSuperAdmin)
	}
}

// handleTokenError は cognito.TokenExchanger が返したエラーを HTTP レスポンスに変換する。
// returned ok=true なら呼び元は早期 return する想定。
func (h *AuthHandler) handleTokenError(c *gin.Context, op string, err error) (int, gin.H, bool) {
	if err == nil {
		return 0, nil, false
	}

	var exErr *cognito.TokenExchangeError
	switch {
	case errors.Is(err, cognito.ErrNotConfigured):
		return http.StatusInternalServerError, gin.H{"error": "cognito_not_configured"}, true
	case errors.As(err, &exErr):
		// 本物の理由 (invalid_grant / invalid_client / redirect_uri_mismatch 等) を残す。
		// クライアントには簡素なエラーだけ返す。
		log.Printf("cognito %s: token exchange status=%d body=%s redirect_uri=%s client_id_set=%t client_secret_set=%t",
			op, exErr.HTTPStatus, exErr.Body, h.cognitoCfg.RedirectURI, h.cognitoCfg.ClientID != "", h.cognitoCfg.ClientSecret != "")
		return http.StatusUnauthorized, gin.H{"error": "token_exchange_failed"}, true
	case errors.Is(err, cognito.ErrUnreachable):
		log.Printf("cognito %s: token endpoint unreachable: %v", op, err)
		return http.StatusBadGateway, gin.H{"error": "cognito_unreachable"}, true
	case errors.Is(err, cognito.ErrInvalidResponse):
		log.Printf("cognito %s: invalid token response: %v", op, err)
		return http.StatusBadGateway, gin.H{"error": "invalid_token_response"}, true
	default:
		log.Printf("cognito %s: unexpected error: %v", op, err)
		return http.StatusInternalServerError, gin.H{"error": "internal_error"}, true
	}
}

// upsertUserFromIDToken は id_token の claim を見て users 行を新規作成 / role 更新する。
// Cognito group "admin" 所属時のみ super_admin に昇格する。降格は AdminInvitation 経由のみ。
func (h *AuthHandler) upsertUserFromIDToken(c *gin.Context, idToken string) {
	claims, err := middleware.DecodeClaims(idToken)
	if err != nil || h.users == nil {
		return
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return
	}
	email, _ := claims["email"].(string)
	groups := middleware.ToStringSliceFromClaim(claims["cognito:groups"])
	desiredRole := domain.RoleTrainee
	if middleware.IsAdminFromGroups(groups) {
		desiredRole = domain.RoleSuperAdmin
	}

	existing, _ := h.users.FindByCognitoSub(c.Request.Context(), sub)
	if existing == nil {
		_ = h.users.Create(c.Request.Context(), &domain.User{
			CognitoSub:  sub,
			Email:       email,
			DisplayName: email,
			Role:        desiredRole,
		})
		return
	}
	// Cognito group で admin に昇格された場合のみ DB role を上書きする。
	// 既に DB で super_admin / company_admin が設定されているケースは触らない
	// （AdminInvitation 系の手動付与が他にある可能性があるため、降格は手動で）。
	if existing.Role != desiredRole && desiredRole == domain.RoleSuperAdmin {
		_ = h.users.UpdateRole(c.Request.Context(), existing.ID, desiredRole)
	}
}
