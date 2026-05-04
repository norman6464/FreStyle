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
	return nil
}
func (s *stubAdminInvRepo) UpdateStatus(_ context.Context, _ uint64, _ string) error { return s.err }
func (s *stubAdminInvRepo) FindPendingByEmail(_ context.Context, _ string) (*domain.AdminInvitation, error) {
	return nil, s.err
}

type stubCognitoAdmin struct{ err error }

func (c *stubCognitoAdmin) InviteUser(_ context.Context, email, _, _ string) (string, error) {
	if c.err != nil {
		return "", c.err
	}
	return "sub-" + email, nil
}

func TestListAdminInvitations_ListByCompanyID_RequiresCompanyID(t *testing.T) {
	uc := NewListAdminInvitationsUseCase(&stubAdminInvRepo{})
	if _, err := uc.ListByCompanyID(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestListAdminInvitations_ListByCompanyID_DelegatesToRepo(t *testing.T) {
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

func TestListAdminInvitations_ListAll_DelegatesToRepo(t *testing.T) {
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

func TestCreateAdminInvitation_Validates(t *testing.T) {
	uc := NewCreateAdminInvitationUseCase(&stubAdminInvRepo{}, &stubCognitoAdmin{})
	if _, err := uc.Execute(context.Background(), CreateAdminInvitationInput{Email: "a@b"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateAdminInvitation_OK(t *testing.T) {
	uc := NewCreateAdminInvitationUseCase(&stubAdminInvRepo{}, &stubCognitoAdmin{})
	got, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		CompanyID: 1, Email: "u@example.com", Role: domain.RoleTrainee,
	})
	if err != nil || got.ID != 91 || got.Status != domain.InvitationStatusPending {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestCreateAdminInvitation_CognitoError(t *testing.T) {
	uc := NewCreateAdminInvitationUseCase(&stubAdminInvRepo{}, &stubCognitoAdmin{err: errors.New("cognito")})
	if _, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		CompanyID: 1, Email: "u@example.com", Role: domain.RoleTrainee,
	}); err == nil {
		t.Fatal("expected error")
	}
}

func TestCancelAdminInvitation_RequiresID(t *testing.T) {
	uc := NewCancelAdminInvitationUseCase(&stubAdminInvRepo{})
	if err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}
