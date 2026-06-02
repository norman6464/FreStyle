package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/bedrock"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// StreamingBedrockClient は Bedrock の streaming 呼び出し抽象。
// テストでは fake を差し替える。infra/bedrock.Client が満たす想定。
type StreamingBedrockClient interface {
	ConverseStream(ctx context.Context, systemPrompt string, history []domain.AiChatMessage) (<-chan bedrock.StreamEvent, error)
}

// AttachmentDownloader は S3 から attachment のバイト列を取得する抽象（infra/s3.Downloader が満たす）。
// nil でも動作し、その場合は添付なしと同等にテキストだけ送る（優雅に degrade）。
type AttachmentDownloader interface {
	Download(ctx context.Context, key string) ([]byte, error)
}

// StreamEvent は handler に流すイベント。token 配信に加え、session 作成 / message 完成などの状態も返す。
type StreamEvent struct {
	Delta        string                // 途中チャンク（token 追加）
	NewSession   *domain.AiChatSession // 新規セッション作成時のみ（最初の 1 回）
	FinalMessage *domain.AiChatMessage // アシスタント返答が完成し DB 保存済のとき（末尾で 1 回）
	Err          error
}

// SendAiMessageStreamUseCase は AI チャットの SSE streaming 版。
// session 確認 → メッセージ保存 → Bedrock streaming → 返答保存、をオーケストレーションする。
type SendAiMessageStreamUseCase struct {
	sessions      repository.AiChatSessionRepository
	messages      repository.AiChatMessageRepository
	bedrockClient StreamingBedrockClient
	attachments   AttachmentDownloader
}

func NewSendAiMessageStreamUseCase(
	sessions repository.AiChatSessionRepository,
	messages repository.AiChatMessageRepository,
	bc StreamingBedrockClient,
	dl AttachmentDownloader,
) *SendAiMessageStreamUseCase {
	return &SendAiMessageStreamUseCase{
		sessions:      sessions,
		messages:      messages,
		bedrockClient: bc,
		attachments:   dl,
	}
}

// Execute は handler がレスポンス書き込みに使う StreamEvent channel を返す。
// 新規セッション作成 → ユーザーメッセージ保存 → Bedrock streaming → 返答保存、を goroutine で進める。
// ctx の cancel で全工程が中断される（クライアント切断時の goroutine リーク防止）。
func (u *SendAiMessageStreamUseCase) Execute(ctx context.Context, in SendAiMessageInput) (<-chan StreamEvent, error) {
	out := make(chan StreamEvent, 16)

	go func() {
		defer close(out)

		sessionID := in.SessionID
		var newSession *domain.AiChatSession
		if sessionID == 0 {
			// 添付のみで本文が空の場合に session タイトルが空文字にならないようフォールバック。
			title := truncateTitle(in.Content, 30)
			if title == "" {
				if len(in.Attachments) > 0 {
					title = "添付ファイルを送信"
				} else {
					title = "新しいチャット"
				}
			}
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
			SessionID:   sessionID,
			MessageID:   uuid.New().String(),
			Role:        domain.AiChatRoleUser,
			Content:     in.Content,
			Attachments: in.Attachments,
			CreatedAt:   time.Now().UTC(),
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

		// 最新ユーザー発話の attachment 実体だけを S3 から詰める（過去履歴の画像は再送しない）。
		if u.attachments != nil && len(in.Attachments) > 0 {
			if last := lastUserMessageIndex(history); last >= 0 {
				history[last].Attachments = u.fetchAttachmentBlobs(ctx, in.Attachments)
			}
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

		// 空応答を保存すると history が「連続 user → 空応答」で壊れるため、保存せずエラー通知する。
		if assistantBuf.Len() == 0 {
			emit(ctx, out, StreamEvent{Err: fmt.Errorf("bedrock returned empty response")})
			return
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

// lastUserMessageIndex は history 末尾から最も近い user メッセージの index を返す（無ければ -1）。
func lastUserMessageIndex(history []domain.AiChatMessage) int {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == domain.AiChatRoleUser {
			return i
		}
	}
	return -1
}

// fetchAttachmentBlobs は各 attachment を S3 から取得して BlobData を埋める。
// 取得失敗した分は落とし、残りの添付 + テキストで送る（部分劣化）。
func (u *SendAiMessageStreamUseCase) fetchAttachmentBlobs(ctx context.Context, in []domain.Attachment) []domain.Attachment {
	out := make([]domain.Attachment, 0, len(in))
	for _, a := range in {
		if a.Key == "" {
			continue
		}
		data, err := u.attachments.Download(ctx, a.Key)
		if err != nil {
			log.Printf("ai-chat attachment download failed: key=%s err=%v", a.Key, err)
			continue
		}
		a.BlobData = data
		out = append(out, a)
	}
	return out
}
