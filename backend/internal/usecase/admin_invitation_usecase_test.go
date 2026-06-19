package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubAdminInvRepo struct {
	rows []domain.AdminInvitation
	err  error
	// 直近に呼ばれた絞り込み条件を記録する。"all" / "company:42" のような形式。
	calledWith string
	// Create で渡された invitation を保持し、token / status を assert に使う。
	created *domain.AdminInvitation
}

func (s *stubAdminInvRepo) ListAll(_ context.Context) ([]domain.AdminInvitation, error) {
	s.calledWith = "all"
	return s.rows, s.err
}

func (s *stubAdminInvRepo) ListByCompanyID(_ context.Context, companyID uint64) ([]domain.AdminInvitation, error) {
	s.calledWith = "company:" + strconv.FormatUint(companyID, 10)
	return s.rows, s.err
}

func (s *stubAdminInvRepo) Create(_ context.Context, inv *domain.AdminInvitation) error {
	if s.err != nil {
		return s.err
	}
	inv.ID = 91
	s.created = inv
	return nil
}

func (s *stubAdminInvRepo) UpdateStatus(_ context.Context, _ uint64, _ string) error { return s.err }

func (s *stubAdminInvRepo) FindPendingByEmail(_ context.Context, _ string) (*domain.AdminInvitation, error) {
	return nil, s.err
}

func (s *stubAdminInvRepo) FindPendingByToken(_ context.Context, _ string) (*domain.AdminInvitation, error) {
	return nil, s.err
}

// stubMailSender は SES SendEmail の代わり。送信内容をフィールドに記録するだけで実際にネットワークアクセスしない。
type stubMailSender struct {
	err     error
	calls   int
	to      string
	subject string
	html    string
	text    string
}

func (s *stubMailSender) SendInvitationEmail(_ context.Context, to, subject, htmlBody, textBody string) error {
	s.calls++
	s.to, s.subject, s.html, s.text = to, subject, htmlBody, textBody
	return s.err
}

// テスト用のシンプルな builder。フォーマットの正しさは ses パッケージのテストで担保し、
// usecase 層では「正しく呼ばれたか」だけ検証する。
func fakeBuildLink(token string) string {
	return "https://test.example/invitations/accept?token=" + token
}

func fakeBuildMail(link, displayName, _, role string) (string, string, string) {
	return "subject", "html-" + link + "-" + displayName + "-" + role, "text-" + link
}

func Test_招待一覧_会社ID指定_会社IDが必須(t *testing.T) {
	uc := NewListAdminInvitationsUseCase(&stubAdminInvRepo{})
	if _, err := uc.ListByCompanyID(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func Test_招待一覧_会社ID指定_リポジトリへ委譲(t *testing.T) {
	repo := &stubAdminInvRepo{rows: []domain.AdminInvitation{{ID: 1}}}
	uc := NewListAdminInvitationsUseCase(repo)
	got, err := uc.ListByCompanyID(context.Background(), 42)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("unexpected rows: %+v", got)
	}
	if repo.calledWith != "company:42" {
		t.Fatalf("expected company:42 query, got %q", repo.calledWith)
	}
}

func Test_招待一覧_全件_リポジトリへ委譲(t *testing.T) {
	repo := &stubAdminInvRepo{rows: []domain.AdminInvitation{{ID: 7}, {ID: 8}}}
	uc := NewListAdminInvitationsUseCase(repo)
	got, err := uc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("unexpected rows: %+v", got)
	}
	if repo.calledWith != "all" {
		t.Fatalf("expected ListAll path, got %q", repo.calledWith)
	}
}

func Test_招待作成_バリデーション(t *testing.T) {
	uc := NewCreateAdminInvitationUseCase(&stubAdminInvRepo{}, &stubMailSender{}, fakeBuildLink, fakeBuildMail)
	if _, err := uc.Execute(context.Background(), CreateAdminInvitationInput{Email: "a@b"}); err == nil {
		t.Fatal("expected error")
	}
}

func Test_招待作成_正常系_token生成とメール送信(t *testing.T) {
	repo := &stubAdminInvRepo{}
	sender := &stubMailSender{}
	uc := NewCreateAdminInvitationUseCase(repo, sender, fakeBuildLink, fakeBuildMail)

	got, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		CompanyID: 1, Email: "u@example.com", Role: domain.RoleTrainee, Name: "山田",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.ID != 91 || got.Status != domain.InvitationStatusPending {
		t.Fatalf("unexpected: %+v", got)
	}
	if got.Token == nil || *got.Token == "" {
		t.Fatalf("Token must be generated, got %+v", got.Token)
	}
	if repo.created == nil || repo.created.Token == nil || *repo.created.Token != *got.Token {
		t.Fatalf("repo.Create must receive token, got %+v", repo.created)
	}
	if sender.calls != 1 {
		t.Fatalf("expected 1 SES send, got %d", sender.calls)
	}
	if sender.to != "u@example.com" {
		t.Fatalf("expected to=u@example.com, got %q", sender.to)
	}
	// magicLink がメール本文に含まれているか（fake builder が link を埋め込む実装）
	if !contains(sender.html, "/invitations/accept?token="+*got.Token) {
		t.Fatalf("html body must contain magic link, got %q", sender.html)
	}
}

func Test_招待作成_SESエラーを伝播(t *testing.T) {
	uc := NewCreateAdminInvitationUseCase(&stubAdminInvRepo{}, &stubMailSender{err: errors.New("ses down")}, fakeBuildLink, fakeBuildMail)
	if _, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		CompanyID: 1, Email: "u@example.com", Role: domain.RoleTrainee,
	}); err == nil {
		t.Fatal("expected ses error to be propagated")
	}
}

// sender が nil の場合（ローカル開発でフォールバック）は invitation だけ作成して終わる。
func Test_招待作成_送信者nilはメールをスキップ(t *testing.T) {
	repo := &stubAdminInvRepo{}
	uc := NewCreateAdminInvitationUseCase(repo, nil, nil, nil)
	got, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		CompanyID: 1, Email: "u@example.com", Role: domain.RoleTrainee,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Token == nil {
		t.Fatalf("Token must still be generated even when sender is nil")
	}
}

func Test_招待取消_IDが必須(t *testing.T) {
	uc := NewCancelAdminInvitationUseCase(&stubAdminInvRepo{})
	if err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
