package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubAdminInvRepo struct {
	rows []domain.AdminInvitation
	err  error
}

func (s *stubAdminInvRepo) ListByCompanyID(_ context.Context, _ uint64) ([]domain.AdminInvitation, error) {
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

type stubCognitoAdmin struct{ err error }

func (c *stubCognitoAdmin) InviteUser(_ context.Context, email, _, _ string) (string, error) {
	if c.err != nil {
		return "", c.err
	}
	return "sub-" + email, nil
}

func TestListAdminInvitations_RequiresCompanyID(t *testing.T) {
	uc := NewListAdminInvitationsUseCase(&stubAdminInvRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
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
