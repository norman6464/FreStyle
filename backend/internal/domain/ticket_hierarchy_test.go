package domain_test

import (
	"testing"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_種別の階層レベル(t *testing.T) {
	for _, ok := range []int{-1, 0, 1} {
		assert.True(t, domain.ValidTicketHierarchyLevel(ok), ok)
	}
	for _, ng := range []int{-2, 2, 3, 100} {
		assert.False(t, domain.ValidTicketHierarchyLevel(ng), ng)
	}
}

// 親子の階層規則（設計 Ⅳ-D）: 子の段 <= 親の段。同じ段どうしの親子は 0 だけ許す
// （束ね 1 の下に束ね 1 は作れない。標準 0 の下に標準 0 は作れる＝サブタスク的な下位分割）。
// -1（小作業）は親になれない。
func Test_親子の階層規則(t *testing.T) {
	cases := []struct {
		name        string
		childLevel  int
		parentLevel int
		ok          bool
	}{
		{name: "束ね(1)の下に標準(0)", childLevel: 0, parentLevel: 1, ok: true},
		{name: "束ね(1)の下に小作業(-1)", childLevel: -1, parentLevel: 1, ok: true},
		{name: "標準(0)の下に標準(0)", childLevel: 0, parentLevel: 0, ok: true},
		{name: "標準(0)の下に小作業(-1)", childLevel: -1, parentLevel: 0, ok: true},
		{name: "束ね(1)の下に束ね(1)は拒否", childLevel: 1, parentLevel: 1, ok: false},
		{name: "小作業(-1)の下に小作業(-1)は拒否", childLevel: -1, parentLevel: -1, ok: false},
		{name: "小作業(-1)は親になれない（子が標準でも）", childLevel: 0, parentLevel: -1, ok: false},
		{name: "小作業(-1)は親になれない（子が束ねでも）", childLevel: 1, parentLevel: -1, ok: false},
		{name: "子が親より段上は拒否", childLevel: 1, parentLevel: 0, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateTicketParentChild(tc.childLevel, tc.parentLevel)
			if tc.ok {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, domain.ErrTicketHierarchyRejected)
			}
		})
	}
}

func Test_日付の順序(t *testing.T) {
	d := func(s string) *string { return &s }

	assert.True(t, domain.ValidTicketDateOrder(nil, nil), "両方無ければ制約なし")
	assert.True(t, domain.ValidTicketDateOrder(d("2026-09-01"), nil), "開始日だけならOK")
	assert.True(t, domain.ValidTicketDateOrder(nil, d("2026-09-01")), "期限だけならOK")
	assert.True(t, domain.ValidTicketDateOrder(d("2026-09-01"), d("2026-09-01")), "同日はOK")
	assert.True(t, domain.ValidTicketDateOrder(d("2026-09-01"), d("2026-09-10")), "開始 < 期限")
	assert.False(t, domain.ValidTicketDateOrder(d("2026-09-10"), d("2026-09-01")), "開始 > 期限は拒否")
}

// 状態変更時の closed_at / resolution は、状態の category から必ず導く。
// 呼び出し側が resolution を直接指定できてしまうと、category=todo なのに
// resolution が入った矛盾行が作れてしまう（DB の ck_tickets_closed_pair は
// 「両方揃っているか」しか見ておらず、category との整合は usecase の責務）。
func Test_状態変更時のclosed_atとresolutionの導出(t *testing.T) {
	now, err := time.Parse(time.RFC3339, "2026-09-09T10:00:00Z")
	require.NoError(t, err)
	done := domain.TicketResolutionDone
	wontDo := domain.TicketResolutionWontDo

	t.Run("todoへ変更するとclosed_atとresolutionは両方nil", func(t *testing.T) {
		closedAt, resolution := domain.ResolveTicketClosedFields(domain.TicketStatusCategoryTodo, nil, now)
		assert.Nil(t, closedAt)
		assert.Nil(t, resolution)
	})

	t.Run("in_progressへ変更しても両方nil", func(t *testing.T) {
		closedAt, resolution := domain.ResolveTicketClosedFields(domain.TicketStatusCategoryInProgress, &done, now)
		assert.Nil(t, closedAt, "in_progressはresolutionを指定していてもclosed_atを持たない")
		assert.Nil(t, resolution)
	})

	t.Run("doneへ変更しresolution未指定ならdoneが既定", func(t *testing.T) {
		closedAt, resolution := domain.ResolveTicketClosedFields(domain.TicketStatusCategoryDone, nil, now)
		require.NotNil(t, closedAt)
		assert.Equal(t, now, *closedAt)
		require.NotNil(t, resolution)
		assert.Equal(t, domain.TicketResolutionDone, *resolution)
	})

	t.Run("doneへ変更しresolution指定ありならそれを使う", func(t *testing.T) {
		closedAt, resolution := domain.ResolveTicketClosedFields(domain.TicketStatusCategoryDone, &wontDo, now)
		assert.NotNil(t, closedAt)
		assert.Equal(t, domain.TicketResolutionWontDo, *resolution)
	})
}
