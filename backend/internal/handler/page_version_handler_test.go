package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ページ本文の版 API（FRESTYLE-433 段 3）の handler テスト。
//
// 判定の軸は kbEndpoints と同じ domain.Capability（view/edit の 2 値）だが、単体取得・復元は
// {seq} という kbEndpoints に無いプレースホルダを要る（comment_handler_test.go が {thread} の
// ために表を分けているのと同じ理由）。

const (
	kbVersionsPath       = "/api/v2/kb/workspaces/{slug}/pages/{page}/versions"
	kbVersionGetPath     = "/api/v2/kb/workspaces/{slug}/pages/{page}/versions/{seq}"
	kbVersionRestorePath = "/api/v2/kb/workspaces/{slug}/pages/{page}/versions/{seq}/restore"
)

// kbVersionEndpoint は 1 エンドポイントと、それが要求するケイパビリティ。
type kbVersionEndpoint struct {
	name       string
	method     string
	path       string // {slug} / {page} / {seq} プレースホルダ
	body       string
	capability domain.Capability
	okStatus   int
}

// kbVersionEndpoints は版 API の全エンドポイント。
// ルートを足したら kb_page_handler_test.go の登録漏れ検査（kbVersionEndpoints を
// covered map へ混ぜている箇所）にも足すこと。
var kbVersionEndpoints = []kbVersionEndpoint{
	{
		name: "版一覧", method: http.MethodGet, path: kbVersionsPath,
		capability: domain.CapabilityView, okStatus: http.StatusOK,
	},
	{
		name: "版取得", method: http.MethodGet, path: kbVersionGetPath,
		capability: domain.CapabilityView, okStatus: http.StatusOK,
	},
	{
		name: "版を残す", method: http.MethodPost, path: kbVersionsPath, body: `{}`,
		capability: domain.CapabilityEdit, okStatus: http.StatusCreated,
	},
	{
		name: "復元", method: http.MethodPost, path: kbVersionRestorePath,
		capability: domain.CapabilityEdit, okStatus: http.StatusOK,
	},
}

// request はプレースホルダを埋めて叩く。{seq} を含む経路は、叩く前に f.versions へ直接
// 版を 1 本作って埋める（usecase 経由の作成に依存すると、作成そのものを検証する経路と
// 循環してしまうため fake を直接使う — comment_handler_test.go の {thread} と同じ理由）。
func (e kbVersionEndpoint) request(f kbFixture, t *testing.T, slug, pageID string) *httptest.ResponseRecorder {
	t.Helper()
	seq := "1" // {seq} を使わない経路では無視される
	if strings.Contains(e.path, "{seq}") {
		_, v, err := f.versions.CreateVersionIfDue(context.Background(), kbWorkspaceID, pageID, kbValidDoc, kbUserID, nil, true)
		require.NoError(t, err)
		require.NotNil(t, v)
		seq = strconv.FormatInt(v.Seq, 10)
	}
	path := strings.NewReplacer("{slug}", slug, "{page}", pageID, "{seq}", seq).Replace(e.path)
	return f.do(t, e.method, path, e.body)
}

func Test_バージョンAPI_編集できるユーザーは全経路を通れる(t *testing.T) {
	for _, e := range kbVersionEndpoints {
		t.Run(e.name, func(t *testing.T) {
			f := newKbFixture(kbCanEdit, kbUserID)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			assert.Equal(t, e.okStatus, w.Code, "body=%s", w.Body.String())
		})
	}
}

func Test_バージョンAPI_閲覧だけのユーザーは書き込み経路で403(t *testing.T) {
	for _, e := range kbVersionEndpoints {
		t.Run(e.name, func(t *testing.T) {
			f := newKbFixture(kbCanView, kbUserID)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			if e.capability == domain.CapabilityView {
				assert.Equal(t, e.okStatus, w.Code, "閲覧経路は通る")
				return
			}
			assert.Equal(t, http.StatusForbidden, w.Code, "body=%s", w.Body.String())
			assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
		})
	}
}

func Test_バージョンAPI_メンバーでない相手は全経路で404(t *testing.T) {
	for _, e := range kbVersionEndpoints {
		t.Run(e.name, func(t *testing.T) {
			// kbOutsiderUserID はどのワークスペースにも所属しない（kb_permission_handler_test.go 参照）。
			f := newKbFixture(kbCanEdit, kbOutsiderUserID)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
		})
	}
}

// Test_バージョンAPI_無関係ページの版は覗けない は、別ページに作った版の seq を、
// このページの URL に流用しても 404 になることを確認する（page_id で隔離されていることの
// handler 越しの確認）。
func Test_バージョンAPI_無関係ページの版は覗けない(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	// kbDestPageID に版を 1 本作る。
	_, v, err := f.versions.CreateVersionIfDue(context.Background(), kbWorkspaceID, kbDestPageID, kbValidDoc, kbUserID, nil, true)
	require.NoError(t, err)
	require.NotNil(t, v)
	seq := strconv.FormatInt(v.Seq, 10)

	// 同じ seq を kbChildPageID の URL で覗く。
	path := strings.NewReplacer("{slug}", kbWorkspaceSlug, "{page}", kbChildPageID, "{seq}", seq).Replace(kbVersionGetPath)
	w := f.do(t, http.MethodGet, path, "")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
}

// Test_バージョンAPI_不正なseqは404 は、数値でない seq・実在しない seq のどちらも
// domain.ErrPageVersionNotFound と同じ 404 に落ちることを確認する。
func Test_バージョンAPI_不正なseqは404(t *testing.T) {
	t.Run("数値でない", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		path := strings.NewReplacer("{slug}", kbWorkspaceSlug, "{page}", kbChildPageID, "{seq}", "not-a-number").
			Replace(kbVersionGetPath)
		w := f.do(t, http.MethodGet, path, "")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
	})

	t.Run("存在しない", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		path := strings.NewReplacer("{slug}", kbWorkspaceSlug, "{page}", kbChildPageID, "{seq}", "999").
			Replace(kbVersionGetPath)
		w := f.do(t, http.MethodGet, path, "")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
	})
}

// Test_バージョンAPI_不正なnoteは400 は、501 バイトの note（domain.ValidateVersionNote の
// 上限超え）を 400 invalid_version_note へマップすることを確認する。
func Test_バージョンAPI_不正なnoteは400(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	note := strings.Repeat("a", domain.PageVersionNoteMaxBytes+1)
	body := `{"note":"` + note + `"}`

	w := f.do(t, http.MethodPost, kbFill(kbVersionsPath, kbWorkspaceSlug, kbChildPageID), body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid_version_note"}`, w.Body.String())
}

// Test_バージョンAPI_版を残すのレスポンス形 はレスポンスの形（フィールド名）をフロントと
// 合わせるための実例固定。著者名は kbFakeUsers に登録が無いので空文字になる。
func Test_バージョンAPI_版を残すのレスポンス形(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	w := f.do(t, http.MethodPost, kbFill(kbVersionsPath, kbWorkspaceSlug, kbChildPageID), `{"note":"手動保存"}`)
	require.Equal(t, http.StatusCreated, w.Code, "body=%s", w.Body.String())
	t.Logf("CreatePageVersion response: %s", w.Body.String())

	assert.Contains(t, w.Body.String(), `"seq"`)
	assert.Contains(t, w.Body.String(), `"author"`)
	assert.Contains(t, w.Body.String(), `"note":"手動保存"`)
	assert.Contains(t, w.Body.String(), `"doc"`)
}

// Test_バージョンAPI_復元のレスポンス形 は PUT .../content と同じ形（doc / builtAt /
// lastEditedBy / lastEditedAt）で返ることの実例固定。
func Test_バージョンAPI_復元のレスポンス形(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	_, v, err := f.versions.CreateVersionIfDue(context.Background(), kbWorkspaceID, kbChildPageID, kbValidDoc, kbUserID, nil, true)
	require.NoError(t, err)
	seq := strconv.FormatInt(v.Seq, 10)

	path := strings.NewReplacer("{slug}", kbWorkspaceSlug, "{page}", kbChildPageID, "{seq}", seq).Replace(kbVersionRestorePath)
	w := f.do(t, http.MethodPost, path, "")
	require.Equal(t, http.StatusOK, w.Code, "body=%s", w.Body.String())
	t.Logf("RestorePageVersion response: %s", w.Body.String())

	assert.Contains(t, w.Body.String(), `"doc"`)
	assert.Contains(t, w.Body.String(), `"builtAt"`)
	assert.Contains(t, w.Body.String(), `"lastEditedBy"`)
	assert.Contains(t, w.Body.String(), `"lastEditedAt"`)
}
