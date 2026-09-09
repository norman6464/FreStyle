package domain

import (
	"errors"
	"time"
)

// ErrTicketHierarchyRejected は種別の階層規則（設計 Ⅳ-D）に反する親子関係を表す。
// usecase が親チェーンを DB から読んだうえでこれを判定に使う（行をまたぐ規則なので
// CHECK 制約では書けない）。handler はこれを 409 にマップする。
var ErrTicketHierarchyRejected = errors.New("domain: ticket hierarchy rejected")

// ErrTicketDateRangeInverted は開始日が期限より後（ck_tickets_dates_ordered 違反）を表す。
// DB の CHECK 制約に到達させると素の Postgres エラーになり 500 へ落ちてしまうため、
// usecase が保存前にこれで断って 400 にマップできるようにする。
var ErrTicketDateRangeInverted = errors.New("domain: ticket start date is after due date")

// チケット・状態・種別の入力検証で使うセンチネル。handler がこれらを見て 400 に
// マップできるように（gin の binding:"required" は空白だけの文字列や、色・カテゴリ・
// 階層レベルのような固定集合の妥当性までは見ないため、usecase 側の検証がここを通る）。
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

// ValidateTicketParentChild は「子の種別の段」と「親の種別の段」の組み合わせが
// 許されるかを検証する。
//
// 規則（設計 Ⅳ-D）: 子の段 <= 親の段。同じ段どうしの親子は 0（標準）だけ許す
// （1 = 束ねの下に束ねは作れない。−1 = 小作業の下に小作業は作れない）。
// −1 は親になれない（小作業はこれ以上分割しない末端）。
//
// 呼び出し側（usecase）は、これに加えて「深さ最大 3」「周期を作らない」を
// 親チェーンを辿って別途検証する（行をまたぐ探索が要るため、ここには置けない）。
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
// ck_tickets_dates_ordered を満たすかを返す。文字列比較で足りる
// （'YYYY-MM-DD' の辞書順は日付順と一致する）。
func ValidTicketDateOrder(startDate, dueDate *string) bool {
	if startDate == nil || dueDate == nil {
		return true
	}
	return *startDate <= *dueDate
}

// ResolveTicketClosedFields は状態変更時の closed_at / resolution を、状態の
// category から必ず導く。呼び出し側（usecase）が resolution を直接指定できる形に
// すると、category=todo なのに resolution が入った矛盾行を作れてしまう
// （DB の ck_tickets_closed_pair は「両方揃っているか」しか見ておらず、
// category との整合はここが担う）。
//
// category が done 以外なら常に (nil, nil)。done なら closedAt は now、
// resolution は requested があればそれを、無ければ TicketResolutionDone を返す。
//
// 純粋関数にするため time.Now() は呼ばない（呼び出し側が now を渡す）。
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
