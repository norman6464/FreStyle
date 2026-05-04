package cognito

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// ErrUserAlreadyConfirmed は既にパスワード変更済み（登録完了済み）のメールアドレスへの再招待を試みた場合に返る。
var ErrUserAlreadyConfirmed = errors.New("user already confirmed")

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

// InviteUser は Cognito AdminCreateUser で招待メールを送信し、CognitoSub を返す。
// 同一メールアドレスでの再招待にも対応:
//   - 未確認ユーザー (FORCE_CHANGE_PASSWORD): MessageAction=RESEND で招待メールを再送
//   - 確認済みユーザー (CONFIRMED): ErrUserAlreadyConfirmed を返す
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
		UserPoolId:     aws.String(c.userPoolID),
		Username:       aws.String(email),
		UserAttributes: attrs,
		DesiredDeliveryMediums: []types.DeliveryMediumType{
			types.DeliveryMediumTypeEmail,
		},
	})
	if err == nil {
		return extractSub(out.User), nil
	}

	var existsErr *types.UsernameExistsException
	if !errors.As(err, &existsErr) {
		return "", fmt.Errorf("cognito admin create user: %w", err)
	}

	// 既存ユーザーの状態を確認して再送信できるか判定する
	got, getErr := c.client.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(email),
	})
	if getErr != nil {
		return "", fmt.Errorf("cognito admin get user: %w", getErr)
	}
	if got.UserStatus == types.UserStatusTypeConfirmed {
		return "", ErrUserAlreadyConfirmed
	}

	// 未確認状態（FORCE_CHANGE_PASSWORD など）なら招待を再送信
	resendOut, resendErr := c.client.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:    aws.String(c.userPoolID),
		Username:      aws.String(email),
		MessageAction: types.MessageActionTypeResend,
		DesiredDeliveryMediums: []types.DeliveryMediumType{
			types.DeliveryMediumTypeEmail,
		},
	})
	if resendErr != nil {
		return "", fmt.Errorf("cognito resend invitation: %w", resendErr)
	}
	return extractSub(resendOut.User), nil
}

func extractSub(user *types.UserType) string {
	if user == nil {
		return ""
	}
	for _, attr := range user.Attributes {
		if aws.ToString(attr.Name) == "sub" {
			return aws.ToString(attr.Value)
		}
	}
	return aws.ToString(user.Username)
}
