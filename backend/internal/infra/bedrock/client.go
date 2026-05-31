// Package bedrock は AWS Bedrock Converse API のラッパーを提供する Infra 層。
package bedrock

import (
	"context"
	"fmt"
	"strings"

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
	messages := buildBedrockMessages(history)

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

// ConverseStream は ConverseStream API の token を逐次 StreamEvent channel に流す。
// SDK の EventStream を usecase が扱いやすい形に変換し、brtypes の詳細を漏らさない。
// ctx が cancel されると channel は閉じる。
func (c *Client) ConverseStream(ctx context.Context, systemPrompt string, history []domain.AiChatMessage) (<-chan StreamEvent, error) {
	messages := buildBedrockMessages(history)

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

// buildBedrockMessages は domain.AiChatMessage を Bedrock Converse の Message 列に変換する。
// 添付は image / document ブロックとして text の前に並べ、BlobData が空のものは無視する。
// Bedrock は user/assistant の交互を要求するため、空メッセージは skip せず placeholder を入れる
// （skip すると連続 user → 空応答の負ループに陥るため）。
func buildBedrockMessages(history []domain.AiChatMessage) []brtypes.Message {
	messages := make([]brtypes.Message, 0, len(history))
	for _, m := range history {
		role := brtypes.ConversationRoleUser
		if m.Role == domain.AiChatRoleAssistant {
			role = brtypes.ConversationRoleAssistant
		}
		blocks := make([]brtypes.ContentBlock, 0, len(m.Attachments)+1)
		for _, a := range m.Attachments {
			block := attachmentToBlock(a)
			if block != nil {
				blocks = append(blocks, block)
			}
		}
		text := m.Content
		if text == "" {
			switch {
			case hasDocumentBlock(blocks):
				// document ブロックは text 必須なので placeholder を付ける。
				text = "(添付ファイルを参照してください)"
			case len(blocks) > 0:
				// image だけのメッセージは Bedrock が text 無しでも受け付ける。
				// 何も追加しない（text "" のまま skip される）。
			default:
				// 何も無いメッセージ。役割交互制約のため placeholder を入れる。
				if role == brtypes.ConversationRoleAssistant {
					text = "(応答が空でした)"
				} else {
					text = "(空のメッセージ)"
				}
			}
		}
		if text != "" {
			blocks = append(blocks, &brtypes.ContentBlockMemberText{Value: text})
		}
		if len(blocks) == 0 {
			// image only かつ BlobData も無い → 何もない実質空メッセージ。skip。
			continue
		}
		messages = append(messages, brtypes.Message{Role: role, Content: blocks})
	}
	return messages
}

func hasDocumentBlock(blocks []brtypes.ContentBlock) bool {
	for _, b := range blocks {
		if _, ok := b.(*brtypes.ContentBlockMemberDocument); ok {
			return true
		}
	}
	return false
}

// attachmentToBlock は 1 つの添付を Bedrock の ContentBlock に変換する。
// BlobData が空（S3 取得失敗 / 未取得）なら nil を返してスキップする。
func attachmentToBlock(a domain.Attachment) brtypes.ContentBlock {
	if len(a.BlobData) == 0 {
		return nil
	}
	switch a.Kind {
	case domain.AttachmentKindImage:
		return &brtypes.ContentBlockMemberImage{
			Value: brtypes.ImageBlock{
				Format: brtypes.ImageFormat(a.Format),
				Source: &brtypes.ImageSourceMemberBytes{Value: a.BlobData},
			},
		}
	case domain.AttachmentKindDocument:
		return &brtypes.ContentBlockMemberDocument{
			Value: brtypes.DocumentBlock{
				Format: brtypes.DocumentFormat(a.Format),
				Name:   strPtr(documentName(a.Filename)),
				Source: &brtypes.DocumentSourceMemberBytes{Value: a.BlobData},
			},
		}
	}
	return nil
}

func strPtr(s string) *string { return &s }

// documentName は Bedrock document.name 用の identifier を返す。
// Bedrock 仕様: 1〜200 文字、英数字 / 空白 / ハイフン / 括弧 / 角括弧 のみ許可。
// 拡張子は除外し、許可外文字は '-' に置換する。空になったら "document" にフォールバック。
func documentName(filename string) string {
	base := filename
	if i := strings.LastIndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	out := make([]byte, 0, len(base))
	for i := 0; i < len(base); i++ {
		c := base[i]
		switch {
		case c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == ' ', c == '-', c == '(', c == ')', c == '[', c == ']':
			out = append(out, c)
		default:
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "document"
	}
	if len(out) > 200 {
		out = out[:200]
	}
	return string(out)
}
