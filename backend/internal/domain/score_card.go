package domain

import "time"

// ScoreCard は AI が評価したスコア結果。
type ScoreCard struct {
	ID                uint64    `gorm:"primaryKey" json:"id"`
	UserID            uint64    `gorm:"column:user_id;index" json:"userId"`
	SessionID         uint64    `gorm:"column:session_id;index" json:"sessionId"`
	OverallScore      float64   `gorm:"column:overall_score" json:"overallScore"`
	LogicalScore      float64   `gorm:"column:logical_score" json:"logicalScore"`
	ConsiderationScore float64  `gorm:"column:consideration_score" json:"considerationScore"`
	SummaryScore      float64   `gorm:"column:summary_score" json:"summaryScore"`
	ProposalScore     float64   `gorm:"column:proposal_score" json:"proposalScore"`
	ListeningScore    float64   `gorm:"column:listening_score" json:"listeningScore"`
	Feedback          string    `gorm:"column:feedback" json:"feedback"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (ScoreCard) TableName() string { return "score_cards" }
