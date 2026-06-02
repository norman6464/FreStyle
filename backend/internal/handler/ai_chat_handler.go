package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// AiChatHandler は AI チャット関連のエンドポイントを提供する。
type AiChatHandler struct {
	getSessions   *usecase.GetAiChatSessionsByUserIDUseCase
	createSession *usecase.CreateAiChatSessionUseCase
	getSession    *usecase.GetAiChatSessionUseCase
	updateTitle   *usecase.UpdateAiChatSessionTitleUseCase
	deleteSession *usecase.DeleteAiChatSessionUseCase
	getMessages   *usecase.GetAiChatMessagesUseCase
}

func NewAiChatHandler(
	getSessions *usecase.GetAiChatSessionsByUserIDUseCase,
	createSession *usecase.CreateAiChatSessionUseCase,
	getSession *usecase.GetAiChatSessionUseCase,
	updateTitle *usecase.UpdateAiChatSessionTitleUseCase,
	deleteSession *usecase.DeleteAiChatSessionUseCase,
	getMessages *usecase.GetAiChatMessagesUseCase,
) *AiChatHandler {
	return &AiChatHandler{
		getSessions:   getSessions,
		createSession: createSession,
		getSession:    getSession,
		updateTitle:   updateTitle,
		deleteSession: deleteSession,
		getMessages:   getMessages,
	}
}

// @Summary      AI チャット セッション 一覧
// @Description  current user の AI チャット セッション を 新しい 順 で 返す。
// @Tags         ai-chat
// @Produce      json
// @Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.AiChatSession
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /ai-chat/sessions [get]
// @Security     CookieAuth
func (h *AiChatHandler) GetSessions(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.getSessions.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type createSessionReq struct {
	Title       string  `json:"title" binding:"required"`
	SessionType string  `json:"sessionType"`
	ScenarioID  *uint64 `json:"scenarioId"`
}

// @Summary      AI チャット セッション 作成
// @Description  current user 名義 で 新規 セッション を 作成。 IDOR 対策 で userId は body から 受け取らない。
// @Tags         ai-chat
// @Accept       json
// @Produce      json
// @Param        body  body      createSessionReq  true  "title 必須"
// @Success      201   {object}  github_com_norman6464_FreStyle_backend_internal_domain.AiChatSession
// @Failure      400   {object}  errorResponse  "バリデーション"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Router       /ai-chat/sessions [post]
// @Security     CookieAuth
func (h *AiChatHandler) CreateSession(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req createSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.createSession.Execute(c.Request.Context(), usecase.CreateAiChatSessionInput{
		UserID:      uid,
		Title:       req.Title,
		SessionType: req.SessionType,
		ScenarioID:  req.ScenarioID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// parseSessionID は :id パスパラメータを uint64 に変換する共通ヘルパー。
func parseSessionID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return 0, false
	}
	return id, true
}

// @Summary      AI チャット セッション 詳細
// @Description  指定 id の セッション を 返す。
// @Tags         ai-chat
// @Produce      json
// @Param        id  path      int  true  "セッション ID"
// @Success      200  {object}  github_com_norman6464_FreStyle_backend_internal_domain.AiChatSession
// @Failure      400  {object}  errorResponse  "id 不正"
// @Failure      404  {object}  errorResponse  "セッション が ない"
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /ai-chat/sessions/{id} [get]
// @Security     CookieAuth
func (h *AiChatHandler) GetSession(c *gin.Context) {
	id, ok := parseSessionID(c)
	if !ok {
		return
	}
	s, err := h.getSession.Execute(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, s)
}

type updateSessionTitleReq struct {
	Title string `json:"title" binding:"required"`
}

// @Summary      AI チャット セッション タイトル 更新
// @Description  指定 id の セッション の title を 更新。
// @Tags         ai-chat
// @Accept       json
// @Produce      json
// @Param        id    path      int                      true  "セッション ID"
// @Param        body  body      updateSessionTitleReq    true  "title 必須"
// @Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_domain.AiChatSession
// @Failure      400   {object}  errorResponse  "バリデーション"
// @Failure      500   {object}  errorResponse  "DB 失敗"
// @Router       /ai-chat/sessions/{id} [put]
// @Security     CookieAuth
func (h *AiChatHandler) UpdateSessionTitle(c *gin.Context) {
	id, ok := parseSessionID(c)
	if !ok {
		return
	}
	var req updateSessionTitleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.updateTitle.Execute(c.Request.Context(), id, req.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	s, _ := h.getSession.Execute(c.Request.Context(), id)
	c.JSON(http.StatusOK, s)
}

// @Summary      AI チャット セッション 削除
// @Description  指定 id の セッション を 削除。 所有者 検証 込み。
// @Tags         ai-chat
// @Produce      json
// @Param        id  path  int  true  "セッション ID"
// @Success      204  "成功 (本文 なし)"
// @Failure      400  {object}  errorResponse  "id 不正"
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      403  {object}  errorResponse  "他人 の セッション"
// @Failure      404  {object}  errorResponse  "セッション が ない"
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /ai-chat/sessions/{id} [delete]
// @Security     CookieAuth
func (h *AiChatHandler) DeleteSession(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, ok := parseSessionID(c)
	if !ok {
		return
	}
	if err := h.deleteSession.Execute(c.Request.Context(), id, uid); err != nil {
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary      AI チャット メッセージ 一覧
// @Description  指定 セッション の 会話 履歴 (DynamoDB から) を 古い 順 で 返す。
// @Tags         ai-chat
// @Produce      json
// @Param        id  path      int  true  "セッション ID"
// @Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.AiChatMessage
// @Failure      400  {object}  errorResponse  "id 不正"
// @Failure      500  {object}  errorResponse  "DynamoDB 失敗"
// @Router       /ai-chat/sessions/{id}/messages [get]
// @Security     CookieAuth
func (h *AiChatHandler) GetMessages(c *gin.Context) {
	id, ok := parseSessionID(c)
	if !ok {
		return
	}
	msgs, err := h.getMessages.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, msgs)
}
