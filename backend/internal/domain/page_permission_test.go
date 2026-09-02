package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func role(r domain.GrantRole) *domain.GrantRole { return &r }

func capability(c domain.Capability) *domain.Capability { return &c }

// rank は届いた役割の強さ。届いていなければ 0（何もできない）。
func rank(r *domain.GrantRole) int {
	if r == nil {
		return 0
	}
	return r.Rank()
}

func Test_実効権限_届いた役割どおりに決まる(t *testing.T) {
	cases := []struct {
		name      string
		facts     domain.PagePermissionFacts
		canView   bool
		canEdit   bool
		canManage bool
	}{
		{name: "付与が無ければ何もできない", facts: domain.PagePermissionFacts{Member: true}},
		{
			name:    "viewer は閲覧のみ",
			facts:   domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleViewer)},
			canView: true,
		},
		{
			name:    "commenter は閲覧のみ（編集は不可）",
			facts:   domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleCommenter)},
			canView: true,
		},
		{
			name:    "editor は閲覧と編集",
			facts:   domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleEditor)},
			canView: true, canEdit: true,
		},
		{
			name:    "admin は閲覧と編集に加えて権限も変えられる",
			facts:   domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRoleAdmin)},
			canView: true, canEdit: true, canManage: true,
		},
		{
			name:  "未知の役割は届いていないのと同じ",
			facts: domain.PagePermissionFacts{Member: true, Role: role(domain.GrantRole("owner"))},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.ResolvePagePermission(tc.facts)
			assert.Equal(t, tc.canView, got.CanView, "閲覧")
			assert.Equal(t, tc.canEdit, got.CanEdit, "編集")
			assert.Equal(t, tc.canManage, got.CanManage, "権限の変更")
		})
	}
}

// 編集できるなら必ず閲覧もできる。閲覧できないページを編集できる状態は UI でも監査でも
// 説明できない。
//
// **この不変条件は、いまの入力では破りようがない。** 役割の並び（GrantRole.Rank）では
// editor 以上が必ず viewer 以上で、共有リンクも edit のリンクは view を含む。だから
// ResolvePagePermission の `canView &&` を外しても答えは変わらず、**このテストは
// その掛け合わせを守っていない**（変異が生き残ることを確認済み）。守れるのは
// 「編集できる入力では閲覧もできる」という結果の側だけで、それをここで固定する。
//
// 掛け合わせ自体は、役割や共有リンクの種類を増やしたときに崩れないための保険として
// 実装に残してある（そのとき初めてこのテストが破れる側に回る）。
func Test_実効権限_編集できるなら閲覧もできる(t *testing.T) {
	roles := append([]domain.GrantRole{}, domain.ValidGrantRoles...)
	roles = append(roles, domain.GrantRole("owner"))
	for _, r := range roles {
		t.Run(string(r), func(t *testing.T) {
			got := domain.ResolvePagePermission(domain.PagePermissionFacts{Member: true, Role: role(r)})
			if got.CanEdit {
				assert.True(t, got.CanView, "編集できるのに閲覧できない役割がある")
			}
		})
	}

	// 共有リンク経由も同じ（編集のリンクは閲覧もできる）。
	editLink := domain.ResolvePagePermission(domain.PagePermissionFacts{
		ShareLinkCapability: capability(domain.CapabilityEdit),
	})
	assert.True(t, editLink.CanEdit)
	assert.True(t, editLink.CanView)
}

// 共有リンクは広げる方向にしか働かない。ログインしていない相手に「見せる」を足すだけで、
// 役割は持たないので権限そのものは変えられない。
func Test_実効権限_共有リンクは既定をリンク自身から得る(t *testing.T) {
	viewLink := domain.ResolvePagePermission(domain.PagePermissionFacts{
		ShareLinkCapability: capability(domain.CapabilityView),
	})
	assert.True(t, viewLink.CanView)
	assert.False(t, viewLink.CanEdit, "閲覧のリンクでは編集できない")
	assert.False(t, viewLink.CanManage, "リンクは役割を持たない")

	editLink := domain.ResolvePagePermission(domain.PagePermissionFacts{
		ShareLinkCapability: capability(domain.CapabilityEdit),
	})
	assert.True(t, editLink.CanView)
	assert.True(t, editLink.CanEdit)
	assert.False(t, editLink.CanManage, "編集のリンクでも権限は変えられない")
}

// 所属していない相手には役割が 1 つも届かない（役割は principals の kind='user' の行から
// 集めるので、その行が無ければ集めようがない）。事実がこの形になることは
// repository 側の結合テストが持ち、ここでは「その事実なら何もできない」を固定する。
func Test_実効権限_非メンバーは何もできない(t *testing.T) {
	got := domain.ResolvePagePermission(domain.PagePermissionFacts{Member: false})
	assert.False(t, got.CanView)
	assert.False(t, got.CanEdit)
	assert.False(t, got.CanManage)
}

// 一覧（役割の列しか集めない経路）と 1 ページ解決が食い違わないことを、
// 届きうる役割の全種類で固定する。片方だけ直すと「開けるのに一覧に出ない」ずれになる。
func Test_実効権限_一覧の閲覧判定は1ページ解決とつねに一致する(t *testing.T) {
	roles := []*domain.GrantRole{
		nil,
		role(domain.GrantRoleViewer),
		role(domain.GrantRoleCommenter),
		role(domain.GrantRoleEditor),
		role(domain.GrantRoleAdmin),
		role(domain.GrantRole("owner")),
	}
	for _, r := range roles {
		want := domain.ResolvePagePermission(domain.PagePermissionFacts{Member: true, Role: r}).CanView
		assert.Equal(t, want, domain.ResolvePageView(r), "役割 %v", r)
	}
}

// 経路に付与を足しても役割は弱くならない、という合成規則の性質を固定する。
//
// **これは domain の合成（StrongestGrantRole）についての主張で、本番のページ経路の
// 証明ではない。** ページ 1 枚 / 一覧の役割は SQL 側が `GREATEST(...)` で畳んだ強さを
// persistence が `GrantRoleByRank` で戻して作るので、この関数を通らない。
// SQL 側が同じ性質を持つことは結合テスト
// （TestKnowledgeBasePageGrantAPI_祖先に付与を足すと子孫も強くなる_Integration）が確かめる。
//
// ここで固定するのは「規則の側は単調である」こと。SQL とこの規則の両方が単調でなければ、
// 「親は編集できるが子は編集できない」が起きないとは言えない。
func Test_ページ権限_経路に付与を足しても役割は弱くならない(t *testing.T) {
	pool := []domain.GrantRole{
		domain.GrantRoleAdmin,
		domain.GrantRoleEditor,
		domain.GrantRoleCommenter,
		domain.GrantRoleViewer,
		domain.GrantRole("owner"), // 未知の値。数えないが、足しても弱くしてはいけない
	}
	for mask := 0; mask < 1<<len(pool); mask++ {
		ancestor := make([]domain.GrantRole, 0, len(pool))
		for i, r := range pool {
			if mask&(1<<i) != 0 {
				ancestor = append(ancestor, r)
			}
		}
		ancestorRole := domain.StrongestGrantRole(ancestor)
		ancestorPerm := domain.ResolvePagePermission(
			domain.PagePermissionFacts{Member: true, Role: ancestorRole},
		)

		for _, added := range pool {
			descendant := make([]domain.GrantRole, 0, len(ancestor)+1)
			descendant = append(descendant, ancestor...)
			descendant = append(descendant, added)

			descendantRole := domain.StrongestGrantRole(descendant)
			assert.GreaterOrEqual(t, rank(descendantRole), rank(ancestorRole),
				"祖先 %v に %s を足したら弱くなった", ancestor, added)

			got := domain.ResolvePagePermission(
				domain.PagePermissionFacts{Member: true, Role: descendantRole},
			)
			if ancestorPerm.CanView {
				assert.True(t, got.CanView, "祖先 %v で閲覧できたのに子孫で閲覧できない", ancestor)
			}
			if ancestorPerm.CanEdit {
				assert.True(t, got.CanEdit, "祖先 %v で編集できたのに子孫で編集できない", ancestor)
			}
			if ancestorPerm.CanManage {
				assert.True(t, got.CanManage, "祖先 %v で権限を変えられたのに子孫で変えられない", ancestor)
			}
		}
	}
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
	assert.Nil(t, domain.GrantRoleByRank(0), "0 は付与なし")
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
}

func Test_ページ権限_管理は役割だけで決まる(t *testing.T) {
	t.Run("admin が届いていれば管理できる", func(t *testing.T) {
		got := domain.ResolvePagePermission(domain.PagePermissionFacts{
			Member: true, Role: role(domain.GrantRoleAdmin),
		})
		assert.True(t, got.CanManage)
	})

	t.Run("editor では管理できない", func(t *testing.T) {
		got := domain.ResolvePagePermission(domain.PagePermissionFacts{
			Member: true, Role: role(domain.GrantRoleEditor),
		})
		assert.False(t, got.CanManage)
		assert.True(t, got.CanEdit, "編集はできる")
	})

	t.Run("共有リンクの来訪者は管理できない", func(t *testing.T) {
		// リンクは役割を持たない（Role が nil）ので、編集のリンクでも管理には届かない。
		got := domain.ResolvePagePermission(domain.PagePermissionFacts{
			ShareLinkCapability: capability(domain.CapabilityEdit),
		})
		assert.True(t, got.CanEdit)
		assert.False(t, got.CanManage)
	})

	t.Run("役割が無ければ管理できない", func(t *testing.T) {
		assert.False(t, domain.ResolvePagePermission(domain.PagePermissionFacts{Member: true}).CanManage)
	})
}
