package localauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserSource struct {
	user        *domain.User
	findErr     error
	subject     string
	subjectErr  error
	ensured     []string // "userID:provider:subject"
	ensureErr   error
	ensuredCall int
}

func (f *fakeUserSource) FindActiveByEmail(_ context.Context, _ string) (*domain.User, error) {
	return f.user, f.findErr
}

func (f *fakeUserSource) CognitoSubjectByUserID(_ context.Context, _ uint64) (string, error) {
	return f.subject, f.subjectErr
}

func (f *fakeUserSource) EnsureOidcIdentity(_ context.Context, userID uint64, provider, subject string) error {
	f.ensuredCall++
	f.ensured = append(f.ensured, provider+":"+subject)
	return f.ensureErr
}

func hashOf(t *testing.T, password string) *string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	s := string(h)
	return &s
}

// decodeClaims は発行された JWT の payload を検証用に取り出す（署名検証はしない =
// router.buildJWTVerify の local 経路と同じ扱い）。
func decodeClaims(t *testing.T, token string) map[string]any {
	t.Helper()
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3, "JWT は header.payload.signature の 3 パートであること")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	var claims map[string]any
	require.NoError(t, json.Unmarshal(payload, &claims))
	return claims
}

func Test_New_local以外ではfail_closed(t *testing.T) {
	for _, env := range []string{"production", "staging", ""} {
		_, err := New(&fakeUserSource{}, env)
		assert.Error(t, err, "APP_ENV=%q で作れてはいけない", env)
	}
	_, err := New(nil, "local")
	assert.Error(t, err, "user source なしで作れてはいけない")
}

func Test_Authenticate_正しいパスワードでCognito互換トークンを返す(t *testing.T) {
	users := &fakeUserSource{
		user: &domain.User{
			ID: 42, Email: "seed1@example.test", Name: "シード利用者1",
			Role: domain.RoleTrainee, PasswordHash: hashOf(t, "password"),
		},
		subject: "seed-sub-1",
	}
	a, err := New(users, "local")
	require.NoError(t, err)
	a.now = func() time.Time { return time.Unix(1_700_000_000, 0) }

	tok, err := a.Authenticate(context.Background(), "seed1@example.test", "password")
	require.NoError(t, err)
	require.NotEmpty(t, tok.AccessToken)
	require.NotEmpty(t, tok.IDToken)
	require.True(t, strings.HasPrefix(tok.RefreshToken, "local-"))

	claims := decodeClaims(t, tok.IDToken)
	assert.Equal(t, "seed-sub-1", claims["sub"], "既存の OIDC identity の subject を sub に使う")
	assert.Equal(t, "seed1@example.test", claims["email"])
	assert.Equal(t, "シード利用者1", claims["name"])
	assert.Equal(t, float64(1_700_000_000+24*3600), claims["exp"], "exp は 24 時間後")
	assert.Equal(t, 0, users.ensuredCall, "identity が既にあるとき Ensure は呼ばない")
	_, hasGroups := claims["cognito:groups"]
	assert.False(t, hasGroups, "groups クレームを付けない（グループ由来の権限昇格を起こさない）")
}

func Test_Authenticate_identityが無ければ決定的なsubjectを生成して張る(t *testing.T) {
	users := &fakeUserSource{
		user: &domain.User{ID: 7, Email: "pw@example.test", PasswordHash: hashOf(t, "password")},
	}
	a, err := New(users, "local")
	require.NoError(t, err)

	tok, err := a.Authenticate(context.Background(), "pw@example.test", "password")
	require.NoError(t, err)

	claims := decodeClaims(t, tok.IDToken)
	assert.Equal(t, "local-pw-7", claims["sub"])
	require.Equal(t, []string{"cognito:local-pw-7"}, users.ensured, "cognito provider として identity を張る")
}

func Test_Authenticate_資格情報エラーは詳細を区別しない(t *testing.T) {
	cases := map[string]*fakeUserSource{
		"ユーザー不在":   {user: nil},
		"パスワード未設定": {user: &domain.User{ID: 1, Email: "x@example.test"}},
		"パスワード不一致": {user: &domain.User{ID: 1, Email: "x@example.test", PasswordHash: hashOf(t, "other")}},
	}
	for name, users := range cases {
		t.Run(name, func(t *testing.T) {
			a, err := New(users, "local")
			require.NoError(t, err)
			_, err = a.Authenticate(context.Background(), "x@example.test", "password")
			assert.True(t, errors.Is(err, cognito.ErrInvalidCredentials),
				"handler が 401 に写せるよう cognito.ErrInvalidCredentials を返すこと（got %v）", err)
		})
	}
}

func Test_Authenticate_DBエラーは資格情報エラーにしない(t *testing.T) {
	users := &fakeUserSource{findErr: errors.New("db down")}
	a, err := New(users, "local")
	require.NoError(t, err)

	_, err = a.Authenticate(context.Background(), "x@example.test", "password")
	require.Error(t, err)
	assert.False(t, errors.Is(err, cognito.ErrInvalidCredentials), "障害を 401 に見せない")
}

func Test_Authenticate_identity張り失敗はログイン失敗(t *testing.T) {
	users := &fakeUserSource{
		user:      &domain.User{ID: 9, Email: "y@example.test", PasswordHash: hashOf(t, "password")},
		ensureErr: errors.New("conflict"),
	}
	a, err := New(users, "local")
	require.NoError(t, err)

	_, err = a.Authenticate(context.Background(), "y@example.test", "password")
	require.Error(t, err)
}

func Test_VerifyToken_署名と失効を検証する(t *testing.T) {
	users := &fakeUserSource{
		user: &domain.User{ID: 1, Email: "v@example.test", PasswordHash: hashOf(t, "password")},
	}
	a, err := New(users, "local")
	require.NoError(t, err)

	tok, err := a.Authenticate(context.Background(), "v@example.test", "password")
	require.NoError(t, err)

	t.Run("正しいトークンは claims が返る", func(t *testing.T) {
		claims, err := VerifyToken(tok.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, "local-pw-1", claims["sub"])
	})

	t.Run("改ざんは ErrNotLocalToken（通常経路へフォールバック）", func(t *testing.T) {
		parts := strings.Split(tok.AccessToken, ".")
		forged := parts[0] + "." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"attacker","iss":"frestyle-localauth","exp":9999999999}`)) + "." + parts[2]
		_, err := VerifyToken(forged)
		assert.True(t, errors.Is(err, ErrNotLocalToken))
	})

	t.Run("他所発行の JWT は ErrNotLocalToken", func(t *testing.T) {
		_, err := VerifyToken("eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ4In0.c2ln")
		assert.True(t, errors.Is(err, ErrNotLocalToken))
	})

	t.Run("失効はフォールバックさせないエラー", func(t *testing.T) {
		expired, err := mintJWT(map[string]any{"sub": "x", "iss": Issuer, "exp": time.Now().Add(-time.Minute).Unix()})
		require.NoError(t, err)
		_, err = VerifyToken(expired)
		require.Error(t, err)
		assert.False(t, errors.Is(err, ErrNotLocalToken), "localauth 発行と分かるトークンの失効は他経路に回さない")
	})
}
