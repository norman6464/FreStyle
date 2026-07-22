package repository

import (
	"context"
	"time"
)

// MemberLearningActivity は自社 trainee 1 人分の学習アクティビティ集計。
type MemberLearningActivity struct {
	UserID uint64
	Name   string
	// LastActiveDate は最後に学習活動があった日(user_daily_activities の activity_date、UTC 基準)。
	// 一度も活動が無ければ nil。
	LastActiveDate *time.Time
	// RecentActivityCount は fromDate 以降の活動回数合計(演習 + レッスン + AI チャット + ノート)。
	RecentActivityCount int
}

// CompanyLearningActivitySummarizer は自社メンバーの学習状況を集計する単一責務 port
// (CompanyMemberCounter と同じ Effective Go 流の -er 命名)。
type CompanyLearningActivitySummarizer interface {
	// ListMemberActivities は company の trainee(論理削除済みを除く)ごとの学習アクティビティを、
	// 最終活動日の新しい順(未活動は末尾)で返す。活動が一度も無い trainee も件数 0 で含む。
	ListMemberActivities(ctx context.Context, companyID uint64, fromDate time.Time) ([]MemberLearningActivity, error)
}
