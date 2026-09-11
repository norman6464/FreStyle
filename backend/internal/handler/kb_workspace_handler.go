package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
)

// KnowledgeBaseWorkspaceHandler はナレッジのワークスペース / スペースの操作を受ける。
//
// ページ操作（KnowledgeBasePageHandler）と分けているのは、テナントの確定の仕方が違うため。
// 一覧と作成は URL に slug を持たず middleware.KnowledgeBaseWorkspace を通れない
// （通したら「まだ所属していない・まだ存在しない」ワークスペースを扱えない）。
type KnowledgeBaseWorkspaceHandler struct {
	listWorkspaces  *kb.ListMemberWorkspacesUseCase
	createWorkspace *kb.CreateWorkspaceUseCase
	deleteWorkspace *kb.DeleteWorkspaceUseCase
	checkWorkspace  *kb.CheckWorkspacePermissionUseCase
	createSpace     *kb.CreateSpaceUseCase
	listSpaces      *kb.ListViewableSpacesUseCase
	checkSpace      *kb.CheckSpacePermissionUseCase
	renameSpace     *kb.RenameSpaceUseCase
	searchPages     *kb.SearchViewablePagesUseCase
	listMembers     *kb.ListWorkspaceMembersUseCase
}

// NewKnowledgeBaseWorkspaceHandler は KnowledgeBaseWorkspaceHandler を組み立てる。
func NewKnowledgeBaseWorkspaceHandler(
	listWorkspaces *kb.ListMemberWorkspacesUseCase,
	createWorkspace *kb.CreateWorkspaceUseCase,
	deleteWorkspace *kb.DeleteWorkspaceUseCase,
	checkWorkspace *kb.CheckWorkspacePermissionUseCase,
	createSpace *kb.CreateSpaceUseCase,
	listSpaces *kb.ListViewableSpacesUseCase,
	checkSpace *kb.CheckSpacePermissionUseCase,
	renameSpace *kb.RenameSpaceUseCase,
	searchPages *kb.SearchViewablePagesUseCase,
	listMembers *kb.ListWorkspaceMembersUseCase,
) *KnowledgeBaseWorkspaceHandler {
	return &KnowledgeBaseWorkspaceHandler{
		listWorkspaces:  listWorkspaces,
		createWorkspace: createWorkspace,
		deleteWorkspace: deleteWorkspace,
		checkWorkspace:  checkWorkspace,
		createSpace:     createSpace,
		listSpaces:      listSpaces,
		checkSpace:      checkSpace,
		renameSpace:     renameSpace,
		searchPages:     searchPages,
		listMembers:     listMembers,
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
	workspaces, err := h.listWorkspaces.Execute(c.Request.Context(), kb.ListMemberWorkspacesInput{UserID: uid})
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
	ws, err := h.createWorkspace.Execute(c.Request.Context(), kb.CreateWorkspaceInput{
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
	spaces, err := h.listSpaces.Execute(c.Request.Context(), kb.ListViewableSpacesInput{
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

// kbWorkspaceMemberResponse はワークスペースに属する人 1 件の返却形。
type kbWorkspaceMemberResponse struct {
	// PrincipalID は担当の割り当て先として使う ID（主体を指す）。
	PrincipalID string `json:"principalId" example:"0198a000-0000-7000-8000-00000000000a"`
	// UserID は発言中の名指しが指す ID（ユーザーを指す）。用途が違うので両方返す。
	UserID uint64 `json:"userId" example:"42"`
	// Name は表示名。引けなかった場合は空文字（行は落とさない）。
	Name string `json:"name" example:"田中 太郎"`
}

// ListMembers はワークスペースに属する人を表示名つきで返す。
//
// 所属していれば誰でも叩ける。担当の名前を出すことと発言で人を名指すことは、
// 権限を変えられない人にも要るため（権限を張る相手を選ぶ ListGrantablePrincipals とは
// 別の口にしてある。あちらはページの管理権限を要求する）。
func (h *KnowledgeBaseWorkspaceHandler) ListMembers(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	members, err := h.listMembers.Execute(c.Request.Context(), scope.workspaceID)
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	out := make([]kbWorkspaceMemberResponse, 0, len(members))
	for _, m := range members {
		out = append(out, kbWorkspaceMemberResponse{PrincipalID: m.PrincipalID, UserID: m.UserID, Name: m.Name})
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
// ナレッジが入るうえ、消しても起動時のバックフィルが作り直すため、中身だけが消えた
// 空のワークスペースが残る。
//
// # 消えるもの
//
// 配下のスペース・ページ・本文・所属・権限・共有リンクがすべて消える（FK の CASCADE）。
// ユーザー（users）は消えない — ナレッジの片付けで人を消さない。
func (h *KnowledgeBaseWorkspaceHandler) Delete(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	perm, err := h.checkWorkspace.Execute(c.Request.Context(), kb.CheckWorkspacePermissionInput{
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
	if err := h.deleteWorkspace.Execute(c.Request.Context(), kb.DeleteWorkspaceInput{
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
		perm, err := h.checkWorkspace.Execute(c.Request.Context(), kb.CheckWorkspacePermissionInput{
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
	space, err := h.createSpace.Execute(c.Request.Context(), kb.CreateSpaceInput{
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
	perm, err := h.checkSpace.Execute(c.Request.Context(), kb.CheckSpacePermissionInput{
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
	space, err := h.renameSpace.Execute(c.Request.Context(), kb.RenameSpaceInput{
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

// kbSearchPageResponse は検索結果 1 件の返却形。kbPageResponse に「どこにヒットしたか」
// を足したもの（本文検索）。
//
// kbPageResponse を埋め込むのは、ページとしての形（id / title / icon …）は既存のツリー・
// 一覧の応答と完全に同じにするため。フロントは検索結果もページ一覧と同じ描画に流用できる。
type kbSearchPageResponse struct {
	kbPageResponse
	// MatchField はヒットした場所（"title" | "body"）。
	MatchField string `json:"matchField" example:"title"`
	// Excerpt は MatchField が "body" のときだけ返す、ヒット周辺の抜粋
	// （前後 30 文字程度。rune 境界を壊さずに切り出してある）。
	Excerpt string `json:"excerpt,omitempty" example:"…この段落には設計メモが含まれている…"`
	// MatchStart / MatchLen は **Excerpt の中での** ヒット位置・長さ（rune 単位。
	// フロントが mark で囲むための材料）。MatchField が "title" のときは出さない。
	MatchStart int `json:"matchStart,omitempty" example:"6"`
	MatchLen   int `json:"matchLen,omitempty" example:"4"`
}

func toKbSearchPageResponse(r *kb.SearchViewablePageResult) kbSearchPageResponse {
	resp := kbSearchPageResponse{
		kbPageResponse: toKbPageResponse(&r.Page),
		MatchField:     r.MatchField,
	}
	if r.MatchField == kb.SearchMatchFieldBody {
		resp.Excerpt = r.Excerpt
		resp.MatchStart = r.MatchStart
		resp.MatchLen = r.MatchLen
	}
	return resp
}

// SearchPages はワークスペース全体を題名 **または本文** で検索する
// （閲覧できるページだけが返る。本文検索に対応）。
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
	results, err := h.searchPages.Execute(c.Request.Context(), kb.SearchViewablePagesInput{
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
	out := make([]kbSearchPageResponse, 0, len(results))
	for i := range results {
		out = append(out, toKbSearchPageResponse(&results[i]))
	}
	c.JSON(http.StatusOK, out)
}
