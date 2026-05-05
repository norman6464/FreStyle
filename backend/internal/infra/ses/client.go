// Package ses は AWS SES v2 SendEmail のラッパーを提供する Infra 層。
// 招待マジックリンクメールの送信に使う。
package ses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// Client は SES v2 SendEmail を呼び出す。
type Client struct {
	svc         *sesv2.Client
	fromAddress string
}

// NewClient は ECS Task Role / 環境変数の認証情報チェーンで SES クライアントを組み立てる。
// fromAddress は SES で検証済の送信元（例: "FreStyle <noreply@normanblog.com>"）を渡す。
func NewClient(ctx context.Context, region, fromAddress string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("ses: load aws config: %w", err)
	}
	return &Client{
		svc:         sesv2.NewFromConfig(cfg),
		fromAddress: fromAddress,
	}, nil
}

// SendInvitationEmail は招待マジックリンクメールを送る。
// SES サンドボックス中は to が「検証済 ID リスト」に登録されたアドレスのみ送れる点に注意。
func (c *Client) SendInvitationEmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	_, err := c.svc.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(c.fromAddress),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(subject),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data:    aws.String(htmlBody),
						Charset: aws.String("UTF-8"),
					},
					Text: &types.Content{
						Data:    aws.String(textBody),
						Charset: aws.String("UTF-8"),
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("ses send email: %w", err)
	}
	return nil
}
