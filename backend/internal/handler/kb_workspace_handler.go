package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// KnowledgeBaseWorkspaceHandler はノートのワークスペース / スペースの操作を受ける。
//
// ページ操作（KnowledgeBasePageHandler）と分けているのは、テナントの確定の仕方が違うため。
// 一覧と作成は URL に slug を持たず middleware.KnowledgeBaseWorkspace を通れない
// （通したら「まだ所属していない・まだ存在しない」ワークスペースを扱えない）。
type KnowledgeBaseWorkspaceHandler struct {
	listWorkspaces  *usecase.ListMemberWorkspacesUseCase
	joinCompany     *usecase.JoinCompanyWorkspaceUseCase
	createWorkspace *usecase.CreateWorkspaceUseCase
	deleteWorkspace *usecase.DeleteWorkspaceUseCase
	checkWorkspace  *usecase.CheckWorkspacePermissionUseCase
	createSpace     *usecase.CreateSpaceUseCase
	listSpaces      *usecase.ListViewableSpacesUseCase
	checkSpace      *usecase.CheckSpacePermissionUseCase
	renameSpace     *usecase.RenameSpaceUseCase
	searchPages     *usecase.SearchViewablePagesUseCase
}

// NewKnowledgeBaseWorkspaceHandler は KnowledgeBaseWorkspaceHandler を組み立てる。
func NewKnowledgeBaseWorkspaceHandler(
	listWorkspaces *usecase.ListMemberWorkspacesUseCase,
	joinCompany *usecase.JoinCompanyWorkspaceUseCase,
	createWorkspace *usecase.CreateWorkspaceUseCase,
	deleteWorkspace *usecase.DeleteWorkspaceUseCase,
	checkWorkspace *usecase.CheckWorkspacePermissionUseCase,
	createSpace *usecase.CreateSpaceUseCase,
	listSpaces *usecase.ListViewableSpacesUseCase,
	checkSpace *usecase.CheckSpacePermissionUseCase,
	renameSpace *usecase.RenameSpaceUseCase,
	searchPages *usecase.SearchViewablePagesUseCase,
) *KnowledgeBaseWorkspaceHandler {
	return &KnowledgeBaseWorkspaceHandler{
		listWorkspaces:  listWorkspaces,
		joinCompany:     joinCompany,
		createWorkspace: createWorkspace,
		deleteWorkspace: deleteWorkspace,
		checkWorkspace:  checkWorkspace,
		createSpace:     createSpace,
		listSpaces:      listSpaces,
		checkSpace:      checkSpace,
		renameSpace:     renameSpace,
		searchPages:     searchPages,
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
	// CanManage は自分がこのワークスペースの admin か（削除操作を出してよいかの判定に使う。
	// DeleteWorkspace が要求する権限と同じ）。
	CanManage bool `json:"canManage"`
}

func toKbWorkspaceResponse(w *domain.Workspace, canManage bool) kbWorkspaceResponse {
	return kbWorkspaceResponse{Slug: w.Slug, Name: w.Name, CreatedAt: w.CreatedAt, CanManage: canManage}
}

// kbSpaceResponse はスペース 1 件の返却形。
// id は載せる（ページ一覧・作成の URL がスペース ID を取るため）。
type kbSpaceResponse struct {
	ID   string `json:"id"  example:"0198a000-0000-7000-8000-000000000002"`
	Key  string `json:"key" example:"eng"`
	Name string `json:"name" example:"開発部"`
	// Visibility はサイドバーの節分けに使う（workspace = チーム / private = プライベート）。
	Visibility string    `json:"visibility" example:"workspace"`
	CreatedAt  time.Time `json:"createdAt"`
}

func toKbSpaceResponse(s *domain.Space) kbSpaceResponse {
	return kbSpaceResponse{
		ID: s.ID, Key: s.Key, Name: s.Name,
		Visibility: string(s.Visibility), CreatedAt: s.CreatedAt,
	}
}

// List は自分が所属するワークスペースの一覧を返す。
func (h *KnowledgeBaseWorkspaceHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	// 会社のワークスペースへは自動で入る。一覧はノートに入る最初の口なので、
	// ここで所属を用意しておけば以降の経路（木・ページ・検索）は既存のままで通る。
	//
	// 会社に属さないユーザー（運営管理者など）は入れる先が無いだけなので、
	// ErrWorkspaceNotFound は一覧の失敗にしない。それ以外の失敗は握り潰さず 500 にする
	// （所属を用意できていないのに空の一覧を返すと「会社のページが無い」に見える）。
	if _, err := h.joinCompany.Execute(c.Request.Context(), usecase.JoinCompanyWorkspaceInput{
		UserID: uid,
	}); err != nil && !errors.Is(err, repository.ErrWorkspaceNotFound) {
		respondKnowledgeBaseErr(c, err)
		return
	}
	workspaces, err := h.listWorkspaces.Execute(c.Request.Context(), usecase.ListMemberWorkspacesInput{UserID: uid})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	out := make([]kbWorkspaceResponse, 0, len(workspaces))
	for i := range workspaces {
		out = append(out, toKbWorkspaceResponse(&workspaces[i].Workspace, workspaces[i].CanManage))
	}
	c.JSON(http.StatusOK, out)
}

// kbCreateWorkspaceRequest はワークスペース作成の入力。
// slug は空でよく、空ならサーバーが自動採番する（URL 名を人に決めさせない）。
type kbCreateWorkspaceRequest struct {
	Slug string `json:"slug" example:"acme"`
	Name string `json:"name" binding:"required,max=200" example:"Acme 社"`
}

// Create はワークスペースを作り、作成者をその admin にする。
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
	// 作成者は同じトランザクションで admin の grant を受け取る（ProvisionWorkspace の契約）。
	c.JSON(http.StatusCreated, toKbWorkspaceResponse(ws, true))
}

// ListSpaces はワークスペース配下のスペースのうち、自分が閲覧できるものだけを返す。
//
// スペース ID を知る唯一の口。ページの木を取る API（GET .../spaces/{spaceId}/pages）が
// spaceId を要求するので、これが無いとスペースを作った本人以外は木にたどり着けない。
//
// # 権限のふるい（この口が守っているもの）
//
// 返すのは「そのスペースの中身を既定で閲覧できる」相手にだけ。スペースは権限を分けるための
// 入れ物なので、key と name が並ぶだけでも中で何が進んでいるかが伝わってしまう。
// ふるいは usecase（domain.ResolveScopePermission）が掛け、handler は結果を並べるだけ。
//
// # 存在オラクルを作らない
//
// 権限の無いワークスペースと存在しないワークスペースは、どちらも middleware が 404 に
// 畳んでいる（middleware.KnowledgeBaseWorkspace）。ここに到達した時点で呼び出し元は
// 必ずそのワークスペースのメンバーなので、あとは「見えるスペースだけを並べる」で足りる。
// 1 件も見えなくても 404 にはしない（空配列）。スペースの実在を撃ち分けないのは
// ページの木（Tree）と同じ扱い。
//
// # ページは含めない
//
// サイドバーはスペースごとに木を取るので、この一覧はスペースだけでよい。ページまで
// 抱き合わせると、開いていないスペースの中身まで毎回引くことになる。
func (h *KnowledgeBaseWorkspaceHandler) ListSpaces(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	// スペースごとに権限を引くと N+1 になるので、一覧はまとめて 1 回で解決する
	// （ページの木を返す Tree と同じ作り）。
	spaces, err := h.listSpaces.Execute(c.Request.Context(), usecase.ListViewableSpacesInput{
		WorkspaceID: scope.workspaceID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	// 0 件でも null ではなく [] を返す（make で長さ 0 のスライスを作ってある）。
	// null になるとフロントの .map / for-of が TypeError で落ちる。
	out := make([]kbSpaceResponse, 0, len(spaces))
	for i := range spaces {
		out = append(out, toKbSpaceResponse(&spaces[i]))
	}
	c.JSON(http.StatusOK, out)
}

// Delete はワークスペースを配下ごと消す（戻せない）。
//
// # 誰が消せるか
//
// そのワークスペースの admin だけ。所属は middleware が確かめており、ここに来る相手は
// 必ずメンバーなので、admin でなければ 403 で理由を返してよい（実在は既に知っている）。
//
// # 会社のワークスペースは誰にも消せない
//
// 判定は repository（さらに SQL の WHERE）が持つ。認可と違って**誰であっても消しては
// いけない**ものなので、入口ではなく最も内側で守る。会社のワークスペースには全員の
// ノートが入るうえ、消しても起動時のバックフィルが作り直すため、中身だけが消えた
// 空のワークスペースが残る。
//
// # 消えるもの
//
// 配下のスペース・ページ・本文・所属・権限・共有リンクがすべて消える（FK の CASCADE）。
// ユーザー（users）は消えない — ノートの片付けで人を消さない。
func (h *KnowledgeBaseWorkspaceHandler) Delete(c *gin.Context) {
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
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	if err := h.deleteWorkspace.Execute(c.Request.Context(), usecase.DeleteWorkspaceInput{
		WorkspaceID: scope.workspaceID,
	}); err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// kbCreateSpaceRequest はスペース作成の入力。
// key は空でよく、空ならサーバーが自動採番する（URL 名を人に決めさせない）。
type kbCreateSpaceRequest struct {
	Key  string `json:"key" example:"eng"`
	Name string `json:"name" binding:"required,max=200" example:"開発部"`
	// Visibility は省略時 workspace（チームスペース）。private は自分だけの区画で、
	// メンバーなら誰でも作れる（作れる範囲の非対称は handler が判定する）。
	Visibility string `json:"visibility,omitempty" binding:"omitempty,oneof=workspace private" example:"workspace"`
}

// CreateSpace はワークスペース配下にスペースを作る（ワークスペースの admin が要る）。
func (h *KnowledgeBaseWorkspaceHandler) CreateSpace(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	// **プライベートでは key を人に決めさせない（必ず自動採番）。**
	// key はワークスペース内で一意で、チームとプライベートで同じ名前空間を共有する。
	// 明示指定を許すと、メンバーが任意の key で作成を試して「409 が返るか」だけで
	// 一覧にも木にも出ないはずの他人のプライベートスペースの実在を言い当てられる
	// （作成という書き込みの口が、伏せた実在を読む口になる）。あわせて、意味のある
	// key（"eng" など）を先に取られて admin がチームスペースを作れなくなる占有も防ぐ。
	// 自動採番の key は衝突しても usecase が引き直すので、409 自体が表に出ない。
	if req.Visibility == string(domain.SpaceVisibilityPrivate) && req.Key != "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	// 作れる範囲は非対称: チームスペース（workspace）は全員に見える入れ物が増えるので
	// admin だけ。プライベートは自分の区画が増えるだけ（他人の見えるものは変わらない）
	// なので、メンバーなら誰でも作れる。所属は middleware が確かめ済み。
	if req.Visibility != string(domain.SpaceVisibilityPrivate) {
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
	}
	space, err := h.createSpace.Execute(c.Request.Context(), usecase.CreateSpaceInput{
		WorkspaceID:   scope.workspaceID,
		Key:           req.Key,
		Name:          req.Name,
		Visibility:    domain.SpaceVisibility(req.Visibility),
		CreatorUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toKbSpaceResponse(space))
}

type kbRenameSpaceRequest struct {
	Name string `json:"name" binding:"required,max=200" example:"開発部 (改組)"`
}

// RenameSpace はスペースの表示名を変える（key は変えない）。
func (h *KnowledgeBaseWorkspaceHandler) RenameSpace(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	perm, err := h.checkSpace.Execute(c.Request.Context(), usecase.CheckSpacePermissionInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     spaceID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	if !perm.CanView {
		// 中身を 1 つも見られない相手にはスペースの実在を教えない（他の口と同じ畳み方）。
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}
	if !perm.CanManage {
		// 見えている相手には理由を返してよい。入れ物そのものの変更は管理権限。
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbRenameSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	space, err := h.renameSpace.Execute(c.Request.Context(), usecase.RenameSpaceInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     spaceID,
		Name:        req.Name,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbSpaceResponse(space))
}

// SearchPages はワークスペース全体を題名で検索する（閲覧できるページだけが返る）。
func (h *KnowledgeBaseWorkspaceHandler) SearchPages(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	q := strings.TrimSpace(c.Query("q"))
	// 空は「全件」ではなく誤りとして返す。空で全件を返すと、この口が
	// 「見えるページの全数を数える口」になってしまう（見せてよいのは一致した分だけ）。
	if q == "" || utf8.RuneCountInString(q) > 100 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_query"})
		return
	}
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	pages, err := h.searchPages.Execute(c.Request.Context(), usecase.SearchViewablePagesInput{
		WorkspaceID: scope.workspaceID,
		UserID:      scope.userID,
		Query:       q,
		Limit:       limit,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	// 0 件でも [] を返す（null だとフロントの .map が落ちる）。
	out := make([]kbPageResponse, 0, len(pages))
	for i := range pages {
		out = append(out, toKbPageResponse(&pages[i]))
	}
	c.JSON(http.StatusOK, out)
}
