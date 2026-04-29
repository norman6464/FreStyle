package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AuthHandler は Cognito 関連の認証エンドポイントを提供する。
// Spring Boot の CognitoAuthController に相当。
type AuthHandler struct {
	getCurrentUser *usecase.GetCurrentUserUseCase
	users          repository.UserRepository
	cognito        *config.CognitoConfig
	httpClient     *http.Client
}

func NewAuthHandler(getCurrentUser *usecase.GetCurrentUserUseCase, users repository.UserRepository, cognito *config.CognitoConfig) *AuthHandler {
	return &AuthHandler{
		getCurrentUser: getCurrentUser,
		users:          users,
		cognito:        cognito,
		// Cognito token endpoint への通信が無限待ちにならないよう必ず timeout を設定する
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Me は現在ログイン中のユーザー情報を返す。
// レスポンスは domain.User の各フィールド + 派生 `isAdmin` / `groups` を含める。
// isAdmin の判定:
//  1. Cognito の `cognito:groups` claim に "ADMIN" が含まれている (Spring Boot 時代と同等)
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
	c.SetCookie(middleware.CookieAccessToken, "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "ログアウトしました。"})
}

type cognitoCallbackReq struct {
	Code string `json:"code" binding:"required"`
}

type cognitoTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Callback は Cognito Hosted UI から戻ってきた認可コードをアクセストークンに交換し、
// HttpOnly Cookie に格納する。Spring Boot の CognitoAuthController#callback 相当。
func (h *AuthHandler) Callback(c *gin.Context) {
	var req cognitoCallbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.cognito == nil || h.cognito.TokenURI == "" || h.cognito.ClientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cognito_not_configured"})
		return
	}

	// Cognito の OAuth2 token endpoint には「Authorization Basic header 方式」と
	// 「body 方式 (client_id + client_secret を form に入れる)」があるが、両方送ると
	// invalid_client を返すケースがある。本実装では body 方式に統一する（AWS docs 推奨）。
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", h.cognito.ClientID)
	if h.cognito.ClientSecret != "" {
		form.Set("client_secret", h.cognito.ClientSecret)
	}
	form.Set("code", req.Code)
	form.Set("redirect_uri", h.cognito.RedirectURI)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost,
		h.cognito.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "request_build_failed"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("cognito callback: token exchange request failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "cognito_unreachable"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		// Cognito 側が拒否した本当の理由 (invalid_grant / invalid_client / redirect_uri_mismatch 等) を
		// CloudWatch Logs に必ず残す。クライアントには簡素なエラーだけ返す。
		log.Printf("cognito callback: token exchange status=%d body=%s redirect_uri=%s client_id_set=%t client_secret_set=%t",
			resp.StatusCode, string(body), h.cognito.RedirectURI, h.cognito.ClientID != "", h.cognito.ClientSecret != "")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token_exchange_failed"})
		return
	}

	var tok cognitoTokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		log.Printf("cognito callback: invalid token response: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid_token_response"})
		return
	}

	// HttpOnly + Secure + SameSite=None でアプリ全体に Cookie を渡す
	maxAge := tok.ExpiresIn
	if maxAge <= 0 {
		maxAge = 3600
	}
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(middleware.CookieAccessToken, tok.AccessToken, maxAge, "/", "", true, true)
	if tok.RefreshToken != "" {
		c.SetCookie("refresh_token", tok.RefreshToken, 30*24*3600, "/", "", true, true)
	}

	// 初回ログインで users 行が無いと /auth/me が 404 になるため自動で upsert する。
	// id_token の payload から sub / email / cognito:groups を取り出して同期する:
	//   - 未登録なら role = (group に ADMIN なら super_admin / 無ければ trainee) で create
	//   - 既存ユーザーは Cognito group が変わっていれば role を update する
	// これで Spring Boot 時代と同じ「Cognito group が真のソース」で運用できる。
	if claims, err := decodeJWTClaims(tok.IDToken); err == nil {
		sub, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)
		groups := middleware.ToStringSliceFromClaim(claims["cognito:groups"])
		desiredRole := domain.RoleTrainee
		if middleware.IsAdminFromGroups(groups) {
			desiredRole = domain.RoleSuperAdmin
		}
		if sub != "" && h.users != nil {
			existing, _ := h.users.FindByCognitoSub(c.Request.Context(), sub)
			if existing == nil {
				_ = h.users.Create(c.Request.Context(), &domain.User{
					CognitoSub:  sub,
					Email:       email,
					DisplayName: email,
					Role:        desiredRole,
				})
			} else if existing.Role != desiredRole && desiredRole == domain.RoleSuperAdmin {
				// Cognito group で admin に昇格された場合のみ DB role を上書きする。
				// 既に DB で super_admin / company_admin が設定されているケースは触らない
				// （AdminInvitation 系の手動付与が他にある可能性があるため、降格は手動で）。
				_ = h.users.UpdateRole(c.Request.Context(), existing.ID, desiredRole)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "ログインしました。"})
}

// decodeJWTClaims は JWT (3 セグメント、署名検証なし) の payload 部だけ base64 デコードする。
// callback 直後に Cognito 発行 token から sub / email を取り出して users 自動作成するためだけに使う。
// 認証 middleware の本格的な署名検証 (JWKS) は別 issue。
func decodeJWTClaims(idToken string) (map[string]any, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt format")
	}
	payload, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

func base64URLDecode(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	s = strings.NewReplacer("-", "+", "_", "/").Replace(s)
	return base64.StdEncoding.DecodeString(s)
}

// Refresh は HttpOnly Cookie の refresh_token を使ってアクセストークンを再発行する。
func (h *AuthHandler) Refresh(c *gin.Context) {
	rt, err := c.Cookie("refresh_token")
	if err != nil || rt == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_token_missing"})
		return
	}
	if h.cognito == nil || h.cognito.TokenURI == "" || h.cognito.ClientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cognito_not_configured"})
		return
	}

	// Callback と同じく body 方式に統一する。
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", h.cognito.ClientID)
	if h.cognito.ClientSecret != "" {
		form.Set("client_secret", h.cognito.ClientSecret)
	}
	form.Set("refresh_token", rt)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost,
		h.cognito.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "request_build_failed"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("cognito refresh: token endpoint unreachable: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "cognito_unreachable"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("cognito refresh: status=%d body=%s", resp.StatusCode, string(body))
		// refresh が無効ならログイン状態をクリアして 401 を返し、フロントは login へ誘導する
		c.SetCookie(middleware.CookieAccessToken, "", -1, "/", "", true, true)
		c.SetCookie("refresh_token", "", -1, "/", "", true, true)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_failed"})
		return
	}

	var tok cognitoTokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid_token_response"})
		return
	}
	maxAge := tok.ExpiresIn
	if maxAge <= 0 {
		maxAge = 3600
	}
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(middleware.CookieAccessToken, tok.AccessToken, maxAge, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "refreshed"})
}
