package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/ratelimit"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// KnowledgeBaseShareLinkHandler はページの公開 URL（共有リンク）の発行・一覧・失効と、
// 受け取った側の検証を受ける。
//
// 発行・一覧・失効は認証必須でページの属するスペースの admin だけが通る。
// 検証（Verify）だけは未認証で通す — リンクを受け取った人はログインしていない。
// 認可の形が 1 つだけ違うので、ルート登録も認証済み group の外に置く
// （routes_knowledge_base.go の registerKnowledgeBasePublicRoutes）。
type KnowledgeBaseShareLinkHandler struct {
	*kbPermissionGate
	issue  *usecase.IssueShareLinkUseCase
	revoke *usecase.RevokeShareLinkUseCase
	list   *usecase.ListPageShareLinksUseCase
	verify *usecase.VerifyShareLinkUseCase
	// verifyAttempts はリンク 1 本あたりの検証試行の上限（kbShareLinkAttemptKey を参照）。
	verifyAttempts *ratelimit.Limiter
}

// NewKnowledgeBaseShareLinkHandler は KnowledgeBaseShareLinkHandler を組み立てる。
// verifyAttempts はリンク 1 本あたりの検証試行を絞る limiter（VerifyShareLink だけが使う）。
func NewKnowledgeBaseShareLinkHandler(
	gate *kbPermissionGate,
	issue *usecase.IssueShareLinkUseCase,
	revoke *usecase.RevokeShareLinkUseCase,
	list *usecase.ListPageShareLinksUseCase,
	verify *usecase.VerifyShareLinkUseCase,
	verifyAttempts *ratelimit.Limiter,
) *KnowledgeBaseShareLinkHandler {
	return &KnowledgeBaseShareLinkHandler{
		kbPermissionGate: gate,
		issue:            issue,
		revoke:           revoke,
		list:             list,
		verify:           verify,
		verifyAttempts:   verifyAttempts,
	}
}

// kbShareLinkAttemptKey は共有リンクの検証回数を数えるときの鍵を作る。
//
// # なぜ IP ではなくトークンを鍵にするのか
//
// パスワード付きリンクのパスワードは人が選ぶ短い値で、総当たりに弱い。それを抑える上限の
// 鍵に IP を選ぶと、**攻撃者が鍵を自由に変えられる**（gin の ClientIP は X-Forwarded-For の
// 最左を読み、このリポジトリは SetTrustedProxies を呼んでいないので詐称できる。実測でも
// XFF を毎回変えれば 200 回連続で通った）。鍵を変えられる上限は、上限として機能しない。
//
// 守りたいのは「このリンクのパスワードを当てられないこと」なので、鍵は**守る対象そのもの**、
// すなわちリンクに取る。こうすると IP をいくら変えても、リンク 1 本あたりの試行回数は
// 必ず頭打ちになる。IP 単位の上限はルート側に残してあるが、あれは素直な大量アクセスを
// 薄める層でしかなく、秘密を守る根拠にはしない。
//
// # なぜトークンそのものではなくハッシュを鍵にするのか
//
// 鍵は limiter の map にしばらく残る。平文トークンを置くと、その map を読めた相手が
// そのままリンクを開ける。ハッシュなら鍵からリンクは開けない。
//
// # なぜ保存されているハッシュ（SHA-256 そのもの）と別の値にするのか
//
// 前置きの文字列を混ぜて、share_links.token_hash と一致しない値にしてある。
// 一致させると、メモリ上の鍵がそのまま DB を引ける値になる（鍵は照合に使うだけで、
// DB と同じである必要はない）。用途が違う値は別の値にしておく。
func kbShareLinkAttemptKey(token string) string {
	sum := sha256.Sum256([]byte("kb-share-link-verify\x00" + token))
	return hex.EncodeToString(sum[:])
}

// kbShareLinkResponse は共有リンク 1 件の返却形。
//
// トークンは載せない。domain.ShareLink が持つのは SHA-256（TokenHash）だけで、
// それ自体 json:"-" で隠してあるが、この応答型を別に定義することで
// 「domain の構造体をそのまま返したらいつの間にか秘密が増えていた」という事故を防ぐ。
// principalId も載せない（リンクの来訪者を表す内部の主体で、クライアントは使わない）。
// パスワードは有無だけを載せる（入力欄を出すかの判断に要るが、値は出さない）。
type kbShareLinkResponse struct {
	ID     string `json:"id"     example:"0198a000-0000-7000-8000-00000000000c"`
	PageID string `json:"pageId" example:"0198a000-0000-7000-8000-000000000003"`
	// Capability はリンク経由でできることの既定（view / edit）。
	Capability string `json:"capability" example:"view"`
	// RequiresPassword はパスワード付きのリンクか。
	RequiresPassword bool       `json:"requiresPassword" example:"false"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
	CreatedByUserID  uint64     `json:"createdByUserId" example:"42"`
	CreatedAt        time.Time  `json:"createdAt"`
}

func toKbShareLinkResponse(l *domain.ShareLink) kbShareLinkResponse {
	return kbShareLinkResponse{
		ID:               l.ID,
		PageID:           l.PageID,
		Capability:       string(l.Capability),
		RequiresPassword: l.RequiresPassword(),
		ExpiresAt:        l.ExpiresAt,
		RevokedAt:        l.RevokedAt,
		CreatedByUserID:  l.CreatedByUserID,
		CreatedAt:        l.CreatedAt,
	}
}

// kbIssuedShareLinkResponse は発行直後だけ返る形。
//
// token は平文で、返るのはこの 1 回だけ（DB には SHA-256 しか残らない）。
// 失うと同じリンクは二度と取り出せず再発行になる。一覧（kbShareLinkResponse）には
// 出ないので、発行時の応答をそのまま保存する運用にしないこと。
type kbIssuedShareLinkResponse struct {
	Link kbShareLinkResponse `json:"link"`
	// Token は共有 URL に載せる平文トークン。
	Token string `json:"token" example:"3q2-7uMBEjRWeJq83vzMzQ"`
}

// kbVerifiedShareLinkResponse は検証に通ったリンクの、来訪者に見せてよい範囲。
// ワークスペースや主体の ID は出さない（リンクの持ち主が知る必要が無い内部の識別子）。
type kbVerifiedShareLinkResponse struct {
	PageID     string     `json:"pageId"     example:"0198a000-0000-7000-8000-000000000003"`
	Capability string     `json:"capability" example:"view"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
}

// kbIssueShareLinkRequest は共有リンク発行の入力。
type kbIssueShareLinkRequest struct {
	// Capability はリンク経由でできることの既定（view / edit）。
	Capability string `json:"capability" binding:"required" example:"view"`
	// Password が空でなければパスワード付きにする。応答にもログにも出さない。
	Password string `json:"password,omitempty"`
	// ExpiresAt が未指定なら無期限。過去の時刻は usecase が弾く。
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// kbVerifyShareLinkRequest は共有リンク検証の入力。
//
// トークンをクエリや path ではなくボディで受けるのは、URL に載せると
// アクセスログ・プロキシのログ・ブラウザの履歴・Referer に平文で残るため。
// このリポジトリのアクセスログは c.FullPath()（ルートのパターン）しか出さないので、
// ボディで受ける限りトークンはどこにも記録されない。
type kbVerifyShareLinkRequest struct {
	Token string `json:"token" binding:"required"`
	// Password はパスワード付きリンクのときに要る。
	Password string `json:"password,omitempty"`
}

// ListShareLinks はページに発行済みの共有リンクを返す（失効済みも含む）。
//
//	@Summary      ノート の 共有 リンク 一覧
//	@Description  ページ に 発行 済み の 共有 リンク を 失効 済み も 含め て 返す。 トークン は 発行 時 の 1 回 しか 返ら ない の で、 ここ に は 出 ない (DB に も SHA-256 しか 無い)。 呼べる の は その ページ が 属する スペース の admin (ワークスペース の admin を 含む) だけ で、 権限 が 無い 場合 と ページ が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Param        pageId         path      string  true  "ページ ID (UUID)"
//	@Success      200            {array}   kbShareLinkResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/share-links [get]
//	@Security     CookieAuth
func (h *KnowledgeBaseShareLinkHandler) ListShareLinks(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePageAdmin(c, scope, pageID) {
		return
	}
	links, err := h.list.Execute(c.Request.Context(), usecase.ListPageShareLinksInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	out := make([]kbShareLinkResponse, 0, len(links))
	for i := range links {
		out = append(out, toKbShareLinkResponse(&links[i]))
	}
	c.JSON(http.StatusOK, out)
}

// IssueShareLink はページの公開 URL を発行する。
//
//	@Summary      ノート の 共有 リンク 発行
//	@Description  ページ と その 子孫 を ログイン 不要 で 開ける URL を 発行 する。 応答 の token は 平文 で、 返る の は この 1 回 だけ (DB に は SHA-256 しか 残ら ない)。 失う と 再 発行 に なる。 パスワード を 付ける と 開く 際 に 必要 に なる (値 は 応答 に も ログ に も 出 ない)。 呼べる の は その ページ が 属する スペース の admin (ワークスペース の admin を 含む) だけ で、 権限 が 無い 場合 と ページ が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string                   true  "ワークスペース の slug"
//	@Param        pageId         path      string                   true  "ページ ID (UUID)"
//	@Param        body           body      kbIssueShareLinkRequest  true  "発行 内容 (capability 必須 / password / expiresAt は 任意)"
//	@Success      201            {object}  kbIssuedShareLinkResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/share-links [post]
//	@Security     CookieAuth
func (h *KnowledgeBaseShareLinkHandler) IssueShareLink(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePageAdmin(c, scope, pageID) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbIssueShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	out, err := h.issue.Execute(c.Request.Context(), usecase.IssueShareLinkInput{
		WorkspaceID:     scope.workspaceID,
		PageID:          pageID,
		Capability:      domain.Capability(req.Capability),
		Password:        req.Password,
		ExpiresAt:       req.ExpiresAt,
		CreatedByUserID: scope.userID,
	})
	if err != nil {
		respondKbShareLinkIssueErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, kbIssuedShareLinkResponse{
		Link:  toKbShareLinkResponse(out.Link),
		Token: out.Token,
	})
}

// RevokeShareLink は共有リンクを失効させる（冪等）。
//
//	@Summary      ノート の 共有 リンク 失効
//	@Description  共有 リンク を 失効 さ せる。 行 は 消さ ず revoked_at を 立てる の で、 誰 が いつ 止め た か は 残る。 既に 失効 済み なら 何 も せ ず 成功 する (冪等)。 URL の ページ に 属さ ない リンク ID を 渡し た 場合 は 権限 が 無い の と 同じ 404 (ページ の 権限 で 判断 する 以上、 別 の スペース の リンク を この 口 から 止め られ て は なら ない)。 呼べる の は その ページ が 属する スペース の admin だけ。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        pageId         path  string  true  "ページ ID (UUID)"
//	@Param        shareLinkId    path  string  true  "共有 リンク ID (UUID)"
//	@Success      204            "失効 済み"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/share-links/{shareLinkId} [delete]
//	@Security     CookieAuth
func (h *KnowledgeBaseShareLinkHandler) RevokeShareLink(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePageAdmin(c, scope, pageID) {
		return
	}
	shareLinkID := c.Param("shareLinkId")
	// 認可はページ（が属するスペース）で判断しているので、リンクが本当にそのページの
	// ものかをここで必ず確かめる。確かめないと、自分が admin のスペースのページ ID と
	// 他スペースのリンク ID を組み合わせるだけで、他スペースの共有リンクを止められる
	// （RevokeShareLinkUseCase はワークスペースとリンク ID しか見ない）。
	links, err := h.list.Execute(c.Request.Context(), usecase.ListPageShareLinksInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	belongs := false
	for i := range links {
		if links[i].ID == shareLinkID {
			belongs = true
			break
		}
	}
	if !belongs {
		// 存在しないリンクと、他のページのリンクを、どちらも拒否と同じ応答にする。
		respondKbPermissionDenied(c)
		return
	}
	if err := h.revoke.Execute(c.Request.Context(), usecase.RevokeShareLinkInput{
		WorkspaceID: scope.workspaceID,
		ShareLinkID: shareLinkID,
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// VerifyShareLink は共有 URL のトークン（とパスワード）を検証する。**認証は要らない。**
//
//	@Summary      ノート の 共有 リンク 検証
//	@Description  受け取っ た 共有 リンク の トークン (と パスワード) を 検証 し、 開ける なら 対象 ページ と できる こと を 返す。 リンク を 受け取っ た 人 は ログイン し て い ない の で、 この 経路 だけ は 認証 を 要求 し ない。 トークン は URL で は なく ボディ で 受ける (URL に 載せる と アクセス ログ や Referer に 平文 で 残る ため)。 応答 に トークン は 含め ない。 総当たり と パスワード 推測 を 抑える ため、 リンク 1 本 あたり の 試行 回数 に 上限 が ある (要求 元 の IP を 変え て も 頭打ち に なる)。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        body  body      kbVerifyShareLinkRequest  true  "トークン と (必要 なら) パスワード"
//	@Success      200   {object}  kbVerifiedShareLinkResponse
//	@Failure      400   {object}  errorResponse  "バリデーション エラー"
//	@Failure      401   {object}  errorResponse  "パスワード が 必要 / 一致 し ない"
//	@Failure      404   {object}  errorResponse  "その トークン の リンク は 無い"
//	@Failure      410   {object}  errorResponse  "失効 済み / 期限 切れ"
//	@Failure      429   {object}  errorResponse  "レート制限超過"
//	@Header       429   {string}  Retry-After    "再試行までの秒数 (例: 60)"
//	@Failure      500   {object}  errorResponse  "DB 失敗"
//	@Router       /kb/share-links/verify [post]
func (h *KnowledgeBaseShareLinkHandler) VerifyShareLink(c *gin.Context) {
	limitKnowledgeBaseBody(c)
	var req kbVerifyShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	// リンク 1 本あたりの試行回数を、判定より**前に** 1 つ消費する。あとから数えると、
	// bcrypt の照合を待つあいだに並んだ要求が全部素通りしてしまう（並列化されると
	// 上限が意味を失う）。鍵の作り方と「なぜ IP ではないのか」は kbShareLinkAttemptKey を参照。
	attemptKey := kbShareLinkAttemptKey(req.Token)
	if !h.verifyAttempts.Allow(attemptKey) {
		middleware.RespondRateLimited(c)
		return
	}
	link, err := h.verify.Execute(c.Request.Context(), usecase.VerifyShareLinkInput{
		Token:    req.Token,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, repository.ErrShareLinkNotFound) {
			// そのトークンのリンクは無かった。守る対象が無いので鍵ごと捨てる。
			// 残すと、でたらめなトークンを投げ続けるだけで limiter の中身を攻撃者に
			// 好きなだけ太らせられる（トークンは要求ごとに変えられる）。
			// パスワードの総当たりには実在するトークンが要る（256 bit の乱数は当てられない）ので、
			// ここを数えないことで守りが緩むことはない。
			h.verifyAttempts.Forget(attemptKey)
		}
		respondKbShareLinkVerifyErr(c, err)
		return
	}
	c.JSON(http.StatusOK, kbVerifiedShareLinkResponse{
		PageID:     link.PageID,
		Capability: string(link.Capability),
		ExpiresAt:  link.ExpiresAt,
	})
}

// respondKbShareLinkIssueErr は発行時のエラーを応答へ落とす。
// 期限が過去・ケイパビリティが未知は入力の誤りなので 400（ここへ来る相手は admin）。
func respondKbShareLinkIssueErr(c *gin.Context, err error) {
	if errors.Is(err, usecase.ErrInvalidCapability) {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	respondKbPermissionOperationErr(c, err)
}

// respondKbShareLinkVerifyErr は検証時のエラーを応答へ落とす。
//
// ここだけは理由ごとに撃ち分ける。ほかの権限操作 API が 404 に揃えるのは
// 「ID を総当たりして対象の実在を数え上げられる」ことを防ぐためだが、共有リンクの
// トークンは 256 bit の乱数（推測は現実的でない）で、それを提示できている相手は
// そのリンクを渡された本人。「期限が切れているので再発行を頼む」「パスワードが違う」を
// 区別できないと、受け取った側が次に何をすればよいか分からない。
//
// パスワードは人が選ぶ短い値で総当たりに弱いので、撃ち分けを許すぶんの担保として
// **リンク 1 本あたりの試行回数**に上限をかけている（VerifyShareLink 本体と
// kbShareLinkAttemptKey を参照）。鍵はリンクなので、要求元の IP をいくら変えても
// 同じリンクへの試行は必ず頭打ちになる。
//
// ルート登録側にも IP 単位の上限があるが、あちらは素直な大量アクセスを薄める層でしかない
// （XFF を詐称すれば鍵が変わる）。**撃ち分けを許してよい根拠はこちらの上限**であって、
// あちらではない。対策がこの関数の外にあることに注意。
func respondKbShareLinkVerifyErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrShareLinkNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
	case errors.Is(err, usecase.ErrShareLinkRevoked):
		c.JSON(http.StatusGone, errorResponse{Error: "share_link_revoked"})
	case errors.Is(err, usecase.ErrShareLinkExpired):
		c.JSON(http.StatusGone, errorResponse{Error: "share_link_expired"})
	case errors.Is(err, usecase.ErrShareLinkPasswordRequired):
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "password_required"})
	case errors.Is(err, usecase.ErrShareLinkPasswordMismatch):
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "password_mismatch"})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
	}
}
