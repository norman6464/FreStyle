package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type FriendshipHandler struct {
	list      *usecase.ListFriendshipsUseCase
	request   *usecase.RequestFriendshipUseCase
	respond   *usecase.RespondFriendshipUseCase
	follow    *usecase.FollowUserUseCase
	unfollow  *usecase.UnfollowUserUseCase
	following *usecase.ListFollowingUseCase
	followers *usecase.ListFollowersUseCase
	status    *usecase.GetFollowStatusUseCase
}

func NewFriendshipHandler(
	l *usecase.ListFriendshipsUseCase,
	r *usecase.RequestFriendshipUseCase,
	p *usecase.RespondFriendshipUseCase,
	f *usecase.FollowUserUseCase,
	uf *usecase.UnfollowUserUseCase,
	lf *usecase.ListFollowingUseCase,
	lfr *usecase.ListFollowersUseCase,
	gs *usecase.GetFollowStatusUseCase,
) *FriendshipHandler {
	return &FriendshipHandler{
		list: l, request: r, respond: p,
		follow: f, unfollow: uf,
		following: lf, followers: lfr, status: gs,
	}
}

// List は current user に紐づく Friendship 全件 (pending 含む) を返す。
func (h *FriendshipHandler) List(c *gin.Context) {
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

type requestFriendshipReq struct {
	AddresseeID uint64 `json:"addresseeId" binding:"required"`
}

// Request は current user 起点で friendship 申請を作る。
func (h *FriendshipHandler) Request(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req requestFriendshipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.request.Execute(c.Request.Context(), uid, req.AddresseeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

type respondFriendshipReq struct {
	Accept bool `json:"accept"`
}

func (h *FriendshipHandler) Respond(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req respondFriendshipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.respond.Execute(c.Request.Context(), id, req.Accept); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Following は current user が follow している相手の Friendship を返す。
func (h *FriendshipHandler) Following(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.following.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// Followers は current user を follow している相手の Friendship を返す。
func (h *FriendshipHandler) Followers(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.followers.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// Follow は current user → :userId の単方向 follow を成立させる。
func (h *FriendshipHandler) Follow(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	target, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	got, err := h.follow.Execute(c.Request.Context(), uid, target)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

// Unfollow は current user → :userId の friendship を削除する。
func (h *FriendshipHandler) Unfollow(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	target, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err := h.unfollow.Execute(c.Request.Context(), uid, target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Status は current user と :userId の follow 関係を返す。
func (h *FriendshipHandler) Status(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	target, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	out, err := h.status.Execute(c.Request.Context(), uid, target)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}
