package handler

import (
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/infra/oidc"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/norman6464/frestyle/backend/internal/usecase/user"
)

// AuthHandler は認証エンドポイントを提供する。
// 発行者との通信は infra/oidc に切り出し、ここは HTTP の境界とユーザーの upsert だけを持つ。
type AuthHandler struct {
	getCurrentUser          *user.GetCurrentUserUseCase
	upsertUser              *user.UpsertUserFromIDTokenUseCase
	ensurePersonalWorkspace *kb.EnsurePersonalWorkspaceUseCase
	verifier                *oidc.Verifier
}

// NewAuthHandler は AuthHandler を組み立てる。
func NewAuthHandler(
	getCurrentUser *user.GetCurrentUserUseCase,
	upsertUser *user.UpsertUserFromIDTokenUseCase,
	ensurePersonalWorkspace *kb.EnsurePersonalWorkspaceUseCase,
	verifier *oidc.Verifier,
) *AuthHandler {
	return &AuthHandler{
		getCurrentUser:          getCurrentUser,
		upsertUser:              upsertUser,
		ensurePersonalWorkspace: ensurePersonalWorkspace,
		verifier:                verifier,
	}
}

// Me は現在ログイン中のユーザー情報を返す。
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
	// workspaceId は段 2 で撤去（1 人が複数のワークスペースに所属できるため、単一の
	// 所属先という概念が無い。所属一覧は GET /kb/workspaces が返す）。
	resp := gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
	}
	c.JSON(http.StatusOK, resp)
}

// Login は Authorization: Bearer で渡された ID トークンを検証し、
// 初回サインインなら users 行と個人ワークスペースを作る（自己サインアップ）。
//
// verifier は発行者非依存（本番は GCIP、ローカルは Dex）。どちらもクライアント側で
// 直接発行者とやり取りして ID トークンを得る設計で、backend が仲介する認可コード交換は
// 存在しない。以降の API 呼び出しは同じ ID トークンをそのまま Bearer で送るだけでよく、
// backend 側で発行する Cookie は無い（JWTAuth がリクエストのたびに同じトークンを検証する）。
func (h *AuthHandler) Login(c *gin.Context) {
	idToken, ok := middleware.BearerToken(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}

	_, err := h.upsertUserFromIDToken(c, idToken)
	if err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "email_taken",
				"message": "同じメールアドレスでの登録が別のリクエストで同時に完了しました。もう一度ログインし直してください。",
			})
			return
		}
		if errors.Is(err, errIDTokenRejected) {
			log.Printf("login: id_token rejected: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_id_token"})
			return
		}
		log.Printf("login: upsert failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ログインしました。"})
}

// errIDTokenRejected は id_token の署名・クレーム検証に落ちたことを表す。
// Login がこれを 401 に変換する（DB 障害の 500 と区別する）。
var errIDTokenRejected = errors.New("handler: id_token rejected")

// upsertUserFromIDToken は id_token を検証してユーザー更新を usecase へ委譲する。
// 続けて個人ワークスペースの確保まで行う（無ければ作る。既存なら 1 回の SELECT で終わる）。
func (h *AuthHandler) upsertUserFromIDToken(c *gin.Context, idToken string) (u *domain.User, err error) {
	if h.upsertUser == nil {
		return nil, errors.New("upsert user usecase not configured")
	}

	// **署名とクレームを検証してから読む。**
	// ここで作られるのはユーザーそのもの（sub / email）で、検証せずに読むと
	// 「好きな sub と email を名乗って新しいユーザーを作る」ことができてしまう。
	//
	// nonce は空文字（照合しない）。nonce は「認可を始めたブラウザ本人か」を確かめる
	// もので、リダイレクトを介した認可要求に対応する。GCIP はクライアント SDK が直接
	// 発行者とやり取りするため、対応する認可要求そのものが存在しない。
	claims, verifyErr := h.verifier.VerifyIDToken(c.Request.Context(), idToken, "")
	if verifyErr != nil {
		return nil, errors.Join(errIDTokenRejected, verifyErr)
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	// email_verified が無いクレームは「未検証」に倒す（ゼロ値 false）。この型アサーションは
	// クレームが欠けている・bool でない場合に false, false を返すため、それで正しい。
	emailVerified, _ := claims["email_verified"].(bool)

	u, err = h.upsertUser.Execute(
		c.Request.Context(),
		user.UpsertUserFromIDTokenInput{
			Subject:       sub,
			Email:         email,
			Name:          name,
			EmailVerified: emailVerified,
		},
	)
	if err != nil || u == nil {
		return nil, err
	}

	// 失敗してもログインは失敗させない（次回ログイン時に自己修復する）。
	if h.ensurePersonalWorkspace != nil {
		if _, wsErr := h.ensurePersonalWorkspace.Execute(
			c.Request.Context(),
			kb.EnsurePersonalWorkspaceInput{UserID: u.ID, Name: u.Name},
		); wsErr != nil {
			slog.ErrorContext(c.Request.Context(), "ensure personal workspace failed (non-fatal)", "userID", u.ID, "err", wsErr)
		}
	}

	return u, nil
}
