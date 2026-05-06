package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/bedrock"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// StreamingBedrockClient は Bedrock の streaming 呼び出し抽象。
// テストでは fake を差し替える。infra/bedrock.Client が満たす想定。
type StreamingBedrockClient interface {
	ConverseStream(ctx context.Context, systemPrompt string, history []domain.AiChatMessage) (<-chan bedrock.StreamEvent, error)
}

// StreamEvent は handler に流すイベント。bedrock.StreamEvent と同形だが、
// ここでは「session 作成済」「message 完成済」など usecase レベルの状態も合わせて返す。
type StreamEvent struct {
	// Delta が non-empty なら token 追加（途中チャンク）
	Delta string
	// NewSession が non-nil なら新規セッションが作成されたことを示す（最初の 1 回のみ）
	NewSession *domain.AiChatSession
	// FinalMessage が non-nil ならアシスタント返答が完成して DB 保存済（末尾で 1 回のみ）
	FinalMessage *domain.AiChatMessage
	// Err が non-nil ならエラー終了
	Err error
}

// SendAiMessageStreamUseCase は WebSocket 版 (SendAiMessageUseCase) の streaming 版。
// 互換のため両方を残し、フロントが対応した順から SSE に切り替える。
type SendAiMessageStreamUseCase struct {
	sessions      repository.AiChatSessionRepository
	messages      repository.AiChatMessageRepository
	bedrockClient StreamingBedrockClient
}

func NewSendAiMessageStreamUseCase(
	sessions repository.AiChatSessionRepository,
	messages repository.AiChatMessageRepository,
	bc StreamingBedrockClient,
) *SendAiMessageStreamUseCase {
	return &SendAiMessageStreamUseCase{sessions: sessions, messages: messages, bedrockClient: bc}
}

// Execute は handler 側でレスポンス書き込みに使う channel を返す。
//
// 流れ:
//  1. sessionID == 0 なら新規セッション作成 → NewSession イベントを emit
//  2. ユーザーメッセージを DB に保存 → 履歴を取得
//  3. Bedrock ConverseStream を呼び、 token を Delta イベントで emit
//  4. token を全て連結したアシスタント返答を DB に保存 → FinalMessage イベントを emit
//  5. channel を close
//
// ctx の cancel で全工程が中断される（クライアント切断時の goroutine リーク防止）。
func (u *SendAiMessageStreamUseCase) Execute(ctx context.Context, in SendAiMessageInput) (<-chan StreamEvent, error) {
	out := make(chan StreamEvent, 16)

	go func() {
		defer close(out)

		sessionID := in.SessionID
		var newSession *domain.AiChatSession
		if sessionID == 0 {
			title := truncateTitle(in.Content, 30)
			s, err := NewCreateAiChatSessionUseCase(u.sessions).Execute(ctx, CreateAiChatSessionInput{
				UserID:      in.UserID,
				Title:       title,
				SessionType: in.SessionType,
				ScenarioID:  in.ScenarioID,
			})
			if err != nil {
				emit(ctx, out, StreamEvent{Err: fmt.Errorf("create session: %w", err)})
				return
			}
			newSession = s
			sessionID = s.ID
			emit(ctx, out, StreamEvent{NewSession: newSession})
		}

		userMsg := &domain.AiChatMessage{
			SessionID: sessionID,
			MessageID: uuid.New().String(),
			Role:      domain.AiChatRoleUser,
			Content:   in.Content,
			CreatedAt: time.Now().UTC(),
		}
		if err := u.messages.Save(ctx, userMsg); err != nil {
			emit(ctx, out, StreamEvent{Err: fmt.Errorf("save user message: %w", err)})
			return
		}

		history, err := u.messages.ListBySessionID(ctx, sessionID)
		if err != nil {
			emit(ctx, out, StreamEvent{Err: fmt.Errorf("list messages: %w", err)})
			return
		}

		systemPrompt := buildSystemPrompt(in.SessionType, in.Scene)
		stream, err := u.bedrockClient.ConverseStream(ctx, systemPrompt, history)
		if err != nil {
			emit(ctx, out, StreamEvent{Err: fmt.Errorf("bedrock converse stream: %w", err)})
			return
		}

		var assistantBuf strings.Builder
		for ev := range stream {
			if ev.Err != nil {
				emit(ctx, out, StreamEvent{Err: ev.Err})
				return
			}
			if ev.Delta != "" {
				assistantBuf.WriteString(ev.Delta)
				emit(ctx, out, StreamEvent{Delta: ev.Delta})
			}
			if ev.Done {
				break
			}
		}

		final := &domain.AiChatMessage{
			SessionID: sessionID,
			MessageID: uuid.New().String(),
			Role:      domain.AiChatRoleAssistant,
			Content:   assistantBuf.String(),
			CreatedAt: time.Now().UTC(),
		}
		if err := u.messages.Save(ctx, final); err != nil {
			emit(ctx, out, StreamEvent{Err: fmt.Errorf("save ai message: %w", err)})
			return
		}
		emit(ctx, out, StreamEvent{FinalMessage: final})
	}()

	return out, nil
}

// emit は ctx 終了時に block しないように選択的送信する。
func emit(ctx context.Context, out chan<- StreamEvent, ev StreamEvent) {
	select {
	case out <- ev:
	case <-ctx.Done():
	}
}
