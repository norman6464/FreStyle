package domain

import "time"

// AiChatSession は AI チャットの 1 セッション
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
//
// Attachments はユーザー発話に紐付く画像 / ドキュメント等の添付ファイル群。
// S3 のオブジェクトキーとメタデータだけを保存し、実体バイト列 (BlobData) は
// DynamoDB / JSON への永続化対象外（Bedrock 呼び出し直前に S3 から取得して詰める）。
type AiChatMessage struct {
	SessionID   uint64       `json:"sessionId"`
	MessageID   string       `json:"messageId"`
	Role        string       `json:"role"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments,omitempty"`
	CreatedAt   time.Time    `json:"createdAt"`
}

// Attachment は AI チャット送信時の添付ファイル。
//
// Kind:
//   - "image": Bedrock の content[].image ブロックに詰めて vision 入力にする
//   - "document": 将来の PR-G2 で PDF / CSV を扱う際の値（PR-G1 では未使用）
//
// Format は AWS Bedrock Converse API が要求する短い文字列。
//   - 画像: "png" / "jpeg" / "gif" / "webp"
//   - ドキュメント (PR-G2 以降): "pdf" / "csv" / "txt" など
//
// SizeBytes は S3 アップロード後の確認用メタ。Bedrock 上限超過時は handler で 400 を返す。
//
// BlobData は usecase が Bedrock に渡す直前に S3 GetObject で詰める一時フィールド。
// JSON / DynamoDB には永続化しないため `json:"-"` と repository 側 dynamoItem 除外で扱う。
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
