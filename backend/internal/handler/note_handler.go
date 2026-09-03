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
// 1 行も消えなかった場合は 404 を返す（以前はすべて 204 だった）。
// 0 行削除を 204 にしていると、呼び出し側は「自分の note を 1 件消した」と「何も起きなかった」を
// 区別できず、削除が効いていないのに画面からは行が消える。
//
// 存在オラクルとの関係（Update と同じ）:
//
//	DELETE 文の WHERE に user_id が入っているので、「他人の note」も「存在しない id」も
//	どちらも 0 行 = 同じ domain.ErrNotFound になり、ここから返る応答は
//	ステータス（404）も本文（noteNotFoundMsg）もバイト単位で同一になる。
//	応答が分かれるのは「自分の note を実際に消せたか」だけで、他人の note の実在は漏れない。
func (h *NoteHandler) Delete(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.del.Execute(c.Request.Context(), uid, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// 他人の note と存在しない id はどちらもここに来る（本文も同一）。
			c.JSON(http.StatusNotFound, gin.H{"error": noteNotFoundMsg})
			return
		}
		// 不在以外の失敗（バリデーション・DB 障害）。note の実在に依存しないエラーだけ。
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
