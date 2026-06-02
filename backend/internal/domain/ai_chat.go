package domain

import "time"

// AiChatSession は AI チャットの 1 セッション。
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

// AiChatMessage は AI チャットの 1 メッセージ。実体は DynamoDB に保存し、API ではこの形で返す。
type AiChatMessage struct {
	SessionID   uint64       `json:"sessionId"`
	MessageID   string       `json:"messageId"`
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
}

// Attachment は AI チャット送信時の添付ファイル。
// Format は Bedrock Converse API が要求する短い文字列（"png" / "jpeg" など）。
// BlobData は Bedrock 呼び出し直前に S3 から詰める一時フィールドで、永続化はしない。
type Attachment struct {
	Key         string `json:"key"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Format      string `json:"format"`
	Kind        string `json:"kind"`
	SizeBytes   int64  `json:"sizeBytes"`
	BlobData    []byte `json:"-"`
}

const (
	AiChatRoleUser      = "user"
	AiChatRoleAssistant = "assistant"

	AttachmentKindImage    = "image"
	AttachmentKindDocument = "document"
)
