package handler

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 権限操作 API（grant / restriction / メンバー / グループ / 共有リンク）の handler テスト。
//
// 見るのは 2 つだけ:
//
//  1. admin 以外は 1 本も通らないこと（viewer / editor / commenter / 非メンバー /
//     別ワークスペースの admin の 5 通り）
//  2. 拒否の応答が、対象が実在するかどうかで変わらないこと（存在オラクルを作らない）
//
// 実効権限そのものの規則は domain と usecase のテストが持つ。ここで固定するのは
// 「その規則が HTTP の入口で必ず適用されること」。

const (
	// kbSecondUserID は権限を張られる側のユーザー。
	kbSecondUserID = uint64(43)
	// kbOutsiderUserID はどのワークスペースにも所属しないユーザー。
	kbOutsiderUserID = uint64(97)
	// kbRivalAdminUserID は別ワークスペース（rival）の admin。acme には所属しない。
	kbRivalAdminUserID = uint64(98)
	// kbMissingID は存在しない UUID。存在オラクルの検証に使う。
	kbMissingID = "0198a000-0000-7000-8000-0000000000ff"
	// kbMissingUserID は存在しないユーザー ID。
	kbMissingUserID = "987654"
	// kbShareLinkVerifyPath は認証不要の共有リンク検証。
	kbShareLinkVerifyPath = "/api/v2/kb/share-links/verify"
)

// kbDenied は権限操作 API の唯一の拒否応答（バイト列で固定する）。
const kbDenied = `{"error":"not_found"}`

// kbPermissionEndpoint は権限操作の 1 エンドポイント。
type kbPermissionEndpoint struct {
	name   string
	method string
	// pattern は gin に登録されるルートのパターン（登録漏れ検査の照合に使う）。
	pattern string
	// path は実リクエストのパス（プレースホルダ入り）。
	path string
	// missing は対象を存在しない ID に差し替えたパス。拒否応答が
	// 対象の実在で変わらないことを見るために使う。
	missing  []string
	body     string
	okStatus int
}

// kbPermissionEndpoints は「権限そのものを変える」全エンドポイント。
// ルートを足したらここにも足す（kb_page_handler_test.go の登録漏れ検査が照合する）。
var kbPermissionEndpoints = []kbPermissionEndpoint{
	{
		name: "ワークスペース権限付与", method: http.MethodPut,
		pattern: "/api/v2/kb/workspaces/:workspaceSlug/grants/:principalId",
		path:    "/api/v2/kb/workspaces/{slug}/grants/{target}",
		missing: []string{"/api/v2/kb/workspaces/{slug}/grants/" + kbMissingID},
		body:    `{"role":"editor"}`, okStatus: http.StatusOK,
	},
	{
		name: "ワークスペース権限取り消し", method: http.MethodDelete,
		pattern:  "/api/v2/kb/workspaces/:workspaceSlug/grants/:principalId",
		path:     "/api/v2/kb/workspaces/{slug}/grants/{target}",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/grants/" + kbMissingID},
		okStatus: http.StatusNoContent,
	},
	{
		name: "スペース権限付与", method: http.MethodPut,
		pattern: "/api/v2/kb/workspaces/:workspaceSlug/spaces/:spaceId/grants/:principalId",
		path:    "/api/v2/kb/workspaces/{slug}/spaces/{space}/grants/{target}",
		missing: []string{
			"/api/v2/kb/workspaces/{slug}/spaces/" + kbMissingID + "/grants/{target}",
			"/api/v2/kb/workspaces/{slug}/spaces/{space}/grants/" + kbMissingID,
		},
		body: `{"role":"editor"}`, okStatus: http.StatusOK,
	},
	{
		name: "スペース権限取り消し", method: http.MethodDelete,
		pattern: "/api/v2/kb/workspaces/:workspaceSlug/spaces/:spaceId/grants/:principalId",
		path:    "/api/v2/kb/workspaces/{slug}/spaces/{space}/grants/{target}",
		missing: []string{
			"/api/v2/kb/workspaces/{slug}/spaces/" + kbMissingID + "/grants/{target}",
		},
		okStatus: http.StatusNoContent,
	},
	{
		name: "ページ例外の設定", method: http.MethodPut,
		pattern: "/api/v2/kb/workspaces/:workspaceSlug/pages/:pageId/restrictions/:principalId/:capability",
		path:    "/api/v2/kb/workspaces/{slug}/pages/{page}/restrictions/{target}/view",
		missing: []string{
			"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingID + "/restrictions/{target}/view",
			"/api/v2/kb/workspaces/{slug}/pages/{page}/restrictions/" + kbMissingID + "/view",
		},
		body: `{"mode":"deny"}`, okStatus: http.StatusOK,
	},
	{
		name: "ページ例外の解除", method: http.MethodDelete,
		pattern: "/api/v2/kb/workspaces/:workspaceSlug/pages/:pageId/restrictions/:principalId/:capability",
		path:    "/api/v2/kb/workspaces/{slug}/pages/{page}/restrictions/{target}/view",
		missing: []string{
			"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingID + "/restrictions/{target}/view",
		},
		okStatus: http.StatusNoContent,
	},
	{
		name: "メンバー追加", method: http.MethodPut,
		pattern:  "/api/v2/kb/workspaces/:workspaceSlug/members/:userId",
		path:     "/api/v2/kb/workspaces/{slug}/members/" + strconv.FormatUint(kbSecondUserID, 10),
		missing:  []string{"/api/v2/kb/workspaces/{slug}/members/" + kbMissingUserID},
		okStatus: http.StatusOK,
	},
	{
		name: "メンバー削除", method: http.MethodDelete,
		pattern:  "/api/v2/kb/workspaces/:workspaceSlug/members/:userId",
		path:     "/api/v2/kb/workspaces/{slug}/members/" + strconv.FormatUint(kbSecondUserID, 10),
		missing:  []string{"/api/v2/kb/workspaces/{slug}/members/" + kbMissingUserID},
		okStatus: http.StatusNoContent,
	},
	{
		name: "グループ作成", method: http.MethodPost,
		pattern: "/api/v2/kb/workspaces/:workspaceSlug/groups",
		path:    "/api/v2/kb/workspaces/{slug}/groups",
		body:    `{"name":"運用チーム"}`, okStatus: http.StatusCreated,
	},
	{
		name: "グループメンバー追加", method: http.MethodPut,
		pattern:  "/api/v2/kb/workspaces/:workspaceSlug/groups/:groupPrincipalId/members/:userId",
		path:     "/api/v2/kb/workspaces/{slug}/groups/{group}/members/" + strconv.FormatUint(kbSecondUserID, 10),
		missing:  []string{"/api/v2/kb/workspaces/{slug}/groups/" + kbMissingID + "/members/" + strconv.FormatUint(kbSecondUserID, 10)},
		okStatus: http.StatusNoContent,
	},
	{
		name: "グループメンバー削除", method: http.MethodDelete,
		pattern:  "/api/v2/kb/workspaces/:workspaceSlug/groups/:groupPrincipalId/members/:userId",
		path:     "/api/v2/kb/workspaces/{slug}/groups/{group}/members/" + strconv.FormatUint(kbSecondUserID, 10),
		missing:  []string{"/api/v2/kb/workspaces/{slug}/groups/" + kbMissingID + "/members/" + strconv.FormatUint(kbSecondUserID, 10)},
		okStatus: http.StatusNoContent,
	},
	{
		name: "スペース全員主体の用意", method: http.MethodPut,
		pattern:  "/api/v2/kb/workspaces/:workspaceSlug/spaces/:spaceId/principals/everyone",
		path:     "/api/v2/kb/workspaces/{slug}/spaces/{space}/principals/everyone",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/spaces/" + kbMissingID + "/principals/everyone"},
		okStatus: http.StatusOK,
	},
	{
		name: "共有リンク一覧", method: http.MethodGet,
		pattern:  "/api/v2/kb/workspaces/:workspaceSlug/pages/:pageId/share-links",
		path:     "/api/v2/kb/workspaces/{slug}/pages/{page}/share-links",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingID + "/share-links"},
		okStatus: http.StatusOK,
	},
	{
		name: "共有リンク発行", method: http.MethodPost,
		pattern: "/api/v2/kb/workspaces/:workspaceSlug/pages/:pageId/share-links",
		path:    "/api/v2/kb/workspaces/{slug}/pages/{page}/share-links",
		missing: []string{"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingID + "/share-links"},
		body:    `{"capability":"view"}`, okStatus: http.StatusCreated,
	},
	{
		name: "共有リンク失効", method: http.MethodDelete,
		pattern: "/api/v2/kb/workspaces/:workspaceSlug/pages/:pageId/share-links/:shareLinkId",
		path:    "/api/v2/kb/workspaces/{slug}/pages/{page}/share-links/{link}",
		missing: []string{
			"/api/v2/kb/workspaces/{slug}/pages/{page}/share-links/" + kbMissingID,
			"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingID + "/share-links/{link}",
		},
		okStatus: http.StatusNoContent,
	},
}

// kbPermFixture は権限操作 API の検証環境。
type kbPermFixture struct {
	kbFixture
	// callerPrincipalID は呼び出し元の主体（メンバーでなければ空）。
	callerPrincipalID string
	// targetPrincipalID は権限を張られる側（別のユーザー）の主体。
	targetPrincipalID string
	// groupPrincipalID はグループの主体。
	groupPrincipalID string
	// shareLinkID は child ページに発行済みの共有リンク。
	shareLinkID string
	// shareToken は shareLinkID の平文トークン（検証経路の入力）。
	shareToken string
}

// newKbPermFixture は uid の立場を作って権限操作 API を叩ける環境を返す。
// role が nil なら uid はワークスペースのメンバーにしない（非メンバーの再現）。
//
// 呼び出し元の役割を setScopeRole ではなく UpsertWorkspaceGrant で張るのは、
// 「最後の admin」の検査が grant の行を数えるため。テストだけ別経路で役割を作ると、
// その検査が本番と違う入力で動いてしまう。
func newKbPermFixture(t *testing.T, uid uint64, role *domain.GrantRole) kbPermFixture {
	t.Helper()
	f := newKbFixture(kbNoPerm, uid)
	ctx := context.Background()

	out := kbPermFixture{kbFixture: f}
	if role != nil {
		caller, err := f.perms.EnsureUserPrincipal(ctx, kbWorkspaceID, uid)
		require.NoError(t, err)
		_, err = f.perms.UpsertWorkspaceGrant(ctx, kbWorkspaceID, caller.ID, *role)
		require.NoError(t, err)
		out.callerPrincipalID = caller.ID
	}

	// 権限を張られる側。呼び出し元自身を対象にすると「最後の admin」の検査に
	// ぶつかって、認可の検証と別の理由で落ちる。
	target, err := f.perms.EnsureUserPrincipal(ctx, kbWorkspaceID, kbSecondUserID)
	require.NoError(t, err)
	out.targetPrincipalID = target.ID

	group, err := f.perms.CreateGroupPrincipal(ctx, kbWorkspaceID, "開発チーム")
	require.NoError(t, err)
	out.groupPrincipalID = group.ID

	out.shareToken = "token-for-test"
	link, err := f.perms.CreateShareLink(ctx, repository.ShareLinkWrite{
		WorkspaceID:     kbWorkspaceID,
		PageID:          kbChildPageID,
		Capability:      domain.CapabilityView,
		TokenHash:       kbTestTokenHash(out.shareToken),
		CreatedByUserID: kbUserID,
	})
	require.NoError(t, err)
	out.shareLinkID = link.ID
	return out
}

// kbTestTokenHash は usecase 側と同じ SHA-256 でトークンを縮める
// （fake に入れた共有リンクを、本物の検証経路から引けるようにするため）。
func kbTestTokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func (f kbPermFixture) fill(s string) string {
	return strings.NewReplacer(
		"{slug}", kbWorkspaceSlug,
		"{space}", kbSpaceID,
		"{page}", kbChildPageID,
		"{target}", f.targetPrincipalID,
		"{group}", f.groupPrincipalID,
		"{link}", f.shareLinkID,
	).Replace(s)
}

func (f kbPermFixture) call(t *testing.T, e kbPermissionEndpoint, path string) (int, []byte) {
	t.Helper()
	w := f.do(t, e.method, f.fill(path), f.fill(e.body))
	return w.Code, w.Body.Bytes()
}

func kbGrantRolePtr(r domain.GrantRole) *domain.GrantRole { return &r }

func Test_ノート権限API_adminだけが通る(t *testing.T) {
	// admin 以外の 5 通り。どれも同じ 404 + 同じ本文で断られなければならない。
	//
	// super_admin 等のアプリ内ロールをここに 1 つも足していないことも同時に固定している
	// （fixture が注入する domain.User にロールを持たせても結果は変わらない）。
	personas := []struct {
		name  string
		uid   uint64
		role  *domain.GrantRole
		setup func(f kbPermFixture)
	}{
		{name: "viewer", uid: kbUserID, role: kbGrantRolePtr(domain.GrantRoleViewer)},
		{name: "commenter", uid: kbUserID, role: kbGrantRolePtr(domain.GrantRoleCommenter)},
		{name: "editor", uid: kbUserID, role: kbGrantRolePtr(domain.GrantRoleEditor)},
		{name: "非メンバー", uid: kbOutsiderUserID, role: nil},
		{
			name: "別ワークスペースのadmin", uid: kbRivalAdminUserID, role: nil,
			setup: func(f kbPermFixture) {
				ctx := context.Background()
				p, err := f.perms.EnsureUserPrincipal(ctx, kbOtherWorkspaceID, kbRivalAdminUserID)
				if err != nil {
					panic(err)
				}
				if _, err := f.perms.UpsertWorkspaceGrant(ctx, kbOtherWorkspaceID, p.ID, domain.GrantRoleAdmin); err != nil {
					panic(err)
				}
			},
		},
	}

	for _, persona := range personas {
		for _, e := range kbPermissionEndpoints {
			t.Run(persona.name+"/"+e.name, func(t *testing.T) {
				f := newKbPermFixture(t, persona.uid, persona.role)
				if persona.setup != nil {
					persona.setup(f)
				}
				code, body := f.call(t, e, e.path)
				assert.Equal(t, http.StatusNotFound, code)
				assert.JSONEq(t, kbDenied, string(body))
			})
		}
	}
}

func Test_ノート権限API_adminは全経路を通れる(t *testing.T) {
	for _, e := range kbPermissionEndpoints {
		t.Run(e.name, func(t *testing.T) {
			f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
			code, body := f.call(t, e, e.path)
			assert.Equal(t, e.okStatus, code, "body=%s", string(body))
		})
	}
}

func Test_ノート権限API_拒否の応答は対象の実在で変わらない(t *testing.T) {
	// 存在オラクル対策の本命。権限の無い相手から見て、実在する対象と存在しない対象の
	// 応答がバイト単位で一致することを固定する（片方だけ別の応答を返すと、ID を
	// 総当たりするだけで中身を読まずに実在を数え上げられる）。
	for _, e := range kbPermissionEndpoints {
		if len(e.missing) == 0 {
			continue // 対象 ID を受け取らない経路（グループ作成）は総当たりの的が無い
		}
		t.Run(e.name, func(t *testing.T) {
			f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleEditor))
			wantCode, wantBody := f.call(t, e, e.path)
			require.Equal(t, http.StatusNotFound, wantCode)

			for _, missing := range e.missing {
				gotCode, gotBody := f.call(t, e, missing)
				assert.Equal(t, wantCode, gotCode, "path=%s", missing)
				assert.Equal(t, wantBody, gotBody, "path=%s（本文がバイト単位で一致すること）", missing)
			}
		})
	}
}

func Test_ノート権限API_ページを名指しする入口は結果によらず同じ回数だけ引く(t *testing.T) {
	// 応答のバイト列を揃えても、返るまでの時間が違えば「そのページ ID が実在するか」が読める。
	// 以前は「ページを引く → スペースの実在を確かめる → 役割を集める」の 3 段で、
	// 落ちる段によって DB の往復が 0 / 1 / 3 回に分かれていた。
	//
	// 数えるのは問い合わせの回数そのもの。時間を測るテストは環境のノイズで揺れるので、
	// 揺れない量で固定する。**特定の数と比べるのではなく、4 通りの内訳が互いに一致すること**を
	// 見る。こうしておくと、別のメソッドで前段の確認が復活しても（回数が結果で変われば）落ちる。
	restrictionPath := func(pageID string) string {
		return "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + pageID + "/restrictions/"
	}

	// snapshot は「権限の読み取り（メソッド名ごと）」と「ページの読み取り」の内訳。
	type snapshot struct {
		permReads map[string]int
		findPage  int
	}
	take := func(f kbPermFixture) snapshot {
		reads := map[string]int{}
		for k, v := range f.perms.permReadCalls {
			reads[k] = v
		}
		return snapshot{permReads: reads, findPage: f.pages.findPageCalls}
	}

	cases := []struct {
		name   string
		role   *domain.GrantRole
		pageID string
	}{
		{"実在するページ・admin ではない", kbGrantRolePtr(domain.GrantRoleEditor), kbChildPageID},
		{"存在しないページ・admin ではない", kbGrantRolePtr(domain.GrantRoleEditor), kbMissingID},
		{"実在するページ・admin", kbGrantRolePtr(domain.GrantRoleAdmin), kbChildPageID},
		{"存在しないページ・admin", kbGrantRolePtr(domain.GrantRoleAdmin), kbMissingID},
	}

	var want *snapshot
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newKbPermFixture(t, kbUserID, tc.role)
			before := take(f)

			w := f.do(t, http.MethodPut,
				restrictionPath(tc.pageID)+f.targetPrincipalID+"/view", `{"mode":"deny"}`)
			require.NotEqual(t, http.StatusInternalServerError, w.Code)

			after := take(f)
			got := snapshot{permReads: map[string]int{}, findPage: after.findPage - before.findPage}
			for k, v := range after.permReads {
				if d := v - before.permReads[k]; d != 0 {
					got.permReads[k] = d
				}
			}

			if want == nil {
				want = &got
				// 認可の前に対象を読まないこと自体も押さえる（読むと必ず回数が結果で揺れる）。
				assert.Equal(t, 0, got.findPage, "認可より先にページを読まない")
				// **絶対値も固定する。** 一致だけを見ると、入口が権限を一切引かずに
				// 一律拒否する退行（全ケース 0 回）でも通ってしまう。
				assert.Equal(t, map[string]int{"PageSpaceScopeFactsForUser": 1}, got.permReads,
					"引くのはページ経由の 1 回だけ")
				return
			}
			assert.Equal(t, *want, got, "結果が違っても引く回数と内訳は同じであること")
		})
	}
}

func Test_ノート権限API_未認証は通らない(t *testing.T) {
	// current user を注入しないルータ。middleware.KnowledgeBaseWorkspace が 401 を返す。
	for _, e := range kbPermissionEndpoints {
		t.Run(e.name, func(t *testing.T) {
			f := newKbPermFixture(t, 0, nil)
			w := f.do(t, e.method, f.fill(e.path), f.fill(e.body))
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func Test_ノート権限API_最後のadminは外せない(t *testing.T) {
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
	path := "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/grants/" + f.callerPrincipalID

	w := f.do(t, http.MethodDelete, path, "")
	assert.Equal(t, http.StatusConflict, w.Code, "取り消しで admin が 0 人になる")
	assert.JSONEq(t, `{"error":"last_workspace_admin"}`, w.Body.String())

	w = f.do(t, http.MethodPut, path, `{"role":"editor"}`)
	assert.Equal(t, http.StatusConflict, w.Code, "降格も admin を外す操作")

	w = f.do(t, http.MethodDelete,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/members/"+strconv.FormatUint(kbUserID, 10), "")
	assert.Equal(t, http.StatusConflict, w.Code, "メンバー削除でも principal ごと消える")
}

func Test_ノート権限API_admin2人目が居れば外せる(t *testing.T) {
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
	ctx := context.Background()
	_, err := f.perms.UpsertWorkspaceGrant(ctx, kbWorkspaceID, f.targetPrincipalID, domain.GrantRoleAdmin)
	require.NoError(t, err)

	w := f.do(t, http.MethodDelete,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/grants/"+f.callerPrincipalID, "")
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_ノート権限API_競合で断られた取り消しも409(t *testing.T) {
	// 手前の検査（CanRemoveWorkspaceAdminUseCase）は読み取りだけなので、admin 2 人を
	// ほぼ同時に外す要求は両方ともそこを通り抜ける。実際に 0 人を止めているのは
	// repository 側で、判定と書き換えを同じトランザクションに入れて断る。
	// そのときの応答が「先に断られた」ときと同じ 409 であることを固定する
	// （撃ち分けると、呼び出し側は競合かどうかで別の分岐を持たされる）。
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
	ctx := context.Background()
	_, err := f.perms.UpsertWorkspaceGrant(ctx, kbWorkspaceID, f.targetPrincipalID, domain.GrantRoleAdmin)
	require.NoError(t, err) // 手前の検査は通る（admin は 2 人）
	f.perms.revokeGrantErr = repository.ErrLastWorkspaceAdmin

	w := f.do(t, http.MethodDelete,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/grants/"+f.callerPrincipalID, "")
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.JSONEq(t, `{"error":"last_workspace_admin"}`, w.Body.String())
}

func Test_ノート権限API_共有リンクは発行時の1回だけトークンを返す(t *testing.T) {
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
	base := "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/" + kbChildPageID + "/share-links"

	issued := f.do(t, http.MethodPost, base, `{"capability":"view"}`)
	require.Equal(t, http.StatusCreated, issued.Code, issued.Body.String())
	var out kbIssuedShareLinkResponse
	require.NoError(t, json.Unmarshal(issued.Body.Bytes(), &out))
	require.NotEmpty(t, out.Token, "発行時だけ平文トークンが返る")

	listed := f.do(t, http.MethodGet, base, "")
	require.Equal(t, http.StatusOK, listed.Code)
	assert.NotContains(t, listed.Body.String(), out.Token, "一覧に平文トークンは出ない")
	assert.NotContains(t, listed.Body.String(), "tokenHash", "ハッシュも出さない")
	assert.NotContains(t, listed.Body.String(), "principalId", "内部の主体 ID も出さない")
}

func Test_ノート権限API_別ページの共有リンクは失効させられない(t *testing.T) {
	// 認可はページ（が属するスペース）で判断するので、リンクが本当にそのページのもので
	// あることを確かめないと、ページ ID とリンク ID を組み替えるだけで
	// 別のページのリンクを止められる。
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
	w := f.do(t, http.MethodDelete,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbRootPageID+"/share-links/"+f.shareLinkID, "")
	assert.Equal(t, http.StatusNotFound, w.Code, "リンクは child ページのもの")
	assert.JSONEq(t, kbDenied, w.Body.String())
}

func Test_ノート権限API_共有リンク検証は未認証で通る(t *testing.T) {
	// current user を注入しないルータでも通ること（リンクを受け取った人はログインしていない）。
	f := newKbPermFixture(t, 0, nil)
	w := f.do(t, http.MethodPost, kbShareLinkVerifyPath, `{"token":"`+f.shareToken+`"}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var got kbVerifiedShareLinkResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, kbChildPageID, got.PageID)
	assert.Equal(t, string(domain.CapabilityView), got.Capability)
	assert.NotContains(t, w.Body.String(), f.shareToken, "応答にトークンを反射しない")
}

func Test_ノート権限API_共有リンク検証は知らないトークンを404にする(t *testing.T) {
	f := newKbPermFixture(t, 0, nil)
	w := f.do(t, http.MethodPost, kbShareLinkVerifyPath, `{"token":"unknown-token"}`)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, kbDenied, w.Body.String())
}

// kbVerifyWithXFF は共有リンク検証を、X-Forwarded-For を毎回変えて叩く。
//
// gin の ClientIP() は XFF の最左を読み、このリポジトリは SetTrustedProxies を
// 呼んでいない（gin の既定は全 IP を信頼する）。つまり **要求元は攻撃者が自由に名乗れる**。
// 「IP を変えれば上限を抜けられるのか」を、その前提のまま再現するためのヘルパ。
func kbVerifyWithXFF(t *testing.T, f kbPermFixture, token, xff string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, kbShareLinkVerifyPath,
		strings.NewReader(`{"token":"`+token+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "198.51.100.7:1234"
	req.Header.Set("X-Forwarded-For", xff)
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func Test_ノート権限API_共有リンク検証はIPを変えても頭打ちになる(t *testing.T) {
	// パスワード付きリンクのパスワードは人が選ぶ短い値で、総当たりに弱い。
	// 上限の鍵を IP に取ると攻撃者が鍵ごと変えられるので、鍵はリンクそのものに取っている。
	// ここで固定するのは「IP を毎回変えても、同じリンクへの試行は必ず頭打ちになる」こと。
	f := newKbPermFixture(t, 0, nil)

	codes := make([]int, 0, kbShareLinkVerifyBurst+5)
	for i := 0; i < kbShareLinkVerifyBurst+5; i++ {
		w := kbVerifyWithXFF(t, f, f.shareToken, "203.0.113."+strconv.Itoa(i))
		codes = append(codes, w.Code)
	}
	for i := 0; i < kbShareLinkVerifyBurst; i++ {
		require.Equal(t, http.StatusOK, codes[i], "burst 内は通る: %v", codes)
	}
	assert.Equal(t, http.StatusTooManyRequests, codes[kbShareLinkVerifyBurst],
		"IP を変えても同じリンクへの試行は頭打ちになる: %v", codes)
	last := kbVerifyWithXFF(t, f, f.shareToken, "192.0.2.250")
	assert.Equal(t, http.StatusTooManyRequests, last.Code)
	assert.Equal(t, "60", last.Header().Get("Retry-After"))
}

func Test_ノート権限API_共有リンクの上限は別のリンクを巻き込まない(t *testing.T) {
	// 上限が「リンク 1 本ごと」であることの裏側。1 本を叩き切っても、他のリンクを
	// 受け取った人は開ける（鍵がリンクなので、巻き添えが起きるとしたらここ）。
	f := newKbPermFixture(t, 0, nil)
	other := "another-token-for-test"
	_, err := f.perms.CreateShareLink(context.Background(), repository.ShareLinkWrite{
		WorkspaceID:     kbWorkspaceID,
		PageID:          kbChildPageID,
		Capability:      domain.CapabilityView,
		TokenHash:       kbTestTokenHash(other),
		CreatedByUserID: kbUserID,
	})
	require.NoError(t, err)

	for i := 0; i < kbShareLinkVerifyBurst+3; i++ {
		kbVerifyWithXFF(t, f, f.shareToken, "203.0.113."+strconv.Itoa(i))
	}
	require.Equal(t, http.StatusTooManyRequests,
		kbVerifyWithXFF(t, f, f.shareToken, "203.0.113.99").Code, "1 本目は頭打ち")
	assert.Equal(t, http.StatusOK,
		kbVerifyWithXFF(t, f, other, "203.0.113.99").Code, "別のリンクは巻き添えにならない")
}

func Test_ノート権限API_存在しないトークンは上限の的にならない(t *testing.T) {
	// 鍵は要求ごとに変えられる（トークンは攻撃者が名乗る値）ので、実在しないリンクの
	// 鍵まで limiter に残すと、でたらめなトークンを投げ続けるだけで中身を太らせられる。
	// 実在しないトークンは 1 回ごとに鍵ごと捨てるので、何度投げても 404 のまま
	// （パスワードの総当たりには実在するトークンが要るので、これで守りは緩まない）。
	f := newKbPermFixture(t, 0, nil)
	for i := 0; i < kbShareLinkVerifyBurst+10; i++ {
		w := kbVerifyWithXFF(t, f, "unknown-"+strconv.Itoa(i), "203.0.113."+strconv.Itoa(i))
		require.Equal(t, http.StatusNotFound, w.Code, "%d 回目", i+1)
	}
	// 実在するリンクの上限は 1 つも減っていない。
	assert.Equal(t, http.StatusOK, kbVerifyWithXFF(t, f, f.shareToken, "192.0.2.1").Code)
}

func Test_ノート権限API_同じ存在しないトークンでも上限の的にならない(t *testing.T) {
	f := newKbPermFixture(t, 0, nil)
	for i := 0; i < kbShareLinkVerifyBurst+10; i++ {
		w := kbVerifyWithXFF(t, f, "always-unknown", "203.0.113."+strconv.Itoa(i))
		require.Equal(t, http.StatusNotFound, w.Code, "%d 回目", i+1)
	}
}

func Test_ノート権限API_メンバー追加はユーザー単位で頭打ちになる(t *testing.T) {
	// この口は users.id をそのまま受け取り、200 と 404 の差で実在が分かる。
	// ワークスペースは誰でも作れて作った本人が admin になるので、放っておくと
	// 全ログインユーザーが使えるユーザー ID の走査器になる。
	// 鍵は検証済み JWT 由来のユーザー ID なので、XFF を変えても抜けられない。
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
	base := "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/members/"

	call := func(userID int, xff string) int {
		req := httptest.NewRequest(http.MethodPut, base+strconv.Itoa(userID), nil)
		req.RemoteAddr = "198.51.100.7:1234"
		req.Header.Set("X-Forwarded-For", xff)
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, req)
		return w.Code
	}

	for i := 0; i < kbAddMemberBurst; i++ {
		require.NotEqual(t, http.StatusTooManyRequests, call(1000+i, "203.0.113."+strconv.Itoa(i)),
			"burst 内は通る: %d 回目", i+1)
	}
	assert.Equal(t, http.StatusTooManyRequests, call(2000, "203.0.113.200"),
		"IP を変えても同じユーザーなら頭打ちになる")
}

// Test_ノート権限API_共有リンクの発行応答以外に平文トークンが出ない は、平文トークンが
// 応答の中だけに現れることを固定する。監査ログは全廃したが、トークンが他所へ漏れない
// ことの確認は残す（発行は応答に平文が載る唯一の経路なので、扱いを間違えやすい）。
func Test_ノート権限API_共有リンクの発行応答以外に平文トークンが出ない(t *testing.T) {
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
	const password = "sup3r-secret-passphrase"
	w := f.do(t, http.MethodPost,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/pages/"+kbChildPageID+"/share-links",
		`{"capability":"view","password":"`+password+`"}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var issued kbIssuedShareLinkResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &issued))
	require.NotEmpty(t, issued.Token)
	// 受け取ったパスワードは応答に出さない（保存はハッシュで、平文は持ち回らない）。
	assert.NotContains(t, w.Body.String(), password)
}

func Test_ノート権限API_未知の役割は400(t *testing.T) {
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleAdmin))
	w := f.do(t, http.MethodPut,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/grants/"+f.targetPrincipalID,
		`{"role":"super_admin"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"アプリ内ロールは grant の役割として通らない（権限の出どころを 2 系統にしない）")
}

// kbSuperAdminRouter は current user に super_admin を持たせたルータを組む
// （本番と同じ registerKnowledgeBaseRoutesWith を通す）。
func kbSuperAdminRouter(f kbFixture, uid uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v2")
	g.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUserID, uid)
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{
			ID: uid, Role: domain.RoleSuperAdmin, RoleID: domain.RoleIDSuperAdmin,
		})
		c.Next()
	})
	registerKnowledgeBaseRoutesWith(g, f.pages, f.perms, f.provisioner, f.users)
	return r
}

// アプリ内ロール（super_admin）を持っていても権限操作は通らないことを固定する。
// ノートの権限は grant だけで決まり、「特権ロールなら全部できる」という
// 抜け道を持たない（domain/grant.go・kb_permission_gate.go）。
func Test_ノート権限API_アプリ内のsuperadminでも通らない(t *testing.T) {
	f := newKbPermFixture(t, kbUserID, kbGrantRolePtr(domain.GrantRoleViewer))
	router := kbSuperAdminRouter(f.kbFixture, kbUserID)

	req := httptest.NewRequest(http.MethodPut,
		"/api/v2/kb/workspaces/"+kbWorkspaceSlug+"/grants/"+f.targetPrincipalID,
		strings.NewReader(`{"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, kbDenied, w.Body.String())
}
