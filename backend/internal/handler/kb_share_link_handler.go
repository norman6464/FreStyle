package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/infra/ratelimit"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// KnowledgeBaseShareLinkHandler はページの公開 URL（共有リンク）の発行・一覧・失効と、
// 受け取った側の検証を受ける。発行・一覧・失効は認証必須でページの属するスペースの admin
// だけが通る。検証（Verify）だけは未認証で通す — リンクを受け取った人はログインしていない
// ため、ルート登録も認証済み group の外に置く（routes_knowledge_base.go）。
type KnowledgeBaseShareLinkHandler struct {
	*kbPermissionGate
	issue  *kb.IssueShareLinkUseCase
	revoke *kb.RevokeShareLinkUseCase
	list   *kb.ListPageShareLinksUseCase
	verify *kb.VerifyShareLinkUseCase
	// verifyAttempts はリンク 1 本あたりの検証試行の上限（kbShareLinkAttemptKey を参照）。
	verifyAttempts *ratelimit.Limiter
}

func NewKnowledgeBaseShareLinkHandler(
	gate *kbPermissionGate,
	issue *kb.IssueShareLinkUseCase,
	revoke *kb.RevokeShareLinkUseCase,
	list *kb.ListPageShareLinksUseCase,
	verify *kb.VerifyShareLinkUseCase,
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
// 鍵は IP ではなくトークンにする。パスワードは人が選ぶ短い値で総当たりに弱く、鍵に IP を
// 選ぶと家庭・オフィスの NAT の裏にいる無関係な複数人が同じ鍵を共有してしまう（RealClientIP
// で詐称は防いでいても、IP が「1 人」を表すとは限らない）。守りたいのは「このリンクの
// パスワードを当てられないこと」なので、鍵は守る対象そのもの＝リンクに取る。IP 単位の
// 上限はルート側に別で残すが、あれは素直な大量アクセスを薄める層でしかなく秘密を守る
// 根拠にはしない。
//
// トークンそのものではなくハッシュを鍵にするのは、limiter の map に平文トークンを残すと
// それを読めた相手がそのままリンクを開けてしまうため。さらに保存済みハッシュ（token_hash）
// とも違う値にするため前置き文字列を混ぜている — 一致させるとメモリ上の鍵がそのまま DB を
// 引ける値になる。
func kbShareLinkAttemptKey(token string) string {
	sum := sha256.Sum256([]byte("kb-share-link-verify\x00" + token))
	return hex.EncodeToString(sum[:])
}

// kbShareLinkResponse は共有リンク 1 件の返却形。トークンは載せない — domain.ShareLink が
// 持つ SHA-256（TokenHash）は json:"-" 済みだが、応答型を別に定義することで「domain をそのまま
// 返したら秘密が増えていた」という事故を防ぐ。principalId も載せない（クライアントは使わない
// 内部主体）。パスワードは有無だけ載せ、値は出さない。
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

// kbIssuedShareLinkResponse は発行直後だけ返る形。token は平文で返るのはこの 1 回だけ
// （DB には SHA-256 しか残らない）。失うと同じリンクは二度と取り出せず再発行になる。
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

// kbVerifyShareLinkRequest は共有リンク検証の入力。トークンをクエリや path ではなくボディで
// 受けるのは、URL に載せるとアクセスログ・プロキシのログ・履歴・Referer に平文で残るため。
type kbVerifyShareLinkRequest struct {
	Token string `json:"token" binding:"required"`
	// Password はパスワード付きリンクのときに要る。
	Password string `json:"password,omitempty"`
}

// ListShareLinks はページに発行済みの共有リンクを返す（失効済みも含む）。
func (h *KnowledgeBaseShareLinkHandler) ListShareLinks(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePageAdmin(c, scope, pageID) {
		return
	}
	links, err := h.list.Execute(c.Request.Context(), kb.ListPageShareLinksInput{
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
	out, err := h.issue.Execute(c.Request.Context(), kb.IssueShareLinkInput{
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
	// 認可はページ単位で判断しているので、リンクが本当にそのページのものかを必ず確かめる。
	// 確かめないと、自分が admin のスペースのページ ID と他スペースのリンク ID を
	// 組み合わせるだけで、他スペースの共有リンクを止められる（RevokeShareLinkUseCase は
	// ワークスペースとリンク ID しか見ない）。
	links, err := h.list.Execute(c.Request.Context(), kb.ListPageShareLinksInput{
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
	if err := h.revoke.Execute(c.Request.Context(), kb.RevokeShareLinkInput{
		WorkspaceID: scope.workspaceID,
		ShareLinkID: shareLinkID,
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// VerifyShareLink は共有 URL のトークン（とパスワード）を検証する。**認証は要らない。**
func (h *KnowledgeBaseShareLinkHandler) VerifyShareLink(c *gin.Context) {
	limitKnowledgeBaseBody(c)
	var req kbVerifyShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	// 試行回数は判定より前に消費する。あとから数えると bcrypt の照合待ちの間に並んだ要求が
	// 全部素通りしてしまう（鍵の作り方は kbShareLinkAttemptKey を参照）。
	attemptKey := kbShareLinkAttemptKey(req.Token)
	if !h.verifyAttempts.Allow(attemptKey) {
		middleware.RespondRateLimited(c)
		return
	}
	link, err := h.verify.Execute(c.Request.Context(), kb.VerifyShareLinkInput{
		Token:    req.Token,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, repository.ErrShareLinkNotFound) {
			// 守る対象が無いので鍵ごと捨てる。残すと、でたらめなトークンを投げ続けるだけで
			// limiter の中身を太らせられる。パスワードの総当たりには実在するトークン（256 bit
			// の乱数、当てられない）が要るので、ここを数えないことで守りが緩むことはない。
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
	if errors.Is(err, kb.ErrInvalidCapability) {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	respondKbPermissionOperationErr(c, err)
}

// respondKbShareLinkVerifyErr は検証時のエラーを応答へ落とす。
//
// ここだけは理由ごとに撃ち分ける。他の権限操作 API が 404 に揃えるのは ID の総当たりで
// 実在を数え上げられるのを防ぐためだが、共有リンクのトークンは 256 bit の乱数で、
// それを提示できる相手はリンクを渡された本人。「期限切れなので再発行」「パスワードが違う」を
// 区別できないと次に何をすべきか分からない。
//
// 撃ち分けを許してよい根拠は、リンク 1 本あたりの試行回数の上限（VerifyShareLink /
// kbShareLinkAttemptKey）であって、ルート登録側の IP 単位の上限（XFF 詐称で鍵が変わる、
// 素直な大量アクセスを薄める層でしかない）ではない。
func respondKbShareLinkVerifyErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrShareLinkNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
	case errors.Is(err, kb.ErrShareLinkRevoked):
		c.JSON(http.StatusGone, errorResponse{Error: "share_link_revoked"})
	case errors.Is(err, kb.ErrShareLinkExpired):
		c.JSON(http.StatusGone, errorResponse{Error: "share_link_expired"})
	case errors.Is(err, kb.ErrShareLinkPasswordRequired):
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "password_required"})
	case errors.Is(err, kb.ErrShareLinkPasswordMismatch):
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "password_mismatch"})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
	}
}
