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
