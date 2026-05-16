package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type SessionNoteHandler struct {
	get    *usecase.GetSessionNoteUseCase
	upsert *usecase.UpsertSessionNoteUseCase
}

func NewSessionNoteHandler(g *usecase.GetSessionNoteUseCase, u *usecase.UpsertSessionNoteUseCase) *SessionNoteHandler {
	return &SessionNoteHandler{get: g, upsert: u}
}

// Get は AI チャット セッション に 紐づく ノート を 取得 する。
//
//	@Summary      セッション ノート 取得
//	@Description  AI チャット セッション に 紐づく ノート (= 学習 者 が セッション ごと に 残した メモ) を 取得。 存在 し ない 場合 は 404。
//	@Tags         session-notes
//	@Produce      json
//	@Param        sessionId  path      int  true  "AI チャット セッション ID"
//	@Success      200        {object}  github_com_norman6464_FreStyle_backend_internal_domain.SessionNote
//	@Failure      400        {object}  errorResponse  "DB 失敗"
//	@Failure      401        {object}  errorResponse  "未 認証"
//	@Failure      404        {object}  errorResponse  "未 作成"
//	@Router       /sessions/{sessionId}/note [get]
//	@Security     CookieAuth
func (h *SessionNoteHandler) Get(c *gin.Context) {
	sid, _ := strconv.ParseUint(c.Param("sessionId"), 10, 64)
	n, err := h.get.Execute(c.Request.Context(), sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if n == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, n)
}

// sessionNoteUpsertReq は body 受け取り 形。 userId は **受け取らない** (current user
// を middleware から 取って 強制 する)。 これ で 他人 名義 の 作成 / 更新 を 防ぐ
// (CodeRabbit PR #1727 で 指摘 さ れた IDOR 修正)。
type sessionNoteUpsertReq struct {
	Content string `json:"content"`
}

// Upsert は セッション ノート を 作成 or 更新 する。
//
//	@Summary      セッション ノート 作成 / 更新
//	@Description  指定 セッション の ノート を upsert。 userId は body で 受け取らず current user 固定 (IDOR 対策)。
//	@Tags         session-notes
//	@Accept       json
//	@Produce      json
//	@Param        sessionId  path      int                    true   "AI チャット セッション ID"
//	@Param        body       body      sessionNoteUpsertReq   true   "content (userId は current user 固定 / body で 受け取らない)"
//	@Success      200        {object}  github_com_norman6464_FreStyle_backend_internal_domain.SessionNote
//	@Failure      400        {object}  errorResponse  "バリデーション or DB 失敗"
//	@Failure      401        {object}  errorResponse  "未 認証"
//	@Router       /sessions/{sessionId}/note [put]
//	@Security     CookieAuth
func (h *SessionNoteHandler) Upsert(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sid, _ := strconv.ParseUint(c.Param("sessionId"), 10, 64)
	var req sessionNoteUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.upsert.Execute(c.Request.Context(), usecase.UpsertSessionNoteInput{
		SessionID: sid, UserID: uid, Content: req.Content,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}
