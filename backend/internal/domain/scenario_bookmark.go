package domain

import "time"

// ScenarioBookmark は練習シナリオへのブックマーク。
type ScenarioBookmark struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"column:user_id;index" json:"userId"`
	ScenarioID uint64    `gorm:"column:scenario_id;index" json:"scenarioId"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (ScenarioBookmark) TableName() string { return "scenario_bookmarks" }
