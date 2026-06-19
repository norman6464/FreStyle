package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// passwordAuthenticator は email / password を Cognito で検証して token を返す境界。
// infra/cognito.PasswordAuthenticator が実装し、テストでは fake を注入する。
type passwordAuthenticator interface {
	Authenticate(ctx context.Context, email, password string) (*cognito.Token, error)
}

// AuthHandler は Cognito 関連の認証エンドポイントを提供する。
// OAuth2 通信は infra/cognito.TokenExchanger に切り出し、ここは HTTP 境界と user upsert だけを持つ。
type AuthHandler struct {
	getCurrentUser *usecase.GetCurrentUserUseCase
	users          repository.UserRepository
	invitations    repository.AdminInvitationRepository
	cognitoCfg     *config.CognitoConfig
	tokens         *cognito.TokenExchanger
	passwordAuth   passwordAuthenticator
	aiChatAccess   *usecase.AiChatEnabledForUserUseCase
}

// NewAuthHandler は本番用に http.Client + 10s timeout の TokenExchanger を組み立てて DI する。
// invitations は招待受諾フロー（初回ログイン時に invitations から role/companyId を反映）に使う。nil 可。
// aiChatAccess は /auth/me で aiChatEnabledForTrainees を算出するのに使う。nil 可（その場合は既定 true）。
func NewAuthHandler(
	getCurrentUser *usecase.GetCurrentUserUseCase,
	users repository.UserRepository,
	invitations repository.AdminInvitationRepository,
	cognitoCfg *config.CognitoConfig,
	passwordAuth passwordAuthenticator,
	aiChatAccess *usecase.AiChatEnabledForUserUseCase,
) *AuthHandler {
	return &AuthHandler{
		getCurrentUser: getCurrentUser,
		users:          users,
		invitations:    invitations,
		cognitoCfg:     cognitoCfg,
		passwordAuth:   passwordAuth,
		aiChatAccess:   aiChatAccess,
		tokens: cognito.NewTokenExchanger(cognito.Config{
			ClientID:     cognitoCfg.ClientID,
			ClientSecret: cognitoCfg.ClientSecret,
			RedirectURI:  cognitoCfg.RedirectURI,
			TokenURI:     cognitoCfg.TokenURI,
		}),
	}
}

// Me は現在ログイン中のユーザー情報（+ 派生 isAdmin / groups）を返す。
// isAdmin は Cognito groups に "admin" を含むか、DB role が super_admin / company_admin なら true。
//
//	@Summary      current user 情報 取得
//	@Description  Cookie 認証 を 元 に 現在 ログイン 中 の user 情報 (id / email / role / isAdmin / onboarded 等) を 返す。
//	@Tags         auth
//	@Produce      json
//	@Success      200  {object}  meResponse
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      404  {object}  errorResponse  "DB に user が ない (Cognito 側 だけ 存在)"
//	@Failure      500  {object}  errorResponse  "DB / repository 取得 失敗"
//	@Router       /auth/me [get]
//	@Security     CookieAuth
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
	// Cognito group admin だが DB role が未昇格なら同期する（federated ユーザー対策）。
	if middleware.IsAdminFromGroups(groups) && user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleCompanyAdmin {
		if h.users != nil {
			_ = h.users.UpdateRole(c.Request.Context(), user.ID, domain.RoleSuperAdmin)
		}
	}
	// trainee への AI チャット表示判定。会社設定が無効なら false。算出失敗・未配線時は既定 true。
	aiEnabled := true
	if h.aiChatAccess != nil {
		if ok, err := h.aiChatAccess.Execute(c.Request.Context(), user); err == nil {
			aiEnabled = ok
		}
	}
	resp := gin.H{
		"id":                       user.ID,
		"cognitoSub":               user.CognitoSub,
		"email":                    user.Email,
		"name":                     user.Name,
		"role":                     user.Role,
		"createdAt":                user.CreatedAt,
		"updatedAt":                user.UpdatedAt,
		"groups":                   groups,
		"isAdmin":                  isAdmin,
		"onboarded":                user.OnboardedAt != nil,
		"aiChatEnabledForTrainees": aiEnabled,
	}
	// companyId は nil 時に JSON フィールド自体を省略する（omitempty 相当）。
	if user.CompanyID != nil {
		resp["companyId"] = user.CompanyID
	}
	c.JSON(http.StatusOK, resp)
}

// Logout はリフレッシュ・アクセストークンの Cookie を消去する。
//
//	@Summary      ログアウト
//	@Description  HttpOnly Cookie の access / refresh token を 消去 する。 Cognito 側 の セッション は 別途 hosted UI で 切る。
//	@Tags         auth
//	@Produce      json
//	@Success      200  {object}  messageResponse
//	@Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	middleware.ClearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "ログアウトしました。"})
}

type passwordLoginReq struct {
	Email    string `json:"email" binding:"required,email" format:"email"`
	Password string `json:"password" binding:"required"`
}

// Login は email / password を Cognito(USER_PASSWORD_AUTH)で検証して HttpOnly Cookie を発行する。
// Hosted UI を経由しないアプリ内ログインフォーム用。招待ゲートは Callback と同じく upsert 側で効かせる。
//
//	@Summary      ログイン (メール / パスワード)
//	@Description  email / password を Cognito の USER_PASSWORD_AUTH で 検証 し、 access / refresh token を HttpOnly Cookie で 返す。 新規 user は 招待 or Cognito admin group 必須。
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      passwordLoginReq  true  "メール / パスワード"
//	@Success      200   {object}  messageResponse
//	@Failure      400   {object}  errorResponse  "入力 不正 (email 形式 / password 欠落)"
//	@Failure      401   {object}  errorResponse  "資格 情報 誤り"
//	@Failure      403   {object}  errorResponse  "招待 なし の 新規 user"
//	@Failure      500   {object}  errorResponse  "内部 エラー (Cognito 未 設定 / DB 失敗 等)"
//	@Failure      502   {object}  errorResponse  "Cognito 到達 不可"
//	@Failure      429   {object}  errorResponse  "レート制限超過"
//	@Header       429  {string}  Retry-After  "再試行までの秒数 (例: 60)"
//	@Router       /auth/cognito/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req passwordLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.passwordAuth == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cognito_not_configured"})
		return
	}

	tok, err := h.passwordAuth.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, cognito.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials"})
		case errors.Is(err, cognito.ErrNotConfigured):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cognito_not_configured"})
		default:
			log.Printf("cognito password login: unexpected error: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "cognito_unreachable"})
		}
		return
	}

	middleware.SetAccessTokenCookie(c, tok.AccessToken, tok.ExpiresIn)
	middleware.SetRefreshTokenCookie(c, tok.RefreshToken)

	// Callback と同じ招待ゲート。内部エラー(DB/decode)は 500、招待拒否は 403 に切り分ける。
	allowed, upErr := h.upsertUserFromIDToken(c, tok.IDToken, "")
	if upErr != nil {
		log.Printf("cognito password login: upsert failed: %v", upErr)
		middleware.ClearAuthCookies(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !allowed {
		middleware.ClearAuthCookies(c)
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "invitation_required",
			"message": "FreStyle のご利用には管理者からの招待が必要です。招待メールに記載されたリンクからログインしてください。",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ログインしました。"})
}

type cognitoCallbackReq struct {
	Code string `json:"code" binding:"required"`
	// InvitationToken は招待マジックリンク経由の UUID（任意）。指定時は email 検索より優先して照合する。
	InvitationToken string `json:"invitationToken"`
}

// Callback は認可コードを token に交換して HttpOnly Cookie に格納する。
// 招待ゲート: 新規ユーザーは Cognito group admin か pending invitation 受信者でなければ 403 で拒否する。
//
//	@Summary      ログイン (認可 コード → token 交換)
//	@Description  Cognito Hosted UI から の callback。 authorization code を access / refresh / id token に 交換 し HttpOnly Cookie で 返す。 新規 user は 招待 or Cognito admin group 必須。
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      cognitoCallbackReq  true  "Cognito callback (code 必須、 invitationToken 任意)"
//	@Success      200   {object}  messageResponse
//	@Failure      400   {object}  errorResponse  "code 欠落 等"
//	@Failure      401   {object}  errorResponse  "token 交換 失敗"
//	@Failure      403   {object}  errorResponse  "招待 なし の 新規 user"
//	@Failure      500   {object}  errorResponse  "Cognito 未 設定 等 の 内部 エラー"
//	@Failure      502   {object}  errorResponse  "Cognito 到達 不可"
//	@Failure      429   {object}  errorResponse  "レート制限超過"
//	@Header       429  {string}  Retry-After  "再試行までの秒数 (例: 60)"
//	@Router       /auth/login [post]
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

	// 初回ログインで users 行が無いと /auth/me が 404 になるため upsert する。
	// 内部エラー(DB/decode)は 500、招待拒否は 403。拒否時は Cookie をクリアしてセッションを残さない。
	allowed, upErr := h.upsertUserFromIDToken(c, tok.IDToken, req.InvitationToken)
	if upErr != nil {
		log.Printf("cognito callback: upsert failed: %v", upErr)
		middleware.ClearAuthCookies(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if !allowed {
		middleware.ClearAuthCookies(c)
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "invitation_required",
			"message": "FreStyle のご利用には管理者からの招待が必要です。招待メールに記載されたリンクからログインしてください。",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ログインしました。"})
}

// Refresh は HttpOnly Cookie の refresh_token を使ってアクセストークンを再発行する。
//
//	@Summary      アクセス トークン リフレッシュ
//	@Description  refresh_token Cookie で access_token を 再 発行 し HttpOnly Cookie に セット する。 失敗 (refresh 切れ 等) は 401 で Cookie クリア。
//	@Tags         auth
//	@Produce      json
//	@Success      200  {object}  messageResponse
//	@Failure      401  {object}  errorResponse  "refresh_token 欠落 / 無効"
//	@Failure      502  {object}  errorResponse  "Cognito 到達 不可"
//	@Failure      429   {object}  errorResponse  "レート制限超過"
//	@Header       429  {string}  Retry-After  "再試行までの秒数 (例: 60)"
//	@Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	rt, err := c.Cookie(middleware.CookieRefreshToken)
	if err != nil || rt == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_token_missing"})
		return
	}

	tok, err := h.tokens.RefreshAccessToken(c.Request.Context(), rt)
	if err != nil {
		// refresh_token 無効は Cookie クリアして 401。それ以外（502 等）は Cookie を残す。
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
	// id_token があれば DB role を同期する。無ければ access_token の claims から昇格を試みる
	// （federated ユーザーは id_token に groups が無いことがある）。refresh は既存ユーザー前提。
	if tok.IDToken != "" {
		// refresh は既存ユーザー前提。role 同期の best-effort なので結果は無視する。
		_, _ = h.upsertUserFromIDToken(c, tok.IDToken, "")
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
		// 本物の理由は log に残し、クライアントには簡素なエラーだけ返す。
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

// upsertUserFromIDToken は id_token の claim から users 行を新規作成 / role 更新する。
// 戻り値は (allowed, err)。allowed=false かつ err=nil は「招待なし & Cognito admin でもない」新規
// ユーザーの **ログイン拒否（403）** を表す。err!=nil は decode 失敗 / DB 失敗等の **内部エラー（500）**で、
// 招待拒否と切り分けて呼び元がレスポンスを出し分けられるようにする。
// invitationToken 指定時は email より優先して照合する（同 email 複数 pending での誤一致防止）。
// 招待ゲートを handler 層に置くのは、Cookie 発行 / 401-403 を返す認可境界が HTTP 層だから。
func (h *AuthHandler) upsertUserFromIDToken(c *gin.Context, idToken, invitationToken string) (allowed bool, err error) {
	claims, derr := middleware.DecodeClaims(idToken)
	if derr != nil {
		return false, fmt.Errorf("decode id_token: %w", derr)
	}
	if h.users == nil {
		return false, errors.New("user repository not configured")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return false, errors.New("id_token missing sub")
	}
	email, _ := claims["email"].(string)
	// OIDC 経由は id_token に name を含む（Cognito SRP では無いこともあるので空許容）。
	oidcName, _ := claims["name"].(string)
	groups := middleware.ToStringSliceFromClaim(claims["cognito:groups"])
	isCognitoAdmin := middleware.IsAdminFromGroups(groups)

	// 招待検索: invitationToken 優先、無ければ email でフォールバック。
	var inv *domain.AdminInvitation
	if h.invitations != nil {
		if invitationToken != "" {
			inv, _ = h.invitations.FindPendingByToken(c.Request.Context(), invitationToken)
		}
		if inv == nil && email != "" {
			inv, _ = h.invitations.FindPendingByEmail(c.Request.Context(), email)
		}
	}

	existing, _ := h.users.FindByCognitoSub(c.Request.Context(), sub)
	if existing != nil {
		// Name が email のまま（旧フローの仮値）なら OIDC name で自動補正する。
		// ユーザーが明示的に書き換えている場合は触らない。
		if oidcName != "" && existing.Email != "" && existing.Name == existing.Email {
			if err := h.users.UpdateName(c.Request.Context(), existing.ID, oidcName); err != nil {
				log.Printf("upsertUserFromIDToken: backfill name failed userID=%d: %v", existing.ID, err)
			}
		}
		// role 更新の制約: super_admin は降格しない / 招待昇格は trainee → company_admin のみ /
		// Cognito group admin は super_admin に昇格する。
		if isCognitoAdmin && existing.Role != domain.RoleSuperAdmin {
			_ = h.users.UpdateRole(c.Request.Context(), existing.ID, domain.RoleSuperAdmin)
		}
		if inv != nil && existing.Role != domain.RoleSuperAdmin {
			// 昇格は trainee → company_admin だけ反映する。
			if existing.Role == domain.RoleTrainee && inv.Role == domain.RoleCompanyAdmin {
				if err := h.users.UpdateRole(c.Request.Context(), existing.ID, domain.RoleCompanyAdmin); err != nil {
					log.Printf("upsertUserFromIDToken: existing user role upgrade failed userID=%d: %v", existing.ID, err)
				} else {
					log.Printf("upsertUserFromIDToken: existing user upgraded trainee→company_admin userID=%d email=%s", existing.ID, email)
				}
			}
			// company 紐付け: 招待の company_id と異なる / 未設定なら更新する。
			if inv.CompanyID != 0 && (existing.CompanyID == nil || *existing.CompanyID != inv.CompanyID) {
				if err := h.users.UpdateCompanyID(c.Request.Context(), existing.ID, inv.CompanyID); err != nil {
					log.Printf("upsertUserFromIDToken: existing user company update failed userID=%d: %v", existing.ID, err)
				}
			}
			// 招待を accepted にマーク（再利用防止 + 監査）。
			_ = h.invitations.UpdateStatus(c.Request.Context(), inv.ID, domain.InvitationStatusAccepted)
		}
		return true, nil
	}

	// 新規ユーザー: 招待 OR Cognito group "admin" のいずれかが必要。両方無ければ拒否（403、内部エラーではない）。
	if !isCognitoAdmin && inv == nil {
		log.Printf("upsertUserFromIDToken: signup blocked — no invitation and not Cognito admin sub=%s email=%s token_provided=%t", sub, email, invitationToken != "")
		return false, nil
	}

	role := domain.RoleTrainee
	var companyID *uint64
	var acceptedInvID uint64
	// name の優先順位: 招待 name > OIDC name > email（フォールバック）。
	name := email
	if oidcName != "" {
		name = oidcName
	}

	if isCognitoAdmin {
		role = domain.RoleSuperAdmin
	}
	if inv != nil {
		if inv.Role == domain.RoleCompanyAdmin || inv.Role == domain.RoleTrainee {
			role = inv.Role
		}
		cid := inv.CompanyID
		companyID = &cid
		acceptedInvID = inv.ID
		if inv.Name != "" {
			name = inv.Name
		}
	}

	if err := h.users.Create(c.Request.Context(), &domain.User{
		CognitoSub: sub,
		Email:      email,
		Name:       name,
		Role:       role,
		CompanyID:  companyID,
	}); err != nil {
		log.Printf("upsertUserFromIDToken: create user failed sub=%s email=%s err=%v", sub, email, err)
		return false, fmt.Errorf("create user: %w", err)
	}
	// 招待を accepted にマーク（履歴・監査）。
	if h.invitations != nil && acceptedInvID != 0 {
		_ = h.invitations.UpdateStatus(c.Request.Context(), acceptedInvID, domain.InvitationStatusAccepted)
	}
	return true, nil
}
