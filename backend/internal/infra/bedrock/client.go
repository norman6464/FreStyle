// Package bedrock は AWS Bedrock Converse API のラッパーを提供する Infra 層。
package bedrock

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// Client は Bedrock Converse API を呼び出す。
type Client struct {
	svc     *bedrockruntime.Client
	modelID string
}

// NewClient は ECS Task Role / 環境変数の認証情報チェーンで Bedrock クライアントを組み立てる。
func NewClient(ctx context.Context, region, modelID string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("bedrock: load aws config: %w", err)
	}
	return &Client{
		svc:     bedrockruntime.NewFromConfig(cfg),
		modelID: modelID,
	}, nil
}

// Converse は会話履歴と新規ユーザーメッセージを Bedrock に送り、アシスタント返答テキストを返す。
// history には直近の会話（user/assistant 交互）を渡す。最後の要素はユーザーの最新発話。
func (c *Client) Converse(ctx context.Context, systemPrompt string, history []domain.AiChatMessage) (string, error) {
	messages := make([]brtypes.Message, 0, len(history))
	for _, m := range history {
		role := brtypes.ConversationRoleUser
		if m.Role == domain.AiChatRoleAssistant {
			role = brtypes.ConversationRoleAssistant
		}
		messages = append(messages, brtypes.Message{
			Role: role,
			Content: []brtypes.ContentBlock{
				&brtypes.ContentBlockMemberText{Value: m.Content},
			},
		})
	}

	input := &bedrockruntime.ConverseInput{
		ModelId:  &c.modelID,
		Messages: messages,
	}
	if systemPrompt != "" {
		input.System = []brtypes.SystemContentBlock{
			&brtypes.SystemContentBlockMemberText{Value: systemPrompt},
		}
	}

	output, err := c.svc.Converse(ctx, input)
	if err != nil {
		return "", fmt.Errorf("bedrock: converse: %w", err)
	}

	msg, ok := output.Output.(*brtypes.ConverseOutputMemberMessage)
	if !ok {
		return "", fmt.Errorf("bedrock: unexpected output type")
	}
	for _, block := range msg.Value.Content {
		if text, ok := block.(*brtypes.ContentBlockMemberText); ok {
			return text.Value, nil
		}
	}
	return "", fmt.Errorf("bedrock: no text content in response")
}

// StreamEvent は ConverseStream が emit する 1 イベント。
// Delta は途中チャンクのテキスト追加分、Done が true なら末尾（StopReason 受信時）。
// Err が non-nil なら ストリーム途中での失敗。
type StreamEvent struct {
	Delta string
	Done  bool
	Err   error
}

// ConverseStream は ConverseStream API を呼び出して、token を逐次 channel に流す。
//
// 呼び出し側は返ってきた channel を range でループするだけで token を消費できる:
//
//	ch, err := bc.ConverseStream(ctx, prompt, history)
//	if err != nil { ... }
//	for ev := range ch {
//	    if ev.Err != nil { ... break }
//	    if ev.Done { break }
//	    write(ev.Delta)
//	}
//
// ctx が cancel されると channel は閉じられる。AWS SDK の EventStream は内部で
// channel を持っているが、本ラッパでは「アプリケーションが扱いやすい形」に変換して提供する
// ことで infra 詳細（brtypes の各 Member 型）を usecase 層に漏らさない。
func (c *Client) ConverseStream(ctx context.Context, systemPrompt string, history []domain.AiChatMessage) (<-chan StreamEvent, error) {
	messages := make([]brtypes.Message, 0, len(history))
	for _, m := range history {
		role := brtypes.ConversationRoleUser
		if m.Role == domain.AiChatRoleAssistant {
			role = brtypes.ConversationRoleAssistant
		}
		messages = append(messages, brtypes.Message{
			Role: role,
			Content: []brtypes.ContentBlock{
				&brtypes.ContentBlockMemberText{Value: m.Content},
			},
		})
	}

	input := &bedrockruntime.ConverseStreamInput{
		ModelId:  &c.modelID,
		Messages: messages,
	}
	if systemPrompt != "" {
		input.System = []brtypes.SystemContentBlock{
			&brtypes.SystemContentBlockMemberText{Value: systemPrompt},
		}
	}

	output, err := c.svc.ConverseStream(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("bedrock: converse stream: %w", err)
	}

	out := make(chan StreamEvent, 16)
	go func() {
		defer close(out)
		defer output.GetStream().Close()

		stream := output.GetStream()
		for ev := range stream.Events() {
			switch v := ev.(type) {
			case *brtypes.ConverseStreamOutputMemberContentBlockDelta:
				if delta, ok := v.Value.Delta.(*brtypes.ContentBlockDeltaMemberText); ok {
					select {
					case out <- StreamEvent{Delta: delta.Value}:
					case <-ctx.Done():
						return
					}
				}
			case *brtypes.ConverseStreamOutputMemberMessageStop:
				// stop reason は本ラッパでは利用しないが、Done を立てて呼び出し側にループ終端を伝える。
				select {
				case out <- StreamEvent{Done: true}:
				case <-ctx.Done():
				}
				return
			case *brtypes.ConverseStreamOutputMemberMessageStart,
				*brtypes.ConverseStreamOutputMemberContentBlockStart,
				*brtypes.ConverseStreamOutputMemberContentBlockStop,
				*brtypes.ConverseStreamOutputMemberMetadata:
				// メタイベントは UI に流さない（メタデータは保存しない）
			}
		}
		if err := stream.Err(); err != nil {
			select {
			case out <- StreamEvent{Err: fmt.Errorf("bedrock: stream err: %w", err)}:
			case <-ctx.Done():
			}
		}
	}()

	return out, nil
}
