package domain

import (
	"strconv"
	"strings"
)

// チケット（Jira / Nulab Backlog でいう「課題」）の語彙。
//
// 表・API・画面が増えるのは段 1 以降で、ここには**どの層も同じ言葉で話すための値**だけを置く。
// ここにある集合は DB の CHECK 制約と対になっている。片方だけ増やすと
// 「Go では通るが保存で落ちる」「保存はできるが Go の知らない値が読み出される」になるので、
// 変えるときは schema.hcl の CHECK と同時に変え、ticket_test.go の集合も直すこと。

// TicketStatusCategory は状態の「枠」。
//
// 状態の名前はスペースごとに自由に足せるが、必ずこの 3 つのどれかに属する。
// 進捗の集計・ボードの列・「終わったか」の判定は、名前ではなく**枠だけ**で書く。
// 名前で書くと、利用者が状態を 1 つ足した日に集計が黙って合わなくなる。
type TicketStatusCategory string

const (
	// TicketStatusCategoryTodo は未着手。
	TicketStatusCategoryTodo TicketStatusCategory = "todo"
	// TicketStatusCategoryInProgress は作業中。
	TicketStatusCategoryInProgress TicketStatusCategory = "in_progress"
	// TicketStatusCategoryDone は完了。この枠に入ったときだけ closed_at と resolution を持つ。
	TicketStatusCategoryDone TicketStatusCategory = "done"
)

// ValidTicketStatusCategories は保存を許す枠の一覧（未着手から完了への順）。
var ValidTicketStatusCategories = []TicketStatusCategory{
	TicketStatusCategoryTodo,
	TicketStatusCategoryInProgress,
	TicketStatusCategoryDone,
}

// Valid は既知の枠かを返す（保存前の検証に使う）。
func (c TicketStatusCategory) Valid() bool {
	for _, v := range ValidTicketStatusCategories {
		if v == c {
			return true
		}
	}
	return false
}

// TicketPriority は優先度。小さいほど高い（1 = 高 / 2 = 中 / 3 = 低）。
//
// Jira は 5 段だが、実データでは直近 50 件すべてが Medium だった。Backlog と同じ 3 段にする。
// 数値の向きが「小さいほど高い」なのは、一覧を priority 昇順で並べると
// 高い順になり、SQL 側に CASE を書かずに済むため。
type TicketPriority int

const (
	// TicketPriorityHigh は高。
	TicketPriorityHigh TicketPriority = 1
	// TicketPriorityNormal は中。
	TicketPriorityNormal TicketPriority = 2
	// TicketPriorityLow は低。
	TicketPriorityLow TicketPriority = 3
)

// TicketPriorityDefault は指定が無いときの優先度。列の DEFAULT と同じ値にする。
const TicketPriorityDefault = TicketPriorityNormal

// ValidTicketPriorities は保存を許す優先度の一覧（高い順）。
var ValidTicketPriorities = []TicketPriority{
	TicketPriorityHigh,
	TicketPriorityNormal,
	TicketPriorityLow,
}

// Valid は既知の優先度かを返す（保存前の検証に使う）。
func (p TicketPriority) Valid() bool {
	for _, v := range ValidTicketPriorities {
		if v == p {
			return true
		}
	}
	return false
}

// TicketResolution は「なぜ終わったか」。完了枠の状態にあるときだけ持つ。
//
// 「理由が無い」は空文字ではなく NULL（*TicketResolution の nil）で表す。
// 空文字を許すと「未設定」と「空という理由」の二通りができる。
type TicketResolution string

const (
	// TicketResolutionDone は対応済み。完了枠へ移すときに指定が無ければこれになる。
	TicketResolutionDone TicketResolution = "done"
	// TicketResolutionWontDo は対応しない（「作業中断」のような状態に付く）。
	TicketResolutionWontDo TicketResolution = "wont_do"
	// TicketResolutionInvalid は無効。
	TicketResolutionInvalid TicketResolution = "invalid"
	// TicketResolutionDuplicate は重複。
	TicketResolutionDuplicate TicketResolution = "duplicate"
	// TicketResolutionCannotReproduce は再現しない。
	TicketResolutionCannotReproduce TicketResolution = "cannot_reproduce"
)

// ValidTicketResolutions は保存を許す完了理由の一覧（Backlog と同じ 5 種）。
var ValidTicketResolutions = []TicketResolution{
	TicketResolutionDone,
	TicketResolutionWontDo,
	TicketResolutionInvalid,
	TicketResolutionDuplicate,
	TicketResolutionCannotReproduce,
}

// Valid は既知の完了理由かを返す（保存前の検証に使う）。
func (r TicketResolution) Valid() bool {
	for _, v := range ValidTicketResolutions {
		if v == r {
			return true
		}
	}
	return false
}

// TicketChangeField は変更履歴の 1 行が「何を変えたか」。
//
// 段 3・段 4 でしか書かれない値（category / milestone / link）も**最初から含める**。
// 段ごとに CHECK を DROP して ADD し直すのを避けるための判断で、
// 「まだ書かれない値が CHECK にある」状態を許す代わりに、制約の作り直しをしない。
type TicketChangeField string

const (
	TicketChangeFieldTitle      TicketChangeField = "title"
	TicketChangeFieldDoc        TicketChangeField = "doc"
	TicketChangeFieldStatus     TicketChangeField = "status"
	TicketChangeFieldType       TicketChangeField = "type"
	TicketChangeFieldPriority   TicketChangeField = "priority"
	TicketChangeFieldAssignee   TicketChangeField = "assignee"
	TicketChangeFieldParent     TicketChangeField = "parent"
	TicketChangeFieldStartDate  TicketChangeField = "start_date"
	TicketChangeFieldDueDate    TicketChangeField = "due_date"
	TicketChangeFieldResolution TicketChangeField = "resolution"
	TicketChangeFieldPosition   TicketChangeField = "position"
	TicketChangeFieldArchived   TicketChangeField = "archived"
	TicketChangeFieldCategory   TicketChangeField = "category"
	TicketChangeFieldMilestone  TicketChangeField = "milestone"
	TicketChangeFieldLink       TicketChangeField = "link"
)

// ValidTicketChangeFields は履歴に書いてよい項目の一覧。
var ValidTicketChangeFields = []TicketChangeField{
	TicketChangeFieldTitle,
	TicketChangeFieldDoc,
	TicketChangeFieldStatus,
	TicketChangeFieldType,
	TicketChangeFieldPriority,
	TicketChangeFieldAssignee,
	TicketChangeFieldParent,
	TicketChangeFieldStartDate,
	TicketChangeFieldDueDate,
	TicketChangeFieldResolution,
	TicketChangeFieldPosition,
	TicketChangeFieldArchived,
	TicketChangeFieldCategory,
	TicketChangeFieldMilestone,
	TicketChangeFieldLink,
}

// Valid は履歴に書いてよい項目かを返す（保存前の検証に使う）。
func (f TicketChangeField) Valid() bool {
	for _, v := range ValidTicketChangeFields {
		if v == f {
			return true
		}
	}
	return false
}

// チケットが作る通知の種別（notifications.type に入れる値）。
//
// notifications.type は自由文字列で、ほかの機能も自分の値を入れる。ここに並ぶのは
// チケット由来の分だけで、通知種別の全体像ではない。
const (
	// NotificationTypeTicketAssigned は自分が担当になったとき。
	NotificationTypeTicketAssigned = "ticket_assigned"
	// NotificationTypeTicketStatusChanged はウォッチ中のチケットの状態が変わったとき。
	NotificationTypeTicketStatusChanged = "ticket_status_changed"
	// NotificationTypeTicketCommented はウォッチ中のチケットにコメントが付いたとき。
	NotificationTypeTicketCommented = "ticket_commented"
	// NotificationTypeTicketMentioned はコメントで自分が名指しされたとき。
	NotificationTypeTicketMentioned = "ticket_mentioned"
)

// TicketNumberMin はチケット番号の最小値。採番は 1 から始まり、0 と負の番号は無い。
const TicketNumberMin = 1

// FormatTicketKey は人が見る表示キー（例: FRESTYLE-12）を組み立てる。
//
// 表示キーは spaces.key と tickets.number からの**派生値で、保存しない**。
// spaces.key は変えない決まりなので二重に持つ理由が無く、SQL 側で連結すると
// sqlc の生成型が interface{} になる（実測）ため、組み立ては Go のここだけで行う。
func FormatTicketKey(spaceKey string, number int64) string {
	return strings.ToUpper(spaceKey) + "-" + strconv.FormatInt(number, 10)
}

// ParseTicketKey は表示キーをスペースの key と番号へ分解する。
// 分解できなければ ok が false で、そのとき key と番号は返さない。
//
// **最後のハイフンで区切る。** spaces.key は小文字の英数字と内側のハイフンを許すので
// `my-app-12` の形が普通に起こり、最初のハイフンで割ると key が `my` になってしまう。
// key 自体が数字で終わる場合（`frestyle-12-3`）も、最後で割れば正しく `frestyle-12` と 3 になる。
//
// 大文字で来ても小文字で来ても受ける（表示は大文字だが、人は小文字でも打つ）。
// 返す key は必ず小文字で、そのまま spaces.key と突き合わせられる。
func ParseTicketKey(s string) (string, int64, bool) {
	i := strings.LastIndex(s, "-")
	if i < 0 {
		return "", 0, false
	}
	spaceKey := strings.ToLower(s[:i])
	if !ValidSpaceKey(spaceKey) {
		return "", 0, false
	}
	number, ok := parseTicketNumber(s[i+1:])
	if !ok {
		return "", 0, false
	}
	return spaceKey, number, true
}

// parseTicketNumber は表示キーの番号側を読む。
//
// strconv.ParseInt にそのまま渡さないのは、あちらが `+12` や `-12`、
// そして先頭に 0 の付いた `012` を受けてしまうため。表示キーの番号は符号も
// 先頭の 0 も持たない（同じチケットを指す綴りが 2 通りできてしまう）。
// 数字だけであることを先に確かめ、int64 に収まるかの判定は ParseInt に任せる
// （桁数で判定すると 19 桁の 9999999999999999999 が静かに折り返す）。
func parseTicketNumber(s string) (int64, bool) {
	if s == "" || s[0] == '0' {
		return 0, false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// int64 に収まらない番号。採番が到達しない大きさだが、URL から来る値なので拒む。
		return 0, false
	}
	if n < TicketNumberMin {
		return 0, false
	}
	return n, true
}
