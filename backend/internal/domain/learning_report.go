package domain

import "time"

// LearningReport はユーザーの学習レポート (週次・月次集計を非同期で生成)。
type LearningReport struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"column:user_id;index" json:"userId"`
	PeriodFrom time.Time `gorm:"column:period_from" json:"periodFrom"`
	PeriodTo   time.Time `gorm:"column:period_to" json:"periodTo"`
	Status     string    `gorm:"column:status" json:"status"`
	S3Key      string    `gorm:"column:s3_key" json:"s3Key,omitempty"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (LearningReport) TableName() string { return "learning_reports" }

const (
	LearningReportStatusPending = "pending"
	LearningReportStatusReady   = "ready"
	LearningReportStatusFailed  = "failed"
)
