package domain

import "time"

// LearningReport はユーザーの学習レポート (週次・月次集計を非同期で生成)。
type LearningReport struct {
	ID         uint64    `json:"id"`
	UserID     uint64    `json:"userId"`
	PeriodFrom time.Time `json:"periodFrom"`
	PeriodTo   time.Time `json:"periodTo"`
	Status     string    `json:"status"`
	S3Key      string    `json:"s3Key,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

const (
	LearningReportStatusPending = "pending"
	LearningReportStatusReady   = "ready"
	LearningReportStatusFailed  = "failed"
)
