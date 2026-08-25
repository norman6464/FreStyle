package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ワークスペース / スペースの API は判定対象がページではないので、kbEndpoints の表
// （ページ 1 枚の権限を軸に回す）とは別にここで検証する。

func Test_ナレッジ基盤API_所属ワークスペース一覧は所属しているものだけを返す(t *testing.T) {
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

func Test_ナレッジ基盤API_所属ワークスペース一覧は未認証なら401(t *testing.T) {
	f := newKbFixture(kbCanEdit, 0)

	w := f.do(t, http.MethodGet, kbWorkspacesPath, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_ナレッジ基盤API_所属ワークスペース一覧は0件でも空配列(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	// 所属を消す（principals の行が唯一の表現なので、消せば非メンバー）。
	f.perms.principals = map[string]*domain.Principal{}

	w := f.do(t, http.MethodGet, kbWorkspacesPath, "")

	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `[]`, w.Body.String(), "null ではなく空配列")
}

func Test_ナレッジ基盤API_ワークスペース作成は作成者をメンバーにする(t *testing.T) {
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

func Test_ナレッジ基盤API_ワークスペース作成はslugの重複を409で断る(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)

	w := f.do(t, http.MethodPost, kbWorkspacesPath,
		`{"slug":"`+kbOtherWorkspaceSlug+`","name":"横取り"}`)

	assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	assert.JSONEq(t, `{"error":"slug_taken"}`, w.Body.String())
}

func Test_ナレッジ基盤API_ワークスペース作成は未認証なら401(t *testing.T) {
	f := newKbFixture(kbCanEdit, 0)

	w := f.do(t, http.MethodPost, kbWorkspacesPath, `{"slug":"new-team","name":"新チーム"}`)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_ナレッジ基盤API_ワークスペース作成の失敗は500(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.provisioner.failWith = errors.New("db down")

	w := f.do(t, http.MethodPost, kbWorkspacesPath, `{"slug":"new-team","name":"新チーム"}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// スペース作成は「ワークスペース単位」の判定で、admin だけが通る。
func Test_ナレッジ基盤API_スペース作成はワークスペースのadminだけが通る(t *testing.T) {
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

func Test_ナレッジ基盤API_スペース作成はkeyの重複を409で断る(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleAdmin)
	spacesPath := kbFill(kbSpacesPath, kbWorkspaceSlug, "")

	require.Equal(t, http.StatusCreated,
		f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`).Code)
	w := f.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"別の部署"}`)

	assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	assert.JSONEq(t, `{"error":"space_key_taken"}`, w.Body.String())
}

func Test_ナレッジ基盤API_スペース作成は不正なkeyと長すぎる名前を400で断る(t *testing.T) {
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
func Test_ナレッジ基盤API_スペース直下のページ作成はスペースの権限で判断する(t *testing.T) {
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
