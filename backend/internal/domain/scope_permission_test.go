package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ScopeFacts → ScopePermission の写像を役割ごとに固定する。
// 入れ物（ワークスペース / スペース）には例外の層が無いので、答えは既定の役割だけで決まる。
func Test_入れ物の実効権限は役割から決まる(t *testing.T) {
	cases := []struct {
		name      string
		roles     []domain.GrantRole
		canView   bool
		canEdit   bool
		canManage bool
	}{
		{name: "役割が無ければ何もできない"},
		{name: "viewer は閲覧だけ", roles: []domain.GrantRole{domain.GrantRoleViewer}, canView: true},
		{name: "commenter は閲覧だけ", roles: []domain.GrantRole{domain.GrantRoleCommenter}, canView: true},
		{
			name: "editor は閲覧と編集", roles: []domain.GrantRole{domain.GrantRoleEditor},
			canView: true, canEdit: true,
		},
		{
			name: "admin は構成も変えられる", roles: []domain.GrantRole{domain.GrantRoleAdmin},
			canView: true, canEdit: true, canManage: true,
		},
		{
			name:    "未知の役割は数えない",
			roles:   []domain.GrantRole{domain.GrantRole("owner"), domain.GrantRole("")},
			canView: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.ResolveScopePermission(domain.ScopeFacts{Roles: tc.roles})

			assert.Equal(t, tc.canView, got.CanView)
			assert.Equal(t, tc.canEdit, got.CanEdit)
			assert.Equal(t, tc.canManage, got.CanManage)
		})
	}
}

// 複数の経路（自分 / 所属グループ / スペース全員 / ワークスペースの grant）で役割を得たときは
// 最も強いものを採る。弱い方を採ると、スペースに viewer を 1 つ張るだけで
// ワークスペース管理者を締め出せてしまう。
func Test_入れ物の実効権限は最も強い役割を採り並び順に依存しない(t *testing.T) {
	forward := domain.ResolveScopePermission(domain.ScopeFacts{Roles: []domain.GrantRole{
		domain.GrantRoleViewer, domain.GrantRoleAdmin, domain.GrantRoleCommenter,
	}})
	backward := domain.ResolveScopePermission(domain.ScopeFacts{Roles: []domain.GrantRole{
		domain.GrantRoleCommenter, domain.GrantRoleAdmin, domain.GrantRoleViewer,
	}})

	assert.True(t, forward.CanManage, "admin が混ざっていれば admin として扱う")
	assert.Equal(t, forward, backward, "grant を張った順で答えが変わってはいけない")

	// 未知の値が混ざっても、既知の中で最も強いものが残る。
	withUnknown := domain.ResolveScopePermission(domain.ScopeFacts{Roles: []domain.GrantRole{
		domain.GrantRole("root"), domain.GrantRoleEditor,
	}})
	assert.True(t, withUnknown.CanEdit)
	assert.False(t, withUnknown.CanManage)
}

func Test_最も強い役割を返す(t *testing.T) {
	assert.Nil(t, domain.StrongestGrantRole(nil), "1 つも無ければ nil（最弱の役割と型で区別する）")
	assert.Nil(t, domain.StrongestGrantRole([]domain.GrantRole{domain.GrantRole("owner")}),
		"未知の値だけなら「役割が無い」と同じ")

	got := domain.StrongestGrantRole([]domain.GrantRole{domain.GrantRoleViewer, domain.GrantRoleEditor})
	require.NotNil(t, got)
	assert.Equal(t, domain.GrantRoleEditor, *got)
}

// SQL が強さ（整数）で返す経路と、役割の集合を畳む経路が同じ答えになることを型の上で固定する。
// ページ 1 枚の解決は GrantRoleByRank、入れ物の解決は StrongestGrantRole を通るので、
// 両者の対応が崩れると同じ既定に対して経路ごとに違う答えが出る。
func Test_強さと役割の対応は両経路で一致する(t *testing.T) {
	for _, role := range domain.ValidGrantRoles {
		t.Run(string(role), func(t *testing.T) {
			byRank := domain.GrantRoleByRank(role.Rank())
			require.NotNil(t, byRank)
			assert.Equal(t, role, *byRank)

			strongest := domain.StrongestGrantRole([]domain.GrantRole{role})
			require.NotNil(t, strongest)
			assert.Equal(t, *byRank, *strongest)
		})
	}
	assert.Nil(t, domain.GrantRoleByRank(0), "0 は「grant が無い」を表す")
}

func Test_入れ物の実効権限のAllows(t *testing.T) {
	perm := domain.ResolveScopePermission(domain.ScopeFacts{
		Roles: []domain.GrantRole{domain.GrantRoleCommenter},
	})

	assert.True(t, perm.Allows(domain.CapabilityView))
	assert.False(t, perm.Allows(domain.CapabilityEdit))
}
