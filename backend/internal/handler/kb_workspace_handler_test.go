package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ワークスペース / スペースの API は判定対象がページではないので、kbEndpoints の表
// （ページ 1 枚の権限を軸に回す）とは別にここで検証する。

func Test_ノートAPI_所属ワークスペース一覧は所属しているものだけを返す(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)

	w := f.do(t, http.MethodGet, kbWorkspacesPath, "")

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got []kbWorkspaceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got, 1, "所属しているのは 1 つだけ")
	assert.Equal(t, kbWorkspaceSlug, got[0].Slug)
	assert.NotContains(t, w.Body.String(), kbOtherWorkspaceSlug,
		"所属していないワークスペースは 1 件も漏らさない")
	assert.NotContains(t, w.Body.String(), kbWorkspaceID, "内部 UUID は返さない")
}

func Test_ノートAPI_所属ワークスペース一覧は未認証なら401(t *testing.T) {
	f := newKbFixture(kbCanEdit, 0)

	w := f.do(t, http.MethodGet, kbWorkspacesPath, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_ノートAPI_所属ワークスペース一覧は0件でも空配列(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	// 所属を消す（principals の行が唯一の表現なので、消せば非メンバー）。
	f.perms.principals = map[string]*domain.Principal{}

	w := f.do(t, http.MethodGet, kbWorkspacesPath, "")

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[]`, w.Body.String(), "null ではなく空配列")
}

func Test_ノートAPI_ワークスペース作成は作成者をメンバーにする(t *testing.T) {
	const otherUser = uint64(777)
	f := newKbFixture(kbCanEdit, otherUser)

	w := f.do(t, http.MethodPost, kbWorkspacesPath, `{"slug":"new-team","name":"新チーム"}`)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var created kbWorkspaceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, "new-team", created.Slug)

	// 作成者が自分の作ったワークスペースに入れること。ここが崩れると
	// middleware が所属を確かめて 404 にするため、誰も入れないワークスペースが残る。
	listed := f.do(t, http.MethodGet, kbWorkspacesPath, "")
	require.Equal(t, http.StatusOK, listed.Code)
	assert.Contains(t, listed.Body.String(), "new-team", "作成者は所属一覧に出る")

	ws, err := f.pages.FindWorkspaceBySlug(t.Context(), "new-team")
	require.NoError(t, err)
	member, err := f.perms.IsWorkspaceMember(t.Context(), ws.ID, otherUser)
	require.NoError(t, err)
	assert.True(t, member, "作成者は principal を持つ（＝ メンバー）")

	facts, err := f.perms.WorkspacePermissionFactsForUser(t.Context(), ws.ID, otherUser)
	require.NoError(t, err)
	assert.True(t, domain.ResolveScopePermission(*facts).CanManage,
		"作成者は admin なので自分のワークスペースを設定できる")
}

func Test_ノートAPI_ワークスペース作成はslugの重複を409で断る(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)

	w := f.do(t, http.MethodPost, kbWorkspacesPath,
		`{"slug":"`+kbOtherWorkspaceSlug+`","name":"横取り"}`)

	assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	assert.JSONEq(t, `{"error":"slug_taken"}`, w.Body.String())
}

func Test_ノートAPI_ワークスペース作成は未認証なら401(t *testing.T) {
	f := newKbFixture(kbCanEdit, 0)

	w := f.do(t, http.MethodPost, kbWorkspacesPath, `{"slug":"new-team","name":"新チーム"}`)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_ノートAPI_ワークスペース作成は連打をレート制限で断る(t *testing.T) {
	// slug はテナントをまたいで一意で、取られた slug を取り返す口が無い。
	// 上限が無いと 1 人で短い slug を掴み取れてしまうので、作成だけは流量を絞る。
	f := newKbFixture(kbCanEdit, kbUserID)

	for i := range 5 {
		slug := "team-" + strconv.Itoa(i)
		w := f.do(t, http.MethodPost, kbWorkspacesPath, `{"slug":"`+slug+`","name":"新チーム"}`)
		require.Equal(t, http.StatusCreated, w.Code, "burst の範囲内: %s", w.Body.String())
	}

	over := f.do(t, http.MethodPost, kbWorkspacesPath, `{"slug":"team-over","name":"新チーム"}`)

	assert.Equal(t, http.StatusTooManyRequests, over.Code, over.Body.String())
	assert.Equal(t, "60", over.Header().Get("Retry-After"), "再試行の目安を返す")
	_, err := f.pages.FindWorkspaceBySlug(t.Context(), "team-over")
	assert.ErrorIs(t, err, repository.ErrWorkspaceNotFound, "断った要求は slug を取らない")

	// 一覧は絞らない（読みは掴み取りに使えないため）。
	assert.Equal(t, http.StatusOK, f.do(t, http.MethodGet, kbWorkspacesPath, "").Code)
}

// fake が本番より緩いと、本番では通らない作成要求で緑になるテストが書けてしまう。
// 実 PostgreSQL 側の対応する検証は knowledge_base_provision_integration_test.go にある。
func Test_ノートテスト用fake_存在しないワークスペースへのスペース作成を拒む(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)

	err := f.pages.CreateSpace(t.Context(), &domain.Space{
		WorkspaceID: "workspace-missing", Key: "eng", Name: "開発部",
	})

	assert.ErrorIs(t, err, repository.ErrWorkspaceNotFound)
}

func Test_ノートAPI_ワークスペース作成の失敗は500(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.provisioner.failWith = errors.New("db down")

	w := f.do(t, http.MethodPost, kbWorkspacesPath, `{"slug":"new-team","name":"新チーム"}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// スペース作成は「ワークスペース単位」の判定で、admin だけが通る。
func Test_ノートAPI_スペース作成はワークスペースのadminだけが通る(t *testing.T) {
	spacesPath := kbFill(kbSpacesPath, kbWorkspaceSlug, "")

	t.Run("admin なら作れる", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)

		w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`)

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var got kbSpaceResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, "eng", got.Key)
		assert.NotEmpty(t, got.ID, "以降の URL で使うのでスペース ID は返す")
	})

	t.Run("editor では作れない", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)

		w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`)

		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
	})

	t.Run("役割が無ければ作れない", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)

		w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`)

		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	})

	t.Run("スペースのadminではワークスペースの操作は通らない", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		// あるスペースの admin であっても、ワークスペース単位の判定には効かない。
		f.perms.setScopeRole(kbSpaceID, kbUserID, domain.GrantRoleAdmin)

		w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`)

		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	})

	t.Run("別ワークスペースのslugは404", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)

		w := f.do(t, http.MethodPost, kbFill(kbSpacesPath, kbOtherWorkspaceSlug, ""),
			`{"key":"eng","name":"開発部"}`)

		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("未認証は401", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, 0)

		w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func Test_ノートAPI_スペース作成はkeyの重複を409で断る(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)
	spacesPath := kbFill(kbSpacesPath, kbWorkspaceSlug, "")

	require.Equal(t, http.StatusCreated,
		f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`).Code)
	w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"別の部署"}`)

	assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	assert.JSONEq(t, `{"error":"space_key_taken"}`, w.Body.String())
}

func Test_ノートAPI_スペース作成は不正なkeyと長すぎる名前を400で断る(t *testing.T) {
	spacesPath := kbFill(kbSpacesPath, kbWorkspaceSlug, "")
	cases := []struct {
		name string
		body string
	}{
		{name: "key に大文字や記号", body: `{"key":"ENG!","name":"開発部"}`},
		{name: "key の先頭がハイフン", body: `{"key":"-eng","name":"開発部"}`},
		{name: "name が 201 文字", body: `{"key":"eng","name":"` + strings.Repeat("あ", 201) + `"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newKbFixture(kbCanEdit, kbUserID)
			f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)

			w := f.do(t, http.MethodPost, spacesPath, tc.body)

			assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		})
	}
}

// スペース直下へのページ作成（parentId 省略）は「スペースの編集権限」で判断する。
func Test_ノートAPI_スペース直下のページ作成はスペースの権限で判断する(t *testing.T) {
	pagesPath := "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/spaces/" + kbSpaceID + "/pages"

	t.Run("スペースのeditorならルートページを作れる", func(t *testing.T) {
		f := newKbFixture(kbNoPerm, kbUserID) // ページ単位の既定は「何もできない」
		f.perms.setScopeRole(kbSpaceID, kbUserID, domain.GrantRoleEditor)

		w := f.do(t, http.MethodPost, pagesPath, `{"title":"最初のページ"}`)

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var page kbPageResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &page))
		assert.Nil(t, page.ParentID, "スペース直下（ルート）に作る")
		assert.Equal(t, kbSpaceID, page.SpaceID)
	})

	t.Run("スペースのviewerでは作れない", func(t *testing.T) {
		f := newKbFixture(kbNoPerm, kbUserID)
		f.perms.setScopeRole(kbSpaceID, kbUserID, domain.GrantRoleViewer)

		w := f.do(t, http.MethodPost, pagesPath, `{"title":"最初のページ"}`)

		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
	})

	t.Run("役割が無ければスペースの実在を漏らさず404", func(t *testing.T) {
		f := newKbFixture(kbNoPerm, kbUserID)

		w := f.do(t, http.MethodPost, pagesPath, `{"title":"最初のページ"}`)

		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
	})

	t.Run("存在しないスペースは404", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)
		missing := "/api/v2/kb/workspaces/" + kbWorkspaceSlug +
			"/spaces/0198a000-0000-7000-8000-00000000dead/pages"

		w := f.do(t, http.MethodPost, missing, `{"title":"最初のページ"}`)

		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("親を指定した作成はスペースの権限では通らない", func(t *testing.T) {
		// スペースでは editor だが、親ページは閲覧すらできない。
		// スペース単位の判定でページ作成を通してしまうと、この経路が開く。
		f := newKbFixture(kbNoPerm, kbUserID)
		f.perms.setScopeRole(kbSpaceID, kbUserID, domain.GrantRoleEditor)

		w := f.do(t, http.MethodPost, pagesPath, `{"parentId":"`+kbRootPageID+`","title":"子"}`)

		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})

	t.Run("別ワークスペースのslugでは作れない", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbSpaceID, kbUserID, domain.GrantRoleAdmin)
		other := "/api/v2/kb/workspaces/" + kbOtherWorkspaceSlug + "/spaces/" + kbSpaceID + "/pages"

		w := f.do(t, http.MethodPost, other, `{"title":"最初のページ"}`)

		assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
	})
}

// kbSecondSpaceID はスペース一覧のテストで使う 2 つ目のスペース。
// 1 つしか無いと「権限のあるものだけを返す」と「全部返す」が同じ結果になり、
// ふるいを外しても緑のままになる。
const kbSecondSpaceID = "0198a000-0000-7000-8000-0000000000a2"

// プライベートスペース: 自分の区画が増えるだけなので、admin でないメンバーでも作れる。
// チームスペース（省略時）は今までどおり admin だけ（上のテスト）。この非対称が仕様。
func Test_ノートAPI_プライベートスペースはメンバーなら作れる(t *testing.T) {
	spacesPath := kbFill(kbSpacesPath, kbWorkspaceSlug, "")

	t.Run("editor でも private なら作れて、応答に visibility が載る", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)

		w := f.do(t, http.MethodPost, spacesPath, `{"name":"自分のメモ","visibility":"private"}`)

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var got kbSpaceResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, "private", got.Visibility)
		assert.NotEmpty(t, got.Key, "key は自動採番される")

		// 作成者の一覧に出る（provisioner が作成者へ grant を張っている）。
		// ワークスペース既定（editor）は private に届かないので、
		// この grant が無ければ作った本人にも見えない。
		_, list := kbListSpaces(t, f, kbWorkspaceSlug)
		found := false
		for _, s := range list {
			if s.ID == got.ID {
				found = true
				assert.Equal(t, "private", s.Visibility, "一覧の応答にも visibility が載る（節分けの判定材料）")
			}
		}
		assert.True(t, found, "作った本人の一覧に出る")
	})

	t.Run("private では key を指定できない（409 から他人のスペースの実在を読ませない）", func(t *testing.T) {
		// key はチームとプライベートで同じ名前空間。明示指定を許すと「409 が返るか」で
		// 一覧にも出ないはずの他人のプライベートスペースの実在を言い当てられる。
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)

		w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"探り","visibility":"private"}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"invalid_request"}`, w.Body.String())
	})

	t.Run("チームスペースでは今までどおり key を指定できる（admin だけが到達する）", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)

		w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`)

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	})

	t.Run("visibility が未知の値なら 400", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)

		w := f.do(t, http.MethodPost, spacesPath, `{"name":"x","visibility":"secret"}`)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("チームスペースの応答は visibility=workspace", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)

		w := f.do(t, http.MethodPost, spacesPath, `{"name":"開発部"}`)

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var got kbSpaceResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, "workspace", got.Visibility)
	})
}

// 会社のワークスペースには、その会社のメンバーが自動で入る。
//
// 会社ごとのワークスペースは起動時のバックフィルが用意するが、ノートの所属
// （principals の行）は作成者にしか無かった。そのため同じ会社の他のメンバーは
// 一覧にも出ず、URL を叩いても 404 になっていた（実際に踏んだ形の回帰）。
func Test_ノートAPI_会社のワークスペースには同じ会社のメンバーが自動で入る(t *testing.T) {
	t.Run("一覧を開くと所属が用意され、会社のワークスペースが出る", func(t *testing.T) {
		const newcomer = uint64(777)
		f := newKbFixture(kbCanEdit, newcomer)
		// この人はまだ principals の行を持たない（＝ 非メンバー）が、会社は同じ。
		require.Nil(t, f.perms.userPrincipal(kbWorkspaceID, newcomer), "前提: まだ非メンバー")
		f.perms.setCompanyWorkspace(newcomer, kbWorkspaceID)

		w := f.do(t, http.MethodGet, "/api/v2/kb/workspaces", "")

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got []kbWorkspaceResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		require.Len(t, got, 1)
		assert.Equal(t, kbWorkspaceSlug, got[0].Slug)
		assert.NotNil(t, f.perms.userPrincipal(kbWorkspaceID, newcomer), "所属が用意される")
	})

	t.Run("URL を直に開いても入れる（一覧を経由しない経路）", func(t *testing.T) {
		const newcomer = uint64(778)
		f := newKbFixture(kbCanEdit, newcomer)
		f.perms.setCompanyWorkspace(newcomer, kbWorkspaceID)

		w := f.do(t, http.MethodGet, kbFill(kbSpacesPath, kbWorkspaceSlug, ""), "")

		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.NotNil(t, f.perms.userPrincipal(kbWorkspaceID, newcomer))
	})

	t.Run("別の会社のワークスペースには入らない", func(t *testing.T) {
		const outsider = uint64(779)
		f := newKbFixture(kbCanEdit, outsider)
		// 会社のワークスペースは別のもの。URL を知っていても入れない。
		f.perms.setCompanyWorkspace(outsider, kbOtherWorkspaceID)

		w := f.do(t, http.MethodGet, kbFill(kbSpacesPath, kbWorkspaceSlug, ""), "")

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Nil(t, f.perms.userPrincipal(kbWorkspaceID, outsider), "所属は作られない")
	})

	t.Run("会社に属さないユーザーの一覧は空で、失敗にしない", func(t *testing.T) {
		const staff = uint64(780)
		f := newKbFixture(kbCanEdit, staff)
		// 会社の紐づけを設定しない（運営管理者のように会社を持たない人）。

		w := f.do(t, http.MethodGet, "/api/v2/kb/workspaces", "")

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.JSONEq(t, `[]`, w.Body.String())
	})

	t.Run("取り消した役割は一覧を開いても戻らない", func(t *testing.T) {
		// 役割の取り消しは grant の行を消すだけで、主体の行は残る。自動所属が
		// 「主体があっても役割を足す」作りだと、admin が取り消した権限が
		// その人の次の読み取りで戻ってしまい、権限管理が効かなくなる。
		f := newKbFixture(kbCanEdit, kbUserID)
		principal, err := f.perms.EnsureUserPrincipal(context.Background(), kbWorkspaceID, kbUserID)
		require.NoError(t, err)
		require.NoError(t, f.perms.GrantWorkspaceRoleIfAbsent(
			context.Background(), kbWorkspaceID, principal.ID, domain.GrantRoleEditor,
		))
		// admin が役割を取り消す（主体は残る）。
		require.NoError(t, f.perms.DeleteWorkspaceGrant(
			context.Background(), kbWorkspaceID, principal.ID,
		))
		f.perms.setCompanyWorkspace(kbUserID, kbWorkspaceID)

		w := f.do(t, http.MethodGet, "/api/v2/kb/workspaces", "")

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		facts, err := f.perms.WorkspacePermissionFactsForUser(
			context.Background(), kbWorkspaceID, kbUserID,
		)
		require.NoError(t, err)
		assert.Empty(t, facts.Roles, "取り消した役割が読み取りで戻ってはいけない")
	})

	t.Run("既にある役割は踏み潰さない（admin を editor へ落とさない）", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		// 本番と同じ形で admin の grant 行を張る（実効権限の写しではなく行を作る）。
		principal, err := f.perms.EnsureUserPrincipal(context.Background(), kbWorkspaceID, kbUserID)
		require.NoError(t, err)
		require.NoError(t, f.perms.GrantWorkspaceRoleIfAbsent(
			context.Background(), kbWorkspaceID, principal.ID, domain.GrantRoleAdmin,
		))
		f.perms.setCompanyWorkspace(kbUserID, kbWorkspaceID)

		w := f.do(t, http.MethodGet, "/api/v2/kb/workspaces", "")
		require.Equal(t, http.StatusOK, w.Code)

		// admin のままなのでチームスペースを作れる。
		created := f.do(t, http.MethodPost, kbFill(kbSpacesPath, kbWorkspaceSlug, ""),
			`{"name":"開発部"}`)
		assert.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	})
}

// kbListSpaces はスペース一覧を叩いて応答をデコードする。
func kbListSpaces(t *testing.T, f kbFixture, slug string) (*httptest.ResponseRecorder, []kbSpaceResponse) {
	t.Helper()
	w := f.do(t, http.MethodGet, kbFill(kbSpacesPath, slug, ""), "")
	if w.Code != http.StatusOK {
		return w, nil
	}
	var got []kbSpaceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	return w, got
}

func Test_ノートAPI_スペース一覧は閲覧できるスペースだけを返す(t *testing.T) {
	// スペースは「誰に何を見せるか」を分ける入れ物なので、key と name が並ぶだけでも
	// 中で何が進んでいるかが伝わる。役割が届いていないスペースは 1 件も出さない。
	roles := []domain.GrantRole{
		domain.GrantRoleViewer, domain.GrantRoleCommenter,
		domain.GrantRoleEditor, domain.GrantRoleAdmin,
	}
	for _, role := range roles {
		t.Run(string(role)+"はそのスペースだけ見える", func(t *testing.T) {
			f := newKbFixture(kbCanEdit, kbUserID)
			f.pages.addSpace(kbWorkspaceID, kbSecondSpaceID)
			// 役割はスペース単位で 1 つ目にだけ張る（2 つ目には何も届かない）。
			f.perms.setScopeRole(kbSpaceID, kbUserID, role)

			w, got := kbListSpaces(t, f, kbWorkspaceSlug)

			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			require.Len(t, got, 1, "役割が届いているスペースだけ")
			assert.Equal(t, kbSpaceID, got[0].ID)
			assert.NotContains(t, w.Body.String(), kbSecondSpaceID,
				"閲覧権限の無いスペースは ID も key も漏らさない")
		})
	}

	t.Run("役割が1つも無ければ1件も返らない", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.pages.addSpace(kbWorkspaceID, kbSecondSpaceID)

		w, got := kbListSpaces(t, f, kbWorkspaceSlug)

		require.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, got, "所属しているだけでは中身は見えない")
		assert.JSONEq(t, `[]`, w.Body.String(), "null ではなく空配列")
	})

	t.Run("ワークスペース全体の役割は配下の全スペースへ届く", func(t *testing.T) {
		f := newKbFixture(kbCanEdit, kbUserID)
		f.pages.addSpace(kbWorkspaceID, kbSecondSpaceID)
		f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleViewer)

		_, got := kbListSpaces(t, f, kbWorkspaceSlug)

		require.Len(t, got, 2, "ワークスペースの grant はスペースを選ばない")
	})
}

func Test_ノートAPI_スペース一覧はスペースが0件でも空配列(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)
	f.pages.spaces = map[string]*domain.Space{}

	w, _ := kbListSpaces(t, f, kbWorkspaceSlug)

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[]`, w.Body.String(),
		"null を返すとフロントの .map が TypeError で落ちる")
}

func Test_ノートAPI_スペース一覧は存在しないワークスペースと権限の無いワークスペースを区別できない(t *testing.T) {
	// slug の総当たりで他社テナントの実在が分からないこと。判定は middleware にあり、
	// この口は「メンバーであること」を前提に動く。
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)

	unknown, _ := kbListSpaces(t, f, "no-such-workspace")
	foreign, _ := kbListSpaces(t, f, kbOtherWorkspaceSlug)

	assert.Equal(t, http.StatusNotFound, unknown.Code)
	assert.Equal(t, unknown.Code, foreign.Code)
	assert.Equal(t, unknown.Body.String(), foreign.Body.String())
}

func Test_ノートAPI_スペース一覧は未認証なら401(t *testing.T) {
	f := newKbFixture(kbCanEdit, 0)

	w, _ := kbListSpaces(t, f, kbWorkspaceSlug)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_ノートAPI_スペース一覧は事実の収集に失敗したら500(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.scopeFactsErr = errors.New("db down")

	w, _ := kbListSpaces(t, f, kbWorkspaceSlug)

	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"確かめられないなら見せない（空配列で「無い」と答えない）")
}
