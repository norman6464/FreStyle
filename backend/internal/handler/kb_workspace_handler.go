package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// KnowledgeBaseWorkspaceHandler はナレッジ基盤のワークスペース / スペースの操作を受ける。
//
// ページ操作（KnowledgeBasePageHandler）と分けているのは、テナントの確定の仕方が違うため。
// 一覧と作成は URL に slug を持たず middleware.KnowledgeBaseWorkspace を通れない
// （通したら「まだ所属していない・まだ存在しない」ワークスペースを扱えない）。
type KnowledgeBaseWorkspaceHandler struct {
	listWorkspaces  *usecase.ListMemberWorkspacesUseCase
	createWorkspace *usecase.CreateWorkspaceUseCase
	checkWorkspace  *usecase.CheckWorkspacePermissionUseCase
	createSpace     *usecase.CreateSpaceUseCase
}

// NewKnowledgeBaseWorkspaceHandler は KnowledgeBaseWorkspaceHandler を組み立てる。
func NewKnowledgeBaseWorkspaceHandler(
	listWorkspaces *usecase.ListMemberWorkspacesUseCase,
	createWorkspace *usecase.CreateWorkspaceUseCase,
	checkWorkspace *usecase.CheckWorkspacePermissionUseCase,
	createSpace *usecase.CreateSpaceUseCase,
) *KnowledgeBaseWorkspaceHandler {
	return &KnowledgeBaseWorkspaceHandler{
		listWorkspaces:  listWorkspaces,
		createWorkspace: createWorkspace,
		checkWorkspace:  checkWorkspace,
		createSpace:     createSpace,
	}
}

// kbWorkspaceResponse はワークスペース 1 件の返却形。
//
// id は載せない。以降の API はすべて URL の slug でテナントを指すので、クライアントが
// 内部 UUID を使う場面が無い（kbPageResponse が workspaceId を出さないのと同じ理由）。
type kbWorkspaceResponse struct {
	Slug      string    `json:"slug" example:"acme"`
	Name      string    `json:"name" example:"Acme 社"`
	CreatedAt time.Time `json:"createdAt"`
}

func toKbWorkspaceResponse(w *domain.Workspace) kbWorkspaceResponse {
	return kbWorkspaceResponse{Slug: w.Slug, Name: w.Name, CreatedAt: w.CreatedAt}
}

// kbSpaceResponse はスペース 1 件の返却形。
// id は載せる（ページ一覧・作成の URL がスペース ID を取るため）。
type kbSpaceResponse struct {
	ID        string    `json:"id"  example:"0198a000-0000-7000-8000-000000000002"`
	Key       string    `json:"key" example:"eng"`
	Name      string    `json:"name" example:"開発部"`
	CreatedAt time.Time `json:"createdAt"`
}

func toKbSpaceResponse(s *domain.Space) kbSpaceResponse {
	return kbSpaceResponse{ID: s.ID, Key: s.Key, Name: s.Name, CreatedAt: s.CreatedAt}
}

// List は自分が所属するワークスペースの一覧を返す。
//
//	@Summary      ナレッジ 基盤 の 所属 ワークスペース 一覧
//	@Description  ログイン 中 の ユーザー が 所属 する ワークスペース を 返す。 所属 は principals (kind='user') の 行 が 唯一 の 表現 で、 所属 し て い ない ワークスペース は 1 件 も 含ま ない。 ほか の ナレッジ 基盤 API が URL に 使う slug を 知る ため の 入口。
//	@Tags         knowledge-base
//	@Produce      json
//	@Success      200  {array}   kbWorkspaceResponse
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces [get]
//	@Security     CookieAuth
func (h *KnowledgeBaseWorkspaceHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	workspaces, err := h.listWorkspaces.Execute(c.Request.Context(), usecase.ListMemberWorkspacesInput{UserID: uid})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	out := make([]kbWorkspaceResponse, 0, len(workspaces))
	for i := range workspaces {
		out = append(out, toKbWorkspaceResponse(&workspaces[i]))
	}
	c.JSON(http.StatusOK, out)
}

// kbCreateWorkspaceRequest はワークスペース作成の入力。
type kbCreateWorkspaceRequest struct {
	Slug string `json:"slug" binding:"required" example:"acme"`
	Name string `json:"name" binding:"required,max=200" example:"Acme 社"`
}

// Create はワークスペースを作り、作成者をその admin にする。
//
//	@Summary      ナレッジ 基盤 の ワークスペース 作成
//	@Description  ワークスペース を 作る。 作成 者 は 同じ トランザクション で メンバー (principal) に なり admin の 権限 を 受け取る (そう し ない と 作成 者 自身 が 入れ ない ワークスペース が でき て しまう)。 slug は 小文字 英数字 と ハイフン だけ で、 全体 で 一意。 認証 済み なら 誰 でも 作れる (中身 が 空 の テナント が 増える だけ で、 既存 の ワークスペース へ の アクセス は 増え ない) が、 slug の 掴み取り を 抑える ため 作成 だけ は レート 制限 が かかる。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        body  body      kbCreateWorkspaceRequest  true  "作成 内容 (slug/name 必須)"
//	@Success      201   {object}  kbWorkspaceResponse
//	@Failure      400   {object}  errorResponse  "バリデーション エラー"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      409   {object}  errorResponse  "slug が 使用 済み"
//	@Failure      429   {object}  errorResponse  "レート制限超過"
//	@Header       429   {string}  Retry-After    "再試行までの秒数 (例: 60)"
//	@Failure      500   {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces [post]
//	@Security     CookieAuth
func (h *KnowledgeBaseWorkspaceHandler) Create(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	ws, err := h.createWorkspace.Execute(c.Request.Context(), usecase.CreateWorkspaceInput{
		Slug:        req.Slug,
		Name:        req.Name,
		OwnerUserID: uid,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toKbWorkspaceResponse(ws))
}

// kbCreateSpaceRequest はスペース作成の入力。
type kbCreateSpaceRequest struct {
	Key  string `json:"key"  binding:"required" example:"eng"`
	Name string `json:"name" binding:"required,max=200" example:"開発部"`
}

// CreateSpace はワークスペース配下にスペースを作る（ワークスペースの admin が要る）。
//
//	@Summary      ナレッジ 基盤 の スペース 作成
//	@Description  ワークスペース 配下 に スペース を 作る。 ワークスペース 全体 で admin の 者 だけ が 作れる。 スペース は 権限 の 既定 を 持つ 入れ物 な の で、 作れる 相手 を 締め た 側 から 始める (あと から 緩める の は 安全 だ が、 緩い まま 出し て から 締める と 既に 作ら れ た スペース を どう 扱う か 決め られ なく なる)。 key は ワークスペース 内 で 一意。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string                true  "ワークスペース の slug"
//	@Param        body           body      kbCreateSpaceRequest  true  "作成 内容 (key/name 必須)"
//	@Success      201            {object}  kbSpaceResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "ワークスペース の admin で は ない"
//	@Failure      404            {object}  errorResponse  "ワークスペース が 無い か 未 所属"
//	@Failure      409            {object}  errorResponse  "key が 使用 済み"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/spaces [post]
//	@Security     CookieAuth
func (h *KnowledgeBaseWorkspaceHandler) CreateSpace(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	perm, err := h.checkWorkspace.Execute(c.Request.Context(), usecase.CheckWorkspacePermissionInput{
		WorkspaceID: scope.workspaceID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	if !perm.CanManage {
		// ここに来る相手はワークスペースのメンバー（middleware が確かめている）なので、
		// 実在は既に知っている。403 で理由を返してよい。
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	space, err := h.createSpace.Execute(c.Request.Context(), usecase.CreateSpaceInput{
		WorkspaceID: scope.workspaceID,
		Key:         req.Key,
		Name:        req.Name,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toKbSpaceResponse(space))
}
