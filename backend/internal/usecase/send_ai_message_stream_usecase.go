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

// AttachmentDownloader は S3 から attachment のバイト列を取得する抽象。
// 本番では infra/s3.Downloader が満たす。usecase テストでは in-memory 実装を差し替える。
//
// nil でも動作する（その場合は添付なしと同等の振る舞い）。Bedrock / DynamoDB の初期化失敗
// などで一時的に S3 利用不可になったときも、テキストだけは送れるよう優雅に degrade する。
type AttachmentDownloader interface {
	Download(ctx context.Context, key string) ([]byte, error)
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
// 互換のため両方を残し、フロントが対応した順から SSE に切り替える。 章 006 で SSE / Bedrock
// streaming を 解説。
//
// 依存 port: [repository.AiChatSessionRepository] (RDB session 確認) +
// [repository.AiChatMessageRepository] (DynamoDB メッセージ 永続化) +
// [StreamingBedrockClient] (Bedrock token 配信) + [AttachmentDownloader] (S3 添付 取得、 nil OK)。
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

		// Bedrock 呼び出し直前に最新ユーザー発話の attachment 実体を S3 から取得して詰める。
		// 過去履歴の画像は再送しない（料金 / latency / Bedrock 上限への配慮）。
		// 取得失敗した attachment は除外して text だけで送る（部分劣化）。
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

		// Bedrock が token を 1 つも emit しなかった場合は空のアシスタントメッセージに
		// なる。これを DDB に保存すると将来の history 構築時に「連続 user → 空応答」の
		// 負のループに陥るため、空応答は保存せずエラーとしてフロントに通知する。
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

// lastUserMessageIndex は history 末尾から最も近い user メッセージの index を返す。
// 見つからなければ -1。最新ユーザー発話だけに今回のリクエスト由来 attachments を貼る用途。
func lastUserMessageIndex(history []domain.AiChatMessage) int {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == domain.AiChatRoleUser {
			return i
		}
	}
	return -1
}

// fetchAttachmentBlobs は in.Attachments それぞれを S3 から GetObject して BlobData を埋める。
//
// 取得失敗した attachment はスライスから落とす（その 1 枚だけ Bedrock に渡らない）。
// すべて失敗したら空スライスを返し、Bedrock はテキストのみで応答する。
// SizeBytes が 0 の attachment は metadata 不整合（フロント側のバグ）なのでスキップする。
func (u *SendAiMessageStreamUseCase) fetchAttachmentBlobs(ctx context.Context, in []domain.Attachment) []domain.Attachment {
	out := make([]domain.Attachment, 0, len(in))
	for _, a := range in {
		if a.Key == "" {
			continue
		}
		data, err := u.attachments.Download(ctx, a.Key)
		if err != nil {
			// 1 件の失敗は致命でない。ログだけ残して続行（Bedrock には残りの添付＋テキストで送る）。
			log.Printf("ai-chat attachment download failed: key=%s err=%v", a.Key, err)
			continue
		}
		a.BlobData = data
		out = append(out, a)
	}
	return out
}
