package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func materialRole(r domain.GrantRole) *domain.GrantRole { return &r }

func Test_教材の権限_読むことに付与を要求しない(t *testing.T) {
	// ここに付与を求めると、研修を受ける人が教材を開くたびに権限を配ることになる。
	got := domain.ResolveMaterialPermission(domain.MaterialFacts{Member: true, Published: true})
	assert.True(t, got.CanView, "公開済みなら一員は誰でも読める")
	assert.False(t, got.CanEdit, "書き換えは別（付与が要る）")
	assert.False(t, got.CanManage)
}

func Test_教材の権限_下書きは編集できる人にしか見せない(t *testing.T) {
	member := domain.MaterialFacts{Member: true, Published: false}
	assert.False(t, domain.ResolveMaterialPermission(member).CanView, "付与の無い一員には見えない")

	editor := domain.MaterialFacts{Member: true, Published: false, Role: materialRole(domain.GrantRoleEditor)}
	got := domain.ResolveMaterialPermission(editor)
	assert.True(t, got.CanView, "編集できる人には下書きも見える")
	assert.True(t, got.CanEdit)

	viewer := domain.MaterialFacts{Member: true, Published: false, Role: materialRole(domain.GrantRoleViewer)}
	assert.True(t, domain.ResolveMaterialPermission(viewer).CanView, "viewer を張れば下書きを覗ける")
	assert.False(t, domain.ResolveMaterialPermission(viewer).CanEdit)
}

func Test_教材の権限_所属していなければ何も見えない(t *testing.T) {
	// 公開済みでも、付与を持っていても、他テナントの教材は読めない。
	for _, c := range []struct {
		name  string
		facts domain.MaterialFacts
	}{
		{"公開済み", domain.MaterialFacts{Member: false, Published: true}},
		{"付与あり", domain.MaterialFacts{Member: false, Role: materialRole(domain.GrantRoleAdmin)}},
		{"ワークスペースの admin", domain.MaterialFacts{Member: false, WorkspaceAdmin: true, Published: true}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := domain.ResolveMaterialPermission(c.facts)
			assert.False(t, got.CanView, "所属していない相手に中身が見えている")
			assert.False(t, got.CanEdit)
		})
	}
}

func Test_教材の権限_ワークスペースのadminだけが届く(t *testing.T) {
	// **この 1 つだけが例外。** ノートの editor が教材の編集権まで得ないこと。
	admin := domain.ResolveMaterialPermission(domain.MaterialFacts{
		Member: true, WorkspaceAdmin: true, Published: false,
	})
	assert.True(t, admin.CanEdit, "ワークスペースの admin は配下すべてを扱える")
	assert.True(t, admin.CanManage)
	assert.True(t, admin.CanView, "下書きも見える")

	// 「ノートの editor が教材へ届かない」は、事実を集める側（SQL）が
	// ワークスペースの admin 以外を Role へ入れないことで守る。ここで見られるのは
	// 「事実として何も届いていなければ何もできない」までで、集め方そのものは
	// 結合テスト（実 PostgreSQL）が固定する。
	notAdmin := domain.ResolveMaterialPermission(domain.MaterialFacts{Member: true, Published: true})
	assert.False(t, notAdmin.CanEdit, "ノートの editor が教材まで編集できてはいけない")
	assert.False(t, notAdmin.CanManage)
}

func Test_教材の権限_役割ごとにできること(t *testing.T) {
	for _, c := range []struct {
		role                 domain.GrantRole
		wantEdit, wantManage bool
	}{
		{domain.GrantRoleAdmin, true, true},
		{domain.GrantRoleEditor, true, false},
		{domain.GrantRoleCommenter, false, false},
		{domain.GrantRoleViewer, false, false},
	} {
		t.Run(string(c.role), func(t *testing.T) {
			got := domain.ResolveMaterialPermission(domain.MaterialFacts{
				Member: true, Published: true, Role: role(c.role),
			})
			assert.Equal(t, c.wantEdit, got.CanEdit)
			assert.Equal(t, c.wantManage, got.CanManage)
			assert.True(t, got.CanView)
		})
	}
}
