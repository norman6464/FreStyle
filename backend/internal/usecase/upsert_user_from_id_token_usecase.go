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
		// user_oidc_identities への冪等な保険。FindByCognitoSub は identity を突き合わせ条件に
		// するため通常この時点で identity は既に存在するが、provider ごとの張り直しを冪等に保証して
		// おく（失敗してもログイン自体は成立しているため致命扱いにしない）。
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
		Email:     email,
		Name:      name,
		Role:      role,
		CompanyID: companyID,
	}

	// users 行と OIDC identity（正規化後のログイン突き合わせの正）を単一トランザクションで作る。
	// 旧カラム users.cognito_sub の撤去（FRESTYLE-311 PR3）で「ユーザーと識別子が同一 INSERT で
	// atomic に書かれる」性質が失われるため、identity 作成を user 作成と不可分にして
	// 識別子を持たない孤児ユーザー（ログイン不能）が生まれないようにする。
	if err := u.users.CreateWithOidcIdentity(ctx, user, domain.OidcProviderCognito, sub); err != nil {
		return false, fmt.Errorf("create user with oidc identity: %w", err)
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
