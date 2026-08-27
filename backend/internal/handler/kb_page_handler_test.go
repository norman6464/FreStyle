package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	kbWorkspaceID        = "0198a000-0000-7000-8000-000000000001"
	kbWorkspaceSlug      = "acme"
	kbOtherWorkspaceID   = "0198a000-0000-7000-8000-0000000000f1"
	kbOtherWorkspaceSlug = "rival"
	kbSpaceID            = "0198a000-0000-7000-8000-000000000002"
	kbRootPageID         = "0198a000-0000-7000-8000-000000000003"
	kbChildPageID        = "0198a000-0000-7000-8000-000000000004"
	kbDestPageID         = "0198a000-0000-7000-8000-000000000005"
	kbUserID             = uint64(42)
)

const kbValidDoc = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"本文"}]}]}`

var (
	kbCanEdit = domain.PagePermission{CanView: true, CanEdit: true}
	kbCanView = domain.PagePermission{CanView: true}
	kbNoPerm  = domain.PagePermission{}
)

// kbAuditRecorder は監査ログの記録先（本番の RecordAuditEventUseCase の代わり）。
//
// 権限操作 API に監査 middleware が本当に掛かっているかは、ルートを実際に叩いて
// 記録が 1 件増えることでしか確かめられない（掛け忘れは配線の穴で、handler 単体を
// 見ても分からない）。
type kbAuditRecorder struct {
	entries []middleware.AuditEntry
}

// handler は記録先をこのレコーダにした監査 middleware を返す。
func (r *kbAuditRecorder) handler() gin.HandlerFunc {
	return middleware.AuditLog(func(_ context.Context, e middleware.AuditEntry) {
		r.entries = append(r.entries, e)
	})
}

// kbFixture は fake repository と、本番と同じ wiring で組んだルータの組。
type kbFixture struct {
	pages       *kbFakePages
	perms       *kbFakePerms
	provisioner *kbFakeProvisioner
	router      *gin.Engine
	audit       *kbAuditRecorder
}

// newKbFixture はワークスペース 2 つ・スペース 1 つ・ページ 3 つ（root / child / dest）の
// 下ごしらえをして、registerKnowledgeBaseRoutesWith で本番と同じルートを張る。
// uid が 0 なら current user を注入せず未認証を再現する。
func newKbFixture(fallback domain.PagePermission, uid uint64) kbFixture {
	gin.SetMode(gin.TestMode)
	pages := newKbFakePages()
	pages.addWorkspace(kbWorkspaceID, kbWorkspaceSlug)
	pages.addWorkspace(kbOtherWorkspaceID, kbOtherWorkspaceSlug)
	pages.addSpace(kbWorkspaceID, kbSpaceID)
	pages.addPage(domain.Page{
		ID: kbRootPageID, WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID,
		Position: "a0", Title: "root", CreatedByUserID: kbUserID,
	})
	rootID := kbRootPageID
	pages.addPage(domain.Page{
		ID: kbChildPageID, WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID, ParentID: &rootID,
		Position: "a1", Title: "child", CreatedByUserID: kbUserID,
	})
	pages.addPage(domain.Page{
		ID: kbDestPageID, WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID,
		Position: "a2", Title: "dest", CreatedByUserID: kbUserID,
	})

	perms := newKbFakePerms(pages, fallback)
	// 所属しているのは本命ワークスペースだけ（もう片方は「他社のテナント」）。
	perms.addMember(kbWorkspaceID, kbUserID)

	r := gin.New()
	g := r.Group("/api/v2")
	if uid != 0 {
		g.Use(func(c *gin.Context) {
			c.Set(middleware.ContextKeyCurrentUserID, uid)
			c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: uid})
			c.Next()
		})
	}
	provisioner := newKbFakeProvisioner(pages, perms)
	audit := &kbAuditRecorder{}
	registerKnowledgeBaseRoutesWith(g, pages, perms, provisioner, audit.handler())
	// 認証不要のルート（共有リンクの検証）は current user を注入しない group に張る。
	// 本番の NewRouter と同じく認証 middleware の外側なので、ここでも外側に置かないと
	// 「未認証でも通ること」を検証できない。
	registerKnowledgeBasePublicRoutesWith(r.Group("/api/v2"), pages, perms)
	return kbFixture{pages: pages, perms: perms, provisioner: provisioner, router: r, audit: audit}
}

func (f kbFixture) do(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

// kbEndpoint は 1 エンドポイントと、それが要求するケイパビリティ。
// {slug} はワークスペースの slug、{page} は「権限の判定対象になるページ」に置換する
// （作成だけは判定対象が親ページなので、body 側に {page} が入る）。
type kbEndpoint struct {
	name       string
	method     string
	path       string
	body       string
	capability domain.Capability
	okStatus   int
}

// kbEndpoints は認可を配線すべき全エンドポイント。
// ルートを足したらここにも足す（この表が「配線漏れが無いこと」の担保になる）。
var kbEndpoints = []kbEndpoint{
	{
		name: "ページ作成", method: http.MethodPost,
		path:       "/api/v2/kb/workspaces/{slug}/spaces/" + kbSpaceID + "/pages",
		body:       `{"parentId":"{page}","title":"新しいページ"}`,
		capability: domain.CapabilityEdit, okStatus: http.StatusCreated,
	},
	{
		name: "ページ取得", method: http.MethodGet,
		path:       "/api/v2/kb/workspaces/{slug}/pages/{page}",
		capability: domain.CapabilityView, okStatus: http.StatusOK,
	},
	{
		name: "ページ改名", method: http.MethodPatch,
		path:       "/api/v2/kb/workspaces/{slug}/pages/{page}",
		body:       `{"title":"改訂"}`,
		capability: domain.CapabilityEdit, okStatus: http.StatusOK,
	},
	{
		name: "ページ移動", method: http.MethodPost,
		path:       "/api/v2/kb/workspaces/{slug}/pages/{page}/move",
		body:       `{"parentId":"` + kbDestPageID + `"}`,
		capability: domain.CapabilityEdit, okStatus: http.StatusOK,
	},
	{
		name: "ページアーカイブ", method: http.MethodPost,
		path:       "/api/v2/kb/workspaces/{slug}/pages/{page}/archive",
		capability: domain.CapabilityEdit, okStatus: http.StatusNoContent,
	},
	{
		name: "ページ復帰", method: http.MethodPost,
		path:       "/api/v2/kb/workspaces/{slug}/pages/{page}/unarchive",
		capability: domain.CapabilityEdit, okStatus: http.StatusOK,
	},
	{
		name: "本文置き換え", method: http.MethodPut,
		path:       "/api/v2/kb/workspaces/{slug}/pages/{page}/content",
		body:       `{"doc":` + kbValidDoc + `}`,
		capability: domain.CapabilityEdit, okStatus: http.StatusOK,
	},
}

// kbTreePath はツリー取得のパス（単一ページを名指ししないので kbEndpoints とは別扱い）。
const kbTreePath = "/api/v2/kb/workspaces/{slug}/spaces/" + kbSpaceID + "/pages"

// ページを名指ししない（＝ 判定対象がページではない）エンドポイント。
// kbEndpoints の表はページ 1 枚の権限を軸に回すので、こちらは別に持って個別に検証する。
const (
	// kbWorkspacesPath は所属ワークスペースの一覧と作成。テナントを URL に持たない。
	kbWorkspacesPath = "/api/v2/kb/workspaces"
	// kbSpacesPath はスペース作成。判定はワークスペース単位。
	kbSpacesPath = "/api/v2/kb/workspaces/{slug}/spaces"
)

func kbFill(s, slug, pageID string) string {
	return strings.NewReplacer("{slug}", slug, "{page}", pageID).Replace(s)
}

func (e kbEndpoint) request(f kbFixture, t *testing.T, slug, pageID string) *httptest.ResponseRecorder {
	t.Helper()
	return f.do(t, e.method, kbFill(e.path, slug, pageID), kbFill(e.body, slug, pageID))
}

// kbRoutePattern は表のパスを gin に登録されるパターンへ戻す（照合用）。
func kbRoutePattern(p string) string {
	return strings.NewReplacer(
		"{slug}", ":workspaceSlug",
		"{page}", ":pageId",
		kbSpaceID, ":spaceId",
	).Replace(p)
}

// 認可テストの表に載っていないルートが増えていないかを見る。
// 表に足し忘れたエンドポイントは認可の検証をすり抜けてしまうので、ここで機械的に塞ぐ。
// 登録されているナレッジ基盤のルートが、1 本残らず認可テストの表に載っていることを見る。
//
// # なぜ結合テストではなく単体テストに置いているのか
//
// 探しているのは「認可を通さないルートが増えたこと」で、それは**ルートの一覧**を見れば
// 分かる（実際に叩いて確かめる必要が無い）。結合テストに置くと、DB が要るぶん
// 開発者の手元では skip され得るし、CI でも専用ジョブでしか走らない。
// 配線の穴は書いた直後に落ちてほしいので、`go test ./...` で必ず走る側に置く。
//
// # この検査が守っている連鎖
//
// gin に登録されたルート ⊆ 表（kbEndpoints / kbPermissionEndpoints）で、その表は
// そのまま「admin 以外は通らない」「拒否の応答は対象の実在で変わらない」「変更操作は
// 監査ログに残る」の各テストが総当たりする入力になっている。つまり **認可も監査も
// 掛けずにルートを 1 本生やすと、まずここで落ちる**（表に足せば、今度は他のテストが
// その 1 本を実際に叩いて落とす）。
//
// # 限界（これで拾えないもの）
//
// gin のルート表からは middleware が見えないので、「このルートに audit を挟んだか」は
// ここでは分からない。それを見ているのは表を総当たりする側のテスト
// （Test_ナレッジ基盤権限API_権限を変える経路は全て監査ログに残る）で、この検査は
// 「新しいルートを必ずその表へ載せさせる」ことでそちらへ橋渡ししている。
func Test_ナレッジ基盤API_登録済みルートは全て認可テストの対象になっている(t *testing.T) {
	covered := map[string]bool{
		http.MethodGet + " " + kbRoutePattern(kbTreePath):    true,
		http.MethodGet + " " + kbWorkspacesPath:              true,
		http.MethodPost + " " + kbWorkspacesPath:             true,
		http.MethodGet + " " + kbRoutePattern(kbSpacesPath):  true,
		http.MethodPost + " " + kbRoutePattern(kbSpacesPath): true,
	}
	for _, e := range kbEndpoints {
		covered[e.method+" "+kbRoutePattern(e.path)] = true
	}
	// 権限操作 API は判定の軸が違う（ページ 1 枚のケイパビリティではなく admin か）ので
	// 表を分けてある。足したら kbPermissionEndpoints 側に足す。
	for _, e := range kbPermissionEndpoints {
		covered[e.method+" "+e.pattern] = true
	}
	covered[http.MethodPost+" "+kbShareLinkVerifyPath] = true

	f := newKbFixture(kbCanEdit, kbUserID)
	registered := map[string]bool{}
	for _, r := range f.router.Routes() {
		if !strings.HasPrefix(r.Path, "/api/v2/kb/") {
			continue
		}
		key := r.Method + " " + r.Path
		registered[key] = true
		assert.True(t, covered[key], "認可テストの表に無いルート: %s", key)
	}
	for key := range covered {
		assert.True(t, registered[key], "表にあるのに登録されていないルート: %s", key)
	}
}

func Test_ナレッジ基盤API_編集できるユーザーは全経路を通れる(t *testing.T) {
	for _, e := range kbEndpoints {
		t.Run(e.name, func(t *testing.T) {
			f := newKbFixture(kbCanEdit, kbUserID)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			assert.Equal(t, e.okStatus, w.Code, "body=%s", w.Body.String())
		})
	}
}

func Test_ナレッジ基盤API_権限が無いユーザーは全経路で404(t *testing.T) {
	for _, e := range kbEndpoints {
		t.Run(e.name, func(t *testing.T) {
			f := newKbFixture(kbNoPerm, kbUserID)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String(),
				"閲覧できないページは存在しないページと同じ応答にする")
		})
	}
	t.Run("ツリー取得", func(t *testing.T) {
		f := newKbFixture(kbNoPerm, kbUserID)
		w := f.do(t, http.MethodGet, kbFill(kbTreePath, kbWorkspaceSlug, ""), "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `[]`, w.Body.String(), "1 件も見えないスペースは空のツリー")
	})
}

func Test_ナレッジ基盤API_閲覧だけのユーザーは書き込み経路で403(t *testing.T) {
	for _, e := range kbEndpoints {
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

func Test_ナレッジ基盤API_別ワークスペースのslugは全経路で404(t *testing.T) {
	for _, e := range kbEndpoints {
		t.Run(e.name, func(t *testing.T) {
			// 権限は最強にしておく。それでも所属していないテナントには触れないことを見る。
			f := newKbFixture(kbCanEdit, kbUserID)
			w := e.request(f, t, kbOtherWorkspaceSlug, kbChildPageID)
			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
		})
	}
	t.Run("ツリー取得", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		w := f.do(t, http.MethodGet, kbFill(kbTreePath, kbOtherWorkspaceSlug, ""), "")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func Test_ナレッジ基盤API_存在しないslugは所属していないslugと区別できない(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	unknown := f.do(t, http.MethodGet,
		kbFill("/api/v2/kb/workspaces/{slug}/pages/{page}", "no-such-workspace", kbChildPageID), "")
	foreign := f.do(t, http.MethodGet,
		kbFill("/api/v2/kb/workspaces/{slug}/pages/{page}", kbOtherWorkspaceSlug, kbChildPageID), "")

	assert.Equal(t, http.StatusNotFound, unknown.Code)
	assert.Equal(t, foreign.Code, unknown.Code)
	assert.Equal(t, foreign.Body.String(), unknown.Body.String(),
		"slug の総当たりでテナントの実在が分からないこと")
}

func Test_ナレッジ基盤API_未認証は全経路で401(t *testing.T) {
	for _, e := range kbEndpoints {
		t.Run(e.name, func(t *testing.T) {
			f := newKbFixture(kbCanEdit, 0)
			w := e.request(f, t, kbWorkspaceSlug, kbChildPageID)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func Test_ナレッジ基盤API_存在しないページと権限の無いページは区別できない(t *testing.T) {
	const missingPageID = "0198a000-0000-7000-8000-00000000dead"
	for _, e := range kbEndpoints {
		t.Run(e.name, func(t *testing.T) {
			// 権限は最強のまま、対象ページだけを存在しない ID にする。
			existing := newKbFixture(kbCanEdit, kbUserID)
			missing := e.request(existing, t, kbWorkspaceSlug, missingPageID)

			// ページは実在するが、そのページだけ閲覧できない。
			hidden := newKbFixture(kbCanEdit, kbUserID)
			hidden.perms.setPagePermission(kbChildPageID, kbUserID, kbNoPerm)
			denied := e.request(hidden, t, kbWorkspaceSlug, kbChildPageID)

			assert.Equal(t, http.StatusNotFound, missing.Code)
			assert.Equal(t, missing.Code, denied.Code)
			assert.Equal(t, missing.Body.String(), denied.Body.String(),
				"ID の総当たりで隠したページの実在が分からないこと")
		})
	}
}

func Test_ナレッジ基盤ツリー_見えない親の子は根に浮かない(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	// root(見えない) → child(見える) → grandchild(見える)、dest(見える) の形にする。
	childID := kbChildPageID
	f.pages.addPage(domain.Page{
		ID: "0198a000-0000-7000-8000-000000000006", WorkspaceID: kbWorkspaceID, SpaceID: kbSpaceID,
		ParentID: &childID, Position: "a15", Title: "grandchild", CreatedByUserID: kbUserID,
	})
	f.perms.setPagePermission(kbRootPageID, kbUserID, kbNoPerm)

	w := f.do(t, http.MethodGet, kbFill(kbTreePath, kbWorkspaceSlug, ""), "")
	require.Equal(t, http.StatusOK, w.Code)

	var tree []kbPageTreeResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tree))
	require.Len(t, tree, 1, "見えない root の配下は子孫ごとツリーに現れない")
	assert.Equal(t, kbDestPageID, tree[0].Page.ID)
	assert.Empty(t, tree[0].Children)
}

func Test_ナレッジ基盤ツリー_見える親の下に子がぶら下がる(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	w := f.do(t, http.MethodGet, kbFill(kbTreePath, kbWorkspaceSlug, ""), "")
	require.Equal(t, http.StatusOK, w.Code)

	var tree []kbPageTreeResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tree))
	require.Len(t, tree, 2, "root と dest が根")
	assert.Equal(t, kbRootPageID, tree[0].Page.ID)
	require.Len(t, tree[0].Children, 1)
	assert.Equal(t, kbChildPageID, tree[0].Children[0].Page.ID)
	assert.Equal(t, kbDestPageID, tree[1].Page.ID)
}

func Test_ナレッジ基盤ツリー_アーカイブ済みページは現れない(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	at := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	f.pages.pages[kbChildPageID].ArchivedAt = &at

	w := f.do(t, http.MethodGet, kbFill(kbTreePath, kbWorkspaceSlug, ""), "")
	require.Equal(t, http.StatusOK, w.Code)

	var tree []kbPageTreeResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tree))
	require.Len(t, tree, 2)
	assert.Empty(t, tree[0].Children)
}

func Test_ナレッジ基盤ツリー_存在しないスペースは空のツリー(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	w := f.do(t, http.MethodGet,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/spaces/0198a000-0000-7000-8000-00000000beef/pages", "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[]`, w.Body.String(),
		"存在しないスペースと中身が見えないスペースを撃ち分けない")
}

func Test_ナレッジ基盤移動_移動先の親の権限も見る(t *testing.T) {
	movePath := "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbChildPageID + "/move"
	body := `{"parentId":"` + kbDestPageID + `"}`

	t.Run("移動先が閲覧だけなら403", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setPagePermission(kbDestPageID, kbUserID, kbCanView)
		w := f.do(t, http.MethodPost, movePath, body)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("移動先が見えないなら404", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setPagePermission(kbDestPageID, kbUserID, kbNoPerm)
		w := f.do(t, http.MethodPost, movePath, body)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
	})

	t.Run("自分の子孫の下へは移せない", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		w := f.do(t, http.MethodPost,
			"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbRootPageID+"/move",
			`{"parentId":"`+kbChildPageID+`"}`)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.JSONEq(t, `{"error":"page_cycle"}`, w.Body.String())
	})
}

func Test_ナレッジ基盤API_アーカイブ済みページの変更は409(t *testing.T) {
	at := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"改名", http.MethodPatch, "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbChildPageID, `{"title":"改訂"}`},
		{"本文置き換え", http.MethodPut, "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbChildPageID + "/content", `{"doc":` + kbValidDoc + `}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newKbFixture(kbCanEdit, kbUserID)
			f.pages.pages[kbChildPageID].ArchivedAt = &at
			w := f.do(t, tc.method, tc.path, tc.body)
			assert.Equal(t, http.StatusConflict, w.Code)
			assert.JSONEq(t, `{"error":"page_archived"}`, w.Body.String())
		})
	}
}

func Test_ナレッジ基盤API_アーカイブは冪等で復帰すると現役に戻る(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	base := "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbRootPageID

	require.Equal(t, http.StatusNoContent, f.do(t, http.MethodPost, base+"/archive", "").Code)
	require.Equal(t, http.StatusNoContent, f.do(t, http.MethodPost, base+"/archive", "").Code)
	assert.NotNil(t, f.pages.pages[kbChildPageID].ArchivedAt, "子孫もまとめてアーカイブされる")

	w := f.do(t, http.MethodPost, base+"/unarchive", "")
	require.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, f.pages.pages[kbRootPageID].ArchivedAt)
	assert.Nil(t, f.pages.pages[kbChildPageID].ArchivedAt, "一緒にアーカイブした子孫も戻る")
}

func Test_ナレッジ基盤API_入力の検証(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "作成にtitleが無ければ400", method: http.MethodPost,
			path:   "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/spaces/" + kbSpaceID + "/pages",
			body:   `{"parentId":"` + kbRootPageID + `"}`,
			status: http.StatusBadRequest,
		},
		{
			name: "ワークスペース作成のslugが不正なら400", method: http.MethodPost,
			path:   kbWorkspacesPath,
			body:   `{"slug":"Acme Inc","name":"Acme"}`,
			status: http.StatusBadRequest,
		},
		{
			name: "タイトルが201文字なら400", method: http.MethodPatch,
			path:   "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbChildPageID,
			body:   `{"title":"` + strings.Repeat("あ", 201) + `"}`,
			status: http.StatusBadRequest,
		},
		{
			name: "本文がdocでなければ400", method: http.MethodPut,
			path:   "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbChildPageID + "/content",
			body:   `{"doc":{"type":"paragraph"}}`,
			status: http.StatusBadRequest,
		},
		{
			name: "本文に未知のノードがあれば400", method: http.MethodPut,
			path:   "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbChildPageID + "/content",
			body:   `{"doc":{"type":"doc","content":[{"type":"未知のノード"}]}}`,
			status: http.StatusBadRequest,
		},
		{
			name: "移動にparentIdが無ければ400", method: http.MethodPost,
			path:   "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbChildPageID + "/move",
			body:   `{}`,
			status: http.StatusBadRequest,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newKbFixture(kbCanEdit, kbUserID)
			w := f.do(t, tc.method, tc.path, tc.body)
			assert.Equal(t, tc.status, w.Code, "body=%s", w.Body.String())
		})
	}
}

func Test_ナレッジ基盤API_取得は本文と作成したページを返す(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)

	created := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/spaces/"+kbSpaceID+"/pages",
		`{"parentId":"`+kbRootPageID+`","title":"新しいページ"}`)
	require.Equal(t, http.StatusCreated, created.Code)

	var page kbPageResponse
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &page))
	assert.Equal(t, "新しいページ", page.Title)
	require.NotNil(t, page.ParentID)
	assert.Equal(t, kbRootPageID, *page.ParentID)
	assert.NotContains(t, created.Body.String(), kbWorkspaceID, "workspaceId は返さない")

	saved := f.do(t, http.MethodPut,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+page.ID+"/content",
		`{"doc":`+kbValidDoc+`}`)
	require.Equal(t, http.StatusOK, saved.Code)

	got := f.do(t, http.MethodGet, "/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+page.ID, "")
	require.Equal(t, http.StatusOK, got.Code)

	var doc kbPageDocResponse
	require.NoError(t, json.Unmarshal(got.Body.Bytes(), &doc))
	assert.Equal(t, page.ID, doc.Page.ID)
	assert.Contains(t, string(doc.Doc), "本文")
}

func Test_ナレッジ基盤API_所属判定が失敗したら500(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.membersErr = errors.New("db down")

	w := f.do(t, http.MethodGet, "/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbChildPageID, "")

	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"DB 障害を 404 に潰すと、落ちていることに気づけない")
}

func Test_ナレッジ基盤API_repositoryの失敗は500(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.pages.failWith = errors.New("db down")

	w := f.do(t, http.MethodPatch,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbChildPageID, `{"title":"改訂"}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"internal_error"}`, w.Body.String())
}

// ワークスペース解決 middleware を通さずにルートを生やす配線ミスを想定した安全網。
// テナント未確定のまま handler が動くと全テナントに触れてしまうので、必ず落ちること。
// kbScope のワークスペース未確定ガードそのものを見る。
//
// handler 越しに「middleware を通らないルート」を叩く形だと、ガードを消しても
// usecase 側の必須チェック（workspaceID is required）で同じ 500 になり、
// ガードが在るか無いかを区別できない。ここでは kbScope を直接呼び、
// ガードを外したときに後続へ進んでしまうことまで見る。
func Test_ナレッジ基盤API_ワークスペース未確定ならkbScopeが止める(t *testing.T) {
	gin.SetMode(gin.TestMode)

	run := func(t *testing.T, withWorkspace bool) (*httptest.ResponseRecorder, bool) {
		t.Helper()
		reached := false
		r := gin.New()
		r.GET("/scope", func(c *gin.Context) {
			c.Set(middleware.ContextKeyCurrentUserID, kbUserID)
			if withWorkspace {
				c.Set(middleware.ContextKeyKnowledgeBaseWorkspace, &domain.Workspace{ID: kbWorkspaceID})
			}
			if _, ok := kbScope(c); !ok {
				return
			}
			reached = true
			c.Status(http.StatusOK)
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/scope", nil))
		return w, reached
	}

	t.Run("未確定なら止まる", func(t *testing.T) {
		w, reached := run(t, false)
		assert.False(t, reached, "テナント未確定のまま後続へ進んではいけない")
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error":"internal_error"}`, w.Body.String())
	})

	t.Run("確定していれば通る", func(t *testing.T) {
		w, reached := run(t, true)
		assert.True(t, reached, "middleware を通っていれば素通しする（常に止めるガードではない）")
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// 配線のミス（middleware を通さない group への登録）が「成功する経路」にならないことを
// 端から端まで見る。どの層で止まるかまでは固定しない（kbScope のガード自体は上のテスト）。
func Test_ナレッジ基盤API_middlewareを通らないルートは成功しない(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pages := newKbFakePages()
	perms := newKbFakePerms(pages, kbCanEdit)
	h := NewKnowledgeBasePageHandler(
		usecase.NewCheckPagePermissionUseCase(perms),
		usecase.NewCheckSpacePermissionUseCase(perms),
		usecase.NewCanEditPageSubtreeUseCase(perms),
		usecase.NewListViewablePagesUseCase(perms),
		usecase.NewGetPageUseCase(pages),
		usecase.NewCreatePageUseCase(pages),
		usecase.NewRenamePageUseCase(pages),
		usecase.NewMovePageUseCase(pages),
		usecase.NewArchivePageUseCase(pages),
		usecase.NewUnarchivePageUseCase(pages),
		usecase.NewReplacePageBlocksUseCase(pages),
	)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUserID, kbUserID)
		c.Next()
	})
	r.GET("/unwired/:pageId", h.Get)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/unwired/"+kbChildPageID, nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_ナレッジ基盤ツリー_事実の収集が失敗したら500(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.listFactsErr = errors.New("db down")

	w := f.do(t, http.MethodGet, kbFill(kbTreePath, kbWorkspaceSlug, ""), "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_ナレッジ基盤アーカイブ_repositoryの失敗は500(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.pages.failWith = errors.New("db down")

	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbChildPageID+"/archive", "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_ナレッジ基盤移動_スペース全員宛ての例外が失効する移動は409(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	// 移動先スペース以外の「全員」宛て例外がサブツリーに残っている状態を repository が
	// 同一トランザクションで検出して中止する経路。move handler は NewSpaceID を渡さないので、
	// 別スペースのページを親に指定するだけでここへ来る。
	f.pages.moveErr = repository.ErrPageMoveVoidsSpaceRestriction

	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbChildPageID+"/move",
		`{"parentId":"`+kbDestPageID+`"}`)

	assert.Equal(t, http.StatusConflict, w.Code,
		"権限設定と両立しないという業務上の衝突であって、DB 障害ではない")
	assert.JSONEq(t, `{"error":"space_restriction_voided"}`, w.Body.String())
}

func Test_ナレッジ基盤アーカイブ_配下に編集できないページがあれば何もせず403(t *testing.T) {
	// 子を直接 rename すれば 403 になる相手が、親のアーカイブ経由なら書き換えられる
	// （見えない子まで巻き込む）状態を塞ぐ。edit を外した場合と view ごと外した場合の両方。
	cases := map[string]domain.Capability{
		"編集だけ外した子": domain.CapabilityEdit,
		"閲覧ごと外した子": domain.CapabilityView,
	}
	for name, capability := range cases {
		t.Run(name, func(t *testing.T) {
			f := newKbFixture(kbCanEdit, kbUserID)
			kbRestrict(t, f, kbChildPageID, kbPrincipalOf(t, f, kbUserID), capability, domain.RestrictionModeDeny)

			w := f.do(t, http.MethodPost,
				"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbRootPageID+"/archive", "")

			assert.Equal(t, http.StatusForbidden, w.Code)
			assert.JSONEq(t, `{"error":"subtree_forbidden"}`, w.Body.String())
			assert.Nil(t, f.pages.pages[kbRootPageID].ArchivedAt, "根は書き換わらない")
			assert.Nil(t, f.pages.pages[kbChildPageID].ArchivedAt, "触れない子も書き換わらない")
		})
	}
}

func Test_ナレッジ基盤アーカイブ_子孫まで編集できるなら通る(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)

	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbRootPageID+"/archive", "")

	require.Equal(t, http.StatusNoContent, w.Code, "例外が無いのが普通なので、通常の運用は止めない")
	assert.NotNil(t, f.pages.pages[kbRootPageID].ArchivedAt)
	assert.NotNil(t, f.pages.pages[kbChildPageID].ArchivedAt)
}

func Test_ナレッジ基盤復帰_配下に編集できないページがあれば何もせず403(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	at := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	f.pages.pages[kbRootPageID].ArchivedAt = &at
	f.pages.pages[kbChildPageID].ArchivedAt = &at
	kbRestrict(t, f, kbChildPageID, kbPrincipalOf(t, f, kbUserID), domain.CapabilityEdit, domain.RestrictionModeDeny)

	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbRootPageID+"/unarchive", "")

	assert.Equal(t, http.StatusForbidden, w.Code, "戻す側も同じ判定にする（片側だけ緩いと結局動かせる）")
	assert.JSONEq(t, `{"error":"subtree_forbidden"}`, w.Body.String())
	assert.NotNil(t, f.pages.pages[kbRootPageID].ArchivedAt, "アーカイブ済みのまま")
}

func Test_ナレッジ基盤アーカイブ_サブツリーの権限確認が失敗したら500(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.subtreeFactsErr = errors.New("db down")

	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbRootPageID+"/archive", "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Nil(t, f.pages.pages[kbRootPageID].ArchivedAt, "確認できないなら書き換えない")
}

func Test_ナレッジ基盤作成_アーカイブ済みの親の下には作れない(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	at := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	f.pages.pages[kbRootPageID].ArchivedAt = &at

	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/spaces/"+kbSpaceID+"/pages",
		`{"parentId":"`+kbRootPageID+`","title":"迷子"}`)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.JSONEq(t, `{"error":"parent_archived"}`, w.Body.String())
}

func Test_ナレッジ基盤作成_親が別スペースなら400(t *testing.T) {
	const otherSpaceID = "0198a000-0000-7000-8000-0000000000c1"
	f := newKbFixture(kbCanEdit, kbUserID)
	f.pages.addSpace(kbWorkspaceID, otherSpaceID)

	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/spaces/"+otherSpaceID+"/pages",
		`{"parentId":"`+kbRootPageID+`","title":"別スペースの親"}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"parent_space_mismatch"}`, w.Body.String())
}

func Test_ナレッジ基盤復帰_親がアーカイブ中なら409(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	at := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	f.pages.pages[kbRootPageID].ArchivedAt = &at
	childArchivedAt := at.Add(time.Hour)
	f.pages.pages[kbChildPageID].ArchivedAt = &childArchivedAt

	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbChildPageID+"/unarchive", "")

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.JSONEq(t, `{"error":"parent_archived"}`, w.Body.String())
}

func Test_ナレッジ基盤取得_本文が未保存でも空のdocを返す(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)

	w := f.do(t, http.MethodGet, "/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbChildPageID, "")

	require.Equal(t, http.StatusOK, w.Code)
	var doc kbPageDocResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &doc))
	assert.JSONEq(t, `{"type":"doc","content":[]}`, string(doc.Doc))
}

// --- 権限の例外（page_restrictions）と許可リスト制の印（page_allow_lists）---
//
// ここから下は「既定の役割は editor のまま、例外だけで見え方が変わる」経路を API 越しに見る。
// fake が印を allow 行の有無で代用していると、主体を消したあとの 2 つのテストが緑にならない。

// kbAlice は許可リストに載せる（そして消す）相手。呼び出し元のユーザーとは別人。
const kbAlice = uint64(43)

func kbPrincipalOf(t *testing.T, f kbFixture, userID uint64) string {
	t.Helper()
	p, err := f.perms.EnsureUserPrincipal(context.Background(), kbWorkspaceID, userID)
	require.NoError(t, err)
	return p.ID
}

func kbRestrict(t *testing.T, f kbFixture, pageID, principalID string, c domain.Capability, m domain.RestrictionMode) {
	t.Helper()
	_, err := f.perms.UpsertPageRestriction(context.Background(), kbWorkspaceID, pageID, principalID, c, m)
	require.NoError(t, err)
}

func kbUnrestrict(t *testing.T, f kbFixture, pageID, principalID string, c domain.Capability) {
	t.Helper()
	require.NoError(t, f.perms.DeletePageRestriction(context.Background(), kbWorkspaceID, pageID, principalID, c))
}

// kbGetStatus はページ取得の HTTP ステータス（見えれば 200、見えなければ 404）。
func kbGetStatus(t *testing.T, f kbFixture, pageID string) int {
	t.Helper()
	return f.do(t, http.MethodGet, "/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+pageID, "").Code
}

// kbTreeIDs はツリー取得に現れるページ ID を親子まとめて返す。
func kbTreeIDs(t *testing.T, f kbFixture) []string {
	t.Helper()
	w := f.do(t, http.MethodGet, kbFill(kbTreePath, kbWorkspaceSlug, ""), "")
	require.Equal(t, http.StatusOK, w.Code)
	var tree []kbPageTreeResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tree))
	ids := make([]string, 0, 4)
	var walk func(nodes []kbPageTreeResponse)
	walk = func(nodes []kbPageTreeResponse) {
		for _, n := range nodes {
			ids = append(ids, n.Page.ID)
			walk(n.Children)
		}
	}
	walk(tree)
	return ids
}

func Test_ナレッジ基盤権限_祖先のdenyは子孫にも効く(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	kbRestrict(t, f, kbRootPageID, kbPrincipalOf(t, f, kbUserID), domain.CapabilityView, domain.RestrictionModeDeny)

	assert.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID))
	assert.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbChildPageID),
		"deny は経路全体で効くので、名指しされていない子も外れたままになる")
	assert.Equal(t, []string{kbDestPageID}, kbTreeIDs(t, f),
		"1 ページの解決と一覧で畳み方が食い違わない")
}

func Test_ナレッジ基盤権限_許可リストに載っていなければ既定がeditorでも見えない(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	kbRestrict(t, f, kbRootPageID, kbPrincipalOf(t, f, kbAlice), domain.CapabilityView, domain.RestrictionModeAllow)

	assert.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID))
	assert.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbChildPageID),
		"限定公開は子孫にも効く")
	assert.Equal(t, []string{kbDestPageID}, kbTreeIDs(t, f))
}

func Test_ナレッジ基盤権限_より近い許可リストが遠い許可リストを上書きする(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	kbRestrict(t, f, kbRootPageID, kbPrincipalOf(t, f, kbAlice), domain.CapabilityView, domain.RestrictionModeAllow)
	kbRestrict(t, f, kbChildPageID, kbPrincipalOf(t, f, kbUserID), domain.CapabilityView, domain.RestrictionModeAllow)

	assert.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID), "root の許可リストには載っていない")
	assert.Equal(t, http.StatusOK, kbGetStatus(t, f, kbChildPageID),
		"最も近い許可リスト制の段（child）に載っていれば見える")
	assert.Equal(t, []string{kbDestPageID}, kbTreeIDs(t, f),
		"見える child も、見えない root の配下なのでツリーには出ない")
}

func Test_ナレッジ基盤権限_許可リストの主体を消しても限定公開は解けない(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	alice := kbPrincipalOf(t, f, kbAlice)
	kbRestrict(t, f, kbRootPageID, alice, domain.CapabilityView, domain.RestrictionModeAllow)
	require.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID))

	// 退職者のオフボーディング。allow 行は主体と一緒に消えるが、印は主体を参照しないので残る。
	require.NoError(t, f.perms.DeletePrincipal(context.Background(), kbWorkspaceID, alice))

	assert.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID),
		"許可リストが空になった段は「誰も載っていない」＝ 閉じたまま")
	assert.Equal(t, []string{kbDestPageID}, kbTreeIDs(t, f))

	rows, err := f.perms.ListPageRestrictions(context.Background(), kbWorkspaceID, kbRootPageID)
	require.NoError(t, err)
	assert.Empty(t, rows, "例外の一覧からは消えている")
	caps, err := f.perms.ListPageAllowListCapabilities(context.Background(), kbWorkspaceID, kbRootPageID)
	require.NoError(t, err)
	assert.Equal(t, []domain.Capability{domain.CapabilityView}, caps,
		"限定公開であることは印にしか残らない（権限設定を見せるときは両方を読む）")
}

func Test_ナレッジ基盤権限_deny行を外しても限定公開は解けない(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	me := kbPrincipalOf(t, f, kbUserID)
	alice := kbPrincipalOf(t, f, kbAlice)
	kbRestrict(t, f, kbRootPageID, alice, domain.CapabilityView, domain.RestrictionModeAllow)
	kbRestrict(t, f, kbRootPageID, me, domain.CapabilityView, domain.RestrictionModeDeny)
	// alice が抜けて許可リストは空になり、印だけが残っている状態を作る。
	require.NoError(t, f.perms.DeletePrincipal(context.Background(), kbWorkspaceID, alice))
	require.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID))

	kbUnrestrict(t, f, kbRootPageID, me, domain.CapabilityView)

	assert.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID),
		"deny 行の解除では印を畳まない（自分宛ての deny を 1 行外すだけで限定公開が解けない）")
	caps, err := f.perms.ListPageAllowListCapabilities(context.Background(), kbWorkspaceID, kbRootPageID)
	require.NoError(t, err)
	assert.Equal(t, []domain.Capability{domain.CapabilityView}, caps)
}

func Test_ナレッジ基盤権限_最後のallowを外すと既定へ戻る(t *testing.T) {
	t.Run("解除", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		alice := kbPrincipalOf(t, f, kbAlice)
		kbRestrict(t, f, kbRootPageID, alice, domain.CapabilityView, domain.RestrictionModeAllow)
		require.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID))

		kbUnrestrict(t, f, kbRootPageID, alice, domain.CapabilityView)

		assert.Equal(t, http.StatusOK, kbGetStatus(t, f, kbRootPageID))
		assert.Equal(t, []string{kbRootPageID, kbChildPageID, kbDestPageID}, kbTreeIDs(t, f))
	})

	t.Run("denyへの書き換え", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		alice := kbPrincipalOf(t, f, kbAlice)
		kbRestrict(t, f, kbRootPageID, alice, domain.CapabilityView, domain.RestrictionModeAllow)
		require.Equal(t, http.StatusNotFound, kbGetStatus(t, f, kbRootPageID))

		kbRestrict(t, f, kbRootPageID, alice, domain.CapabilityView, domain.RestrictionModeDeny)

		assert.Equal(t, http.StatusOK, kbGetStatus(t, f, kbRootPageID),
			"その段に allow が 1 行も残らなければ印も畳まれる（畳むのは解除だけではない）")
		caps, err := f.perms.ListPageAllowListCapabilities(context.Background(), kbWorkspaceID, kbRootPageID)
		require.NoError(t, err)
		assert.Empty(t, caps)
	})
}

func Test_ナレッジ基盤権限_editのdenyは閲覧を残したまま書き込みだけ止める(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	kbRestrict(t, f, kbRootPageID, kbPrincipalOf(t, f, kbUserID), domain.CapabilityEdit, domain.RestrictionModeDeny)

	assert.Equal(t, http.StatusOK, kbGetStatus(t, f, kbChildPageID))
	w := f.do(t, http.MethodPatch,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbChildPageID, `{"title":"改訂"}`)
	assert.Equal(t, http.StatusForbidden, w.Code, "祖先で編集を外された子も書き込めない")
}
