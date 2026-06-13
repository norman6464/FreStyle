package cognito

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cip "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// fakeInitiateAuth は initiateAuthAPI のテスト用スタブ。
type fakeInitiateAuth struct {
	out      *cip.InitiateAuthOutput
	err      error
	gotInput *cip.InitiateAuthInput
}

func (f *fakeInitiateAuth) InitiateAuth(_ context.Context, in *cip.InitiateAuthInput, _ ...func(*cip.Options)) (*cip.InitiateAuthOutput, error) {
	f.gotInput = in
	return f.out, f.err
}

func Test_パスワード認証_成功(t *testing.T) {
	fake := &fakeInitiateAuth{out: &cip.InitiateAuthOutput{
		AuthenticationResult: &types.AuthenticationResultType{
			AccessToken:  aws.String("AT"),
			IdToken:      aws.String("ID"),
			RefreshToken: aws.String("RT"),
			ExpiresIn:    3600,
			TokenType:    aws.String("Bearer"),
		},
	}}
	a := newPasswordAuthenticatorWithClient(fake, "client-id", "secret")

	tok, err := a.Authenticate(context.Background(), "u@example.com", "pw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "AT" || tok.IDToken != "ID" || tok.RefreshToken != "RT" || tok.ExpiresIn != 3600 {
		t.Errorf("unexpected token: %+v", tok)
	}
	// USER_PASSWORD_AUTH フロー + SECRET_HASH が渡る。
	if fake.gotInput.AuthFlow != types.AuthFlowTypeUserPasswordAuth {
		t.Errorf("want USER_PASSWORD_AUTH, got %v", fake.gotInput.AuthFlow)
	}
	if fake.gotInput.AuthParameters["USERNAME"] != "u@example.com" {
		t.Errorf("USERNAME not passed: %v", fake.gotInput.AuthParameters)
	}
	if fake.gotInput.AuthParameters["SECRET_HASH"] == "" {
		t.Errorf("expected SECRET_HASH when client secret is set")
	}
}

func Test_パスワード認証_認証情報不正(t *testing.T) {
	for _, awsErr := range []error{
		&types.NotAuthorizedException{},
		&types.UserNotFoundException{},
		&types.UserNotConfirmedException{},
	} {
		fake := &fakeInitiateAuth{err: awsErr}
		a := newPasswordAuthenticatorWithClient(fake, "client-id", "secret")
		if _, err := a.Authenticate(context.Background(), "u", "pw"); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("%T should map to ErrInvalidCredentials, got %v", awsErr, err)
		}
	}
}

func Test_パスワード認証_secretなしはSecretHashなし(t *testing.T) {
	fake := &fakeInitiateAuth{out: &cip.InitiateAuthOutput{
		AuthenticationResult: &types.AuthenticationResultType{AccessToken: aws.String("AT")},
	}}
	a := newPasswordAuthenticatorWithClient(fake, "client-id", "")
	if _, err := a.Authenticate(context.Background(), "u", "pw"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := fake.gotInput.AuthParameters["SECRET_HASH"]; ok {
		t.Errorf("SECRET_HASH must be omitted when client secret is empty")
	}
}

func Test_パスワード認証_未設定(t *testing.T) {
	a := newPasswordAuthenticatorWithClient(&fakeInitiateAuth{}, "", "secret")
	if _, err := a.Authenticate(context.Background(), "u", "pw"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("want ErrNotConfigured, got %v", err)
	}
}

func Test_パスワード認証_SecretHash算出(t *testing.T) {
	a := newPasswordAuthenticatorWithClient(nil, "id", "secret")
	got := a.secretHash("user")

	mac := hmac.New(sha256.New, []byte("secret"))
	mac.Write([]byte("user" + "id"))
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if got != want {
		t.Errorf("secretHash = %s, want %s", got, want)
	}
}
