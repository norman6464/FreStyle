package cognito

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cip "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type fakeAdminCreate struct {
	out   *cip.AdminCreateUserOutput
	err   error
	input *cip.AdminCreateUserInput
}

func (f *fakeAdminCreate) AdminCreateUser(_ context.Context, in *cip.AdminCreateUserInput, _ ...func(*cip.Options)) (*cip.AdminCreateUserOutput, error) {
	f.input = in
	return f.out, f.err
}

func Test_NewAdminUserCreator_poolID必須(t *testing.T) {
	if _, err := NewAdminUserCreator(context.Background(), "ap-northeast-1", ""); err == nil {
		t.Fatal("空の pool id で作れてはいけない")
	}
}

func Test_CreateWithTemporaryPassword_成功(t *testing.T) {
	fake := &fakeAdminCreate{out: &cip.AdminCreateUserOutput{}}
	a := newAdminUserCreatorWithClient(fake, "pool-1")

	pw, err := a.CreateWithTemporaryPassword(context.Background(), "u@example.com", "山田太郎")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 生成パスワードはポリシー（8+・各文字種）を満たす。
	if len(pw) < 8 {
		t.Errorf("temp password too short: %d", len(pw))
	}
	for name, re := range map[string]string{
		"lower": "[a-z]", "upper": "[A-Z]", "digit": "[0-9]", "symbol": `[!@#$%^&*\-_=+]`,
	} {
		if !regexp.MustCompile(re).MatchString(pw) {
			t.Errorf("temp password missing %s class: %q", name, pw)
		}
	}

	in := fake.input
	if aws.ToString(in.UserPoolId) != "pool-1" {
		t.Errorf("pool id = %q", aws.ToString(in.UserPoolId))
	}
	if aws.ToString(in.Username) != "u@example.com" {
		t.Errorf("username = %q", aws.ToString(in.Username))
	}
	if in.MessageAction != types.MessageActionTypeSuppress {
		t.Errorf("message action = %q, want SUPPRESS（Cognito 自動メールを止める）", in.MessageAction)
	}
	if aws.ToString(in.TemporaryPassword) != pw {
		t.Error("送信した TemporaryPassword と返り値が一致しない")
	}
	// email_verified=true と name が属性に含まれる。
	var hasEmail, hasVerified, hasName bool
	for _, at := range in.UserAttributes {
		if aws.ToString(at.Name) == "email" && aws.ToString(at.Value) == "u@example.com" {
			hasEmail = true
		}
		if aws.ToString(at.Name) == "email_verified" && aws.ToString(at.Value) == "true" {
			hasVerified = true
		}
		if aws.ToString(at.Name) == "name" && aws.ToString(at.Value) == "山田太郎" {
			hasName = true
		}
	}
	if !hasEmail {
		t.Error("email 属性が無い")
	}
	if !hasVerified {
		t.Error("email_verified=true が無い")
	}
	if !hasName {
		t.Error("name 属性が無い")
	}
}

func Test_CreateWithTemporaryPassword_毎回異なる(t *testing.T) {
	a := newAdminUserCreatorWithClient(&fakeAdminCreate{out: &cip.AdminCreateUserOutput{}}, "pool-1")
	pw1, _ := a.CreateWithTemporaryPassword(context.Background(), "a@example.com", "")
	pw2, _ := a.CreateWithTemporaryPassword(context.Background(), "b@example.com", "")
	if pw1 == pw2 {
		t.Error("一時パスワードが毎回同じ（乱数になっていない）")
	}
}

func Test_CreateWithTemporaryPassword_既存ユーザーは専用エラー(t *testing.T) {
	fake := &fakeAdminCreate{err: &types.UsernameExistsException{}}
	a := newAdminUserCreatorWithClient(fake, "pool-1")

	_, err := a.CreateWithTemporaryPassword(context.Background(), "dup@example.com", "")
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("ErrUserAlreadyExists を期待したが: %v", err)
	}
}

func Test_CreateWithTemporaryPassword_その他エラーはラップして返す(t *testing.T) {
	fake := &fakeAdminCreate{err: errors.New("throttled")}
	a := newAdminUserCreatorWithClient(fake, "pool-1")

	_, err := a.CreateWithTemporaryPassword(context.Background(), "x@example.com", "")
	if err == nil || errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("汎用エラーをそのまま返すべき: %v", err)
	}
	if !strings.Contains(err.Error(), "admin create user") {
		t.Errorf("ラップされていない: %v", err)
	}
}
