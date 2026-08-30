package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
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
	// RespondToNewPassword は NEW_PASSWORD_REQUIRED チャレンジに新パスワードで応答する
	// （一時パスワードでの初回ログイン）。session は Authenticate が返した
	// *cognito.NewPasswordRequiredError の Session。
	RespondToNewPassword(ctx context.Context, email, session, newPassword string) (*cognito.Token, error)
}

// AuthHandler は Cognito 関連の認証エンドポイントを提供する。
// OAuth2 通信は infra/cognito.TokenExchanger に切り出し、ここは HTTP 境界と user upsert だけを持つ。
type AuthHandler struct {
	getCurrentUser          *usecase.GetCurrentUserUseCase
	upsertUser              *usecase.UpsertUserFromIDTokenUseCase
	ensurePersonalWorkspace *usecase.EnsurePersonalWorkspaceUseCase
	promoteAdmin            *usecase.PromoteCognitoAdminRoleUseCase
	cognitoCfg              *config.CognitoConfig
	tokens                  *cognito.TokenExchanger
	passwordAuth            passwordAuthenticator
}

// NewAuthHandler は本番用に http.Client + 10s timeout の TokenExchanger を組み立てて DI する。
func NewAuthHandler(
	getCurrentUser *usecase.GetCurrentUserUseCase,
	upsertUser *usecase.UpsertUserFromIDTokenUseCase,
	ensurePersonalWorkspace *usecase.EnsurePersonalWorkspaceUseCase,
	promoteAdmin *usecase.PromoteCognitoAdminRoleUseCase,
	cognitoCfg *config.CognitoConfig,
	passwordAuth passwordAuthenticator,
) *AuthHandler {
	return &AuthHandler{
		getCurrentUser:          getCurrentUser,
		upsertUser:              upsertUser,
		ensurePersonalWorkspace: ensurePersonalWorkspace,
		promoteAdmin:            promoteAdmin,
		cognitoCfg:              cognitoCfg,
		passwordAuth:            passwordAuth,
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
//	@Description  Cookie 認証 を 元 に 現在 ログイン 中 の user 情報 (id / email / role / isAdmin 等) を 返す。
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
	// 昇格できたらこのレスポンスの role も揃える。捨てると、初回だけ
	// 「isAdmin=true / role=trainee」という起き得ない組み合わせをフロントへ返してしまう
	// （次回の /auth/me では直るが、role を見て画面を出し分けている箇所が初回だけ食い違う）。
	if middleware.IsAdminFromGroups(groups) && user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleCompanyAdmin {
		if h.promoteCognitoAdmin(c, sub.(string)) {
			user.Role = domain.RoleSuperAdmin
		}
	}
	resp := gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"role":      user.Role,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
		"groups":    groups,
		"isAdmin":   isAdmin,
	}
	// companyId / workspaceId は nil 時に JSON フィールド自体を省略する（omitempty 相当）。
	if user.CompanyID != nil {
		resp["companyId"] = user.CompanyID
	}
	if user.WorkspaceID != nil {
		resp["workspaceId"] = user.WorkspaceID
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
// Hosted UI を経由しないアプリ内ログインフォーム用。ユーザー作成は Callback と同じく upsert 側で行う。
//
//	@Summary      ログイン (メール / パスワード)
//	@Description  email / password を Cognito の USER_PASSWORD_AUTH で 検証 し、 access / refresh token を HttpOnly Cookie で 返す。 招待が無くても新規 user を自己サインアップとして作成する（Cognito admin group だけでは昇格しない）。
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      passwordLoginReq  true  "メール / パスワード"
//	@Success      200   {object}  messageResponse
//	@Failure      400   {object}  errorResponse  "入力 不正 (email 形式 / password 欠落)"
//	@Failure      401   {object}  errorResponse  "資格 情報 誤り"
//	@Failure      403   {object}  errorResponse  "最初の運営管理者作成の競合負け"
//	@Failure      409   {object}  errorResponse  "同じ email での同時サインアップ競合"
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
		// 一時パスワードでの初回ログインはチャレンジが返る。session をフロントへ渡し、
		// 新パスワード設定（/auth/cognito/new-password）へ誘導する（トークンはまだ発行しない）。
		var challenge *cognito.NewPasswordRequiredError
		if errors.As(err, &challenge) {
			c.JSON(http.StatusOK, gin.H{
				"challenge": "NEW_PASSWORD_REQUIRED",
				"session":   challenge.Session,
			})
			return
		}
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

	h.finishPasswordLogin(c, tok)
}

// finishPasswordLogin は取得済みトークンでユーザーを解決し、Cookie を発行する。
// パスワードログイン（Login）と新パスワード設定（NewPassword）で共有する。
func (h *AuthHandler) finishPasswordLogin(c *gin.Context, tok *cognito.Token) {
	user, upErr := h.upsertUserFromIDToken(c, tok.IDToken, "")
	if !h.respondUpsertOutcome(c, "cognito password login", tok, user, upErr) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ログインしました。"})
}

// respondUpsertOutcome は upsertUserFromIDToken の結果を HTTP レスポンスへ変換する。
// 成功時のみ Cookie を発行して true を返す。false を返したときは呼び出し元がそのまま return する。
//
// upErr は原因ごとに扱いを分ける: 内部エラー(DB/decode)は 500、同じ email での同時サインアップ
// 競合(repository.ErrEmailTaken)は 409、最初の運営管理者作成の競合負け(user == nil, err == nil)は
// 403。ErrEmailTaken を bootstrap 競合負けと同じ 403 に丸めると、招待とは無関係の一時的な
// 二重送信なのに「招待を受けてください」という見当違いの案内を返してしまう。
func (h *AuthHandler) respondUpsertOutcome(c *gin.Context, logPrefix string, tok *cognito.Token, user *domain.User, upErr error) bool {
	if upErr != nil {
		if errors.Is(upErr, repository.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "email_taken",
				"message": "同じメールアドレスでの登録が別のリクエストで同時に完了しました。もう一度ログインし直してください。",
			})
			return false
		}
		log.Printf("%s: upsert failed: %v", logPrefix, upErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return false
	}
	if user == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "bootstrap_admin_race_lost",
			"message": "最初の運営管理者は既に作成されています。招待を受けてログインしてください。",
		})
		return false
	}

	middleware.SetAccessTokenCookie(c, tok.AccessToken, tok.ExpiresIn)
	middleware.SetRefreshTokenCookie(c, tok.RefreshToken)
	return true
}

type newPasswordReq struct {
	Email       string `json:"email" binding:"required,email" format:"email"`
	Session     string `json:"session" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// NewPassword は NEW_PASSWORD_REQUIRED チャレンジに新パスワードで応答し、
// 成功したらパスワードログインと同じく Cookie を発行する（一時パスワードでの初回ログイン）。
//
//	@Summary      初回パスワード設定（一時パスワードログイン）
//	@Description  一時パスワードでの初回ログイン時に返る NEW_PASSWORD_REQUIRED チャレンジへ
//	@Description  新パスワードで応答する。成功で認証 Cookie を発行する。
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      newPasswordReq  true  "email / session / 新パスワード"
//	@Success      200   {object}  messageResponse  "設定してログイン"
//	@Failure      400   {object}  errorResponse    "入力エラー / パスワードポリシー違反"
//	@Failure      401   {object}  errorResponse    "session 失効等"
//	@Failure      403   {object}  errorResponse    "最初の運営管理者作成の競合負け"
//	@Failure      409   {object}  errorResponse    "同じ email での同時サインアップ競合"
//	@Failure      429   {object}  errorResponse    "レート制限超過"
//	@Router       /auth/cognito/new-password [post]
func (h *AuthHandler) NewPassword(c *gin.Context) {
	var req newPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h.passwordAuth == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cognito_not_configured"})
		return
	}

	tok, err := h.passwordAuth.RespondToNewPassword(c.Request.Context(), req.Email, req.Session, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, cognito.ErrInvalidCredentials):
			// session 失効・不正など。再ログインを促す。
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_session"})
		case errors.Is(err, cognito.ErrNotConfigured):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cognito_not_configured"})
		default:
			// パスワードポリシー違反などは 400。詳細メッセージはユーザーに出さず code のみ。
			log.Printf("cognito new password: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_new_password"})
		}
		return
	}

	h.finishPasswordLogin(c, tok)
}

type cognitoCallbackReq struct {
	Code string `json:"code" binding:"required"`
	// InvitationToken は招待マジックリンク経由の UUID（任意）。指定時は email 検索より優先して照合する。
	InvitationToken string `json:"invitationToken"`
}

// Callback は認可コードを token に交換して HttpOnly Cookie に格納する。
// 招待が無くても新規ユーザーを自己サインアップとして作成する（招待は役割・所属先の指定としてだけ働く）。
//
//	@Summary      ログイン (認可 コード → token 交換)
//	@Description  Cognito Hosted UI から の callback。 authorization code を access / refresh / id token に 交換 し HttpOnly Cookie で 返す。 招待が無くても新規 user を自己サインアップとして作成する（Cognito admin group だけでは昇格しない）。
//	@Tags         auth
//	@Accept       json
//	@Produce      json
//	@Param        body  body      cognitoCallbackReq  true  "Cognito callback (code 必須、 invitationToken 任意)"
//	@Success      200   {object}  messageResponse
//	@Failure      400   {object}  errorResponse  "code 欠落 等"
//	@Failure      401   {object}  errorResponse  "token 交換 失敗"
//	@Failure      403   {object}  errorResponse  "最初の運営管理者作成の競合負け"
//	@Failure      409   {object}  errorResponse  "同じ email での同時サインアップ競合"
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

	// 初回ログインで users 行が無いと /auth/me が 404 になるため upsert する
	// （招待が無くても個人サインアップとして新規作成する。招待は役割・所属先の指定としてだけ働く）。
	user, upErr := h.upsertUserFromIDToken(c, tok.IDToken, req.InvitationToken)
	if !h.respondUpsertOutcome(c, "cognito callback", tok, user, upErr) {
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
		// refresh は既存ユーザー前提。role 同期の best-effort なのでレスポンスは変えないが、
		// 失敗は握り潰さずログに残す（恒久的に失敗し続ける状態に気付けるようにする）。
		if _, err := h.upsertUserFromIDToken(c, tok.IDToken, ""); err != nil {
			slog.WarnContext(c.Request.Context(), "refresh: user upsert failed", "err", err)
		}
	} else {
		h.syncRoleFromAccessToken(c, tok.AccessToken)
	}
	c.JSON(http.StatusOK, gin.H{"message": "refreshed"})
}

// syncRoleFromAccessToken は access_token の cognito:groups を見て DB role を super_admin に昇格する。
// ID token に groups が含まれない Google federated ユーザー向けのフォールバック。
func (h *AuthHandler) syncRoleFromAccessToken(c *gin.Context, accessToken string) {
	claims, err := middleware.DecodeClaims(accessToken)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "refresh: access_token decode failed", "err", err)
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
	h.promoteCognitoAdmin(c, sub)
}

// promoteCognitoAdmin は Cognito admin グループのユーザーを super_admin へ同期する（昇格のみ）。
// 戻り値は実際に昇格したか（未配線・失敗・既に管理者なら false）。
// 失敗してもレスポンスのステータスは変えない（本人の閲覧・refresh は妨げない）が、必ずログに残す。
// role 名の解決失敗のような恒久エラーを握り潰すと、そのユーザーは「UI 上は管理者・API は 403」の
// 壊れた状態にログすら残さず留まり続けるため。
func (h *AuthHandler) promoteCognitoAdmin(c *gin.Context, cognitoSub string) bool {
	if h.promoteAdmin == nil {
		return false
	}
	ctx := c.Request.Context()
	promoted, err := h.promoteAdmin.Execute(ctx, usecase.PromoteCognitoAdminRoleInput{
		CognitoSub: cognitoSub,
	})
	if err != nil {
		slog.ErrorContext(ctx, "cognito admin role sync failed", "cognitoSub", cognitoSub, "err", err)
		return false
	}
	return promoted
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

// upsertUserFromIDToken はIDトークンから認証情報を取得し、ユーザー更新をusecaseへ委譲する。
// 続けて個人ワークスペースの確保まで行う（無ければ作る。既存なら 1 回の SELECT で終わる）。
// user が nil かつ err が nil のときだけ、最初の運営管理者作成の競合負けで弾かれたことを表す
// （招待必須のゲートは撤去済みで、それ以外の新規ユーザーはここで弾かれない）。
func (h *AuthHandler) upsertUserFromIDToken(
	c *gin.Context,
	idToken string,
	invitationToken string,
) (user *domain.User, err error) {
	claims, decodeErr := middleware.DecodeClaims(idToken)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode id_token: %w", decodeErr)
	}

	if h.upsertUser == nil {
		return nil, errors.New("upsert user usecase not configured")
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	groups := middleware.ToStringSliceFromClaim(claims["cognito:groups"])
	isCognitoAdmin := middleware.IsAdminFromGroups(groups)

	user, err = h.upsertUser.Execute(
		c.Request.Context(),
		usecase.UpsertUserFromIDTokenInput{
			CognitoSub:      sub,
			Email:           email,
			Name:            name,
			IsCognitoAdmin:  isCognitoAdmin,
			InvitationToken: invitationToken,
		},
	)
	if err != nil || user == nil {
		return nil, err
	}

	// 失敗してもログインは失敗させない（次回ログイン時に自己修復する）。
	if h.ensurePersonalWorkspace != nil {
		if _, wsErr := h.ensurePersonalWorkspace.Execute(
			c.Request.Context(),
			usecase.EnsurePersonalWorkspaceInput{UserID: user.ID, Name: user.Name},
		); wsErr != nil {
			slog.ErrorContext(c.Request.Context(), "ensure personal workspace failed (non-fatal)", "userID", user.ID, "err", wsErr)
		}
	}

	return user, nil
}
