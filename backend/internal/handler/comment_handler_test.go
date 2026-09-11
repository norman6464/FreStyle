package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// コメント API（FRESTYLE-432 段 2・ページ全体へのコメント）の handler テスト。
//
// 判定の軸が domain.Capability（view/edit の 2 値）ではなく
// domain.PagePermission.CanComment という別軸なので、kb_page_handler_test.go の
// kbEndpoints とは別の表（kbCommentEndpoints）を持つ
// （kb_permission_handler_test.go が admin 専用の判定用に kbPermissionEndpoints を
// 分けているのと同じ理由）。

const (
	kbCommentThreadsPath = "/api/v2/kb/workspaces/{slug}/pages/{page}/comment-threads"
	kbCommentAddPath     = "/api/v2/kb/workspaces/{slug}/pages/{page}/comment-threads/{thread}/comments"
	kbCommentResolvePath = "/api/v2/kb/workspaces/{slug}/pages/{page}/comment-threads/{thread}/resolve"
	kbCommentReopenPath  = "/api/v2/kb/workspaces/{slug}/pages/{page}/comment-threads/{thread}/reopen"
)

// kbValidCommentBody は本文検証（domain.ValidateCommentBody）を通る最小の入力。
const kbValidCommentBody = `{"body":[{"type":"text","text":"hello"}]}`

// kbCommentEndpoint は 1 エンドポイントと、それが要求する権限の軸。
type kbCommentEndpoint struct {
	name   string
	method string
	path   string // {slug} / {page} / {thread} プレースホルダ
	body   string
	// needsComment が true なら CanComment が要る（作成・返信・解決・再開）。
	// false なら CanView だけで良い（一覧 — viewer でも既存のコメントは読める）。
	needsComment bool
	okStatus     int
}

// kbCommentEndpoints はコメント API の全エンドポイント。
// ルートを足したら kb_page_handler_test.go の登録漏れ検査（kbCommentEndpoints を
// covered map へ混ぜている箇所）にも足すこと。
var kbCommentEndpoints = []kbCommentEndpoint{
	{
		name: "スレッド作成", method: http.MethodPost, path: kbCommentThreadsPath,
		body: kbValidCommentBody, needsComment: true, okStatus: http.StatusCreated,
	},
	{
		name: "スレッド一覧", method: http.MethodGet, path: kbCommentThreadsPath,
		needsComment: false, okStatus: http.StatusOK,
	},
	{
		name: "返信", method: http.MethodPost, path: kbCommentAddPath,
		body: kbValidCommentBody, needsComment: true, okStatus: http.StatusCreated,
	},
	{
		name: "解決", method: http.MethodPost, path: kbCommentResolvePath,
		needsComment: true, okStatus: http.StatusOK,
	},
	{
		name: "再開", method: http.MethodPost, path: kbCommentReopenPath,
		needsComment: true, okStatus: http.StatusOK,
	},
}

// request はプレースホルダを埋めて叩く。{thread} を含む経路は、叩く前に
// f.comments へ直接スレッドを 1 本作って埋める（usecase 経由のスレッド作成に
// 依存すると、作成そのものを検証する経路と循環してしまうため fake を直接使う）。
func (e kbCommentEndpoint) request(f kbFixture, t *testing.T, slug, pageID string) *httptest.ResponseRecorder {
	t.Helper()
	threadID := "0198a000-0000-7000-8000-0000000000ff" // {thread} を使わない経路では無視される
	if strings.Contains(e.path, "{thread}") {
		th, err := f.comments.CreateCommentThread(context.Background(), kbWorkspaceID, pageID, kbUserID, repository.CommentAnchor{})
		require.NoError(t, err)
		threadID = th.ID
	}
	path := strings.NewReplacer("{slug}", slug, "{page}", pageID, "{thread}", threadID).Replace(e.path)
	return f.do(t, e.method, path, e.body)
}

func Test_コメントAPI_コメントできるユーザーは全経路を通れる(t *testing.T) {
	for _, e := range kbCommentEndpoints {
		t.Run(e.name, func(t *testing.T) {
			f := newKbFixture(domain.PagePermission{CanView: true, CanComment: true}, kbUserID)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			assert.Equal(t, e.okStatus, w.Code, "body=%s", w.Body.String())
		})
	}
}

func Test_コメントAPI_閲覧だけのユーザーは書き込み経路で403(t *testing.T) {
	for _, e := range kbCommentEndpoints {
		t.Run(e.name, func(t *testing.T) {
			// kbCanView は CanView だけを立てた既定（CanComment はゼロ値で false）。
			f := newKbFixture(kbCanView, kbUserID)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			if !e.needsComment {
				assert.Equal(t, e.okStatus, w.Code, "一覧は CanView だけで通る")
				return
			}
			assert.Equal(t, http.StatusForbidden, w.Code, "body=%s", w.Body.String())
			assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
		})
	}
}

func Test_コメントAPI_メンバーでない相手は全経路で404(t *testing.T) {
	for _, e := range kbCommentEndpoints {
		t.Run(e.name, func(t *testing.T) {
			// kbOutsiderUserID はどのワークスペースにも所属しない（kb_permission_handler_test.go
			// で定義済み）。newKbFixture が所属させるのは常に kbUserID なので、
			// uid をこれに差し替えるだけで「無関係な人」を再現できる。
			f := newKbFixture(domain.PagePermission{CanView: true, CanComment: true}, kbOutsiderUserID)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
		})
	}
}

func Test_コメントAPI_本文が空配列なら400(t *testing.T) {
	const emptyBody = `{"body":[]}`

	t.Run("スレッド作成", func(t *testing.T) {
		f := newKbFixture(domain.PagePermission{CanView: true, CanComment: true}, kbUserID)
		w := f.do(t, http.MethodPost, kbFill(kbCommentThreadsPath, kbWorkspaceSlug, kbChildPageID), emptyBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"invalid_request"}`, w.Body.String())
	})

	t.Run("返信", func(t *testing.T) {
		f := newKbFixture(domain.PagePermission{CanView: true, CanComment: true}, kbUserID)
		th, err := f.comments.CreateCommentThread(context.Background(), kbWorkspaceID, kbChildPageID, kbUserID, repository.CommentAnchor{})
		require.NoError(t, err)
		path := strings.NewReplacer(
			"{slug}", kbWorkspaceSlug, "{page}", kbChildPageID, "{thread}", th.ID,
		).Replace(kbCommentAddPath)
		w := f.do(t, http.MethodPost, path, emptyBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"invalid_request"}`, w.Body.String())
	})
}

// レスポンスの形（フィールド名）をフロントと合わせるための実例固定。
// 著者名は kbFakeUsers に登録が無いので空文字になる（LookupUserNameUseCase の doc 参照）。
func Test_コメントAPI_スレッド作成のレスポンス形(t *testing.T) {
	f := newKbFixture(domain.PagePermission{CanView: true, CanComment: true}, kbUserID)
	w := f.do(t, http.MethodPost, kbFill(kbCommentThreadsPath, kbWorkspaceSlug, kbChildPageID), kbValidCommentBody)
	require.Equal(t, http.StatusCreated, w.Code, "body=%s", w.Body.String())
	t.Logf("CreateThread response: %s", w.Body.String())

	assert.Contains(t, w.Body.String(), `"id"`)
	assert.Contains(t, w.Body.String(), `"createdBy"`)
	assert.Contains(t, w.Body.String(), `"comments"`)
}

// Test_コメントAPI_錨付きスレッド作成のレスポンス形 は FRESTYLE-432 段 3 の実例固定。
// page-level（上のテスト）と違い blockId/anchorFrom/anchorTo/quote が応答に乗ることを確認する。
func Test_コメントAPI_錨付きスレッド作成のレスポンス形(t *testing.T) {
	f := newKbFixture(domain.PagePermission{CanView: true, CanComment: true}, kbUserID)
	f.comments.addBlock(kbChildPageID, "block-1")
	const body = `{"body":[{"type":"text","text":"hello"}],"blockId":"block-1","anchorFrom":3,"anchorTo":12,"quote":"錨付けされた引用文"}`

	w := f.do(t, http.MethodPost, kbFill(kbCommentThreadsPath, kbWorkspaceSlug, kbChildPageID), body)
	require.Equal(t, http.StatusCreated, w.Code, "body=%s", w.Body.String())
	t.Logf("CreateThread(anchored) response: %s", w.Body.String())

	// 部分文字列一致（Contains）だと "anchorFrom":3 は "anchorFrom":30 のような値でも
	// 素通りしてしまう。数値として正確に比較する — CodeRabbit 指摘。
	var got struct {
		BlockID    string `json:"blockId"`
		AnchorFrom int    `json:"anchorFrom"`
		AnchorTo   int    `json:"anchorTo"`
		Quote      string `json:"quote"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got), "body=%s", w.Body.String())
	assert.Equal(t, "block-1", got.BlockID)
	assert.Equal(t, 3, got.AnchorFrom)
	assert.Equal(t, 12, got.AnchorTo)
	assert.Equal(t, "錨付けされた引用文", got.Quote)
}

// Test_コメントAPI_錨が不正な組み合わせなら400 は、一部だけ非nilの中途半端な入力
// （ここでは blockId だけ）を domain.ErrInvalidCommentAnchor 経由で 400 invalid_comment_anchor に
// マップすることを確認する。
func Test_コメントAPI_錨が不正な組み合わせなら400(t *testing.T) {
	f := newKbFixture(domain.PagePermission{CanView: true, CanComment: true}, kbUserID)
	const body = `{"body":[{"type":"text","text":"hello"}],"blockId":"block-1"}`

	w := f.do(t, http.MethodPost, kbFill(kbCommentThreadsPath, kbWorkspaceSlug, kbChildPageID), body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid_comment_anchor"}`, w.Body.String())
}

// Test_コメントAPI_他ページの実在するblockIdは400 は、handler越しでも他ページのブロックへ
// 錨を張ろうとする書き込みが拒否されることを確認する（usecase/repositoryの検証がhandlerまで
// きちんと配線されていることの確認）。
func Test_コメントAPI_他ページの実在するblockIdは400(t *testing.T) {
	f := newKbFixture(domain.PagePermission{CanView: true, CanComment: true}, kbUserID)
	// kbRootPageID に実在することにしたブロックを、別ページ kbChildPageID への
	// スレッド作成で錨として指定する。
	f.comments.addBlock(kbRootPageID, "block-on-root")
	const body = `{"body":[{"type":"text","text":"hello"}],"blockId":"block-on-root","anchorFrom":0,"anchorTo":5,"quote":"乗っ取り"}`

	w := f.do(t, http.MethodPost, kbFill(kbCommentThreadsPath, kbWorkspaceSlug, kbChildPageID), body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid_comment_anchor"}`, w.Body.String())
}
