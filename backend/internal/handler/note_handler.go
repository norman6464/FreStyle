package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type NoteHandler struct {
	list   *usecase.ListNotesByUserIDUseCase
	create *usecase.CreateNoteUseCase
	update *usecase.UpdateNoteUseCase
	del    *usecase.DeleteNoteUseCase
}

func NewNoteHandler(
	l *usecase.ListNotesByUserIDUseCase,
	c *usecase.CreateNoteUseCase,
	u *usecase.UpdateNoteUseCase,
	d *usecase.DeleteNoteUseCase,
) *NoteHandler {
	return &NoteHandler{list: l, create: c, update: u, del: d}
}

// List は current user の note 一覧を返す。
// userId はクライアントから受け取らない（IDOR 対策）。
//
//	@Summary      自分 の ノート 一覧
//	@Description  current user の note を 更新 日 降順 で 返す。 IDOR 対策 で userId は 受け取らない。
//	@Tags         notes
//	@Produce      json
//	@Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.Note
//	@Failure      400  {object}  errorResponse  "DB 取得 失敗"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Router       /notes [get]
//	@Security     CookieAuth
func (h *NoteHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type noteCreateReq struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content"`
	IsPublic bool   `json:"isPublic"`
	IsPinned bool   `json:"isPinned"`
}

// Create は current user 名義で note を作る。userId は受け取らず固定する。
//
//	@Summary      ノート 作成
//	@Description  current user 名義 で 新規 note を 作る。 userId は body で 指定 でき ない (current user 固定)。
//	@Tags         notes
//	@Accept       json
//	@Produce      json
//	@Param        body  body      noteCreateReq  true  "作成 内容 (title 必須)"
//	@Success      201   {object}  github_com_norman6464_FreStyle_backend_internal_domain.Note
//	@Failure      400   {object}  errorResponse  "バリデーション エラー or DB 失敗"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Router       /notes [post]
//	@Security     CookieAuth
func (h *NoteHandler) Create(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req noteCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.create.Execute(c.Request.Context(), usecase.CreateNoteInput{
		UserID: uid, Title: req.Title, Content: req.Content,
		IsPublic: req.IsPublic, IsPinned: req.IsPinned,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

type noteUpdateReq struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content"`
	IsPublic bool   `json:"isPublic"`
	IsPinned bool   `json:"isPinned"`
}

// Update は current user 所有の note のみ更新可能。usecase 側で所有者検証する。
//
//	@Summary      ノート 更新
//	@Description  指定 note を 更新。 所有者 検証 を usecase 層 で 行い、 他人 の note は 403。
//	@Tags         notes
//	@Accept       json
//	@Produce      json
//	@Param        id    path      int            true  "ノート ID"
//	@Param        body  body      noteUpdateReq  true  "更新 内容"
//	@Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_domain.Note
//	@Failure      400   {object}  errorResponse  "バリデーション or DB 失敗"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      403   {object}  errorResponse  "他人 の note"
//	@Router       /notes/{id} [put]
//	@Security     CookieAuth
func (h *NoteHandler) Update(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req noteUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.update.Execute(c.Request.Context(), usecase.UpdateNoteInput{
		UserID: uid, ID: id, Title: req.Title, Content: req.Content,
		IsPublic: req.IsPublic, IsPinned: req.IsPinned,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrNoteForbidden) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}

// Delete は WHERE で user_id を絞るため他人の note は消せない。
//
//	@Summary      ノート 削除
//	@Description  current user 所有 の note を 削除。 WHERE user_id 絞り込み で 他人 の note は そもそも 影響 を 受け ない。
//	@Tags         notes
//	@Produce      json
//	@Param        id  path  int  true  "ノート ID"
//	@Success      204  "成功 (本文 なし)"
//	@Failure      400  {object}  errorResponse  "DB 失敗"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Router       /notes/{id} [delete]
//	@Security     CookieAuth
func (h *NoteHandler) Delete(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.del.Execute(c.Request.Context(), uid, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
