package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 定数の集合は DB の CHECK 制約と対になっている。片方だけ増えると
// 「Go では通るが保存で落ちる」「保存はできるが Go が知らない値が読み出される」になるので、
// 値そのものを固定しておく。
func Test_チケットの状態の枠(t *testing.T) {
	assert.Equal(t,
		[]domain.TicketStatusCategory{"todo", "in_progress", "done"},
		domain.ValidTicketStatusCategories,
		"ck_ticket_statuses_category と同じ集合であること")

	for _, c := range domain.ValidTicketStatusCategories {
		assert.True(t, c.Valid(), string(c))
	}
	assert.False(t, domain.TicketStatusCategory("").Valid())
	assert.False(t, domain.TicketStatusCategory("closed").Valid(), "似ているが知らない枠は拒む")
	assert.False(t, domain.TicketStatusCategory("TODO").Valid(), "大文字は別の値")
}

func Test_チケットの優先度(t *testing.T) {
	assert.Equal(t,
		[]domain.TicketPriority{1, 2, 3},
		domain.ValidTicketPriorities,
		"ck_tickets_priority と同じ集合であること")
	assert.Equal(t, domain.TicketPriorityNormal, domain.TicketPriorityDefault,
		"列の DEFAULT 2 と一致すること")

	for _, p := range domain.ValidTicketPriorities {
		assert.True(t, p.Valid(), p)
	}
	assert.False(t, domain.TicketPriority(0).Valid())
	assert.False(t, domain.TicketPriority(4).Valid())
	assert.False(t, domain.TicketPriority(-1).Valid())
}

func Test_チケットの完了理由(t *testing.T) {
	assert.Equal(t,
		[]domain.TicketResolution{"done", "wont_do", "invalid", "duplicate", "cannot_reproduce"},
		domain.ValidTicketResolutions,
		"ck_tickets_resolution と同じ集合であること")

	for _, r := range domain.ValidTicketResolutions {
		assert.True(t, r.Valid(), string(r))
	}
	assert.False(t, domain.TicketResolution("").Valid(),
		"「理由なし」は空文字ではなく NULL（*TicketResolution の nil）で表す")
	assert.False(t, domain.TicketResolution("fixed").Valid())
}

// 履歴の項目は段 3・段 4 で書かれる値（category / milestone / link）も最初から含める。
// 段ごとに CHECK を DROP + ADD し直さないための判断（設計 Ⅳ-F）で、
// Go 側の集合もそれに合わせておく。
func Test_チケットの履歴の項目(t *testing.T) {
	assert.Equal(t, []domain.TicketChangeField{
		"title", "doc", "status", "type", "priority", "assignee", "parent",
		"start_date", "due_date", "resolution", "position", "archived",
		"category", "milestone", "link",
	}, domain.ValidTicketChangeFields, "ck_ticket_change_items_field と同じ集合であること")

	for _, f := range domain.ValidTicketChangeFields {
		assert.True(t, f.Valid(), string(f))
	}
	assert.False(t, domain.TicketChangeField("").Valid())
	assert.False(t, domain.TicketChangeField("watcher").Valid(),
		"ウォッチは履歴に書かない（設計 Ⅳ-F の項目一覧に無い）")
}

func Test_チケットの通知種別(t *testing.T) {
	assert.Equal(t, "ticket_assigned", domain.NotificationTypeTicketAssigned)
	assert.Equal(t, "ticket_status_changed", domain.NotificationTypeTicketStatusChanged)
	assert.Equal(t, "ticket_commented", domain.NotificationTypeTicketCommented)
	assert.Equal(t, "ticket_mentioned", domain.NotificationTypeTicketMentioned)
}

// 表示キーは保存しない派生値なので、組み立てと分解が唯一の正本になる。
func Test_表示キーの組み立て(t *testing.T) {
	assert.Equal(t, "FRESTYLE-12", domain.FormatTicketKey("frestyle", 12))
	assert.Equal(t, "MY-APP-3", domain.FormatTicketKey("my-app", 3),
		"key に含まれるハイフンはそのまま残る")
	assert.Equal(t, "A-1", domain.FormatTicketKey("a", 1))
}

func Test_表示キーの分解(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		spaceKey string
		number   int64
	}{
		{name: "ふつうの表示キー", in: "FRESTYLE-12", spaceKey: "frestyle", number: 12},
		{
			name: "key にハイフンを含む場合は最後のハイフンで区切る",
			in:   "MY-APP-12", spaceKey: "my-app", number: 12,
		},
		{name: "小文字で打たれても受ける", in: "frestyle-12", spaceKey: "frestyle", number: 12},
		{name: "1 文字の key", in: "A-1", spaceKey: "a", number: 1},
		{
			name: "key が数字で終わっても最後のハイフンで割る",
			in:   "FRESTYLE-12-3", spaceKey: "frestyle-12", number: 3,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, number, ok := domain.ParseTicketKey(tc.in)
			require.True(t, ok)
			assert.Equal(t, tc.spaceKey, key)
			assert.Equal(t, tc.number, number)
		})
	}
}

func Test_表示キーの分解の負例(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{name: "空", in: ""},
		{name: "ハイフンが無い", in: "FRESTYLE"},
		{name: "番号が無い", in: "FRESTYLE-"},
		{name: "key が無い", in: "-12"},
		{name: "番号が数字でない", in: "FRESTYLE-abc"},
		{name: "番号が 0（採番は 1 から）", in: "FRESTYLE-0"},
		{name: "ハイフンが連続する（key の末尾がハイフンになる）", in: "FRESTYLE--12"},
		{name: "先頭に 0 が付いた番号（同じチケットを指す綴りを 2 通り作らない）", in: "FRESTYLE-012"},
		{name: "int64 に収まらない番号", in: "FRESTYLE-99999999999999999999"},
		{name: "前後に空白", in: " FRESTYLE-12"},
		{name: "key に使えない文字", in: "FRE_STYLE-12"},
		{name: "番号に符号", in: "FRESTYLE-+12"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, number, ok := domain.ParseTicketKey(tc.in)

			assert.False(t, ok)
			assert.Empty(t, key, "拒むときは値を返さない")
			assert.Zero(t, number)
		})
	}
}

// 組み立てたものは必ず分解できる。表示キーは URL とチケット参照の両方に出るので、
// 片方向だけ通る実装だと「画面には出るが検索から開けない」ずれ方をする。
func Test_表示キーは組み立てと分解が往復する(t *testing.T) {
	cases := []struct {
		spaceKey string
		number   int64
	}{
		{spaceKey: "frestyle", number: 1},
		{spaceKey: "my-app", number: 452},
		{spaceKey: "a", number: 9223372036854775807},
		{spaceKey: "team-2026-q3", number: 10},
	}
	for _, tc := range cases {
		t.Run(tc.spaceKey, func(t *testing.T) {
			key, number, ok := domain.ParseTicketKey(domain.FormatTicketKey(tc.spaceKey, tc.number))

			require.True(t, ok)
			assert.Equal(t, tc.spaceKey, key)
			assert.Equal(t, tc.number, number)
		})
	}
}
