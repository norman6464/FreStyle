package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

type stubUserRepo struct {
	user *domain.User
	err  error
}

func (s *stubUserRepo) FindByOidcSubject(_ context.Context, _ string) (*domain.User, error) {
	return s.user, s.err
}

func (s *stubUserRepo) FindByID(_ context.Context, _ uint64) (*domain.User, error) {
	return s.user, s.err
}

func (s *stubUserRepo) ListByWorkspaceID(_ context.Context, _ string) ([]domain.User, error) {
	return nil, s.err
}

func (s *stubUserRepo) Create(_ context.Context, _ *domain.User) error {
	return s.err
}

func (s *stubUserRepo) UpdateName(_ context.Context, _ uint64, _ string) error {
	return s.err
}

func (s *stubUserRepo) UpdateWorkspaceID(_ context.Context, _ uint64, _ *string) error {
	return s.err
}

func (s *stubUserRepo) UpdateActive(context.Context, uint64, bool) error { return nil }
func (s *stubUserRepo) SoftDelete(context.Context, uint64) error         { return nil }

func (s *stubUserRepo) FindActiveByEmail(context.Context, string) (*domain.User, error) {
	return nil, nil
}

func (s *stubUserRepo) OidcSubjectByUserID(context.Context, uint64) (string, error) {
	return "", nil
}

// fakeTxManager は repository.TxManager のテスト用 no-op 実装。fn(ctx) をそのまま呼ぶだけで、
// 実 DB もトランザクションも介さない。本ファイル内の複数のテストで共有する。
type fakeTxManager struct{}

func (fakeTxManager) DoInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func Test_現在ユーザー取得_見つかる(t *testing.T) {
	want := &domain.User{ID: 1, Email: "u@example.com"}
	uc := NewGetCurrentUserUseCase(&stubUserRepo{user: want})
	got, err := uc.Execute(context.Background(), "abc")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got == nil || got.ID != 1 {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func Test_現在ユーザー取得_見つからない(t *testing.T) {
	uc := NewGetCurrentUserUseCase(&stubUserRepo{user: nil})
	got, err := uc.Execute(context.Background(), "missing")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func Test_現在ユーザー取得_エラー(t *testing.T) {
	uc := NewGetCurrentUserUseCase(&stubUserRepo{err: errors.New("db down")})
	if _, err := uc.Execute(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}

// upsertUserRepoSpy は UpsertUserFromIDTokenUseCase の呼び出しを記録する UserRepository の spy。
type upsertUserRepoSpy struct {
	stubUserRepo
	created *domain.User

	findByOidcSubjectCalls int
	createCalls            int
	createErr              error
	nameUpdateCalls        int
	nameUpdateErr          error
}

func (s *upsertUserRepoSpy) FindByOidcSubject(
	ctx context.Context,
	sub string,
) (*domain.User, error) {
	s.findByOidcSubjectCalls++
	return s.stubUserRepo.FindByOidcSubject(ctx, sub)
}

func (s *upsertUserRepoSpy) Create(
	_ context.Context,
	user *domain.User,
) error {
	s.createCalls++
	if s.createErr != nil {
		return s.createErr
	}

	copied := *user
	s.created = &copied
	return nil
}

func (s *upsertUserRepoSpy) UpdateName(
	_ context.Context,
	_ uint64,
	_ string,
) error {
	s.nameUpdateCalls++
	return s.nameUpdateErr
}

// upsertOidcIdentitySpy は UpsertUserFromIDTokenUseCase の呼び出しを記録する
// UserOidcIdentityRepository の spy。
type upsertOidcIdentitySpy struct {
	ensureIdentityCalls int
	ensuredUserID       uint64
	ensuredProvider     string
	ensuredSubject      string
	err                 error
}

func (s *upsertOidcIdentitySpy) EnsureIdentity(
	_ context.Context,
	userID uint64,
	provider, subject string,
) error {
	s.ensureIdentityCalls++
	s.ensuredUserID = userID
	s.ensuredProvider = provider
	s.ensuredSubject = subject
	return s.err
}

// newUpsertUserUseCase はテスト用の依存（oidc spy は既定・txManager は no-op fake）で
// UpsertUserFromIDTokenUseCase を組み立てる。
func newUpsertUserUseCase(users *upsertUserRepoSpy) (*UpsertUserFromIDTokenUseCase, *upsertOidcIdentitySpy) {
	oidc := &upsertOidcIdentitySpy{}
	return NewUpsertUserFromIDTokenUseCase(users, oidc, fakeTxManager{}), oidc
}

// 招待ゲートは撤去済み（個人サインアップ）。新規ユーザーは所属ワークスペース無しで作られる。
func Test_UpsertUserFromIDToken_新規ユーザーは自己サインアップできる(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "new-sub",
			Email:   "new@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("新規ユーザーは自己サインアップできるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.WorkspaceID != nil {
		t.Fatalf("workspaceID = %v, want nil", users.created.WorkspaceID)
	}
}

// 新規ユーザ作成時に id_token の name claim が Name に使われる（email にフォールバックしない）。
func Test_UpsertUserFromIDToken_新規はOIDC名をメールより優先(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "new-sub",
			Email:   "taro@example.com",
			Name:    "山田 太郎",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("新規ユーザーは許可されるべき")
	}
	if users.created.Name != "山田 太郎" {
		t.Fatalf("name = %q, want %q", users.created.Name, "山田 太郎")
	}
}

// name claim が無いケースは email にフォールバックする。
func Test_UpsertUserFromIDToken_新規でOIDC名なしはメールにフォールバック(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "new-sub",
			Email:   "a@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("新規ユーザーは許可されるべき")
	}
	if users.created.Name != "a@example.com" {
		t.Fatalf("name = %q, want %q (fallback)", users.created.Name, "a@example.com")
	}
}

func Test_UpsertUserFromIDToken_ユーザー検索が失敗する(t *testing.T) {
	userFindErr := errors.New("user lookup failed")
	users := &upsertUserRepoSpy{stubUserRepo: stubUserRepo{err: userFindErr}}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{Subject: "user-error-sub"},
	)

	if user != nil {
		t.Fatal("検索エラー時に許可してはいけない")
	}
	if !errors.Is(err, userFindErr) {
		t.Fatalf("error = %v, want wrapped %v", err, userFindErr)
	}
	if !strings.Contains(err.Error(), "find user by oidc subject") {
		t.Fatalf("error = %q, want message containing %q", err.Error(), "find user by oidc subject")
	}
}

func Test_UpsertUserFromIDToken_Subjectが空なら処理しない(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "",
			Email:   "user@example.com",
		},
	)

	if user != nil {
		t.Fatal("Subjectが空のユーザーを許可してはいけない")
	}
	if err == nil {
		t.Fatal("Subjectが空の場合はエラーを返すべき")
	}
	if !strings.Contains(err.Error(), "id_token missing sub") {
		t.Fatalf("error = %q, want message containing %q", err.Error(), "id_token missing sub")
	}
	if users.findByOidcSubjectCalls != 0 {
		t.Fatalf("FindByOidcSubject calls = %d, want 0", users.findByOidcSubjectCalls)
	}
	if users.createCalls != 0 {
		t.Fatalf("Create calls = %d, want 0", users.createCalls)
	}
}

// Test_UpsertUserFromIDToken_同じemailでの同時サインアップはErrEmailTakenを返す は、
// repository.ErrEmailTaken をそのまま呼び出し元へ返すことを固定する
// （呼び出し元の 403/409 の出し分けが前提にする契約）。
func Test_UpsertUserFromIDToken_同じemailでの同時サインアップはErrEmailTakenを返す(t *testing.T) {
	users := &upsertUserRepoSpy{createErr: repository.ErrEmailTaken}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "race-sub",
			Email:   "race@example.com",
		},
	)

	if user != nil {
		t.Fatal("email 衝突時にユーザーを返してはいけない")
	}
	if !errors.Is(err, repository.ErrEmailTaken) {
		t.Fatalf("error = %v, want wrapped %v", err, repository.ErrEmailTaken)
	}
}

func Test_UpsertUserFromIDToken_ユーザー作成に失敗する(t *testing.T) {
	mutationErr := errors.New("create failed")
	users := &upsertUserRepoSpy{createErr: mutationErr}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "new-user-error",
			Email:   "new-user@example.com",
		},
	)

	if user != nil {
		t.Fatal("作成失敗時に許可してはいけない")
	}
	if !errors.Is(err, mutationErr) {
		t.Fatalf("error = %v, want wrapped %v", err, mutationErr)
	}
	if users.createCalls != 1 {
		t.Fatalf("Create calls = %d, want 1", users.createCalls)
	}
}

func Test_UpsertUserFromIDToken_新規作成でOIDCidentityを対で作る(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc, oidc := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "new-sub-1",
			Email:   "new@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("新規ユーザーは許可されるべき")
	}
	// 新規ユーザーは users 行と identity を同じ DoInTx の中で不可分に作る。
	if users.createCalls != 1 {
		t.Fatalf("Create calls = %d, want 1", users.createCalls)
	}
	if oidc.ensureIdentityCalls != 1 {
		t.Fatalf("EnsureIdentity calls = %d, want 1", oidc.ensureIdentityCalls)
	}
	if oidc.ensuredProvider != domain.OidcProviderDefault {
		t.Fatalf("provider = %q, want %q", oidc.ensuredProvider, domain.OidcProviderDefault)
	}
	if oidc.ensuredSubject != "new-sub-1" {
		t.Fatalf("subject = %q, want %q", oidc.ensuredSubject, "new-sub-1")
	}
}

func Test_UpsertUserFromIDToken_既存ユーザーでもidentityをセルフヒールする(t *testing.T) {
	existing := &domain.User{ID: 77, Email: "e@example.com"}
	users := &upsertUserRepoSpy{stubUserRepo: stubUserRepo{user: existing}}
	uc, oidc := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "old-sub",
			Email:   "e@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("既存ユーザーは許可されるべき")
	}
	if oidc.ensureIdentityCalls != 1 {
		t.Fatalf("EnsureIdentity calls = %d, want 1（セルフヒールされていない）", oidc.ensureIdentityCalls)
	}
	if oidc.ensuredUserID != 77 {
		t.Fatalf("ensured userID = %d, want 77", oidc.ensuredUserID)
	}
	if oidc.ensuredSubject != "old-sub" {
		t.Fatalf("subject = %q, want %q", oidc.ensuredSubject, "old-sub")
	}
}

// 既存ユーザの Name が email と一致 + id_token に name → name で上書きされる。
func Test_UpsertUserFromIDToken_既存ユーザーは表示名をOIDCから補完(t *testing.T) {
	existing := &domain.User{ID: 5, Email: "old@example.com", Name: "old@example.com"}
	users := &upsertUserRepoSpy{stubUserRepo: stubUserRepo{user: existing}}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "exists",
			Email:   "old@example.com",
			Name:    "本名 太郎",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("既存ユーザーは許可されるべき")
	}
	if users.nameUpdateCalls != 1 {
		t.Fatalf("UpdateName calls = %d, want 1", users.nameUpdateCalls)
	}
}

// 既存ユーザが既にプロフィール編集済（Name != email）なら OIDC name で上書きしない。
func Test_UpsertUserFromIDToken_表示名カスタム済みは補完しない(t *testing.T) {
	existing := &domain.User{ID: 5, Email: "u@example.com", Name: "ユーザ自身が編集した名前"}
	users := &upsertUserRepoSpy{stubUserRepo: stubUserRepo{user: existing}}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "exists",
			Email:   "u@example.com",
			Name:    "Google Name",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("既存ユーザーは許可されるべき")
	}
	if users.nameUpdateCalls != 0 {
		t.Fatalf("expected no backfill, but UpdateName called %d times", users.nameUpdateCalls)
	}
}

func Test_UpsertUserFromIDToken_名前補完の更新に失敗する(t *testing.T) {
	mutationErr := errors.New("mutation failed")
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{ID: 7, Email: "existing@example.com", Name: "existing@example.com"},
		},
		nameUpdateErr: mutationErr,
	}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "existing-user",
			Email:   "existing@example.com",
			Name:    "OIDC User",
		},
	)

	if user != nil {
		t.Fatal("名前補完の更新失敗時にユーザーを許可してはいけない")
	}
	if !errors.Is(err, mutationErr) {
		t.Fatalf("error = %v, want wrapped %v", err, mutationErr)
	}
	if users.nameUpdateCalls != 1 {
		t.Fatalf("UpdateName calls = %d, want 1", users.nameUpdateCalls)
	}
}

// 保存されるのは生の claim 値ではなく正規形。生値のまま保存すると、以後の
// byte 一致検索・一意索引と食い違う。
func Test_UpsertUserFromIDToken_emailは正規形で保存する(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc, _ := newUpsertUserUseCase(users)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			Subject: "member-sub",
			Email:   " Member@Example.com ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Email != "member@example.com" {
		t.Fatalf("保存された email = %q, want %q", users.created.Email, "member@example.com")
	}
}
