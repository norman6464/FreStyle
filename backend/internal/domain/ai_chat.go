package domain

import "time"

// AiChatSession は AI チャットの 1 セッション。Spring Boot の entity.AiChatSession に相当。
type AiChatSession struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	UserID      uint64    `gorm:"column:user_id;index" json:"userId"`
	Title       string    `gorm:"column:title" json:"title"`
	SessionType string    `gorm:"column:session_type" json:"sessionType"`
	ScenarioID  *uint64   `gorm:"column:scenario_id" json:"scenarioId,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (AiChatSession) TableName() string { return "ai_chat_sessions" }

const (
	AiChatSessionTypeFree     = "free"
	AiChatSessionTypePractice = "practice"
)

// AiChatMessage は AI チャットの 1 メッセージ。実体は DynamoDB に保存されるが、
// API ではこの形で返却する。
type AiChatMessage struct {
	SessionID uint64    `json:"sessionId"`
	MessageID string    `json:"messageId"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

const (
	AiChatRoleUser      = "user"
	AiChatRoleAssistant = "assistant"
)
