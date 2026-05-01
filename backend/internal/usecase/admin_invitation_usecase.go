package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListAdminInvitationsUseCase struct {
	repo repository.AdminInvitationRepository
}

func NewListAdminInvitationsUseCase(r repository.AdminInvitationRepository) *ListAdminInvitationsUseCase {
	return &ListAdminInvitationsUseCase{repo: r}
}

// ListAll は全社横断で招待一覧を返す。SuperAdmin (運営側) からのみ呼ばれる想定。
// 認可は handler 層で current user.role を見て行う。
func (u *ListAdminInvitationsUseCase) ListAll(ctx context.Context) ([]domain.AdminInvitation, error) {
	return u.repo.ListAll(ctx)
}

// ListByCompanyID は指定 company の招待一覧を返す。CompanyAdmin が自社のみを見る用。
func (u *ListAdminInvitationsUseCase) ListByCompanyID(ctx context.Context, companyID uint64) ([]domain.AdminInvitation, error) {
	if companyID == 0 {
		return nil, errors.New("companyID is required")
	}
	return u.repo.ListByCompanyID(ctx, companyID)
}

type CreateAdminInvitationUseCase struct {
	repo repository.AdminInvitationRepository
	cog  repository.CognitoAdminClient
}

func NewCreateAdminInvitationUseCase(r repository.AdminInvitationRepository, c repository.CognitoAdminClient) *CreateAdminInvitationUseCase {
	return &CreateAdminInvitationUseCase{repo: r, cog: c}
}

type CreateAdminInvitationInput struct {
	CompanyID   uint64
	Email       string
	Role        string
	DisplayName string
}

func (u *CreateAdminInvitationUseCase) Execute(ctx context.Context, in CreateAdminInvitationInput) (*domain.AdminInvitation, error) {
	if in.CompanyID == 0 || in.Email == "" || in.Role == "" {
		return nil, errors.New("companyID, email, role are required")
	}
	if _, err := u.cog.InviteUser(ctx, in.Email, in.DisplayName, in.Role); err != nil {
		return nil, err
	}
	inv := &domain.AdminInvitation{
		CompanyID: in.CompanyID, Email: in.Email, Role: in.Role,
		DisplayName: in.DisplayName, Status: domain.InvitationStatusPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := u.repo.Create(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

type CancelAdminInvitationUseCase struct {
	repo repository.AdminInvitationRepository
}

func NewCancelAdminInvitationUseCase(r repository.AdminInvitationRepository) *CancelAdminInvitationUseCase {
	return &CancelAdminInvitationUseCase{repo: r}
}

func (u *CancelAdminInvitationUseCase) Execute(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("id is required")
	}
	return u.repo.UpdateStatus(ctx, id, domain.InvitationStatusCanceled)
}
