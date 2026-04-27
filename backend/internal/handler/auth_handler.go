package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AuthHandler は Cognito 関連の認証エンドポイントを提供する。
// Spring Boot の CognitoAuthController に相当。
type AuthHandler struct {
	getCurrentUser *usecase.GetCurrentUserUseCase
	cognito        *config.CognitoConfig
	httpClient     *http.Client
}

func NewAuthHandler(getCurrentUser *usecase.GetCurrentUserUseCase, cognito *config.CognitoConfig) *AuthHandler {
	return &AuthHandler{
		getCurrentUser: getCurrentUser,
		cognito:        cognito,
		httpClient:     &http.Client{},
	}
}

// Me は現在ログイン中のユーザー情報を返す。
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
	c.JSON(http.StatusOK, user)
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

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", h.cognito.ClientID)
	form.Set("code", req.Code)
	form.Set("redirect_uri", h.cognito.RedirectURI)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost,
		h.cognito.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "request_build_failed"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if h.cognito.ClientSecret != "" {
		basic := base64.StdEncoding.EncodeToString(
			[]byte(fmt.Sprintf("%s:%s", h.cognito.ClientID, h.cognito.ClientSecret)))
		httpReq.Header.Set("Authorization", "Basic "+basic)
	}

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "cognito_unreachable"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token_exchange_failed", "detail": string(body)})
		return
	}

	var tok cognitoTokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
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
	c.JSON(http.StatusOK, gin.H{"message": "ログインしました。"})
}
