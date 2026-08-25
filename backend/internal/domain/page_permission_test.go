package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func role(r domain.GrantRole) *domain.GrantRole { return &r }

func capability(c domain.Capability) *domain.Capability { return &c }

// exception は「最も近い制限の段」の集計を組み立てる小道具。
func exception(denied, allowed, hasAllowList bool) *domain.RestrictionFacts {
	return &domain.RestrictionFacts{Denied: denied, Allowed: allowed, HasAllowList: hasAllowList}
}

func Test_実効権限_例外が無ければ既定の役割どおり(t *testing.T) {
	cases := []struct {
		name    string
		facts   domain.PagePermissionFacts
		canView bool
		canEdit bool
	}{
		{"grant が無ければ何もできない", domain.PagePermissionFacts{Member: true}, false, false},
		{"viewer は閲覧のみ", domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleViewer)}, true, false},
		{"commenter は閲覧のみ（編集は不可）", domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleCommenter)}, true, false},
		{"editor は閲覧と編集", domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleEditor)}, true, true},
		{"admin は閲覧と編集", domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleAdmin)}, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.ResolvePagePermission(tc.facts)
			assert.Equal(t, tc.canView, got.CanView, "閲覧")
			assert.Equal(t, tc.canEdit, got.CanEdit, "編集")
		})
	}
}

func Test_実効権限_例外の優先規則(t *testing.T) {
	base := func(view *domain.RestrictionFacts) domain.PagePermissionFacts {
		return domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleEditor), View: view}
	}
	cases := []struct {
		name    string
		view    *domain.RestrictionFacts
		canView bool
	}{
		{"制限の段が無ければ既定に従う", nil, true},
		{"自分が deny されていれば不許可", exception(true, false, false), false},
		{"deny と allow の両方に当たったら deny が勝つ", exception(true, true, true), false},
		{"自分が allow されていれば許可", exception(false, true, true), true},
		{"allow リストがあり自分が載っていなければ不許可", exception(false, false, true), false},
		{"deny だけの段で名指しされていなければ既定に戻る", exception(false, false, false), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.ResolvePagePermission(base(tc.view))
			assert.Equal(t, tc.canView, got.CanView)
		})
	}
}

func Test_実効権限_deny_だけの段は既定を弱めない(t *testing.T) {
	// 「この人だけ編集させない」を張ったとき、名指しされていない人の既定（editor）は変わらない。
	denied := domain.ResolvePagePermission(domain.PagePermissionFacts{
		Member: true, Role: role(domain.GrantRoleEditor),
		Edit: exception(true, false, false),
	})
	assert.True(t, denied.CanView, "閲覧の制限は張っていないので見られる")
	assert.False(t, denied.CanEdit, "名指しされた本人は編集できない")

	other := domain.ResolvePagePermission(domain.PagePermissionFacts{
		Member: true, Role: role(domain.GrantRoleEditor),
		Edit: exception(false, false, false),
	})
	assert.True(t, other.CanEdit, "名指しされていない人は既定のまま編集できる")
}

func Test_実効権限_allow_を1つ足すと限定公開に切り替わる(t *testing.T) {
	listed := domain.ResolvePagePermission(domain.PagePermissionFacts{
		Member: true, Role: role(domain.GrantRoleEditor),
		View: exception(false, true, true),
	})
	assert.True(t, listed.CanView, "許可リストに載っている人は見られる")

	unlisted := domain.ResolvePagePermission(domain.PagePermissionFacts{
		Member: true, Role: role(domain.GrantRoleEditor),
		View: exception(false, false, true),
	})
	assert.False(t, unlisted.CanView, "載っていない人は既定が editor でも締め出される")
}

func Test_実効権限_編集は閲覧を含む(t *testing.T) {
	// 閲覧を deny されたページは、編集の既定が editor でも編集できない
	// （編集経路から中身を読めてしまう穴を作らない）。
	got := domain.ResolvePagePermission(domain.PagePermissionFacts{
		Member: true, Role: role(domain.GrantRoleEditor),
		View: exception(true, false, false),
	})
	assert.False(t, got.CanView)
	assert.False(t, got.CanEdit, "閲覧できないページは編集もできない")
}

func Test_実効権限_閲覧と編集は別々の段で解決する(t *testing.T) {
	// 閲覧は allow リストに載って許可、編集はさらに近い段で deny、という組み合わせ。
	got := domain.ResolvePagePermission(domain.PagePermissionFacts{
		Member: true, Role: role(domain.GrantRoleEditor),
		View: exception(false, true, true),
		Edit: exception(true, false, false),
	})
	assert.True(t, got.CanView, "閲覧は許可リストに載っている")
	assert.False(t, got.CanEdit, "編集は別の段で deny されている")
}

func Test_実効権限_共有リンクは既定をリンク自身から得る(t *testing.T) {
	viewLink := domain.ResolvePagePermission(domain.PagePermissionFacts{
		ShareLinkCapability: capability(domain.CapabilityView),
	})
	assert.True(t, viewLink.CanView)
	assert.False(t, viewLink.CanEdit, "閲覧のリンクでは編集できない")

	editLink := domain.ResolvePagePermission(domain.PagePermissionFacts{
		ShareLinkCapability: capability(domain.CapabilityEdit),
	})
	assert.True(t, editLink.CanView)
	assert.True(t, editLink.CanEdit)
}

func Test_実効権限_共有リンクでも例外は効く(t *testing.T) {
	// 公開したページの下の 1 枚だけ deny で隠す使い方。
	hidden := domain.ResolvePagePermission(domain.PagePermissionFacts{
		ShareLinkCapability: capability(domain.CapabilityView),
		View:                exception(true, false, false),
	})
	assert.False(t, hidden.CanView, "リンクの主体を deny した子ページは開けない")
}

func Test_実効権限_非メンバーは既定を持たない(t *testing.T) {
	got := domain.ResolvePagePermission(domain.PagePermissionFacts{Member: false})
	assert.False(t, got.CanView)
	assert.False(t, got.CanEdit)
}

func Test_実効権限_Allows(t *testing.T) {
	p := domain.PagePermission{CanView: true, CanEdit: false}
	assert.True(t, p.Allows(domain.CapabilityView))
	assert.False(t, p.Allows(domain.CapabilityEdit))
	assert.True(t, p.Allows(domain.Capability("unknown")), "既知でない値は閲覧として扱う（最も弱い解釈）")
}

func Test_役割_強さと権限(t *testing.T) {
	assert.Greater(t, domain.GrantRoleAdmin.Rank(), domain.GrantRoleEditor.Rank())
	assert.Greater(t, domain.GrantRoleEditor.Rank(), domain.GrantRoleCommenter.Rank())
	assert.Greater(t, domain.GrantRoleCommenter.Rank(), domain.GrantRoleViewer.Rank())
	assert.Equal(t, 0, domain.GrantRole("unknown").Rank())

	assert.True(t, domain.GrantRoleAdmin.CanManage())
	assert.False(t, domain.GrantRoleEditor.CanManage())
	assert.True(t, domain.GrantRoleCommenter.CanComment())
	assert.False(t, domain.GrantRoleViewer.CanComment())
	assert.False(t, domain.GrantRole("unknown").CanView())
}

func Test_役割_強さからの逆引き(t *testing.T) {
	for _, r := range domain.ValidGrantRoles {
		got := domain.GrantRoleByRank(r.Rank())
		require.NotNil(t, got, "%s の逆引き", r)
		assert.Equal(t, r, *got)
	}
	assert.Nil(t, domain.GrantRoleByRank(0), "0 は grant なし")
	assert.Nil(t, domain.GrantRoleByRank(99))
}

func Test_権限モデルの値の検証(t *testing.T) {
	for _, k := range domain.ValidPrincipalKinds {
		assert.True(t, k.Valid(), string(k))
	}
	assert.False(t, domain.PrincipalKind("robot").Valid())

	for _, r := range domain.ValidGrantRoles {
		assert.True(t, r.Valid(), string(r))
	}
	assert.False(t, domain.GrantRole("owner").Valid())

	for _, c := range domain.ValidCapabilities {
		assert.True(t, c.Valid(), string(c))
	}
	assert.False(t, domain.Capability("comment").Valid())

	for _, m := range domain.ValidRestrictionModes {
		assert.True(t, m.Valid(), string(m))
	}
	assert.False(t, domain.RestrictionMode("ignore").Valid())
}
