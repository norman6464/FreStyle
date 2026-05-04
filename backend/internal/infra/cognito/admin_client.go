package cognito

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// AdminClient は Cognito User Pool の管理 API（AdminCreateUser など）を担う。
type AdminClient struct {
	client     *cognitoidentityprovider.Client
	userPoolID string
}

// NewAdminClient は AWS デフォルト認証情報チェーン（IAM ロール）で AdminClient を初期化する。
func NewAdminClient(ctx context.Context, region, userPoolID string) (*AdminClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("cognito admin client: load aws config: %w", err)
	}
	return &AdminClient{
		client:     cognitoidentityprovider.NewFromConfig(cfg),
		userPoolID: userPoolID,
	}, nil
}

// InviteUser は Cognito AdminCreateUser で一時パスワード付き招待メールを送信し、CognitoSub を返す。
// Cognito の招待メールテンプレートはユーザープール設定で管理する。
func (c *AdminClient) InviteUser(ctx context.Context, email, displayName, _ string) (string, error) {
	attrs := []types.AttributeType{
		{Name: aws.String("email"), Value: aws.String(email)},
		{Name: aws.String("email_verified"), Value: aws.String("true")},
	}
	if displayName != "" {
		attrs = append(attrs, types.AttributeType{
			Name:  aws.String("name"),
			Value: aws.String(displayName),
		})
	}

	out, err := c.client.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        aws.String(c.userPoolID),
		Username:          aws.String(email),
		UserAttributes:    attrs,
		DesiredDeliveryMediums: []types.DeliveryMediumType{
			types.DeliveryMediumTypeEmail,
		},
		// MessageAction を省略することで Cognito が招待メールを自動送信する
	})
	if err != nil {
		return "", fmt.Errorf("cognito admin create user: %w", err)
	}
	if out.User == nil {
		return "", fmt.Errorf("cognito admin create user: nil user in response")
	}

	for _, attr := range out.User.Attributes {
		if aws.ToString(attr.Name) == "sub" {
			return aws.ToString(attr.Value), nil
		}
	}
	return aws.ToString(out.User.Username), nil
}
