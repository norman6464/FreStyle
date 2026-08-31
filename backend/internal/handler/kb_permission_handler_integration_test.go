//go:build integration

package handler

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ノートの権限操作 API を、実 PostgreSQL・本番と同じ配線で確かめる。
//
// 権限を書き換える usecase は認可を一切見ない（受け取った ID をそのまま書く）ので、
// 認可が効いているかどうかは HTTP の入口を実際に叩かないと分からない。ここで固定するのは:
//
//  1. admin 以外は 1 本も通らないこと（viewer / commenter / editor / 非メンバー /
//     別ワークスペースの admin の 5 通り）
//  2. 拒否の応答が、対象が実在するかどうかで変わらないこと（ステータスも本文もバイト単位で同じ）
//  3. admin なら通ること、そして通した結果が実効権限に反映されること
//
// fake ではなく本物の DB を通すのは、権限の解決が SQL（1 本のクエリで事実を集める）に
// 寄っているため。fake は本番より賢くも馬鹿にもなり得るので、認可の最終的な担保はこちら。

// kbPermEnv は権限操作 API の検証環境。kbEnv（ワークスペース + スペース）に、
// 役割の違う利用者・ページ・共有リンクを足したもの。
type kbPermEnv struct {
	*kbEnv
	admin     uint64
	viewer    uint64
	commenter uint64
	editor    uint64
	outsider  uint64
	// rivalAdmin は別ワークスペースの admin（このワークスペースには所属しない）。
	rivalAdmin uint64
	// target は権限を張られる側。呼び出し元自身を対象にすると「最後の admin」の検査に
	// ぶつかり、認可とは別の理由で落ちてしまう。
	target          uint64
	targetPrincipal string
	adminPrincipal  string
	groupPrincipal  string
	rootPage        string
	childPage       string
	shareLinkID     string
	shareToken      string
}

// kbSharedToken は検証経路に渡す平文トークン。usecase と同じ SHA-256 で保存する。
const kbSharedToken = "integration-share-token"

func kbTokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func newKbPermEnv(t *testing.T, sqlDB *sql.DB) *kbPermEnv {
	t.Helper()
	env := newKbEnv(t, sqlDB, "acme")
	ctx := t.Context()

	e := &kbPermEnv{
		kbEnv:     env,
		admin:     kbInsertUser(t, sqlDB, "admin"),
		viewer:    kbInsertUser(t, sqlDB, "viewer"),
		commenter: kbInsertUser(t, sqlDB, "commenter"),
		editor:    kbInsertUser(t, sqlDB, "editor"),
		outsider:  kbInsertUser(t, sqlDB, "outsider"),
		target:    kbInsertUser(t, sqlDB, "target"),
	}
	e.adminPrincipal = env.joinWorkspace(t, e.admin, domain.GrantRoleAdmin).ID
	env.joinWorkspace(t, e.viewer, domain.GrantRoleViewer)
	env.joinWorkspace(t, e.commenter, domain.GrantRoleCommenter)
	env.joinWorkspace(t, e.editor, domain.GrantRoleEditor)
	e.targetPrincipal = env.joinWorkspace(t, e.target, domain.GrantRoleViewer).ID

	// 別ワークスペースの admin。同じ DB に居るが acme には principals の行が無い。
	e.rivalAdmin = kbInsertUser(t, sqlDB, "rival-admin")
	rivalWS := kbInsertWorkspace(t, sqlDB, "rival")
	rivalPrincipal, err := env.permissions.EnsureUserPrincipal(ctx, rivalWS, e.rivalAdmin)
	require.NoError(t, err)
	_, err = env.permissions.UpsertWorkspaceGrant(ctx, rivalWS, rivalPrincipal.ID, domain.GrantRoleAdmin)
	require.NoError(t, err)

	group, err := env.permissions.CreateGroupPrincipal(ctx, env.workspaceID, "開発チーム")
	require.NoError(t, err)
	e.groupPrincipal = group.ID

	e.rootPage = kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, e.admin, "a0", "root")
	e.childPage = kbInsertChildPage(t, sqlDB, env.workspaceID, env.spaceID, e.rootPage, e.admin, "a1", "child")

	e.shareToken = kbSharedToken
	link, err := env.permissions.CreateShareLink(ctx, repository.ShareLinkWrite{
		WorkspaceID:     env.workspaceID,
		PageID:          e.childPage,
		Capability:      domain.CapabilityView,
		TokenHash:       kbTokenHash(e.shareToken),
		CreatedByUserID: e.admin,
	})
	require.NoError(t, err)
	e.shareLinkID = link.ID
	return e
}

// kbInsertChildPage は親を持つページを直接入れる（closure も張る）。
func kbInsertChildPage(t *testing.T, db *sql.DB, workspaceID, spaceID, parentID string, createdBy uint64, position, title string) string {
	t.Helper()
	id := kbNewUUID()
	_, err := db.Exec(
		`INSERT INTO pages (id, workspace_id, space_id, parent_id, "position", title, created_by_user_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, workspaceID, spaceID, parentID, position, title, createdBy,
	)
	require.NoError(t, err)
	// closure は「自分自身（depth 0）+ 親の経路を 1 段ずらしたもの」。
	// UUID 列と比べるパラメータは明示的にキャストする（テキストのまま推論されると
	// uuid = text で落ちる）。
	_, err = db.Exec(
		`INSERT INTO page_paths (workspace_id, page_id, ancestor_id, depth)
		 SELECT $1::uuid, $2::uuid, $2::uuid, 0
		 UNION ALL
		 SELECT $1::uuid, $2::uuid, ancestor_id, depth + 1 FROM page_paths
		 WHERE workspace_id = $1::uuid AND page_id = $3::uuid`,
		workspaceID, id, parentID,
	)
	require.NoError(t, err)
	return id
}

func (e *kbPermEnv) fill(s string) string {
	return strings.NewReplacer(
		"{slug}", e.slug,
		"{space}", e.spaceID,
		"{page}", e.childPage,
		"{target}", e.targetPrincipal,
		"{group}", e.groupPrincipal,
		"{link}", e.shareLinkID,
		"{user}", strconv.FormatUint(e.target, 10),
	).Replace(s)
}

// kbPermCase は権限操作の 1 経路。missing は対象を存在しない ID に差し替えたパス。
type kbPermCase struct {
	name     string
	method   string
	path     string
	missing  []string
	body     string
	okStatus int
}

// kbMissingIntegrationUUID は存在しない UUID（実在する ID と同じ形にする — 形式不正で弾かれると
// 「権限が無い」経路を通っていないのに 404 が返り、テストが空振りする）。
const kbMissingIntegrationUUID = "0198a000-0000-7000-8000-0000000000ff"

// kbMissingIntegrationUserID は存在しないユーザー ID
// （users は他の結合テストと共有するので、実在しそうにない大きな値を使う）。
const kbMissingIntegrationUserID = "987654321"

// kbPermCases は「権限そのものを変える」全経路。
var kbPermCases = []kbPermCase{
	{
		name: "ワークスペース権限付与", method: http.MethodPut,
		path:    "/api/v2/kb/workspaces/{slug}/grants/{target}",
		missing: []string{"/api/v2/kb/workspaces/{slug}/grants/" + kbMissingIntegrationUUID},
		body:    `{"role":"editor"}`, okStatus: http.StatusOK,
	},
	{
		name: "ワークスペース権限取り消し", method: http.MethodDelete,
		path:     "/api/v2/kb/workspaces/{slug}/grants/{target}",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/grants/" + kbMissingIntegrationUUID},
		okStatus: http.StatusNoContent,
	},
	{
		name: "スペース権限付与", method: http.MethodPut,
		path: "/api/v2/kb/workspaces/{slug}/spaces/{space}/grants/{target}",
		missing: []string{
			"/api/v2/kb/workspaces/{slug}/spaces/" + kbMissingIntegrationUUID + "/grants/{target}",
			"/api/v2/kb/workspaces/{slug}/spaces/{space}/grants/" + kbMissingIntegrationUUID,
		},
		body: `{"role":"editor"}`, okStatus: http.StatusOK,
	},
	{
		name: "スペース権限取り消し", method: http.MethodDelete,
		path:     "/api/v2/kb/workspaces/{slug}/spaces/{space}/grants/{target}",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/spaces/" + kbMissingIntegrationUUID + "/grants/{target}"},
		okStatus: http.StatusNoContent,
	},
	{
		name: "ページ例外の設定", method: http.MethodPut,
		path: "/api/v2/kb/workspaces/{slug}/pages/{page}/restrictions/{target}/view",
		missing: []string{
			"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingIntegrationUUID + "/restrictions/{target}/view",
			"/api/v2/kb/workspaces/{slug}/pages/{page}/restrictions/" + kbMissingIntegrationUUID + "/view",
		},
		body: `{"mode":"deny"}`, okStatus: http.StatusOK,
	},
	{
		name: "ページ例外の解除", method: http.MethodDelete,
		path:     "/api/v2/kb/workspaces/{slug}/pages/{page}/restrictions/{target}/view",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingIntegrationUUID + "/restrictions/{target}/view"},
		okStatus: http.StatusNoContent,
	},
	{
		name: "メンバー追加", method: http.MethodPut,
		path:     "/api/v2/kb/workspaces/{slug}/members/{user}",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/members/" + kbMissingIntegrationUserID},
		okStatus: http.StatusOK,
	},
	{
		name: "メンバー削除", method: http.MethodDelete,
		path:     "/api/v2/kb/workspaces/{slug}/members/{user}",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/members/" + kbMissingIntegrationUserID},
		okStatus: http.StatusNoContent,
	},
	{
		name: "グループ作成", method: http.MethodPost,
		path: "/api/v2/kb/workspaces/{slug}/groups",
		body: `{"name":"運用チーム"}`, okStatus: http.StatusCreated,
	},
	{
		name: "グループメンバー追加", method: http.MethodPut,
		path:     "/api/v2/kb/workspaces/{slug}/groups/{group}/members/{user}",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/groups/" + kbMissingIntegrationUUID + "/members/{user}"},
		okStatus: http.StatusNoContent,
	},
	{
		name: "グループメンバー削除", method: http.MethodDelete,
		path:     "/api/v2/kb/workspaces/{slug}/groups/{group}/members/{user}",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/groups/" + kbMissingIntegrationUUID + "/members/{user}"},
		okStatus: http.StatusNoContent,
	},
	{
		name: "スペース全員主体の用意", method: http.MethodPut,
		path:     "/api/v2/kb/workspaces/{slug}/spaces/{space}/principals/everyone",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/spaces/" + kbMissingIntegrationUUID + "/principals/everyone"},
		okStatus: http.StatusOK,
	},
	{
		name: "共有リンク一覧", method: http.MethodGet,
		path:     "/api/v2/kb/workspaces/{slug}/pages/{page}/share-links",
		missing:  []string{"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingIntegrationUUID + "/share-links"},
		okStatus: http.StatusOK,
	},
	{
		name: "共有リンク発行", method: http.MethodPost,
		path:    "/api/v2/kb/workspaces/{slug}/pages/{page}/share-links",
		missing: []string{"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingIntegrationUUID + "/share-links"},
		body:    `{"capability":"view"}`, okStatus: http.StatusCreated,
	},
	{
		name: "共有リンク失効", method: http.MethodDelete,
		path: "/api/v2/kb/workspaces/{slug}/pages/{page}/share-links/{link}",
		missing: []string{
			"/api/v2/kb/workspaces/{slug}/pages/{page}/share-links/" + kbMissingIntegrationUUID,
			"/api/v2/kb/workspaces/{slug}/pages/" + kbMissingIntegrationUUID + "/share-links/{link}",
		},
		okStatus: http.StatusNoContent,
	},
}

// kbDeniedBody は権限操作 API の唯一の拒否応答。バイト列で固定する。
const kbDeniedBody = `{"error":"not_found"}`

func TestKnowledgeBasePermissionAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	t.Run("admin以外は権限操作を1本も通せない", func(t *testing.T) {
		// この 5 通りが 1 つでも通ると、認証さえ済ませれば自分を admin にできる。
		// 応答はすべて同じ 404 + 同じ本文でなければならない（誰なのかも漏らさない）。
		for _, persona := range []string{"viewer", "commenter", "editor", "非メンバー", "別ワークスペースのadmin"} {
			for _, tc := range kbPermCases {
				t.Run(persona+"/"+tc.name, func(t *testing.T) {
					env := newKbPermEnv(t, sqlDB)
					var uid uint64
					switch persona {
					case "viewer":
						uid = env.viewer
					case "commenter":
						uid = env.commenter
					case "editor":
						uid = env.editor
					case "非メンバー":
						uid = env.outsider
					default:
						uid = env.rivalAdmin
					}
					e := env.as(uid)
					w := e.do(t, tc.method, env.fill(tc.path), env.fill(tc.body))
					assert.Equal(t, http.StatusNotFound, w.Code, "body=%s", w.Body.String())
					assert.Equal(t, kbDeniedBody, w.Body.String())
				})
			}
		}
	})

	t.Run("拒否の応答は対象の実在で変わらない", func(t *testing.T) {
		// 存在オラクル対策の本命。権限の無い相手から見て、実在する対象と存在しない対象で
		// ステータスも本文もバイト単位で一致すること。片方だけ違えば、ID を総当たりする
		// だけで中身を読まずに実在を数え上げられる。
		for _, tc := range kbPermCases {
			if len(tc.missing) == 0 {
				continue // 対象 ID を受け取らない経路は総当たりの的が無い
			}
			t.Run(tc.name, func(t *testing.T) {
				env := newKbPermEnv(t, sqlDB)
				e := env.as(env.editor)

				real := e.do(t, tc.method, env.fill(tc.path), env.fill(tc.body))
				require.Equal(t, http.StatusNotFound, real.Code)
				wantBody := real.Body.Bytes()

				for _, missing := range tc.missing {
					got := e.do(t, tc.method, env.fill(missing), env.fill(tc.body))
					assert.Equal(t, real.Code, got.Code, "path=%s", missing)
					assert.Equal(t, wantBody, got.Body.Bytes(),
						"path=%s（本文がバイト単位で一致すること）", missing)
				}
			})
		}
	})

	t.Run("adminは全経路を通れる", func(t *testing.T) {
		for _, tc := range kbPermCases {
			t.Run(tc.name, func(t *testing.T) {
				env := newKbPermEnv(t, sqlDB)
				e := env.as(env.admin)
				w := e.do(t, tc.method, env.fill(tc.path), env.fill(tc.body))
				assert.Equal(t, tc.okStatus, w.Code, "body=%s", w.Body.String())
			})
		}
	})

	t.Run("付与した権限がそのまま実効権限になる", func(t *testing.T) {
		// 認可を通したあとの書き込みが本当に効いているか（配線だけして書けていない、を防ぐ）。
		env := newKbPermEnv(t, sqlDB)
		admin := env.as(env.admin)
		target := env.as(env.target)

		// 前提: target は viewer なので改名できない。
		renamePath := "/api/v2/kb/workspaces/" + env.slug + "/pages/" + env.childPage
		require.Equal(t, http.StatusForbidden,
			target.do(t, http.MethodPatch, renamePath, `{"title":"改訂"}`).Code)

		granted := admin.do(t, http.MethodPut,
			"/api/v2/kb/workspaces/"+env.slug+"/grants/"+env.targetPrincipal, `{"role":"editor"}`)
		require.Equal(t, http.StatusOK, granted.Code, granted.Body.String())

		assert.Equal(t, http.StatusOK,
			target.do(t, http.MethodPatch, renamePath, `{"title":"改訂"}`).Code,
			"editor へ上げたら改名できる")

		// ページ単位の deny を張ると、既定が editor でも見えなくなる（例外が既定に勝つ）。
		denied := admin.do(t, http.MethodPut,
			"/api/v2/kb/workspaces/"+env.slug+"/pages/"+env.childPage+
				"/restrictions/"+env.targetPrincipal+"/view", `{"mode":"deny"}`)
		require.Equal(t, http.StatusOK, denied.Code, denied.Body.String())
		assert.Equal(t, http.StatusNotFound,
			target.do(t, http.MethodGet, renamePath, "").Code,
			"deny されたページは存在しないページと同じ 404")
	})

	t.Run("スペースadminは自分のスペースだけを変えられる", func(t *testing.T) {
		// スペースの admin は「そのスペースで権限を変えられる」だけで、
		// ワークスペース全体の grant やメンバーの出入りには手が届かない。
		env := newKbPermEnv(t, sqlDB)
		spaceAdmin := kbInsertUser(t, sqlDB, "space-admin")
		principal := env.joinWorkspace(t, spaceAdmin, domain.GrantRoleViewer)
		_, err := env.permissions.UpsertSpaceGrant(
			t.Context(), env.workspaceID, env.spaceID, principal.ID, domain.GrantRoleAdmin,
		)
		require.NoError(t, err)
		e := env.as(spaceAdmin)

		assert.Equal(t, http.StatusOK,
			e.do(t, http.MethodPut,
				"/api/v2/kb/workspaces/"+env.slug+"/spaces/"+env.spaceID+"/grants/"+env.targetPrincipal,
				`{"role":"editor"}`).Code,
			"自分のスペースの grant は変えられる")

		w := e.do(t, http.MethodPut,
			"/api/v2/kb/workspaces/"+env.slug+"/grants/"+env.targetPrincipal, `{"role":"admin"}`)
		assert.Equal(t, http.StatusNotFound, w.Code, "ワークスペース全体の grant には届かない")
		assert.Equal(t, kbDeniedBody, w.Body.String())

		w = e.do(t, http.MethodPut,
			"/api/v2/kb/workspaces/"+env.slug+"/members/"+strconv.FormatUint(env.outsider, 10), "")
		assert.Equal(t, http.StatusNotFound, w.Code, "メンバーの追加にも届かない")
	})

	t.Run("別スペースのスペースadminは他スペースの権限を変えられない", func(t *testing.T) {
		env := newKbPermEnv(t, sqlDB)
		otherSpace := kbInsertSpace(t, sqlDB, env.workspaceID, "ops")
		spaceAdmin := kbInsertUser(t, sqlDB, "ops-admin")
		principal := env.joinWorkspace(t, spaceAdmin, domain.GrantRoleViewer)
		_, err := env.permissions.UpsertSpaceGrant(
			t.Context(), env.workspaceID, otherSpace, principal.ID, domain.GrantRoleAdmin,
		)
		require.NoError(t, err)

		w := env.as(spaceAdmin).do(t, http.MethodPut,
			"/api/v2/kb/workspaces/"+env.slug+"/spaces/"+env.spaceID+"/grants/"+env.targetPrincipal,
			`{"role":"editor"}`)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, kbDeniedBody, w.Body.String())
	})

	t.Run("最後のadminは外せない", func(t *testing.T) {
		env := newKbPermEnv(t, sqlDB)
		e := env.as(env.admin)
		grantPath := "/api/v2/kb/workspaces/" + env.slug + "/grants/" + env.adminPrincipal

		w := e.do(t, http.MethodDelete, grantPath, "")
		assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"last_workspace_admin"}`, w.Body.String())

		assert.Equal(t, http.StatusConflict,
			e.do(t, http.MethodPut, grantPath, `{"role":"editor"}`).Code, "降格も admin を外す操作")

		assert.Equal(t, http.StatusConflict,
			e.do(t, http.MethodDelete,
				"/api/v2/kb/workspaces/"+env.slug+"/members/"+strconv.FormatUint(env.admin, 10), "").Code,
			"メンバー削除でも principal ごと消える")

		// 2 人目の admin を立てれば外せる。
		require.Equal(t, http.StatusOK,
			e.do(t, http.MethodPut,
				"/api/v2/kb/workspaces/"+env.slug+"/grants/"+env.targetPrincipal, `{"role":"admin"}`).Code)
		assert.Equal(t, http.StatusNoContent, e.do(t, http.MethodDelete, grantPath, "").Code)
	})

	t.Run("グループ宛てのadminは最後の1人として数えない", func(t *testing.T) {
		// メンバーが 0 人のグループが「最後の admin」として残ると、結局誰も権限を
		// 変えられなくなる。grant の行からは中身が分からないので数に入れない。
		env := newKbPermEnv(t, sqlDB)
		e := env.as(env.admin)
		_, err := env.permissions.UpsertWorkspaceGrant(
			t.Context(), env.workspaceID, env.groupPrincipal, domain.GrantRoleAdmin,
		)
		require.NoError(t, err)

		w := e.do(t, http.MethodDelete,
			"/api/v2/kb/workspaces/"+env.slug+"/grants/"+env.adminPrincipal, "")
		assert.Equal(t, http.StatusConflict, w.Code, "グループの admin では代わりにならない")
	})

	t.Run("存在しないユーザーのメンバー追加は500ではなく404", func(t *testing.T) {
		// principals.user_id は users への FK。実在しない ID を渡すと制約違反になるが、
		// それは入力の誤りであってサーバの故障ではない。
		env := newKbPermEnv(t, sqlDB)
		w := env.as(env.admin).do(t, http.MethodPut,
			"/api/v2/kb/workspaces/"+env.slug+"/members/"+kbMissingIntegrationUserID, "")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, kbDeniedBody, w.Body.String())
	})

	t.Run("グループ名の重複は500ではなく409", func(t *testing.T) {
		env := newKbPermEnv(t, sqlDB)
		e := env.as(env.admin)
		path := "/api/v2/kb/workspaces/" + env.slug + "/groups"
		require.Equal(t, http.StatusCreated, e.do(t, http.MethodPost, path, `{"name":"運用"}`).Code)
		w := e.do(t, http.MethodPost, path, `{"name":"運用"}`)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.JSONEq(t, `{"error":"group_name_taken"}`, w.Body.String())
	})

	t.Run("未知の役割やケイパビリティは400", func(t *testing.T) {
		env := newKbPermEnv(t, sqlDB)
		e := env.as(env.admin)

		// アプリ内ロール（super_admin）は grant の役割ではない。
		assert.Equal(t, http.StatusBadRequest,
			e.do(t, http.MethodPut,
				"/api/v2/kb/workspaces/"+env.slug+"/grants/"+env.targetPrincipal,
				`{"role":"super_admin"}`).Code)
		assert.Equal(t, http.StatusBadRequest,
			e.do(t, http.MethodPut,
				"/api/v2/kb/workspaces/"+env.slug+"/pages/"+env.childPage+
					"/restrictions/"+env.targetPrincipal+"/manage", `{"mode":"deny"}`).Code)
		assert.Equal(t, http.StatusBadRequest,
			e.do(t, http.MethodPut,
				"/api/v2/kb/workspaces/"+env.slug+"/pages/"+env.childPage+
					"/restrictions/"+env.targetPrincipal+"/view", `{"mode":"maybe"}`).Code)
	})
}

func TestKnowledgeBaseShareLinkAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	t.Run("発行時の1回だけトークンを返し一覧には出さない", func(t *testing.T) {
		env := newKbPermEnv(t, sqlDB)
		e := env.as(env.admin)
		base := "/api/v2/kb/workspaces/" + env.slug + "/pages/" + env.childPage + "/share-links"

		issued := e.do(t, http.MethodPost, base, `{"capability":"view"}`)
		require.Equal(t, http.StatusCreated, issued.Code, issued.Body.String())
		var out kbIssuedShareLinkResponse
		require.NoError(t, json.Unmarshal(issued.Body.Bytes(), &out))
		require.NotEmpty(t, out.Token)

		listed := e.do(t, http.MethodGet, base, "")
		require.Equal(t, http.StatusOK, listed.Code)
		body := listed.Body.String()
		assert.NotContains(t, body, out.Token, "平文トークンは一覧に出ない")
		assert.NotContains(t, body, "tokenHash", "ハッシュも出さない")
		assert.NotContains(t, body, "principalId", "内部の主体 ID も出さない")
	})

	t.Run("検証は未認証で通り失効させると410になる", func(t *testing.T) {
		env := newKbPermEnv(t, sqlDB)
		// current user を注入しないルータ（未認証）。
		anonymous := env.as(0)
		verifyPath := "/api/v2/kb/share-links/verify"

		w := anonymous.do(t, http.MethodPost, verifyPath, `{"token":"`+env.shareToken+`"}`)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got kbVerifiedShareLinkResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, env.childPage, got.PageID)
		assert.Equal(t, string(domain.CapabilityView), got.Capability)
		assert.NotContains(t, w.Body.String(), env.shareToken, "トークンを応答へ反射しない")

		revoked := env.as(env.admin).do(t, http.MethodDelete,
			"/api/v2/kb/workspaces/"+env.slug+"/pages/"+env.childPage+"/share-links/"+env.shareLinkID, "")
		require.Equal(t, http.StatusNoContent, revoked.Code)

		w = anonymous.do(t, http.MethodPost, verifyPath, `{"token":"`+env.shareToken+`"}`)
		assert.Equal(t, http.StatusGone, w.Code)
		assert.JSONEq(t, `{"error":"share_link_revoked"}`, w.Body.String())
	})

	t.Run("知らないトークンは404で期限切れは410", func(t *testing.T) {
		env := newKbPermEnv(t, sqlDB)
		anonymous := env.as(0)
		verifyPath := "/api/v2/kb/share-links/verify"

		w := anonymous.do(t, http.MethodPost, verifyPath, `{"token":"unknown"}`)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, kbDeniedBody, w.Body.String())

		past := time.Now().Add(-time.Hour)
		expiredToken := "expired-token"
		_, err := env.permissions.CreateShareLink(t.Context(), repository.ShareLinkWrite{
			WorkspaceID:     env.workspaceID,
			PageID:          env.childPage,
			Capability:      domain.CapabilityView,
			TokenHash:       kbTokenHash(expiredToken),
			ExpiresAt:       &past,
			CreatedByUserID: env.admin,
		})
		require.NoError(t, err)

		w = anonymous.do(t, http.MethodPost, verifyPath, `{"token":"`+expiredToken+`"}`)
		assert.Equal(t, http.StatusGone, w.Code)
		assert.JSONEq(t, `{"error":"share_link_expired"}`, w.Body.String())
	})

	t.Run("パスワード付きは合致するまで通らない", func(t *testing.T) {
		env := newKbPermEnv(t, sqlDB)
		e := env.as(env.admin)
		anonymous := env.as(0)
		verifyPath := "/api/v2/kb/share-links/verify"

		issued := e.do(t, http.MethodPost,
			"/api/v2/kb/workspaces/"+env.slug+"/pages/"+env.childPage+"/share-links",
			`{"capability":"view","password":"s3cret"}`)
		require.Equal(t, http.StatusCreated, issued.Code, issued.Body.String())
		var out kbIssuedShareLinkResponse
		require.NoError(t, json.Unmarshal(issued.Body.Bytes(), &out))
		assert.True(t, out.Link.RequiresPassword)
		assert.NotContains(t, issued.Body.String(), "s3cret", "パスワードを応答へ反射しない")

		w := anonymous.do(t, http.MethodPost, verifyPath, `{"token":"`+out.Token+`"}`)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error":"password_required"}`, w.Body.String())

		w = anonymous.do(t, http.MethodPost, verifyPath, `{"token":"`+out.Token+`","password":"wrong"}`)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.JSONEq(t, `{"error":"password_mismatch"}`, w.Body.String())

		w = anonymous.do(t, http.MethodPost, verifyPath, `{"token":"`+out.Token+`","password":"s3cret"}`)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("別ページのリンクIDを渡しても失効させられない", func(t *testing.T) {
		// 認可はページ（が属するスペース）で判断するので、リンクが本当にそのページの
		// ものであることを確かめないと、ページ ID とリンク ID を組み替えるだけで
		// 別のページのリンクを止められる。
		env := newKbPermEnv(t, sqlDB)
		w := env.as(env.admin).do(t, http.MethodDelete,
			"/api/v2/kb/workspaces/"+env.slug+"/pages/"+env.rootPage+"/share-links/"+env.shareLinkID, "")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, kbDeniedBody, w.Body.String())

		// 本当のページを指せば止まる（上の 404 が「そもそも失効できない」ではないことの確認）。
		assert.Equal(t, http.StatusNoContent,
			env.as(env.admin).do(t, http.MethodDelete,
				"/api/v2/kb/workspaces/"+env.slug+"/pages/"+env.childPage+"/share-links/"+env.shareLinkID, "").Code)
	})
}

// kbSuperAdminEnvRouter は current user に super_admin を持たせたルータを組む
// （kbEnv.as は ID しか載せないので、アプリ内ロールを持つ相手をここで作る）。
func kbSuperAdminEnvRouter(e *kbPermEnv, userID uint64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v2")
	g.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUserID, userID)
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{
			ID: userID, Role: domain.RoleSuperAdmin, RoleID: domain.RoleIDSuperAdmin,
		})
		c.Next()
	})
	registerKnowledgeBaseRoutesWith(g, e.pages, e.permissions, e.provisioner, e.users)
	return r
}

// アプリ内ロール（super_admin）を持っていても権限操作は通らないことを実 DB で固定する。
// ノートの権限は principals / grants だけで閉じており、「特権ロールなら全部できる」
// という抜け道を持たない（domain/grant.go・kb_permission_gate.go）。
func TestKnowledgeBasePermissionAPI_SuperAdminHasNoBypass_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	env := newKbPermEnv(t, sqlDB)

	for _, tc := range []struct {
		name   string
		userID uint64
	}{
		{name: "ワークスペースのviewer", userID: env.viewer},
		{name: "非メンバー", userID: env.outsider},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := kbSuperAdminEnvRouter(env, tc.userID)
			req := httptest.NewRequest(http.MethodPut,
				"/api/v2/kb/workspaces/"+env.slug+"/grants/"+env.targetPrincipal,
				strings.NewReader(`{"role":"admin"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)
			assert.Equal(t, kbDeniedBody, w.Body.String())
		})
	}
}
