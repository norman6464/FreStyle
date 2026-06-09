package cognito

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	cip "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// ErrInvalidCredentials は email / password 不一致・未確認ユーザー等で認証に失敗したときに返す（401 用）。
// 真の理由（ユーザー不在 か パスワード違い か）はユーザー列挙対策のため呼び元に伝えない。
var ErrInvalidCredentials = errors.New("cognito: invalid credentials")

// initiateAuthAPI は InitiateAuth だけを使う最小インターフェイス（テストで fake に差し替えるため）。
type initiateAuthAPI interface {
	InitiateAuth(ctx context.Context, in *cip.InitiateAuthInput, optFns ...func(*cip.Options)) (*cip.InitiateAuthOutput, error)
}

// PasswordAuthenticator は Cognito の USER_PASSWORD_AUTH フローで email / password を検証し、
// access / id / refresh token を取得する。client secret 付きクライアント用に SECRET_HASH を計算する。
// Hosted UI(認可コード交換)とは別経路で、フロントの「メール / パスワードフォーム」から使う。
type PasswordAuthenticator struct {
	client       initiateAuthAPI
	clientID     string
	clientSecret string
}

// NewPasswordAuthenticator は AWS 既定の認証情報チェーンで cognitoidp クライアントを組み立てる。
func NewPasswordAuthenticator(ctx context.Context, region, clientID, clientSecret string) (*PasswordAuthenticator, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &PasswordAuthenticator{
		client:       cip.NewFromConfig(awsCfg),
		clientID:     clientID,
		clientSecret: clientSecret,
	}, nil
}

// newPasswordAuthenticatorWithClient はテスト用に InitiateAuth クライアントを差し替える DI コンストラクタ。
func newPasswordAuthenticatorWithClient(client initiateAuthAPI, clientID, clientSecret string) *PasswordAuthenticator {
	return &PasswordAuthenticator{client: client, clientID: clientID, clientSecret: clientSecret}
}

// Authenticate は email / password を Cognito で検証し Token を返す。
// 資格情報誤りは ErrInvalidCredentials、未設定は ErrNotConfigured、それ以外は元のエラーをラップして返す。
func (a *PasswordAuthenticator) Authenticate(ctx context.Context, email, password string) (*Token, error) {
	if a.clientID == "" {
		return nil, ErrNotConfigured
	}

	params := map[string]string{
		"USERNAME": email,
		"PASSWORD": password,
	}
	if a.clientSecret != "" {
		params["SECRET_HASH"] = a.secretHash(email)
	}

	out, err := a.client.InitiateAuth(ctx, &cip.InitiateAuthInput{
		AuthFlow:       types.AuthFlowTypeUserPasswordAuth,
		ClientId:       aws.String(a.clientID),
		AuthParameters: params,
	})
	if err != nil {
		// ユーザー不在 / パスワード違い / 未確認は資格情報エラーに丸めて 401 にする（ユーザー列挙対策）。
		var notAuth *types.NotAuthorizedException
		var notFound *types.UserNotFoundException
		var notConfirmed *types.UserNotConfirmedException
		if errors.As(err, &notAuth) || errors.As(err, &notFound) || errors.As(err, &notConfirmed) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// MFA 等のチャレンジは本アプリの Cognito 設定では発生しない想定。AuthenticationResult が無ければ失敗扱い。
	if out.AuthenticationResult == nil {
		return nil, ErrInvalidCredentials
	}
	r := out.AuthenticationResult
	return &Token{
		AccessToken:  aws.ToString(r.AccessToken),
		IDToken:      aws.ToString(r.IdToken),
		RefreshToken: aws.ToString(r.RefreshToken),
		ExpiresIn:    int(r.ExpiresIn),
		TokenType:    aws.ToString(r.TokenType),
	}, nil
}

// secretHash は Cognito の SECRET_HASH = Base64(HMAC-SHA256(username + clientId, clientSecret)) を計算する。
func (a *PasswordAuthenticator) secretHash(username string) string {
	mac := hmac.New(sha256.New, []byte(a.clientSecret))
	mac.Write([]byte(username + a.clientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
