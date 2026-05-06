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
	invitations    repository.AdminInvitationRepository
	cognitoCfg     *config.CognitoConfig
	tokens         *cognito.TokenExchanger
}

// NewAuthHandler は本番用に http.Client + 10s timeout の TokenExchanger を組み立てて DI する。
// invitations は招待受諾フロー（初回ログイン時に invitations から role/companyId を反映）に使う。nil 可。
func NewAuthHandler(
	getCurrentUser *usecase.GetCurrentUserUseCase,
	users repository.UserRepository,
	invitations repository.AdminInvitationRepository,
	cognitoCfg *config.CognitoConfig,
) *AuthHandler {
	return &AuthHandler{
		getCurrentUser: getCurrentUser,
		users:          users,
		invitations:    invitations,
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
	// InvitationToken はフロントが sessionStorage から復元してくる、招待マジックリンク経由で
	// 受領した UUID トークン。任意。指定がある場合は upsert 時に email ベースの招待検索より
	// 優先して照合に使う（同じ email に複数 pending invitation がある異常系での誤一致を防ぐ）。
	InvitationToken string `json:"invitationToken"`
}

// Callback は Cognito Hosted UI から戻ってきた認可コードをアクセストークンに交換し、
// HttpOnly Cookie に格納する。Spring Boot の CognitoAuthController#callback 相当。
//
// 招待ゲート: 新規ユーザー（DB に sub が無い）は「Cognito group admin」または
// 「pending な invitation 受信者」のいずれかでない限り、ログインを拒否する（403）。
// これによりトレイニー / company_admin は必ず上位ロールからの招待経由でないとアカウントを作れない。
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
	//   - 招待 OR Cognito group "admin" のいずれかが真なら create（role は招待 / group に従う）
	//   - 既存ユーザーは Cognito group が変わっていれば role を update する
	// 上記いずれでもない（=招待なし & Cognito admin でもない）新規ユーザーはログイン拒否し、
	// 直前にセットした Cookie をクリアしてセッションを残さない。
	if allowed := h.upsertUserFromIDToken(c, tok.IDToken, req.InvitationToken); !allowed {
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
	// refresh は既存ユーザーが対象のため、招待ゲート（戻り値 false）はここでは無視する
	// — DB から削除されたユーザーが refresh する稀ケースは別途 401 にすべきだが、本 PR の
	// スコープ外として既存挙動を維持する。
	if tok.IDToken != "" {
		// refresh は既存ユーザー前提で、新規招待を引く必要はないため invitationToken は空文字を渡す。
		_ = h.upsertUserFromIDToken(c, tok.IDToken, "")
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
// 戻り値 allowed:
//   - true : 既存ユーザー、または「Cognito group admin / pending invitation あり」で新規作成成功
//   - false: 「招待なし & Cognito admin でもない」新規ユーザー（ログイン拒否）または create 失敗
//
// invitationToken はマジックリンク経由の不透明 token（任意）。指定がある場合は email より
// 優先して照合する（同 email に複数 pending invitation がある異常系での誤一致を防ぐ）。
//
// 招待ゲートを usecase ではなく handler 層に置いているのは、認可境界（Cookie 発行 / 401-403 を返す責務）が
// HTTP 層であり、ユーザーが拒否された理由を上位ハンドラが直接知る必要があるため。
func (h *AuthHandler) upsertUserFromIDToken(c *gin.Context, idToken, invitationToken string) (allowed bool) {
	claims, err := middleware.DecodeClaims(idToken)
	if err != nil || h.users == nil {
		// claims を読めない / users repo 無し は内部エラー扱い。Cookie はクリアして拒否する。
		return false
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return false
	}
	email, _ := claims["email"].(string)
	groups := middleware.ToStringSliceFromClaim(claims["cognito:groups"])
	isCognitoAdmin := middleware.IsAdminFromGroups(groups)

	// 招待検索（既存ユーザーの role 昇格・新規ユーザー作成の両方で使う）。
	// 優先順位:
	//   1. invitationToken が指定されていれば FindPendingByToken で照合（マジックリンク経由）
	//   2. 1 で見つからない場合は email でフォールバック検索（旧フロー互換）
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
		// 既存ユーザー: 招待 / Cognito group の状態で role / company を更新する。
		// 重要な制約:
		//   - super_admin は何があっても降格しない（招待での降格は不可、別途 admin API でのみ降格可）
		//   - 招待による昇格は trainee → company_admin のみ（trainee → super_admin / company_admin → super_admin はしない）
		//   - Cognito group admin は最強で super_admin に昇格する
		if isCognitoAdmin && existing.Role != domain.RoleSuperAdmin {
			_ = h.users.UpdateRole(c.Request.Context(), existing.ID, domain.RoleSuperAdmin)
		}
		if inv != nil && existing.Role != domain.RoleSuperAdmin {
			// 役割昇格: trainee → company_admin だけ反映する（同一 role / 降格は skip）。
			if existing.Role == domain.RoleTrainee && inv.Role == domain.RoleCompanyAdmin {
				if err := h.users.UpdateRole(c.Request.Context(), existing.ID, domain.RoleCompanyAdmin); err != nil {
					log.Printf("upsertUserFromIDToken: existing user role upgrade failed userID=%d: %v", existing.ID, err)
				} else {
					log.Printf("upsertUserFromIDToken: existing user upgraded trainee→company_admin userID=%d email=%s", existing.ID, email)
				}
			}
			// company 紐付け: 招待の company_id と異なる、または未設定なら更新する。
			if inv.CompanyID != 0 && (existing.CompanyID == nil || *existing.CompanyID != inv.CompanyID) {
				if err := h.users.UpdateCompanyID(c.Request.Context(), existing.ID, inv.CompanyID); err != nil {
					log.Printf("upsertUserFromIDToken: existing user company update failed userID=%d: %v", existing.ID, err)
				}
			}
			// 招待を accepted にマーク（再利用防止 + 監査）
			_ = h.invitations.UpdateStatus(c.Request.Context(), inv.ID, domain.InvitationStatusAccepted)
		}
		return true
	}

	// 新規ユーザー: 招待 OR Cognito group "admin" のいずれかが必要。両方無ければ拒否。
	if !isCognitoAdmin && inv == nil {
		log.Printf("upsertUserFromIDToken: signup blocked — no invitation and not Cognito admin sub=%s email=%s token_provided=%t", sub, email, invitationToken != "")
		return false
	}

	role := domain.RoleTrainee
	var companyID *uint64
	var acceptedInvID uint64
	displayName := email

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
		if inv.DisplayName != "" {
			displayName = inv.DisplayName
		}
	}

	if err := h.users.Create(c.Request.Context(), &domain.User{
		CognitoSub:  sub,
		Email:       email,
		DisplayName: displayName,
		Role:        role,
		CompanyID:   companyID,
	}); err != nil {
		log.Printf("upsertUserFromIDToken: create user failed sub=%s email=%s err=%v", sub, email, err)
		return false
	}
	// 招待を accepted にマーク（履歴・監査）
	if h.invitations != nil && acceptedInvID != 0 {
		_ = h.invitations.UpdateStatus(c.Request.Context(), acceptedInvID, domain.InvitationStatusAccepted)
	}
	return true
}
