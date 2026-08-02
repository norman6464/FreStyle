package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// UpsertUserFromIDTokenInput はIDトークンから取得したユーザー情報を表す。
type UpsertUserFromIDTokenInput struct {
	CognitoSub      string
	Email           string
	Name            string
	IsCognitoAdmin  bool
	InvitationToken string
}

// UpsertUserFromIDTokenUseCase は認証済みユーザーの作成・更新を行う。
type UpsertUserFromIDTokenUseCase struct {
	users       repository.UserRepository
	invitations repository.AdminInvitationRepository
}

// NewUpsertUserFromIDTokenUseCase はUpsertUserFromIDTokenUseCaseを生成する。
func NewUpsertUserFromIDTokenUseCase(
	users repository.UserRepository,
	invitations repository.AdminInvitationRepository,
) *UpsertUserFromIDTokenUseCase {
	return &UpsertUserFromIDTokenUseCase{
		users:       users,
		invitations: invitations,
	}
}

// Execute はユーザー情報と招待情報を基にユーザーを作成・更新する。
func (u *UpsertUserFromIDTokenUseCase) Execute(
	ctx context.Context,
	in UpsertUserFromIDTokenInput,
) (allowed bool, err error) {
	if u.users == nil {
		return false, errors.New("user repository not configured")
	}

	sub := in.CognitoSub
	if sub == "" {
		return false, errors.New("id_token missing sub")
	}

	email := in.Email
	oidcName := in.Name
	isCognitoAdmin := in.IsCognitoAdmin
	invitationToken := in.InvitationToken

	var inv *domain.AdminInvitation
	if u.invitations != nil {
		if invitationToken != "" {
			inv, _ = u.invitations.FindPendingByToken(ctx, invitationToken)
		}
		if inv == nil && email != "" {
			inv, _ = u.invitations.FindPendingByEmail(ctx, email)
		}
	}

	existing, _ := u.users.FindByCognitoSub(ctx, sub)
	if existing != nil {
		if oidcName != "" && existing.Email != "" && existing.Name == existing.Email {
			if err := u.users.UpdateName(ctx, existing.ID, oidcName); err != nil {
				log.Printf("upsertUserFromIDToken: backfill name failed userID=%d: %v", existing.ID, err)
			}
		}

		if isCognitoAdmin && existing.Role != domain.RoleSuperAdmin {
			_ = u.users.UpdateRole(ctx, existing.ID, domain.RoleSuperAdmin)
		}
		if inv != nil && existing.Role != domain.RoleSuperAdmin {
			if existing.Role == domain.RoleTrainee && inv.Role == domain.RoleCompanyAdmin {
				if err := u.users.UpdateRole(ctx, existing.ID, domain.RoleCompanyAdmin); err != nil {
					log.Printf("upsertUserFromIDToken: existing user role upgrade failed userID=%d: %v", existing.ID, err)
				} else {
					log.Printf("upsertUserFromIDToken: existing user upgraded trainee→company_admin userID=%d email=%s", existing.ID, email)
				}
			}
			if inv.CompanyID != 0 && (existing.CompanyID == nil || *existing.CompanyID != inv.CompanyID) {
				if err := u.users.UpdateCompanyID(ctx, existing.ID, inv.CompanyID); err != nil {
					log.Printf("upsertUserFromIDToken: existing user company update failed userID=%d: %v", existing.ID, err)
				}
			}
			_ = u.invitations.UpdateStatus(ctx, inv.ID, domain.InvitationStatusAccepted)
		}
		return true, nil
	}

	if !isCognitoAdmin && inv == nil {
		log.Printf(
			"upsertUserFromIDToken: signup blocked - no invitation and not Cognito admin sub=%s email=%s token_provided=%t",
			sub,
			email,
			invitationToken != "",
		)
		return false, nil
	}

	role := domain.RoleTrainee
	var companyID *uint64
	var acceptedInvID uint64

	name := email
	if oidcName != "" {
		name = oidcName
	}

	if isCognitoAdmin {
		role = domain.RoleSuperAdmin
	}
	if inv != nil {
		if inv.Role == domain.RoleCompanyAdmin || inv.Role == domain.RoleTrainee {
			role = inv.Role
		}
		cid := inv.CompanyID
		companyID = &cid
		acceptedInvID = inv.ID
		if inv.Name != "" {
			name = inv.Name
		}
	}

	if err := u.users.Create(ctx, &domain.User{
		CognitoSub: sub,
		Email:      email,
		Name:       name,
		Role:       role,
		CompanyID:  companyID,
	}); err != nil {
		log.Printf("upsertUserFromIDToken: create user failed sub=%s email=%s err=%v", sub, email, err)
		return false, fmt.Errorf("create user: %w", err)
	}
	if u.invitations != nil && acceptedInvID != 0 {
		_ = u.invitations.UpdateStatus(ctx, acceptedInvID, domain.InvitationStatusAccepted)
	}
	return true, nil
}
