package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// noteNotFoundMsg は「そのノートは取得できない」ことを表す唯一の応答本文。
//
// 他人のノートでも存在しないノートでも、必ずこの 1 つの文言を 404 で返す。
// ステータスだけ 404 に揃えても本文が違えば、呼び出し元は本文の差で
// 「実在するが自分のものではない」と「そもそも無い」を見分けられてしまい、
// 存在オラクル（ID 空間を総当たりして実在 ID を全数把握できる状態）が残る。
// 定数を 1 つだけ置き、両方の経路がこれを共有することでバイト単位で同一にする。
const noteNotFoundMsg = "note not found"

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

// List は current user の note 一覧を返す（userId は受け取らない、IDOR 対策）。
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

// Update は current user 所有の note のみ更新可能。
//
// 他人の note と存在しない note は撃ち分けない。usecase が両方を domain.ErrNotFound へ
// 畳んで返すので、handler はその 1 本の分岐で 404 + noteNotFoundMsg を返す。
// 分岐が 1 つしか無いこと自体が「ステータスも本文も一致する」ことの担保になる
// （403 と 404 のように分けて書くと、片方を直し忘れた瞬間に存在オラクルが復活する）。
//
//	@Summary      ノート 更新
//	@Description  指定 note を 更新。 更新 できる の は current user 所有 の note だけ。
//	@Description  他人 の note と 存在 し ない note は 撃ち 分け ず、 どちら も 同じ 404 (同一 本文) を 返す。
//	@Description  応答 の 差 で ID の 実在 を 数え 上げ られる (存在 オラクル) の を 防ぐ ため。
//	@Tags         notes
//	@Accept       json
//	@Produce      json
//	@Param        id    path      int            true  "ノート ID"
//	@Param        body  body      noteUpdateReq  true  "更新 内容"
//	@Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_domain.Note
//	@Failure      400   {object}  errorResponse  "バリデーション or DB 失敗"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      404   {object}  errorResponse  "他人 の note or 存在 し ない note (区別 し ない)"
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
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": noteNotFoundMsg})
			return
		}
		// 他人 / 不在 以外の失敗（バリデーション・DB 障害）。err.Error() をそのまま返すが、
		// ここに来るのは note の実在に依存しないエラーだけなので存在は漏れない。
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}

// Delete は WHERE で user_id を絞るため他人の note は消せない。
//
// 削除は「消えた」「元から無い」「他人のものだった」を区別せず、いずれも 204（本文なし）を返す。
// DELETE 文が 0 行に当たっただけではエラーにならないので、これは分岐を足さない限り自然にそうなる。
// 逆に「対象なしだから 404」のような分岐を足すと、そこが存在オラクルになるので入れてはいけない。
//
//	@Summary      ノート 削除
//	@Description  current user 所有 の note を 削除。 WHERE user_id 絞り込み で 他人 の note は そもそも 影響 を 受け ない。
//	@Description  他人 の note・存在 し ない note・自分 の note の いずれ に 対して も 同じ 204 (本文 なし) を 返し、
//	@Description  応答 から ID の 実在 が 分から ない よう に する。
//	@Tags         notes
//	@Produce      json
//	@Param        id  path  int  true  "ノート ID"
//	@Success      204  "成功 (本文 なし。 対象 が 無く て も 同じ)"
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
