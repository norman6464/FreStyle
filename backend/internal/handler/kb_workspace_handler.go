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
}

func toKbWorkspaceResponse(w *domain.Workspace) kbWorkspaceResponse {
	return kbWorkspaceResponse{Slug: w.Slug, Name: w.Name, CreatedAt: w.CreatedAt}
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
//
//	@Summary      ノート の 所属 ワークスペース 一覧
//	@Description  ログイン 中 の ユーザー が 所属 する ワークスペース を 返す。 所属 は principals (kind='user') の 行 が 唯一 の 表現 で、 所属 し て い ない ワークスペース は 1 件 も 含ま ない。 ほか の ノート API が URL に 使う slug を 知る ため の 入口。
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
		out = append(out, toKbWorkspaceResponse(&workspaces[i]))
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
//
//	@Summary      ノート の ワークスペース 作成
//	@Description  ワークスペース を 作る。 作成 者 は 同じ トランザクション で メンバー (principal) に なり admin の 権限 を 受け取る (そう し ない と 作成 者 自身 が 入れ ない ワークスペース が でき て しまう)。 slug は 省略 でき、 空 なら サーバー が 自動 採番 する。 指定 する 場合 は 小文字 英数字 と ハイフン だけ で、 全体 で 一意。 認証 済み なら 誰 でも 作れる (中身 が 空 の テナント が 増える だけ で、 既存 の ワークスペース へ の アクセス は 増え ない) が、 slug の 掴み取り を 抑える ため 作成 だけ は レート 制限 が かかる。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        body  body      kbCreateWorkspaceRequest  true  "作成 内容 (name 必須。 slug は 空 なら 自動 採番)"
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
//
//	@Summary      ノート の スペース 一覧
//	@Description  ワークスペース 配下 の スペース の うち、 呼び出し 元 が 中身 を 閲覧 できる もの だけ を key 順 で 返す。 閲覧 権限 の 無い スペース は 1 件 も 含ま ない (key や name その もの が 情報 に なる ため)。 1 件 も 見え なく て も 空 配列 を 返し、 スペース の 実在 は 撃ち分け ない。 ページ は 含ま ない (木 は スペース ごと に GET /kb/workspaces/{workspaceSlug}/spaces/{spaceId}/pages で 取る)。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Success      200            {array}   kbSpaceResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "ワークスペース が 無い か 未 所属"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/spaces [get]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の スペース 作成
//	@Description  ワークスペース 配下 に スペース を 作る。 チーム スペース (visibility=workspace、 省略 時) は ワークスペース 全体 で admin の 者 だけ が 作れる。 プライベート (visibility=private) は メンバー なら 誰 でも 作れ、 作成 者 だけ に 見える (ワークスペース 既定 の grant が 届か ず、 作成 時 に 作成 者 へ space_grant(admin) を 張る)。 スペース は 権限 の 既定 を 持つ 入れ物 な の で、 作れる 相手 を 締め た 側 から 始める (あと から 緩める の は 安全 だ が、 緩い まま 出し て から 締める と 既に 作ら れ た スペース を どう 扱う か 決め られ なく なる)。 key は 省略 でき、 空 なら サーバー が 自動 採番 する。 指定 する 場合 は ワークスペース 内 で 一意。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string                true  "ワークスペース の slug"
//	@Param        body           body      kbCreateSpaceRequest  true  "作成 内容 (name 必須。 key は 空 なら 自動 採番)"
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
	limitKnowledgeBaseBody(c)
	var req kbCreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
//
//	@Summary      ノート の スペース 改名
//	@Description  表示名 だけ を 変更 する。 key は URL・権限 の 参照 に 使う ため 不変。 スペース の 管理 権限 が 要る。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string                true  "ワークスペース の slug"
//	@Param        spaceId        path      string                true  "スペース ID (UUID)"
//	@Param        body           body      kbRenameSpaceRequest  true  "新しい 表示名"
//	@Success      200            {object}  kbSpaceResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "管理 権限 が 無い"
//	@Failure      404            {object}  errorResponse  "存在 し ない か 閲覧 権限 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/spaces/{spaceId} [patch]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の ページ 題名 検索
//	@Description  ワークスペース 全体 から 題名 の 部分 一致 で 検索 する。 返る の は 閲覧 できる 現役 ページ のみ。 並び は 題名 順。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true   "ワークスペース の slug"
//	@Param        q              query     string  true   "題名 の 部分 一致 (1〜100 文字)"
//	@Param        limit          query     int     false  "最大 件数 (既定 20 / 上限 50)"
//	@Success      200            {array}   kbPageResponse
//	@Failure      400            {object}  errorResponse  "q が 空 か 長 すぎる"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/search [get]
//	@Security     CookieAuth
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
