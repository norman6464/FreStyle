package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// DocumentHandler はリッチテキスト文書（tiptap JSON を jsonb で保持）の CRUD を受ける。
type DocumentHandler struct {
	get    *usecase.GetRichDocumentUseCase
	create *usecase.CreateRichDocumentUseCase
	update *usecase.UpdateRichDocumentUseCase
	del    *usecase.DeleteRichDocumentUseCase
	list   *usecase.ListRichDocumentsUseCase
}

// maxDocumentBodyBytes はリクエストボディの上限。doc 本体の上限（usecase 側 1 MiB）に
// エンベロープ（title 等）の余裕を足した値。bind 前に MaxBytesReader で早期に切り、巨大ボディの
// 全読み込みによるメモリ枯渇を防ぐ（横断的なボディ上限は別途対応）。
const maxDocumentBodyBytes = (1 << 20) + (64 << 10) // 1 MiB + 64 KiB

// limitBody は bind 前にボディサイズ上限を課す。
func limitBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxDocumentBodyBytes)
}

// currentCompanyID は current user の会社 ID を返す（未所属/未設定なら nil）。作成時に文書へ写す用。
// 閲覧側の境界判定は workspace_id で行うため actorWorkspaceFromContext を使う（下記 List/Get）。
func currentCompanyID(c *gin.Context) *uint64 {
	if u := middleware.CurrentUserFromContext(c); u != nil {
		return u.CompanyID
	}
	return nil
}

// NewDocumentHandler は DocumentHandler を組み立てる。
func NewDocumentHandler(
	g *usecase.GetRichDocumentUseCase,
	c *usecase.CreateRichDocumentUseCase,
	u *usecase.UpdateRichDocumentUseCase,
	d *usecase.DeleteRichDocumentUseCase,
	l *usecase.ListRichDocumentsUseCase,
) *DocumentHandler {
	return &DocumentHandler{get: g, create: c, update: u, del: d, list: l}
}

// documentSummaryResponse は一覧の 1 件。doc 本体は含めない（一覧は軽量に保つ・本文は個別取得）。
type documentSummaryResponse struct {
	ID            string    `json:"id"            example:"31400a07-297e-8057-884b-c05dbdf9fa53"`
	OwnerID       uint64    `json:"ownerId"       example:"42"`
	CompanyID     *uint64   `json:"companyId,omitempty" example:"1"`
	WorkspaceID   *string   `json:"workspaceId,omitempty"`
	Kind          string    `json:"kind"          example:"note"`
	Title         string    `json:"title"         example:"学習メモ"`
	IsPublic      bool      `json:"isPublic"      example:"false"`
	SchemaVersion int       `json:"schemaVersion" example:"1"`
	Revision      int       `json:"revision"      example:"1"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func toDocumentSummary(d *domain.RichDocument) documentSummaryResponse {
	return documentSummaryResponse{
		ID:            d.ID,
		OwnerID:       d.OwnerID,
		CompanyID:     d.CompanyID,
		WorkspaceID:   d.WorkspaceID,
		Kind:          string(d.Kind),
		Title:         d.Title,
		IsPublic:      d.IsPublic,
		SchemaVersion: d.SchemaVersion,
		Revision:      d.Revision,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

// documentResponse は文書の返却形。doc は正本 JSON をそのまま埋め込む（json.RawMessage）。
type documentResponse struct {
	ID            string          `json:"id"            example:"31400a07-297e-8057-884b-c05dbdf9fa53"`
	OwnerID       uint64          `json:"ownerId"       example:"42"`
	CompanyID     *uint64         `json:"companyId,omitempty" example:"1"`
	WorkspaceID   *string         `json:"workspaceId,omitempty"`
	Kind          string          `json:"kind"          example:"note"`
	Title         string          `json:"title"         example:"学習メモ"`
	IsPublic      bool            `json:"isPublic"      example:"false"`
	SchemaVersion int             `json:"schemaVersion" example:"1"`
	Doc           json.RawMessage `json:"doc"           swaggertype:"object"`
	Revision      int             `json:"revision"      example:"1"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func toDocumentResponse(d *domain.RichDocument) documentResponse {
	return documentResponse{
		ID:            d.ID,
		OwnerID:       d.OwnerID,
		CompanyID:     d.CompanyID,
		WorkspaceID:   d.WorkspaceID,
		Kind:          string(d.Kind),
		Title:         d.Title,
		IsPublic:      d.IsPublic,
		SchemaVersion: d.SchemaVersion,
		Doc:           json.RawMessage(d.Doc),
		Revision:      d.Revision,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

// normalizeDocumentID は URL の id（ダッシュ有り/無しの UUID どちらも）を正規の UUID 文字列に直す。
func normalizeDocumentID(raw string) (string, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return "", false
	}
	return id.String(), true
}

// respondRichDocErr は usecase のセンチネルを HTTP ステータスへ対応づける（内部詳細は漏らさない）。
func respondRichDocErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrRichDocumentNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
	case errors.Is(err, usecase.ErrRichDocumentConflict):
		c.JSON(http.StatusConflict, errorResponse{Error: "revision_conflict"})
	case errors.Is(err, usecase.ErrRichDocumentForbidden):
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
	case errors.Is(err, usecase.ErrRichDocumentInvalid):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
	}
}

// List は current user 名義の文書一覧を返す（owner スコープ・doc 本体は含まない軽量サマリ）。
//
//	@Summary      自分 の リッチ 文書 一覧
//	@Description  current user が所有する文書を更新日降順で返す (doc 本体は含まない)。 kind で絞り込み可 (note / course-chapter)。
//	@Tags         documents
//	@Produce      json
//	@Param        kind  query     string  false  "用途 で 絞り込み (note / course-chapter)"
//	@Success      200   {array}   documentSummaryResponse
//	@Failure      400   {object}  errorResponse  "不正 な kind"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      500   {object}  errorResponse  "DB 失敗"
//	@Router       /documents [get]
//	@Security     CookieAuth
func (h *DocumentHandler) List(c *gin.Context) {
	uid, workspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	docs, err := h.list.Execute(c.Request.Context(), usecase.ListRichDocumentsInput{
		OwnerID:         uid,
		ViewerWorkspace: workspace,
		Kind:            domain.DocumentKind(c.Query("kind")),
	})
	if err != nil {
		respondRichDocErr(c, err)
		return
	}
	out := make([]documentSummaryResponse, 0, len(docs))
	for i := range docs {
		out = append(out, toDocumentSummary(&docs[i]))
	}
	c.JSON(http.StatusOK, out)
}

type documentCreateReq struct {
	Kind          string          `json:"kind"  binding:"required"`
	Title         string          `json:"title" binding:"required"`
	Doc           json.RawMessage `json:"doc"   binding:"required" swaggertype:"object"`
	IsPublic      bool            `json:"isPublic"`
	SchemaVersion int             `json:"schemaVersion"`
}

// Create は current user 名義で文書を作る。ownerId は body で指定できない（current user 固定・IDOR 対策）。
//
//	@Summary      リッチ 文書 の 作成
//	@Description  current user 名義 で 新規 文書 を 作る。 doc は tiptap の JSON (type='doc')。
//	@Tags         documents
//	@Accept       json
//	@Produce      json
//	@Param        body  body      documentCreateReq  true  "作成 内容 (kind/title/doc 必須)"
//	@Success      201   {object}  documentResponse
//	@Failure      400   {object}  errorResponse  "バリデーション エラー"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      500   {object}  errorResponse  "DB 失敗"
//	@Router       /documents [post]
//	@Security     CookieAuth
func (h *DocumentHandler) Create(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	limitBody(c)
	var req documentCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	doc, err := h.create.Execute(c.Request.Context(), usecase.CreateRichDocumentInput{
		OwnerID:       uid,
		CompanyID:     currentCompanyID(c),
		Kind:          domain.DocumentKind(req.Kind),
		Title:         req.Title,
		Doc:           string(req.Doc),
		IsPublic:      req.IsPublic,
		SchemaVersion: req.SchemaVersion,
	})
	if err != nil {
		respondRichDocErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toDocumentResponse(doc))
}

// Get は文書を 1 件返す。所有者、または同一ワークスペースの公開文書のみ（他ワークスペース・非公開は存在を漏らさず 404）。
//
//	@Summary      リッチ 文書 の 取得
//	@Description  id の 文書 を 返す。 所有者 か 同一 ワークスペース の 公開 文書 のみ (他 ワークスペース の 文書 や 非公開 は 存在 を 漏らさ ず 404)。
//	@Tags         documents
//	@Produce      json
//	@Param        id   path      string  true  "文書 ID (UUID)"
//	@Success      200  {object}  documentResponse
//	@Failure      400  {object}  errorResponse  "不正 な ID"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      404  {object}  errorResponse  "存在 し ない か 権限 無し"
//	@Router       /documents/{id} [get]
//	@Security     CookieAuth
func (h *DocumentHandler) Get(c *gin.Context) {
	uid, workspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	id, ok := normalizeDocumentID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_id"})
		return
	}
	doc, err := h.get.Execute(c.Request.Context(), id, uid, workspace)
	if err != nil {
		respondRichDocErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toDocumentResponse(doc))
}

type documentUpdateReq struct {
	Title         string          `json:"title" binding:"required"`
	Doc           json.RawMessage `json:"doc"   binding:"required" swaggertype:"object"`
	IsPublic      bool            `json:"isPublic"`
	SchemaVersion int             `json:"schemaVersion"`
	Revision      int             `json:"revision" binding:"required"`
}

// Update は文書を更新する（所有者のみ・楽観ロック）。revision 不一致は 409 を返す。
//
//	@Summary      リッチ 文書 の 更新
//	@Description  所有者 のみ 更新 できる (他人 の 文書 は 存在 を 漏らさ ず 404)。 revision が 現在 値 と 一致 する とき だけ 更新 し +1 する。 不一致 は 409。
//	@Tags         documents
//	@Accept       json
//	@Produce      json
//	@Param        id    path      string             true  "文書 ID (UUID)"
//	@Param        body  body      documentUpdateReq  true  "更新 内容 (title/doc/revision 必須)"
//	@Success      200   {object}  documentResponse
//	@Failure      400   {object}  errorResponse  "バリデーション エラー"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      404   {object}  errorResponse  "存在 し ない か 権限 無し"
//	@Failure      409   {object}  errorResponse  "版 不一致 (楽観 ロック)"
//	@Router       /documents/{id} [put]
//	@Security     CookieAuth
func (h *DocumentHandler) Update(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	id, ok := normalizeDocumentID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_id"})
		return
	}
	limitBody(c)
	var req documentUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	doc, err := h.update.Execute(c.Request.Context(), usecase.UpdateRichDocumentInput{
		ID:            id,
		ActorID:       uid,
		Title:         req.Title,
		Doc:           string(req.Doc),
		IsPublic:      req.IsPublic,
		SchemaVersion: req.SchemaVersion,
		Revision:      req.Revision,
	})
	if err != nil {
		respondRichDocErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toDocumentResponse(doc))
}

// Delete は文書を論理削除する（所有者のみ）。
//
//	@Summary      リッチ 文書 の 削除
//	@Description  所有者 の 文書 を 論理 削除 する。 他人 の 文書 は 消せ ない (404)。
//	@Tags         documents
//	@Param        id   path  string  true  "文書 ID (UUID)"
//	@Success      204  "削除 成功 (本文 なし)"
//	@Failure      400  {object}  errorResponse  "不正 な ID"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      404  {object}  errorResponse  "存在 し ない か 権限 無し"
//	@Router       /documents/{id} [delete]
//	@Security     CookieAuth
func (h *DocumentHandler) Delete(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	id, ok := normalizeDocumentID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_id"})
		return
	}
	if err := h.del.Execute(c.Request.Context(), id, uid); err != nil {
		respondRichDocErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
