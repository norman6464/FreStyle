package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
	users             repository.UserRepository
	invitations       repository.AdminInvitationRepository
	transactionRunner repository.UserInvitationTransactionRunner
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
	transactionRunner repository.UserInvitationTransactionRunner,
) *UpsertUserFromIDTokenUseCase {
	return &UpsertUserFromIDTokenUseCase{
		users:                    users,
		invitations:              invitations,
		transactionRunner:        transactionRunner,
		bootstrapSuperAdminEmail: domain.NormalizeEmail(bootstrapSuperAdminEmail),
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
// 3 により、最初の 1 人ができた瞬間にこの経路は自動的に閉じる。ただしこの照会は「作成の直前に
// 一度見た事実」でしかなく、これだけでは同時実行で 2 人目を止められない。実際に閉じることを
// 保証するのは作成側（CreateFirstSuperAdminWithOidcIdentity）で、判定と INSERT を同じ
// トランザクションに入れて直列化する。ここでの照会は、作成まで進まずに早く拒否して
// 経路を記録するための前段。
//
// email は呼び出し元が domain.NormalizeEmail で畳んだ値を渡すこと。比較は正規形どうしの
// 一致で行い、EqualFold のような「畳んでから比べる」比較はしない（アプリでは同一・DB では
// 別行という食い違いを作らないため）。両方が空文字なら一致してしまうので、
// bootstrapSuperAdminEmail == "" と email == "" のガードは比較の前提として効いている。
func (u *UpsertUserFromIDTokenUseCase) bootstrapSignupAllowed(
	ctx context.Context,
	email string,
	isCognitoAdmin bool,
) (bool, error) {
	if u.bootstrapSuperAdminEmail == "" || !isCognitoAdmin || email == "" {
		return false, nil
	}
	if email != u.bootstrapSuperAdminEmail {
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

// updateExistingUser は既存 user へ OIDC / 招待の内容を反映する（氏名・役割・所属・招待の受諾）。
//
// ここで呼ぶ repository の書き込みは、対象行が 1 行も無いと domain.ErrNotFound を返す。
// existing / inv はどちらも直前に読み出したものなので、実際に not-found になるのは
// 「読み出しと書き込みのあいだに user / 招待が消えた」競合だけ。その場合はエラーを
// そのまま返してログインを失敗させる。握り潰して先へ進めると、昇格や所属の付け替えが
// 1 行も当たっていないのに反映済みのものとしてセッションを張ることになる。
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

// Execute はユーザー情報と招待情報を基にユーザーを作成・更新し、解決した user を返す。
// nil, nil は bootstrap の同時実行負けのときに返る。同じ email での同時サインアップ競合は
// nil, repository.ErrEmailTaken を返す（呼び出し元が原因を区別できるよう別扱いにする）。
func (u *UpsertUserFromIDTokenUseCase) Execute(
	ctx context.Context,
	in UpsertUserFromIDTokenInput,
) (user *domain.User, err error) {
	if u.users == nil {
		return nil, errors.New("user repository not configured")
	}

	sub := in.CognitoSub
	if sub == "" {
		return nil, errors.New("id_token missing sub")
	}

	// email はここで 1 度だけ正規形へ畳み、以後の照会・比較・保存すべてでこの値を使う。
	// 生の claim 値のまま保存すると、免除の比較（畳んで一致）と DB の一意索引・byte 一致検索
	// （畳まない）で同一性の定義がずれ、同じアドレスの行が複数作れてしまう。
	email := domain.NormalizeEmail(in.Email)
	oidcName := in.Name
	isCognitoAdmin := in.IsCognitoAdmin
	invitationToken := in.InvitationToken

	var inv *domain.AdminInvitation
	if u.invitations != nil {
		if invitationToken != "" {
			var findErr error
			inv, findErr = u.invitations.FindPendingByToken(ctx, invitationToken)
			if findErr != nil {
				return nil, fmt.Errorf(
					"find pending invitation by token: %w",
					findErr,
				)
			}
		}

		if inv == nil && email != "" {
			var findErr error
			inv, findErr = u.invitations.FindPendingByEmail(ctx, email)
			if findErr != nil {
				return nil, fmt.Errorf(
					"find pending invitation by email: %w",
					findErr,
				)
			}
		}
	}

	existing, findErr := u.users.FindByCognitoSub(ctx, sub)
	if findErr != nil {
		return nil, fmt.Errorf(
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
			return nil, fmt.Errorf("update existing user: %w", err)
		}
		// user_oidc_identities への冪等な保険。FindByCognitoSub は identity を突き合わせ条件に
		// するため通常この時点で identity は既に存在するが、provider ごとの張り直しを冪等に保証して
		// おく（失敗してもログイン自体は成立しているため致命扱いにしない）。
		if err := u.users.EnsureOidcIdentity(ctx, existing.ID, domain.OidcProviderCognito, sub); err != nil {
			slog.WarnContext(ctx, "ensure oidc identity failed (self-heal, non-fatal)", "userID", existing.ID, "err", err)
		}
		return existing, nil
	}

	// bootstrap 以外の招待なし新規ユーザーは、下の通常経路（role は既定の trainee）へ流れる。
	bootstrap := false
	if inv == nil {
		var bootstrapErr error
		bootstrap, bootstrapErr = u.bootstrapSignupAllowed(ctx, email, isCognitoAdmin)
		if bootstrapErr != nil {
			return nil, bootstrapErr
		}
		if bootstrap {
			slog.WarnContext(
				ctx,
				"bootstrap signup allowed: creating the first super admin without invitation",
				"cognitoSub", sub,
				"email", email,
			)
		} else {
			slog.InfoContext(
				ctx,
				"self signup: creating a new user without invitation",
				"cognitoSub", sub,
				"email", email,
			)
		}
	}

	role := domain.RoleTrainee
	var companyID *uint64

	name := email
	if oidcName != "" {
		name = oidcName
	}

	// Cognito admin グループへの所属だけでは super_admin にしない（招待か bootstrap が要る）。
	// 外すとグループ名 1 つで招待統制を経ない super_admin が作れてしまう。
	if isCognitoAdmin && (inv != nil || bootstrap) {
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

	user = &domain.User{
		Email:     email,
		Name:      name,
		Role:      role,
		CompanyID: companyID,
	}

	// users 行と OIDC identity（正規化後のログイン突き合わせの正）を単一トランザクションで作る。
	// 旧カラム users.cognito_sub の撤去（FRESTYLE-311 PR3）で「ユーザーと識別子が同一 INSERT で
	// atomic に書かれる」性質が失われるため、identity 作成を user 作成と不可分にして
	// 識別子を持たない孤児ユーザー（ログイン不能）が生まれないようにする。
	//
	// 招待を経ない bootstrap 経路だけは「super_admin が 0 人」の判定も同じトランザクションに
	// 入れる。判定と作成が別トランザクションだと、同時に来た 2 本がどちらも「0 人」を見て
	// どちらも作れてしまい、「最初の 1 人ができた瞬間に閉じる」という不変条件が破れる。
	if bootstrap {
		created, createErr := u.users.CreateFirstSuperAdminWithOidcIdentity(
			ctx, user, domain.OidcProviderCognito, sub,
		)
		if createErr != nil {
			return nil, fmt.Errorf("create first super admin: %w", createErr)
		}
		if !created {
			// 判定と作成のあいだに別の運営管理者ができた。免除は既に閉じているので拒否する。
			slog.WarnContext(
				ctx,
				"bootstrap signup rejected: another super admin was created concurrently",
				"cognitoSub", sub,
				"email", email,
			)
			return nil, nil
		}
		return user, nil
	}

	if u.transactionRunner == nil {
		return nil, errors.New("user invitation transaction runner not configured")
	}

	if err := u.transactionRunner.WithinTransaction(
		ctx,
		func(
			users repository.UserWithOidcIdentityCreator,
			invitations repository.InvitationStatusUpdater,
		) error {
			if err := users.CreateWithOidcIdentity(
				ctx,
				user,
				domain.OidcProviderCognito,
				sub,
			); err != nil {
				return fmt.Errorf("create user with oidc identity: %w", err)
			}

			// 招待を経ない自己サインアップでは inv が nil のまま（受諾する招待が無い）。
			if inv == nil {
				return nil
			}
			if err := invitations.UpdateStatus(
				ctx,
				inv.ID,
				domain.InvitationStatusAccepted,
			); err != nil {
				return fmt.Errorf("accept invitation: %w", err)
			}
			return nil
		},
	); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			// 同じ email で同時にサインアップが競合した（同一人物の二重送信・招待の
			// 二重受諾など）。別の sub で先に確定しているだけなので、呼び出し元が
			// bootstrap 競合負け（nil, nil）と区別できるよう ErrEmailTaken をそのまま返す。
			slog.WarnContext(ctx, "signup rejected: email already taken by a concurrent signup", "cognitoSub", sub, "email", email)
			return nil, repository.ErrEmailTaken
		}
		return nil, fmt.Errorf("create user and accept invitation: %w", err)
	}

	return user, nil
}
