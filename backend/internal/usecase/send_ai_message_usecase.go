package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// BedrockClient は Bedrock Converse API の呼び出し抽象。テストでモック差し替えを可能にする。
type BedrockClient interface {
	Converse(ctx context.Context, systemPrompt string, history []domain.AiChatMessage) (string, error)
}

// SendAiMessageInput は WebSocket から受け取るメッセージ送信リクエスト。
type SendAiMessageInput struct {
	UserID      uint64
	SessionID   uint64
	Content     string
	Scene       string
	SessionType string
	ScenarioID  *uint64
}

// SendAiMessageOutput は WebSocket ハンドラに返す結果。
type SendAiMessageOutput struct {
	NewSession *domain.AiChatSession
	UserMsg    *domain.AiChatMessage
	AiMsg      *domain.AiChatMessage
}

// SendAiMessageUseCase はユーザー発話を受け取り、Bedrock に送って返答を保存するユースケース。
type SendAiMessageUseCase struct {
	sessions      repository.AiChatSessionRepository
	messages      repository.AiChatMessageRepository
	bedrockClient BedrockClient
}

func NewSendAiMessageUseCase(
	sessions repository.AiChatSessionRepository,
	messages repository.AiChatMessageRepository,
	bc BedrockClient,
) *SendAiMessageUseCase {
	return &SendAiMessageUseCase{sessions: sessions, messages: messages, bedrockClient: bc}
}

func (u *SendAiMessageUseCase) Execute(ctx context.Context, in SendAiMessageInput) (*SendAiMessageOutput, error) {
	var newSession *domain.AiChatSession
	sessionID := in.SessionID

	if sessionID == 0 {
		title := truncateTitle(in.Content, 30)
		s, err := NewCreateAiChatSessionUseCase(u.sessions).Execute(ctx, CreateAiChatSessionInput{
			UserID:      in.UserID,
			Title:       title,
			SessionType: in.SessionType,
			ScenarioID:  in.ScenarioID,
		})
		if err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
		newSession = s
		sessionID = s.ID
	}

	userMsg := &domain.AiChatMessage{
		SessionID: sessionID,
		MessageID: uuid.New().String(),
		Role:      domain.AiChatRoleUser,
		Content:   in.Content,
		CreatedAt: time.Now().UTC(),
	}
	if err := u.messages.Save(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("save user message: %w", err)
	}

	history, err := u.messages.ListBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}

	systemPrompt := buildSystemPrompt(in.SessionType, in.Scene)
	aiText, err := u.bedrockClient.Converse(ctx, systemPrompt, history)
	if err != nil {
		return nil, fmt.Errorf("bedrock converse: %w", err)
	}

	aiMsg := &domain.AiChatMessage{
		SessionID: sessionID,
		MessageID: uuid.New().String(),
		Role:      domain.AiChatRoleAssistant,
		Content:   aiText,
		CreatedAt: time.Now().UTC(),
	}
	if err := u.messages.Save(ctx, aiMsg); err != nil {
		return nil, fmt.Errorf("save ai message: %w", err)
	}

	return &SendAiMessageOutput{NewSession: newSession, UserMsg: userMsg, AiMsg: aiMsg}, nil
}

func truncateTitle(s string, max int) string {
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max]) + "…"
	}
	return s
}

// buildSystemPrompt は汎用 AI チャットの system prompt を返す。
//
// 旧版は「ビジネスコミュニケーションのコーチ」「ロールプレイ練習」など、
// 練習モード前提のプロンプトを混ぜていたが、コア機能整理（PR-A）で練習モード
// 自体を撤去したため、汎用 AI アシスタント向けのプロンプトに書き換える。
//
// sessionType / scene 引数は WebSocket 受信側との互換のため残しているが、
// 現状はどちらもプロンプトに反映しない（次の PR で削除予定）。
func buildSystemPrompt(_, _ string) string {
	return "あなたは新卒 IT エンジニア向け学習プラットフォーム『FreStyle』に組み込まれた汎用 AI アシスタントです。" +
		"ユーザーからの質問・要約依頼・コードレビュー・概念説明など、幅広いトピックに答えます。" +
		"日本語で簡潔・丁寧に応答してください。" +
		"回答は **Markdown 形式** で返してください（見出し / 箇条書き / 表 / コードブロックなど、内容に応じて適切に使い分けます）。" +
		"コードを示すときは ```言語名 で囲み、シンタックスハイライトが効くようにしてください。"
}
