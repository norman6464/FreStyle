package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// MaterialGrantHandler は教材（コース / 章）の権限の読み書きを受ける。
//
// 認可はすべて usecase 側が持つ（判定規則をここに書かない）。拒否の返し方は
// 教材の読み書きと同じで、**見えない相手には 404、見えているが権限が足りない相手には
// 403**。前者を撃ち分けると、ID を総当たりするだけで隠した教材の実在が分かる。
type MaterialGrantHandler struct {
	grantCourse    *usecase.GrantCourseRoleUseCase
	revokeCourse   *usecase.RevokeCourseRoleUseCase
	listCourse     *usecase.ListCourseGrantsUseCase
	grantChapter   *usecase.GrantChapterRoleUseCase
	revokeChapter  *usecase.RevokeChapterRoleUseCase
	listChapter    *usecase.ListChapterGrantsUseCase
	listPrincipals *usecase.ListGrantableMaterialPrincipalsUseCase
}

// NewMaterialGrantHandler は MaterialGrantHandler を組み立てる。
func NewMaterialGrantHandler(
	grantCourse *usecase.GrantCourseRoleUseCase,
	revokeCourse *usecase.RevokeCourseRoleUseCase,
	listCourse *usecase.ListCourseGrantsUseCase,
	grantChapter *usecase.GrantChapterRoleUseCase,
	revokeChapter *usecase.RevokeChapterRoleUseCase,
	listChapter *usecase.ListChapterGrantsUseCase,
	listPrincipals *usecase.ListGrantableMaterialPrincipalsUseCase,
) *MaterialGrantHandler {
	return &MaterialGrantHandler{
		grantCourse:    grantCourse,
		revokeCourse:   revokeCourse,
		listCourse:     listCourse,
		grantChapter:   grantChapter,
		revokeChapter:  revokeChapter,
		listChapter:    listChapter,
		listPrincipals: listPrincipals,
	}
}

// materialGrantResponse は教材の付与 1 件の返却形。
// コースと章で同じ形にする（画面が 1 つの部品で両方を扱えるようにするため）。
type materialGrantResponse struct {
	PrincipalID string    `json:"principalId" example:"0198a000-0000-7000-8000-00000000000a"`
	Role        string    `json:"role"        example:"editor"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// materialPrincipalResponse は権限を張れる相手 1 件の返却形。
type materialPrincipalResponse struct {
	ID   string `json:"id"   example:"0198a000-0000-7000-8000-00000000000a"`
	Kind string `json:"kind" example:"user"`
	// Name は表示名。引けなかった場合は空文字（行は落とさない）。
	Name string `json:"name" example:"田中 太郎"`
}

// materialGrantRoleRequest は付与を張るときの入力。
type materialGrantRoleRequest struct {
	// Role は domain.ValidGrantRoles のいずれか。既知でない値は usecase が弾く。
	Role string `json:"role" binding:"required" example:"editor"`
}

// materialGrantActor は認証情報とパスの ID をまとめて取り出す。
// 取り出せなければ応答を書いて ok=false を返す。
func materialGrantActor(c *gin.Context, param string) (usecase.MaterialActor, uint64, bool) {
	uid, workspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return usecase.MaterialActor{}, 0, false
	}
	id, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return usecase.MaterialActor{}, 0, false
	}
	return usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: workspace}, id, true
}

// respondMaterialGrantErr は付与の操作で起きたエラーを応答へ落とす。
//
// 未知の役割だけ 400（入力の誤り）。**相手が別テナントの主体だった場合は 404** で、
// 「対象が無い」と同じ応答にする。ここを取り違えると、主体 ID を総当たりするだけで
// 他テナントに誰が居るかが分かる（FK 違反のまま 500 にすると、応答の差でそれが読める）。
func respondMaterialGrantErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrInvalidGrantRole):
		c.JSON(http.StatusBadRequest, gin.H{"error": "役割の指定が不正です"})
	case errors.Is(err, repository.ErrPrincipalNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "対象が見つかりません"})
	default:
		respondEntityErr(c, err, "対象が見つかりません", "権限の操作に失敗しました")
	}
}

// ListCourseGrants はコース自身に張られた付与を返す。
func (h *MaterialGrantHandler) ListCourseGrants(c *gin.Context) {
	actor, courseID, ok := materialGrantActor(c, "id")
	if !ok {
		return
	}
	grants, err := h.listCourse.Execute(c.Request.Context(), usecase.ListCourseGrantsInput{
		MaterialActor: actor, CourseID: courseID,
	})
	if err != nil {
		respondMaterialGrantErr(c, err)
		return
	}
	out := make([]materialGrantResponse, 0, len(grants))
	for _, g := range grants {
		out = append(out, materialGrantResponse{
			PrincipalID: g.PrincipalID, Role: string(g.Role),
			CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

// GrantCourseRole はコースでの既定の役割を主体に与える。
func (h *MaterialGrantHandler) GrantCourseRole(c *gin.Context) {
	actor, courseID, ok := materialGrantActor(c, "id")
	if !ok {
		return
	}
	var req materialGrantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	g, err := h.grantCourse.Execute(c.Request.Context(), usecase.GrantCourseRoleInput{
		MaterialActor: actor, CourseID: courseID,
		PrincipalID: c.Param("principalId"), Role: domain.GrantRole(req.Role),
	})
	if err != nil {
		respondMaterialGrantErr(c, err)
		return
	}
	c.JSON(http.StatusOK, materialGrantResponse{
		PrincipalID: g.PrincipalID, Role: string(g.Role),
		CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	})
}

// RevokeCourseRole はコースでの既定の役割を剥がす（冪等）。
func (h *MaterialGrantHandler) RevokeCourseRole(c *gin.Context) {
	actor, courseID, ok := materialGrantActor(c, "id")
	if !ok {
		return
	}
	if err := h.revokeCourse.Execute(c.Request.Context(), usecase.RevokeCourseRoleInput{
		MaterialActor: actor, CourseID: courseID, PrincipalID: c.Param("principalId"),
	}); err != nil {
		respondMaterialGrantErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListGrantablePrincipals はコースに権限を張れる相手を表示名つきで返す。
func (h *MaterialGrantHandler) ListGrantablePrincipals(c *gin.Context) {
	actor, courseID, ok := materialGrantActor(c, "id")
	if !ok {
		return
	}
	principals, err := h.listPrincipals.Execute(c.Request.Context(), usecase.ListGrantableMaterialPrincipalsInput{
		MaterialActor: actor, CourseID: courseID,
	})
	if err != nil {
		respondMaterialGrantErr(c, err)
		return
	}
	out := make([]materialPrincipalResponse, 0, len(principals))
	for _, p := range principals {
		out = append(out, materialPrincipalResponse{ID: p.ID, Kind: string(p.Kind), Name: p.Name})
	}
	c.JSON(http.StatusOK, out)
}

// ListChapterGrants は章自身に張られた付与を返す。
func (h *MaterialGrantHandler) ListChapterGrants(c *gin.Context) {
	actor, chapterID, ok := materialGrantActor(c, "id")
	if !ok {
		return
	}
	grants, err := h.listChapter.Execute(c.Request.Context(), usecase.ListChapterGrantsInput{
		MaterialActor: actor, ChapterID: chapterID,
	})
	if err != nil {
		respondMaterialGrantErr(c, err)
		return
	}
	out := make([]materialGrantResponse, 0, len(grants))
	for _, g := range grants {
		out = append(out, materialGrantResponse{
			PrincipalID: g.PrincipalID, Role: string(g.Role),
			CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

// GrantChapterRole は章 1 つでの既定の役割を主体に与える。
func (h *MaterialGrantHandler) GrantChapterRole(c *gin.Context) {
	actor, chapterID, ok := materialGrantActor(c, "id")
	if !ok {
		return
	}
	var req materialGrantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	g, err := h.grantChapter.Execute(c.Request.Context(), usecase.GrantChapterRoleInput{
		MaterialActor: actor, ChapterID: chapterID,
		PrincipalID: c.Param("principalId"), Role: domain.GrantRole(req.Role),
	})
	if err != nil {
		respondMaterialGrantErr(c, err)
		return
	}
	c.JSON(http.StatusOK, materialGrantResponse{
		PrincipalID: g.PrincipalID, Role: string(g.Role),
		CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	})
}

// RevokeChapterRole は章での既定の役割を剥がす（冪等）。
func (h *MaterialGrantHandler) RevokeChapterRole(c *gin.Context) {
	actor, chapterID, ok := materialGrantActor(c, "id")
	if !ok {
		return
	}
	if err := h.revokeChapter.Execute(c.Request.Context(), usecase.RevokeChapterRoleInput{
		MaterialActor: actor, ChapterID: chapterID, PrincipalID: c.Param("principalId"),
	}); err != nil {
		respondMaterialGrantErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
