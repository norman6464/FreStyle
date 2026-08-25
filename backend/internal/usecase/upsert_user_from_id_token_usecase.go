package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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
	// bootstrapSuperAdminEmail は招待なしのサインアップを許す唯一の例外アドレス
	// （空なら例外なし）。詳しくは bootstrapSignupAllowed を参照。
	bootstrapSuperAdminEmail string
}

// NewUpsertUserFromIDTokenUseCase はUpsertUserFromIDTokenUseCaseを生成する。
// bootstrapSuperAdminEmail は「最初の運営管理者」だけに招待を免除するアドレス（通常は空）。
func NewUpsertUserFromIDTokenUseCase(
	users repository.UserRepository,
	invitations repository.AdminInvitationRepository,
	bootstrapSuperAdminEmail string,
) *UpsertUserFromIDTokenUseCase {
	return &UpsertUserFromIDTokenUseCase{
		users:                    users,
		invitations:              invitations,
		bootstrapSuperAdminEmail: strings.TrimSpace(bootstrapSuperAdminEmail),
	}
}

// bootstrapSignupAllowed は、招待の無い新規サインアップを「最初の運営管理者」に限って許すかを返す。
//
// Cognito の admin グループに属しているだけで招待を迂回できると、グループ名 1 つで会社をまたぐ
// super_admin を、招待（FreStyle 唯一のアカウント発行統制）を通さずに作れてしまう。一方でこの
// 免除は「まだ super_admin が 1 人も居ない環境で最初の 1 人を作る」唯一の経路でもあり、単純に
// 消すと新環境で誰もログインできなくなる。そこで次の 3 つが揃ったときだけ通す:
//
//  1. 運用者が明示した bootstrapSuperAdminEmail と一致する（未設定なら免除は一切効かない）
//  2. Cognito の admin グループに属している
//  3. まだ super_admin が 1 人も居ない
//
// 3 により、最初の 1 人ができた瞬間にこの経路は自動的に閉じる。
func (u *UpsertUserFromIDTokenUseCase) bootstrapSignupAllowed(
	ctx context.Context,
	in UpsertUserFromIDTokenInput,
) (bool, error) {
	if u.bootstrapSuperAdminEmail == "" || !in.IsCognitoAdmin || in.Email == "" {
		return false, nil
	}
	if !strings.EqualFold(strings.TrimSpace(in.Email), u.bootstrapSuperAdminEmail) {
		return false, nil
	}
	// 既存の運営管理者を数える。取得できないときは「居ないこと」を確認できていないので
	// 免除しない（fail closed）。
	admins, err := u.users.ListByRole(ctx, domain.RoleSuperAdmin)
	if err != nil {
		return false, fmt.Errorf("list super admins for bootstrap: %w", err)
	}
	return len(admins) == 0, nil
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
			slog.WarnContext(ctx, "ensure oidc identity failed (self-heal, non-fatal)", "userID", existing.ID, "err", err)
		}
		return true, nil
	}

	if inv == nil {
		bootstrap, bootstrapErr := u.bootstrapSignupAllowed(ctx, in)
		if bootstrapErr != nil {
			return false, bootstrapErr
		}
		if !bootstrap {
			slog.WarnContext(
				ctx,
				"signup blocked: invitation required",
				"cognitoSub", sub,
				"email", email,
				"tokenProvided", invitationToken != "",
				"cognitoAdminGroup", isCognitoAdmin,
			)
			return false, nil
		}
		slog.WarnContext(
			ctx,
			"bootstrap signup allowed: creating the first super admin without invitation",
			"cognitoSub", sub,
			"email", email,
		)
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
