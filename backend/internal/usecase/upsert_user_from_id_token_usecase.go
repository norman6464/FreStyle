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

func (u *UpsertUserFromIDTokenUseCase) shouldBackfillName(
	oidcName string,
	existing *domain.User,
) bool {
	return oidcName != "" &&
		existing != nil &&
		existing.Email != "" &&
		existing.Name == existing.Email
}

func (u *UpsertUserFromIDTokenUseCase) shouldUpdateRoleFromInvitation(
	isCognitoAdmin bool,
	existing *domain.User,
	inv *domain.AdminInvitation,
) bool {
	return !isCognitoAdmin &&
		existing != nil &&
		inv != nil &&
		existing.Role == domain.RoleTrainee &&
		inv.Role == domain.RoleCompanyAdmin
}

func (u *UpsertUserFromIDTokenUseCase) updateExistingUser(
	ctx context.Context,
	existing *domain.User,
	inv *domain.AdminInvitation,
	oidcName string,
	isCognitoAdmin bool,
) error {
	role := existing.Role

	if u.shouldBackfillName(oidcName, existing) {
		if err := u.users.UpdateName(ctx, existing.ID, oidcName); err != nil {
			return fmt.Errorf("update existing user name: %w", err)
		}
	}

	if isCognitoAdmin && existing.Role != domain.RoleSuperAdmin {
		if err := u.users.UpdateRole(
			ctx,
			existing.ID,
			domain.RoleSuperAdmin,
		); err != nil {
			return fmt.Errorf("update existing user admin role: %w", err)
		}
		role = domain.RoleSuperAdmin
	}

	if inv == nil || role == domain.RoleSuperAdmin {
		return nil
	}

	if u.shouldUpdateRoleFromInvitation(
		isCognitoAdmin,
		existing,
		inv,
	) {
		if err := u.users.UpdateRole(
			ctx,
			existing.ID,
			domain.RoleCompanyAdmin,
		); err != nil {
			return fmt.Errorf("update existing user invitation role: %w", err)
		}
	}

	if inv.CompanyID != 0 &&
		(existing.CompanyID == nil ||
			*existing.CompanyID != inv.CompanyID) {
		if err := u.users.UpdateCompanyID(
			ctx,
			existing.ID,
			inv.CompanyID,
		); err != nil {
			return fmt.Errorf("update existing user company: %w", err)
		}
	}

	if err := u.invitations.UpdateStatus(
		ctx,
		inv.ID,
		domain.InvitationStatusAccepted,
	); err != nil {
		return fmt.Errorf("accept invitation: %w", err)
	}

	return nil
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
			var findErr error
			inv, findErr = u.invitations.FindPendingByToken(ctx, invitationToken)
			if findErr != nil {
				return false, fmt.Errorf(
					"find pending invitation by token: %w",
					findErr,
				)
			}
		}

		if inv == nil && email != "" {
			var findErr error
			inv, findErr = u.invitations.FindPendingByEmail(ctx, email)
			if findErr != nil {
				return false, fmt.Errorf(
					"find pending invitation by email: %w",
					findErr,
				)
			}
		}
	}

	existing, findErr := u.users.FindByCognitoSub(ctx, sub)
	if findErr != nil {
		return false, fmt.Errorf(
			"find user by cognito sub: %w",
			findErr,
		)
	}

	if existing != nil {
		if err := u.updateExistingUser(
			ctx,
			existing,
			inv,
			oidcName,
			isCognitoAdmin,
		); err != nil {
			return false, fmt.Errorf("update existing user: %w", err)
		}
		// 正規化テーブル（user_oidc_identities）のセルフヒール。旧カラム経由で見つかった行にも
		// identity を張り直す（冪等）。失敗してもログインは旧カラムのフォールバックで継続できる
		// ため、ここでは致命扱いにしない。
		if err := u.users.EnsureOidcIdentity(ctx, existing.ID, domain.OidcProviderCognito, sub); err != nil {
			log.Printf("upsertUserFromIDToken: ensure oidc identity failed (self-heal, non-fatal): user=%d err=%v", existing.ID, err)
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

	name := email
	if oidcName != "" {
		name = oidcName
	}

	if isCognitoAdmin {
		role = domain.RoleSuperAdmin
	}
	if inv != nil {
		if !isCognitoAdmin &&
			(inv.Role == domain.RoleCompanyAdmin ||
				inv.Role == domain.RoleTrainee) {
			role = inv.Role
		}

		if inv.CompanyID != 0 {
			cid := inv.CompanyID
			companyID = &cid
		}
		if inv.Name != "" {
			name = inv.Name
		}
	}

	user := &domain.User{
		CognitoSub: sub,
		Email:      email,
		Name:       name,
		Role:       role,
		CompanyID:  companyID,
	}

	if err := u.users.Create(ctx, user); err != nil {
		return false, fmt.Errorf("create user: %w", err)
	}

	// OIDC identity を users と対で作る（正規化後のログイン突き合わせの正・FRESTYLE-311）。
	// 失敗しても旧カラム（users.cognito_sub）のフォールバックで次回ログインでき、
	// その際に上のセルフヒールで張り直されるため致命扱いにしない。
	if err := u.users.EnsureOidcIdentity(ctx, user.ID, domain.OidcProviderCognito, sub); err != nil {
		log.Printf("upsertUserFromIDToken: create oidc identity failed (non-fatal): user=%d err=%v", user.ID, err)
	}

	if inv != nil {
		if err := u.invitations.UpdateStatus(
			ctx,
			inv.ID,
			domain.InvitationStatusAccepted,
		); err != nil {
			return false, fmt.Errorf("accept invitation: %w", err)
		}
	}

	return true, nil
}
