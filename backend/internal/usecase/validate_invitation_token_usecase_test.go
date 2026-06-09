package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubCompanies struct {
	byID map[uint64]*domain.Company
	err  error
}

func (s *stubCompanies) ListAll(_ context.Context) ([]domain.Company, error)     { return nil, s.err }
func (s *stubCompanies) UpdateAiChatEnabled(context.Context, uint64, bool) error { return nil }
func (s *stubCompanies) FindByID(_ context.Context, id uint64) (*domain.Company, error) {
	if s.err != nil {
		return nil, s.err
	}
	if c, ok := s.byID[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

// stubAdminInvRepoWithToken は ValidateInvitationTokenUseCase 専用の stub。
// stubAdminInvRepo (admin_invitation_usecase_test.go) は他テストで FindPendingByToken を
// nil 固定で返してしまうため、このテストでは別の stub を持つ。
type stubAdminInvRepoWithToken struct {
	stubAdminInvRepo
	pendingByToken map[string]*domain.AdminInvitation
}

func (s *stubAdminInvRepoWithToken) FindPendingByToken(_ context.Context, token string) (*domain.AdminInvitation, error) {
	if s.err != nil {
		return nil, s.err
	}
	if v, ok := s.pendingByToken[token]; ok {
		return v, nil
	}
	return nil, nil
}

func TestValidateInvitationToken_EmptyToken_ReturnsNil(t *testing.T) {
	uc := NewValidateInvitationTokenUseCase(&stubAdminInvRepoWithToken{}, &stubCompanies{})
	got, err := uc.Execute(context.Background(), "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != nil {
		t.Fatalf("empty token must return nil, got %+v", got)
	}
}

func TestValidateInvitationToken_NotFound_ReturnsNil(t *testing.T) {
	uc := NewValidateInvitationTokenUseCase(&stubAdminInvRepoWithToken{}, &stubCompanies{})
	got, err := uc.Execute(context.Background(), "missing-token")
	if err != nil || got != nil {
		t.Fatalf("missing token must return (nil, nil), got=%+v err=%v", got, err)
	}
}

func TestValidateInvitationToken_OK_AttachesCompanyName(t *testing.T) {
	repo := &stubAdminInvRepoWithToken{
		pendingByToken: map[string]*domain.AdminInvitation{
			"abc-123": {
				ID: 9, CompanyID: 42, Email: "u@example.com",
				Role: domain.RoleCompanyAdmin, DisplayName: "山田",
			},
		},
	}
	companies := &stubCompanies{
		byID: map[uint64]*domain.Company{
			42: {ID: 42, Name: "株式会社FreStyle"},
		},
	}
	uc := NewValidateInvitationTokenUseCase(repo, companies)

	got, err := uc.Execute(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil result")
	}
	if got.Role != domain.RoleCompanyAdmin {
		t.Errorf("Role = %q, want company_admin", got.Role)
	}
	if got.DisplayName != "山田" {
		t.Errorf("DisplayName = %q, want 山田", got.DisplayName)
	}
	if got.CompanyID != 42 || got.CompanyName != "株式会社FreStyle" {
		t.Errorf("Company = %d/%q, want 42/株式会社FreStyle", got.CompanyID, got.CompanyName)
	}
}

func TestValidateInvitationToken_CompanyLookupFails_StillReturnsInvitation(t *testing.T) {
	// company 取得失敗は invitation 自体の有効性を否定しない。
	// CompanyName 空でも受諾画面を表示できる方が UX として良い。
	repo := &stubAdminInvRepoWithToken{
		pendingByToken: map[string]*domain.AdminInvitation{
			"t": {ID: 9, CompanyID: 999, Email: "u@example.com", Role: domain.RoleTrainee},
		},
	}
	companies := &stubCompanies{} // 空 → FindByID で「not found」
	uc := NewValidateInvitationTokenUseCase(repo, companies)

	got, err := uc.Execute(context.Background(), "t")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got == nil || got.CompanyID != 999 {
		t.Fatalf("invitation must be returned with CompanyID even when company lookup fails, got %+v", got)
	}
	if got.CompanyName != "" {
		t.Errorf("CompanyName should be empty on lookup failure, got %q", got.CompanyName)
	}
}

func TestValidateInvitationToken_NormalizesUnknownRole(t *testing.T) {
	repo := &stubAdminInvRepoWithToken{
		pendingByToken: map[string]*domain.AdminInvitation{
			"t": {ID: 9, CompanyID: 1, Role: "garbage_role"},
		},
	}
	uc := NewValidateInvitationTokenUseCase(repo, &stubCompanies{byID: map[uint64]*domain.Company{1: {ID: 1, Name: "x"}}})
	got, _ := uc.Execute(context.Background(), "t")
	if got == nil || got.Role != domain.RoleTrainee {
		t.Errorf("unknown role must fallback to trainee, got %+v", got)
	}
}
