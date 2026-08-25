package usecase_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kbTreePages は root → child → grandchild と、独立した sibling の 4 ページ。
func kbTreePages() []domain.Page {
	root := "p-root"
	child := "p-child"
	return []domain.Page{
		{ID: root, Position: "a0", Title: "root"},
		{ID: child, Position: "a1", Title: "child", ParentID: &root},
		{ID: "p-grandchild", Position: "a2", Title: "grandchild", ParentID: &child},
		{ID: "p-sibling", Position: "a3", Title: "sibling"},
	}
}

func Test_ツリー組み立て_親子が復元される(t *testing.T) {
	roots := usecase.BuildPageTree(kbTreePages(), usecase.PageTreeOrphanHidden)

	require.Len(t, roots, 2)
	assert.Equal(t, "p-root", roots[0].Page.ID)
	require.Len(t, roots[0].Children, 1)
	assert.Equal(t, "p-child", roots[0].Children[0].Page.ID)
	require.Len(t, roots[0].Children[0].Children, 1)
	assert.Equal(t, "p-grandchild", roots[0].Children[0].Children[0].Page.ID)
	assert.Equal(t, "p-sibling", roots[1].Page.ID)
}

func Test_ツリー組み立て_見えない親の子孫はまとめて落ちる(t *testing.T) {
	// 権限のふるいで root が落ちた一覧を再現する。
	pages := kbTreePages()[1:]

	roots := usecase.BuildPageTree(pages, usecase.PageTreeOrphanHidden)

	require.Len(t, roots, 1, "見えない親の子は根に昇格させない（配下の存在も漏らさない）")
	assert.Equal(t, "p-sibling", roots[0].Page.ID)
}

func Test_ツリー組み立て_ふるいにかけない一覧では親無しを根として見せる(t *testing.T) {
	pages := kbTreePages()[1:]

	roots := usecase.BuildPageTree(pages, usecase.PageTreeOrphanAsRoot)

	require.Len(t, roots, 2, "整合が崩れた行はデータを隠さずルート扱いで見せる")
	assert.Equal(t, "p-child", roots[0].Page.ID)
	require.Len(t, roots[0].Children, 1)
	assert.Equal(t, "p-sibling", roots[1].Page.ID)
}

func Test_ツリー組み立て_空の一覧は空のスライス(t *testing.T) {
	roots := usecase.BuildPageTree(nil, usecase.PageTreeOrphanHidden)
	assert.NotNil(t, roots, "nil ではなく空スライス（JSON で null にしない）")
	assert.Empty(t, roots)
}

func Test_ツリー組み立て_兄弟の並びは入力順のまま(t *testing.T) {
	parent := "p-parent"
	pages := []domain.Page{
		{ID: parent, Position: "a0"},
		{ID: "p-1", Position: "a1", ParentID: &parent},
		{ID: "p-2", Position: "a2", ParentID: &parent},
		{ID: "p-3", Position: "a3", ParentID: &parent},
	}

	roots := usecase.BuildPageTree(pages, usecase.PageTreeOrphanHidden)

	require.Len(t, roots, 1)
	got := make([]string, 0, 3)
	for _, c := range roots[0].Children {
		got = append(got, c.Page.ID)
	}
	assert.Equal(t, []string{"p-1", "p-2", "p-3"}, got)
}
