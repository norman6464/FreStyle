// Package cognito の AdminCreateUser ラッパー。招待の「初期パスワード方式」で、
// 管理者が一時パスワード付きの Cognito ユーザーを作る（FRESTYLE-313）。
package cognito

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	cip "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// ErrUserAlreadyExists は AdminCreateUser 対象の email が既に存在するときに返す（409 用）。
var ErrUserAlreadyExists = errors.New("cognito: user already exists")

// adminCreateUserAPI は AdminCreateUser だけを使う最小インターフェイス（テストで fake 差し替え）。
type adminCreateUserAPI interface {
	AdminCreateUser(ctx context.Context, in *cip.AdminCreateUserInput, optFns ...func(*cip.Options)) (*cip.AdminCreateUserOutput, error)
}

// AdminUserCreator は User Pool に一時パスワード付きユーザーを作る。
// 一時パスワードはこの型が生成し、呼び元へ 1 度だけ返す（保存・ログ出力しない）。
// Cognito 側の自動メールは MessageAction=SUPPRESS で止める（提示はアプリが制御する）。
type AdminUserCreator struct {
	client     adminCreateUserAPI
	userPoolID string
}

// NewAdminUserCreator は AWS 既定の認証情報チェーンで組み立てる。userPoolID が空なら
// エラー（本番では COGNITO_USER_POOL_ID が必須）。
func NewAdminUserCreator(ctx context.Context, region, userPoolID string) (*AdminUserCreator, error) {
	if userPoolID == "" {
		return nil, errors.New("cognito: user pool id is required for AdminCreateUser")
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &AdminUserCreator{client: cip.NewFromConfig(awsCfg), userPoolID: userPoolID}, nil
}

// newAdminUserCreatorWithClient はテスト用 DI コンストラクタ。
func newAdminUserCreatorWithClient(client adminCreateUserAPI, userPoolID string) *AdminUserCreator {
	return &AdminUserCreator{client: client, userPoolID: userPoolID}
}

// CreateWithTemporaryPassword は email のユーザーを一時パスワード付きで作成し、その一時パスワードを返す。
// 初回ログイン時に Cognito が NEW_PASSWORD_REQUIRED を要求し、本人が新パスワードへ変更する
// （管理者が恒久パスワードを知る状態を作らない）。既存 email は ErrUserAlreadyExists。
func (a *AdminUserCreator) CreateWithTemporaryPassword(ctx context.Context, email, name string) (temporaryPassword string, err error) {
	tempPw, err := generateTemporaryPassword()
	if err != nil {
		return "", fmt.Errorf("generate temporary password: %w", err)
	}
	attrs := []types.AttributeType{
		{Name: aws.String("email"), Value: aws.String(email)},
		{Name: aws.String("email_verified"), Value: aws.String("true")},
	}
	if name != "" {
		attrs = append(attrs, types.AttributeType{Name: aws.String("name"), Value: aws.String(name)})
	}
	_, err = a.client.AdminCreateUser(ctx, &cip.AdminCreateUserInput{
		UserPoolId:        aws.String(a.userPoolID),
		Username:          aws.String(email),
		TemporaryPassword: aws.String(tempPw),
		// Cognito 側の招待メールは送らない（アプリが提示・送付を制御する）。
		MessageAction:  types.MessageActionTypeSuppress,
		UserAttributes: attrs,
	})
	if err != nil {
		var exists *types.UsernameExistsException
		if errors.As(err, &exists) {
			return "", ErrUserAlreadyExists
		}
		return "", fmt.Errorf("admin create user: %w", err)
	}
	return tempPw, nil
}

// tempPasswordLength / 文字集合は一般的な Cognito パスワードポリシー
// （8 文字以上・大文字・小文字・数字・記号）を確実に満たす長さ・構成にする。
const tempPasswordLength = 16

const (
	pwLower  = "abcdefghijkmnpqrstuvwxyz" // 紛らわしい l/o を除く
	pwUpper  = "ABCDEFGHJKLMNPQRSTUVWXYZ" // 紛らわしい I/O を除く
	pwDigit  = "23456789"                 // 紛らわしい 0/1 を除く
	pwSymbol = "!@#$%^&*-_=+"
)

// generateTemporaryPassword は各文字種を最低 1 つ含む暗号論的乱数のパスワードを生成する。
func generateTemporaryPassword() (string, error) {
	all := pwLower + pwUpper + pwDigit + pwSymbol
	// まず各文字種から 1 文字ずつ確保し、残りを全集合から埋める（ポリシー確実充足）。
	buf := make([]byte, 0, tempPasswordLength)
	for _, set := range []string{pwLower, pwUpper, pwDigit, pwSymbol} {
		ch, err := randChar(set)
		if err != nil {
			return "", err
		}
		buf = append(buf, ch)
	}
	for len(buf) < tempPasswordLength {
		ch, err := randChar(all)
		if err != nil {
			return "", err
		}
		buf = append(buf, ch)
	}
	// 先頭が必ず小文字…の偏りを消すためシャッフルする。
	if err := shuffle(buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func randChar(set string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
	if err != nil {
		return 0, err
	}
	return set[n.Int64()], nil
}

// shuffle は Fisher-Yates で暗号論的乱数を使って並べ替える。
func shuffle(b []byte) error {
	for i := len(b) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		b[i], b[j.Int64()] = b[j.Int64()], b[i]
	}
	return nil
}
