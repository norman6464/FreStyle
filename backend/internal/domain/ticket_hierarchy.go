package domain

import (
	"errors"
	"time"
)

// ErrTicketHierarchyRejected は種別の階層規則（設計 Ⅳ-D）に反する親子関係を表す。
// usecase が親チェーンを DB から読んで判定する（行をまたぐ規則なので CHECK 制約では書けない）。
// handler はこれを 409 にマップする。
var ErrTicketHierarchyRejected = errors.New("domain: ticket hierarchy rejected")

// ErrTicketDateRangeInverted は開始日が期限より後（ck_tickets_dates_ordered 違反）を表す。
// DB の CHECK に到達させると素の Postgres エラーで 500 に落ちるため、usecase が保存前に
// これで断って 400 にマップする。
var ErrTicketDateRangeInverted = errors.New("domain: ticket start date is after due date")

// チケット・状態・種別の入力検証で使うセンチネル。gin の binding:"required" は空白だけの文字列や、
// 色・カテゴリ・階層レベルのような固定集合の妥当性までは見ないため、usecase 側の検証がここを
// 通って handler の 400 マッピングにつなげる。
var (
	ErrInvalidTicketName           = errors.New("domain: invalid ticket name")
	ErrInvalidTicketColor          = errors.New("domain: invalid ticket color")
	ErrInvalidTicketStatusCategory = errors.New("domain: invalid ticket status category")
	ErrInvalidTicketHierarchyLevel = errors.New("domain: invalid ticket hierarchy level")
)

// ValidTicketHierarchyLevel は種別の階層レベルとして保存してよい値かを返す
// （ck_ticket_types_hierarchy_level と同じ判定。1=束ね / 0=標準 / -1=小作業）。
func ValidTicketHierarchyLevel(level int) bool {
	return level >= -1 && level <= 1
}

// ValidateTicketParentChild は「子の種別の段」と「親の種別の段」の組み合わせが許されるかを
// 検証する（設計 Ⅳ-D）。規則: 子の段 <= 親の段。同じ段どうしの親子は 0（標準）だけ許す
// （1=束ねの下に束ねは作れない、-1=小作業の下に小作業は作れない）。-1 は親になれない
// （小作業はこれ以上分割しない末端）。
//
// 呼び出し側（usecase）は「深さ最大 3」「周期を作らない」を親チェーンを辿って別途検証する
// （行をまたぐ探索が要るため、ここには置けない）。
func ValidateTicketParentChild(childLevel, parentLevel int) error {
	if parentLevel == -1 {
		return ErrTicketHierarchyRejected
	}
	if childLevel > parentLevel {
		return ErrTicketHierarchyRejected
	}
	if childLevel == parentLevel && childLevel != 0 {
		return ErrTicketHierarchyRejected
	}
	return nil
}

// ValidTicketDateOrder は開始日・期限（'YYYY-MM-DD' 文字列、Ⅳ-K）の組が
// ck_tickets_dates_ordered を満たすかを返す。文字列比較で足りる（辞書順が日付順と一致する）。
func ValidTicketDateOrder(startDate, dueDate *string) bool {
	if startDate == nil || dueDate == nil {
		return true
	}
	return *startDate <= *dueDate
}

// ResolveTicketClosedFields は状態変更時の closed_at / resolution を状態の category から必ず
// 導く。呼び出し側が resolution を直接指定できる形にすると、category=todo なのに resolution が
// 入った矛盾行を作れてしまう（DB の ck_tickets_closed_pair は両方揃っているかしか見ておらず、
// category との整合はここが担う）。
//
// category が done 以外なら常に (nil, nil)。done なら closedAt は now、resolution は requested
// があればそれを、無ければ TicketResolutionDone を返す。純粋関数にするため time.Now() は
// 呼ばない（呼び出し側が now を渡す）。
func ResolveTicketClosedFields(
	category TicketStatusCategory, requested *TicketResolution, now time.Time,
) (closedAt *time.Time, resolution *TicketResolution) {
	if category != TicketStatusCategoryDone {
		return nil, nil
	}
	r := TicketResolutionDone
	if requested != nil {
		r = *requested
	}
	return &now, &r
}
