package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ページの雛形 API（FRESTYLE-435 段5）の handler テスト。
//
// 判定の軸がエンドポイントごとに違う（一覧=所属のみ、保存・削除=ワークスペース全体の
// CanEdit、使用=既存のページ作成と同じ分岐）ため、kbEndpoints のような単一の表には乗せず
// 個別に検証する（kb_page_handler_test.go の登録漏れ検査には手で足してある）。

const (
	kbTemplatesListPath          = "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/templates"
	kbTemplateCreatePathBase     = "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/pages/"
	kbTemplateDeletePathBase     = "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/templates/"
	kbCreatePageFromTemplatePath = "/api/v2/kb/workspaces/" + kbWorkspaceSlug + "/spaces/" + kbSpaceID + "/pages/from-template"
)

// Test_雛形API_一覧は所属していれば読める は、一覧がワークスペース全体の CanEdit を
// 要求しない（所属者なら誰でも読める）ことを固定する。
func Test_雛形API_一覧は所属していれば読める(t *testing.T) {
	// kbNoPerm はページ 1 枚の権限が空でも、一覧はページの権限を一切見ないので通る。
	f := newKbFixture(kbNoPerm, kbUserID)

	w := f.do(t, http.MethodGet, kbTemplatesListPath, "")

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.JSONEq(t, `[]`, w.Body.String(), "0件でもnullではなく空配列")
}

// Test_雛形API_一覧は非所属なら404 は、ワークスペースに属さないユーザーが
// middleware.KnowledgeBaseWorkspace の段で 404 に落ちることを固定する
// （他の全経路と同じ畳み方）。
func Test_雛形API_一覧は非所属なら404(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbOutsiderUserID)

	w := f.do(t, http.MethodGet, kbTemplatesListPath, "")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
}

// Test_雛形API_保存にはページ編集権限が要る は、対象ページを編集できない相手が
// 403 になることを固定する（閲覧はできるので実在は伏せない）。
func Test_雛形API_保存にはページ編集権限が要る(t *testing.T) {
	f := newKbFixture(kbCanView, kbUserID)
	// ワークスペース全体の CanEdit は満たしておく（もう一段の検査が独立して効いていることを
	// 示すため）。ただし workspace 既定の役割は private スペースには届かない
	// （kbFakePerms.rolesAt の doc 参照）ので、対象ページを private スペースへ移し、
	// このページの権限は fallback（kbCanView = CanView のみ）だけで決まる状態にする。
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)
	f.perms.hideInOwnPrivateSpace(kbWorkspaceID, kbChildPageID)

	w := f.do(t, http.MethodPost, kbTemplateCreatePathBase+kbChildPageID+"/templates", `{"name":"議事録"}`)

	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
}

// Test_雛形API_保存にはワークスペース全体のCanEditも要る は、対象ページは編集できても
// ワークスペース全体への書き込み資格が無ければ 403 になることを固定する。
func Test_雛形API_保存にはワークスペース全体のCanEditも要る(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	// ワークスペース全体の役割は張らない（既定は無権限）。

	w := f.do(t, http.MethodPost, kbTemplateCreatePathBase+kbChildPageID+"/templates", `{"name":"議事録"}`)

	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
}

// Test_雛形API_保存は両方満たせば成功する は、ページ編集権限とワークスペース全体の
// CanEdit の両方を満たしたときに 201 で雛形（軽量な形）が返ることを固定する。
func Test_雛形API_保存は両方満たせば成功する(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)

	w := f.do(t, http.MethodPost, kbTemplateCreatePathBase+kbChildPageID+"/templates",
		`{"name":"議事録","spaceId":"`+kbSpaceID+`"}`)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var got kbPageTemplateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "議事録", got.Name)
	require.NotNil(t, got.SpaceID)
	assert.Equal(t, kbSpaceID, *got.SpaceID)
	assert.NotEmpty(t, got.ID)
}

// Test_雛形API_保存で不正な名前は400 は、空白のみの名前が ErrInvalidTemplateName →
// invalid_template_name に落ちることを固定する。
func Test_雛形API_保存で不正な名前は400(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)

	w := f.do(t, http.MethodPost, kbTemplateCreatePathBase+kbChildPageID+"/templates", `{"name":"   "}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid_template_name"}`, w.Body.String())
}

// Test_雛形API_保存で名前が重複すれば409 は、同じワークスペース内で同名の雛形を
// 2回保存しようとすると duplicate_template_name になることを固定する。
func Test_雛形API_保存で名前が重複すれば409(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)
	body := `{"name":"議事録"}`
	first := f.do(t, http.MethodPost, kbTemplateCreatePathBase+kbChildPageID+"/templates", body)
	require.Equal(t, http.StatusCreated, first.Code, first.Body.String())

	w := f.do(t, http.MethodPost, kbTemplateCreatePathBase+kbChildPageID+"/templates", body)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.JSONEq(t, `{"error":"duplicate_template_name"}`, w.Body.String())
}

// Test_雛形API_保存で存在しないspaceIdは404 は、指定した spaceId が同じワークスペース内に
// 実在しなければ not_found になることを固定する。
func Test_雛形API_保存で存在しないspaceIdは404(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)
	missingSpace := "0198a000-0000-7000-8000-0000000000ff"

	w := f.do(t, http.MethodPost, kbTemplateCreatePathBase+kbChildPageID+"/templates",
		`{"name":"議事録","spaceId":"`+missingSpace+`"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
}

// Test_雛形API_削除にはワークスペース全体のCanEditが要る は、削除がページ単位の権限を
// 見ず、ワークスペース全体の CanEdit だけで判定されることを固定する。
func Test_雛形API_削除にはワークスペース全体のCanEditが要る(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)
	created := f.do(t, http.MethodPost, kbTemplateCreatePathBase+kbChildPageID+"/templates", `{"name":"削除対象"}`)
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	var tpl kbPageTemplateResponse
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &tpl))

	t.Run("CanEditが無ければ403", func(t *testing.T) {
		f2 := newKbFixture(kbCanEdit, kbUserID)
		// このユーザーには workspace 全体の役割を張らない。
		w := f2.do(t, http.MethodDelete, kbTemplateDeletePathBase+tpl.ID, "")
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("CanEditがあれば204で削除される", func(t *testing.T) {
		w := f.do(t, http.MethodDelete, kbTemplateDeletePathBase+tpl.ID, "")
		require.Equal(t, http.StatusNoContent, w.Code)

		_, err := f.templates.Get(t.Context(), kbWorkspaceID, tpl.ID)
		require.ErrorIs(t, err, domain.ErrPageTemplateNotFound)
	})
}

// Test_雛形API_削除で存在しないidは404 は domain.ErrPageTemplateNotFound の応答を固定する。
func Test_雛形API_削除で存在しないidは404(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbWorkspaceID, kbUserID, domain.GrantRoleEditor)

	w := f.do(t, http.MethodDelete, kbTemplateDeletePathBase+"00000000-0000-7000-8000-000000000000", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
}

// Test_雛形API_使用は親ページの編集権限で判定する は、parentId を指定したときの認可が
// 既存のページ作成（Create）の分岐と同じ（親ページの CapabilityEdit）であることを固定する。
// 雛形そのものへの追加の権限は要らない。
func Test_雛形API_使用は親ページの編集権限で判定する(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	tpl := &domain.PageTemplate{WorkspaceID: kbWorkspaceID, Name: "テンプレ", Doc: kbValidDoc, CreatedByUserID: 1}
	require.NoError(t, f.templates.Create(t.Context(), tpl))

	w := f.do(t, http.MethodPost, kbCreatePageFromTemplatePath,
		`{"templateId":"`+tpl.ID+`","parentId":"`+kbChildPageID+`","title":"新しい議事録"}`)

	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	var page kbPageResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &page))
	assert.Equal(t, "新しい議事録", page.Title)
	require.NotNil(t, page.ParentID)
	assert.Equal(t, kbChildPageID, *page.ParentID)
}

// Test_雛形API_使用は親を省くとスペースの編集権限で判定する は、parentId を省いたときの
// 認可が既存のページ作成の分岐と同じ（そのスペースの CapabilityEdit）であることを固定する。
func Test_雛形API_使用は親を省くとスペースの編集権限で判定する(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	tpl := &domain.PageTemplate{WorkspaceID: kbWorkspaceID, Name: "テンプレ", Doc: kbValidDoc, CreatedByUserID: 1}
	require.NoError(t, f.templates.Create(t.Context(), tpl))

	t.Run("スペースを編集できなければ403", func(t *testing.T) {
		// 役割を一切張らない（CanView も false）と、実在を隠す 404 に畳まれてしまう
		// （requireSpacePermissionWith の doc 参照）。閲覧はできるが編集はできない、という
		// 403 を確かめたいので viewer を明示的に張る（kb_page_handler_test.go の
		// deniedCases と同じ形）。
		f.perms.setScopeRole(kbSpaceID, kbUserID, domain.GrantRoleViewer)
		w := f.do(t, http.MethodPost, kbCreatePageFromTemplatePath,
			`{"templateId":"`+tpl.ID+`","title":"新しい議事録"}`)
		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	})

	t.Run("スペースを編集できれば201", func(t *testing.T) {
		f.perms.setScopeRole(kbSpaceID, kbUserID, domain.GrantRoleEditor)
		w := f.do(t, http.MethodPost, kbCreatePageFromTemplatePath,
			`{"templateId":"`+tpl.ID+`","title":"新しい議事録"}`)
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var page kbPageResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &page))
		assert.Nil(t, page.ParentID, "スペース直下に作られる")
	})
}

// Test_雛形API_使用で存在しない雛形は404 は domain.ErrPageTemplateNotFound の応答を固定する。
func Test_雛形API_使用で存在しない雛形は404(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbUserID)
	f.perms.setScopeRole(kbSpaceID, kbUserID, domain.GrantRoleEditor)

	w := f.do(t, http.MethodPost, kbCreatePageFromTemplatePath,
		`{"templateId":"00000000-0000-7000-8000-000000000000","title":"新しい議事録"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())
}
