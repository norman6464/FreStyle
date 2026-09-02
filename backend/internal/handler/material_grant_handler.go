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
//
//	@Summary      コース の 権限 一覧
//	@Description  その コース 自身 に 張ら れ た 付与 を 返す。 **「この コース を 編集 できる 人 の 一覧」 で は ない** — ワークスペース の admin は 含ま れ ず、 空 で も 「誰 も 編集 でき ない」 の 意味 に なら ない。 呼べる の は その コース を 管理 できる 人 だけ で、 読め ない 相手 に は 404、 読める が 管理 でき ない 場合 は 403。
//	@Tags         courses
//	@Produce      json
//	@Param        id   path      int  true  "コース ID"
//	@Success      200  {array}   materialGrantResponse
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      403  {object}  errorResponse  "管理 権限 が 無い"
//	@Failure      404  {object}  errorResponse  "コース が 無い か 読め ない"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /courses/{id}/grants [get]
//	@Security     CookieAuth
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
//
//	@Summary      コース の 権限 付与
//	@Description  コース で の 既定 の 役割 を 与える (同じ 主体 に は 1 行 だけ な の で 上書き)。 配下 の 章 に も 効く。 合成 は 「最も 強い もの を 採る」 な の で、 **ここ に 弱い 役割 を 張っ て も 上位 で 得 て いる 役割 は 下がら ない**。 呼べる の は その コース を 管理 できる 人 だけ。
//	@Tags         courses
//	@Accept       json
//	@Produce      json
//	@Param        id           path      int                        true  "コース ID"
//	@Param        principalId  path      string                     true  "主体 ID (UUID)"
//	@Param        body         body      materialGrantRoleRequest   true  "役割 (admin / editor / commenter / viewer)"
//	@Success      200          {object}  materialGrantResponse
//	@Failure      400          {object}  errorResponse  "バリデーション エラー"
//	@Failure      401          {object}  errorResponse  "未 認証"
//	@Failure      403          {object}  errorResponse  "管理 権限 が 無い"
//	@Failure      404          {object}  errorResponse  "コース か 主体 が 無い"
//	@Failure      500          {object}  errorResponse  "DB 失敗"
//	@Router       /courses/{id}/grants/{principalId} [put]
//	@Security     CookieAuth
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
//
//	@Summary      コース の 権限 取り消し
//	@Description  コース で の 既定 の 役割 を 剥がす。 元 から 無い 相手 に 対し て も 成功 する (冪等)。 呼べる の は その コース を 管理 できる 人 だけ。
//	@Tags         courses
//	@Produce      json
//	@Param        id           path  int     true  "コース ID"
//	@Param        principalId  path  string  true  "主体 ID (UUID)"
//	@Success      204          "取り消し 済み"
//	@Failure      401          {object}  errorResponse  "未 認証"
//	@Failure      403          {object}  errorResponse  "管理 権限 が 無い"
//	@Failure      404          {object}  errorResponse  "コース が 無い か 読め ない"
//	@Failure      500          {object}  errorResponse  "DB 失敗"
//	@Router       /courses/{id}/grants/{principalId} [delete]
//	@Security     CookieAuth
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
//
//	@Summary      教材 の 権限 を 張れる 相手 の 一覧
//	@Description  権限 を 張れる 相手 (ユーザー / グループ / スペース の 全員) を 表示 名 つき で 返す。 共有 の 画面 で 相手 を 選ぶ ため の 口。 中身 は ワークスペース 全体 だ が、 **呼べる か は コース 単位 で 決まる** — ワークスペース の admin に 絞る と、 コース に admin を 張ら れ た 人 が 相手 を 選べ なく なる。 名前 が 引け なかっ た 行 も 空文字 の まま 返す。
//	@Tags         courses
//	@Produce      json
//	@Param        id   path      int  true  "コース ID"
//	@Success      200  {array}   materialPrincipalResponse
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      403  {object}  errorResponse  "管理 権限 が 無い"
//	@Failure      404  {object}  errorResponse  "コース が 無い か 読め ない"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /courses/{id}/principals [get]
//	@Security     CookieAuth
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
//
//	@Summary      教材 の 権限 一覧
//	@Description  その 教材 (章) 自身 に 張ら れ た 付与 を 返す。 **コース から 降り て くる 分 は 含ま ない** の で、 空 で も 「誰 も 編集 でき ない」 の 意味 に なら ない。 呼べる の は その 教材 を 管理 できる 人 だけ。
//	@Tags         teaching-materials
//	@Produce      json
//	@Param        id   path      int  true  "教材 ID"
//	@Success      200  {array}   materialGrantResponse
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      403  {object}  errorResponse  "管理 権限 が 無い"
//	@Failure      404  {object}  errorResponse  "教材 が 無い か 読め ない"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /teaching-materials/{id}/grants [get]
//	@Security     CookieAuth
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
//
//	@Summary      教材 の 権限 付与
//	@Description  その 教材 (章) だけ に 効く 既定 の 役割 を 与える。 コース の 付与 より 弱い 役割 を ここ に 張っ て も 下がら ない (合成 は 最も 強い もの を 採る)。 呼べる の は その 教材 を 管理 できる 人 だけ。
//	@Tags         teaching-materials
//	@Accept       json
//	@Produce      json
//	@Param        id           path      int                       true  "教材 ID"
//	@Param        principalId  path      string                    true  "主体 ID (UUID)"
//	@Param        body         body      materialGrantRoleRequest  true  "役割 (admin / editor / commenter / viewer)"
//	@Success      200          {object}  materialGrantResponse
//	@Failure      400          {object}  errorResponse  "バリデーション エラー"
//	@Failure      401          {object}  errorResponse  "未 認証"
//	@Failure      403          {object}  errorResponse  "管理 権限 が 無い"
//	@Failure      404          {object}  errorResponse  "教材 か 主体 が 無い"
//	@Failure      500          {object}  errorResponse  "DB 失敗"
//	@Router       /teaching-materials/{id}/grants/{principalId} [put]
//	@Security     CookieAuth
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
//
//	@Summary      教材 の 権限 取り消し
//	@Description  その 教材 (章) で の 既定 の 役割 を 剥がす。 元 から 無い 相手 に 対し て も 成功 する (冪等)。 消える の は この 段 で 足し た 分 だけ で、 コース から 降り て いる 役割 は そのまま 残る。
//	@Tags         teaching-materials
//	@Produce      json
//	@Param        id           path  int     true  "教材 ID"
//	@Param        principalId  path  string  true  "主体 ID (UUID)"
//	@Success      204          "取り消し 済み"
//	@Failure      401          {object}  errorResponse  "未 認証"
//	@Failure      403          {object}  errorResponse  "管理 権限 が 無い"
//	@Failure      404          {object}  errorResponse  "教材 が 無い か 読め ない"
//	@Failure      500          {object}  errorResponse  "DB 失敗"
//	@Router       /teaching-materials/{id}/grants/{principalId} [delete]
//	@Security     CookieAuth
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
