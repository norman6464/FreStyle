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
	CognitoSub string
	Email      string
	Name       string
}

// UpsertUserFromIDTokenUseCase は認証済みユーザーの作成・更新を行う。
type UpsertUserFromIDTokenUseCase struct {
	users          repository.UserRepository
	oidcIdentities repository.UserOidcIdentityRepository
	txManager      repository.TxManager
}

// NewUpsertUserFromIDTokenUseCase はUpsertUserFromIDTokenUseCaseを生成する。
func NewUpsertUserFromIDTokenUseCase(
	users repository.UserRepository,
	oidcIdentities repository.UserOidcIdentityRepository,
	txManager repository.TxManager,
) *UpsertUserFromIDTokenUseCase {
	return &UpsertUserFromIDTokenUseCase{
		users:          users,
		oidcIdentities: oidcIdentities,
		txManager:      txManager,
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

// Execute はユーザー情報を基にユーザーを作成・更新し、解決した user を返す。
// 同じ email での同時サインアップ競合は nil, repository.ErrEmailTaken を返す
// （呼び出し元が原因を区別できるよう別扱いにする）。
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
	// 生の claim 値のまま保存すると、DB の一意索引・byte 一致検索（畳まない）と
	// 同一性の定義がずれ、同じアドレスの行が複数作れてしまう。
	email := domain.NormalizeEmail(in.Email)
	oidcName := in.Name

	existing, findErr := u.users.FindByCognitoSub(ctx, sub)
	if findErr != nil {
		return nil, fmt.Errorf(
			"find user by cognito sub: %w",
			findErr,
		)
	}

	if existing != nil {
		if u.shouldBackfillName(oidcName, existing) {
			if err := u.users.UpdateName(ctx, existing.ID, oidcName); err != nil {
				return nil, fmt.Errorf("update existing user name: %w", err)
			}
			existing.Name = oidcName
		}
		// user_oidc_identities への冪等な保険。FindByCognitoSub は identity を突き合わせ条件に
		// するため通常この時点で identity は既に存在するが、provider ごとの張り直しを冪等に保証して
		// おく（失敗してもログイン自体は成立しているため致命扱いにしない）。
		if err := u.oidcIdentities.EnsureIdentity(ctx, existing.ID, domain.OidcProviderCognito, sub); err != nil {
			slog.WarnContext(ctx, "ensure oidc identity failed (self-heal, non-fatal)", "userID", existing.ID, "err", err)
		}
		return existing, nil
	}

	slog.InfoContext(
		ctx,
		"self signup: creating a new user",
		"cognitoSub", sub,
		"email", email,
	)

	name := email
	if oidcName != "" {
		name = oidcName
	}

	user = &domain.User{
		Email: email,
		Name:  name,
	}

	// users 行と OIDC identity は不可分に作る（正規化後は識別子を持たないユーザーは存在し得ない）。
	// identity 側が競合などで失敗すればトランザクションごと巻き戻り、users 行だけが残る
	// （＝ログイン不能な孤児）状態を作らない。
	if err := u.txManager.DoInTx(ctx, func(ctx context.Context) error {
		if err := u.users.Create(ctx, user); err != nil {
			return err
		}
		return u.oidcIdentities.EnsureIdentity(ctx, user.ID, domain.OidcProviderCognito, sub)
	}); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			// 同じ email で同時にサインアップが競合した（同一人物の二重送信など）。
			// 別の sub で先に確定しているだけなので、呼び出し元が区別できるよう
			// ErrEmailTaken をそのまま返す。
			slog.WarnContext(ctx, "signup rejected: email already taken by a concurrent signup", "cognitoSub", sub, "email", email)
			return nil, repository.ErrEmailTaken
		}
		return nil, fmt.Errorf("create user with oidc identity: %w", err)
	}

	return user, nil
}
