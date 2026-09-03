package handler

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/infra/oidc"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// AuthHandler は認証エンドポイントを提供する。
// 発行者との通信は infra/oidc に切り出し、ここは HTTP の境界とユーザーの upsert だけを持つ。
type AuthHandler struct {
	getCurrentUser          *usecase.GetCurrentUserUseCase
	upsertUser              *usecase.UpsertUserFromIDTokenUseCase
	ensurePersonalWorkspace *usecase.EnsurePersonalWorkspaceUseCase
	promoteAdmin            *usecase.PromoteCognitoAdminRoleUseCase
	oidcCfg                 *config.OIDCConfig
	tokens                  *oidc.TokenExchanger
	verifier                *oidc.Verifier
}

// NewAuthHandler は AuthHandler を組み立てる。
func NewAuthHandler(
	getCurrentUser *usecase.GetCurrentUserUseCase,
	upsertUser *usecase.UpsertUserFromIDTokenUseCase,
	ensurePersonalWorkspace *usecase.EnsurePersonalWorkspaceUseCase,
	promoteAdmin *usecase.PromoteCognitoAdminRoleUseCase,
	oidcCfg *config.OIDCConfig,
	verifier *oidc.Verifier,
) *AuthHandler {
	return &AuthHandler{
		getCurrentUser:          getCurrentUser,
		upsertUser:              upsertUser,
		ensurePersonalWorkspace: ensurePersonalWorkspace,
		promoteAdmin:            promoteAdmin,
		oidcCfg:                 oidcCfg,
		verifier:                verifier,
		tokens: oidc.NewTokenExchanger(oidc.ExchangerConfig{
			ClientID:     oidcCfg.ClientID,
			ClientSecret: oidcCfg.ClientSecret,
			RedirectURI:  oidcCfg.RedirectURI,
			TokenURI:     oidcCfg.TokenURI,
		}),
	}
}

// Me は現在ログイン中のユーザー情報（+ 派生 isAdmin / roles）を返す。
// isAdmin は発行者の役割に管理者が含まれるか、DB role が super_admin / company_admin なら true。
func (h *AuthHandler) Me(c *gin.Context) {
	sub, ok := c.Get(middleware.ContextKeySubject)
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
	roles := middleware.RolesFromContext(c)
	isIdPAdmin := oidc.HasRole(roles, h.adminRole())
	isAdmin := isIdPAdmin ||
		user.Role == domain.RoleSuperAdmin ||
		user.Role == domain.RoleCompanyAdmin
	// 発行者側では管理者だが DB role が未昇格なら同期する。捨てると、初回だけ
	// 「isAdmin=true / role=trainee」という起き得ない組み合わせをフロントへ返してしまう
	// （次回の /auth/me では直るが、role を見て画面を出し分けている箇所が初回だけ食い違う）。
	if isIdPAdmin && user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleCompanyAdmin {
		if h.promoteIdPAdmin(c, sub.(string)) {
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
		"groups":    roles,
		"isAdmin":   isAdmin,
	}
	// workspaceId は nil 時に JSON フィールド自体を省略する（omitempty 相当）。
	if user.WorkspaceID != nil {
		resp["workspaceId"] = user.WorkspaceID
	}
	c.JSON(http.StatusOK, resp)
}

// logoutResponse は Cookie 消去に続けてフロントが踏むべき URL を返す。
type logoutResponse struct {
	Message string `json:"message"`
	// EndSessionURL は発行者側のセッションも終わらせるための遷移先（設定が無ければ空）。
	EndSessionURL string `json:"endSessionUrl,omitempty"`
}

// Logout は認証 Cookie を消去し、発行者側のセッション終了先を返す。
//
// Cookie を消すだけでは、発行者の側にはログイン済みのセッションが残る。同じ端末で
// もう一度ログインを始めると、ログイン画面すら出ずにそのまま入り直せてしまう。
// 共用端末では、前の人のアカウントに次の人が入れることになる。
func (h *AuthHandler) Logout(c *gin.Context) {
	middleware.ClearAuthCookies(c)
	c.JSON(http.StatusOK, logoutResponse{
		Message:       "ログアウトしました。",
		EndSessionURL: h.oidcCfg.EndSessionURI,
	})
}

type callbackReq struct {
	Code string `json:"code" binding:"required"`
	// CodeVerifier は PKCE の検証値。認可を始めたブラウザが作って手元に置いた乱数で、
	// 発行者がこれと認可要求に載った要約を突き合わせる。
	CodeVerifier string `json:"codeVerifier" binding:"required"`
	// Nonce は認可を始めたブラウザが作った値。id_token の中身と一致することを確かめる。
	Nonce string `json:"nonce" binding:"required"`
	// InvitationToken は招待マジックリンク経由の UUID（任意）。指定時は email 検索より優先して照合する。
	InvitationToken string `json:"invitationToken"`
}

// Callback は認可コードを token に交換して HttpOnly Cookie に格納する。
// 招待が無くても新規ユーザーを自己サインアップとして作成する（招待は役割・所属先の指定としてだけ働く）。
func (h *AuthHandler) Callback(c *gin.Context) {
	var req callbackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	tok, err := h.tokens.ExchangeAuthorizationCode(c.Request.Context(), req.Code, req.CodeVerifier)
	if status, body, ok := h.handleTokenError(c, "callback", err); ok {
		c.JSON(status, body)
		return
	}

	// 初回ログインで users 行が無いと /auth/me が 404 になるため upsert する
	// （招待が無くても個人サインアップとして新規作成する。招待は役割・所属先の指定としてだけ働く）。
	user, upErr := h.upsertUserFromIDToken(c, tok.IDToken, req.Nonce, req.InvitationToken)
	if !h.respondUpsertOutcome(c, "callback", tok, user, upErr) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ログインしました。"})
}

// respondUpsertOutcome は upsertUserFromIDToken の結果を HTTP レスポンスへ変換する。
// 成功時のみ Cookie を発行して true を返す。false を返したときは呼び出し元がそのまま return する。
//
// upErr は原因ごとに扱いを分ける: id_token の検証失敗は 401、内部エラー(DB)は 500、
// 同じ email での同時サインアップ競合(repository.ErrEmailTaken)は 409、
// 最初の運営管理者作成の競合負け(user == nil, err == nil)は 403。
// ErrEmailTaken を bootstrap 競合負けと同じ 403 に丸めると、招待とは無関係の一時的な
// 二重送信なのに「招待を受けてください」という見当違いの案内を返してしまう。
func (h *AuthHandler) respondUpsertOutcome(c *gin.Context, logPrefix string, tok *oidc.Token, user *domain.User, upErr error) bool {
	if upErr != nil {
		if errors.Is(upErr, repository.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "email_taken",
				"message": "同じメールアドレスでの登録が別のリクエストで同時に完了しました。もう一度ログインし直してください。",
			})
			return false
		}
		if errors.Is(upErr, errIDTokenRejected) {
			log.Printf("%s: id_token rejected: %v", logPrefix, upErr)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_id_token"})
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

// Refresh は HttpOnly Cookie の refresh_token を使ってアクセストークンを再発行する。
func (h *AuthHandler) Refresh(c *gin.Context) {
	rt, err := c.Cookie(middleware.CookieRefreshToken)
	if err != nil || rt == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_token_missing"})
		return
	}

	tok, err := h.tokens.RefreshAccessToken(c.Request.Context(), rt)
	if err != nil {
		var exErr *oidc.TokenExchangeError
		if errors.As(err, &exErr) {
			log.Printf("refresh: status=%d body=%s", exErr.HTTPStatus, exErr.Body)
			// **Cookie を消すのは「この refresh_token はもう使えない」と分かったときだけ。**
			//
			// TokenExchangeError は 200 以外のすべてを表すので、429（絞られた）や
			// 5xx（発行者が一時的に落ちている）でも同じ扱いにすると、
			// 発行者の短い不調が「全利用者の強制ログアウト」に化ける。
			// 手元のトークンはまだ有効なのに、こちらから捨てることになる。
			if isUnrecoverableGrantError(exErr) {
				middleware.ClearAuthCookies(c)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_failed"})
				return
			}
			// 一時的な失敗。Cookie は残し、後でもう一度試せるようにする。
			c.JSON(http.StatusBadGateway, gin.H{"error": "idp_unreachable"})
			return
		}
		status, body, _ := h.handleTokenError(c, "refresh", err)
		c.JSON(status, body)
		return
	}

	middleware.SetAccessTokenCookie(c, tok.AccessToken, tok.ExpiresIn)
	// **回転した refresh_token を必ず書き戻す。**
	//
	// 発行者によっては、交換のたびに refresh_token 自体が新しいものへ入れ替わる。
	// 書き戻さないと Cookie には使用済みの値が残り、次の更新で「使い回し」と見なされて
	// 失敗する。しかも多くの実装は使い回しをトークン窃取の兆候として扱い、
	// そのトークン系列をまとめて失効させる。結果、2 回目の更新で全員が
	// ログイン画面へ飛ばされ、書きかけの内容が消える。
	// 空文字のときは何もしない（SetRefreshTokenCookie 側で握る）。
	middleware.SetRefreshTokenCookie(c, tok.RefreshToken)

	// id_token があれば DB role を同期する。無ければ access_token の検証済みクレームから昇格を試みる。
	// refresh は既存ユーザー前提なので、失敗してもレスポンスは変えないがログには残す
	// （恒久的に失敗し続ける状態に気付けるようにする）。
	//
	// nonce は空を渡す。nonce は「認可を始めた本人か」を確かめるためのもので、
	// 更新の応答には対応する認可要求が無い（そもそも発行者が nonce を載せない）。
	if tok.IDToken != "" {
		if _, err := h.upsertUserFromIDToken(c, tok.IDToken, "", ""); err != nil {
			slog.WarnContext(c.Request.Context(), "refresh: user upsert failed", "err", err)
		}
	} else {
		h.syncRoleFromAccessToken(c, tok.AccessToken)
	}
	c.JSON(http.StatusOK, gin.H{"message": "refreshed"})
}

// syncRoleFromAccessToken は access_token の役割を見て DB role を super_admin に昇格する。
// id_token に役割が含まれない構成向けのフォールバック。
//
// **署名を検証したうえで読む。** 以前はここも id_token の読み取りも署名を確かめずに
// payload をデコードしていた。役割の昇格という、最も強い権限の入口を、
// 誰でも書き換えられる値で駆動していたことになる。
func (h *AuthHandler) syncRoleFromAccessToken(c *gin.Context, accessToken string) {
	ctx := c.Request.Context()
	claims, err := h.verifier.Verify(ctx, accessToken)
	if err != nil {
		slog.WarnContext(ctx, "refresh: access_token verify failed", "err", err)
		return
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return
	}
	roles := oidc.RolesFromClaim(claims[h.rolesClaim()])
	if !oidc.HasRole(roles, h.adminRole()) {
		return
	}
	h.promoteIdPAdmin(c, sub)
}

// promoteIdPAdmin は発行者側で管理者の役割を持つユーザーを super_admin へ同期する（昇格のみ）。
// 戻り値は実際に昇格したか（未配線・失敗・既に管理者なら false）。
// 失敗してもレスポンスのステータスは変えない（本人の閲覧・refresh は妨げない）が、必ずログに残す。
// role 名の解決失敗のような恒久エラーを握り潰すと、そのユーザーは「UI 上は管理者・API は 403」の
// 壊れた状態にログすら残さず留まり続けるため。
func (h *AuthHandler) promoteIdPAdmin(c *gin.Context, subject string) bool {
	if h.promoteAdmin == nil {
		return false
	}
	ctx := c.Request.Context()
	promoted, err := h.promoteAdmin.Execute(ctx, usecase.PromoteCognitoAdminRoleInput{
		CognitoSub: subject,
	})
	if err != nil {
		slog.ErrorContext(ctx, "idp admin role sync failed", "subject", subject, "err", err)
		return false
	}
	return promoted
}

// rolesClaim / adminRole は設定の既定を 1 か所に閉じる。
func (h *AuthHandler) rolesClaim() string { return h.oidcCfg.AdminRoleClaim }
func (h *AuthHandler) adminRole() string  { return h.oidcCfg.AdminRole }

// handleTokenError は TokenExchanger が返したエラーを HTTP レスポンスに変換する。
// returned ok=true なら呼び元は早期 return する想定。
func (h *AuthHandler) handleTokenError(c *gin.Context, op string, err error) (int, gin.H, bool) {
	if err == nil {
		return 0, nil, false
	}

	var exErr *oidc.TokenExchangeError
	switch {
	case errors.Is(err, oidc.ErrNotConfigured):
		return http.StatusInternalServerError, gin.H{"error": "oidc_not_configured"}, true
	case errors.As(err, &exErr):
		// 本物の理由は log に残し、クライアントには簡素なエラーだけ返す。
		log.Printf("%s: token exchange status=%d body=%s redirect_uri=%s client_id_set=%t client_secret_set=%t",
			op, exErr.HTTPStatus, exErr.Body, h.oidcCfg.RedirectURI, h.oidcCfg.ClientID != "", h.oidcCfg.ClientSecret != "")
		return http.StatusUnauthorized, gin.H{"error": "token_exchange_failed"}, true
	case errors.Is(err, oidc.ErrUnreachable):
		log.Printf("%s: token endpoint unreachable: %v", op, err)
		return http.StatusBadGateway, gin.H{"error": "idp_unreachable"}, true
	case errors.Is(err, oidc.ErrInvalidResponse):
		log.Printf("%s: invalid token response: %v", op, err)
		return http.StatusBadGateway, gin.H{"error": "invalid_token_response"}, true
	default:
		log.Printf("%s: unexpected error: %v", op, err)
		return http.StatusInternalServerError, gin.H{"error": "internal_error"}, true
	}
}

// isUnrecoverableGrantError は「その refresh_token はもう使えない」と言い切れる失敗かを返す。
//
// **状態コードだけでは決められない。** OAuth2 は grant が無効なときに 400 を返すが
// （RFC 6749 §5.2）、同じ 400 は invalid_request（こちらの組み立て方が悪い）でも返るし、
// 401 は invalid_client（クライアントの設定が悪い）で返る。どちらも設定の問題であって、
// 利用者の refresh_token は生きている。ここで消すと、設定を 1 つ間違えた瞬間に
// 全利用者がログアウトさせられる。
//
// 消してよいのは error が invalid_grant のときだけ。本文が読めない・別の error なら
// 消さない（分からないときは手元の資格を残す側に倒す）。
func isUnrecoverableGrantError(e *oidc.TokenExchangeError) bool {
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(e.Body), &body); err != nil {
		return false
	}
	return body.Error == "invalid_grant"
}

// errIDTokenRejected は id_token の署名・クレーム検証に落ちたことを表す。
// respondUpsertOutcome がこれを 401 に変換する（DB 障害の 500 と区別する）。
var errIDTokenRejected = errors.New("handler: id_token rejected")

// upsertUserFromIDToken は id_token を検証してユーザー更新を usecase へ委譲する。
// 続けて個人ワークスペースの確保まで行う（無ければ作る。既存なら 1 回の SELECT で終わる）。
// user が nil かつ err が nil のときだけ、最初の運営管理者作成の競合負けで弾かれたことを表す
// （招待必須のゲートは撤去済みで、それ以外の新規ユーザーはここで弾かれない）。
func (h *AuthHandler) upsertUserFromIDToken(
	c *gin.Context,
	idToken string,
	expectedNonce string,
	invitationToken string,
) (user *domain.User, err error) {
	if h.upsertUser == nil {
		return nil, errors.New("upsert user usecase not configured")
	}

	// **署名とクレームを検証してから読む。**
	// ここで作られるのはユーザーそのもの（sub / email / 役割）で、検証せずに読むと
	// 「好きな sub と email を名乗って新しいユーザーを作る」ことができてしまう。
	claims, verifyErr := h.verifier.VerifyIDToken(c.Request.Context(), idToken, expectedNonce)
	if verifyErr != nil {
		return nil, errors.Join(errIDTokenRejected, verifyErr)
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	roles := oidc.RolesFromClaim(claims[h.rolesClaim()])
	isIdPAdmin := oidc.HasRole(roles, h.adminRole())

	user, err = h.upsertUser.Execute(
		c.Request.Context(),
		usecase.UpsertUserFromIDTokenInput{
			CognitoSub:      sub,
			Email:           email,
			Name:            name,
			IsCognitoAdmin:  isIdPAdmin,
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
