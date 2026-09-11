package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// GetCurrentUserUseCase は OIDC の subject から現在のユーザー情報を引く。
type GetCurrentUserUseCase struct {
	users repository.UserRepository
}

func NewGetCurrentUserUseCase(users repository.UserRepository) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{users: users}
}

func (u *GetCurrentUserUseCase) Execute(ctx context.Context, subject string) (*domain.User, error) {
	return u.users.FindByOidcSubject(ctx, subject)
}

// LookupUserDisplayUseCase はユーザー ID から、人を表示するのに要る最小限
// （表示名・アイコン・状態メッセージ）を引く。
//
// チケットの作成者・変更履歴の実行者・発言の投稿者・ページの最終編集者、どの画面も
// これを解決の単位にする（domain.UserDisplay の doc 参照）。kb / ticket / comment の
// どの usecase サブパッケージからも import されない中立の置き場所として user に置く
// （usecase サブパッケージ同士は import しない規約のため、handler 層が各パッケージの
// usecase と並べてこれを直接保持する）。
type LookupUserDisplayUseCase struct {
	users repository.UserRepository
}

func NewLookupUserDisplayUseCase(users repository.UserRepository) *LookupUserDisplayUseCase {
	return &LookupUserDisplayUseCase{users: users}
}

// Execute は表示情報を返す。見つからなければ (nil, nil)（「最終編集者」のような
// 付随情報のために、本体の応答自体を失敗にはしない — handler の判断に委ねる）。
// repository の失敗はそのまま伝える。
func (u *LookupUserDisplayUseCase) Execute(ctx context.Context, userID uint64) (*domain.UserDisplay, error) {
	return u.users.FindDisplayByID(ctx, userID)
}

// UpsertUserFromIDTokenInput はIDトークンから取得したユーザー情報を表す。
type UpsertUserFromIDTokenInput struct {
	Subject string
	Email   string
	Name    string
	// EmailVerified は id_token の email_verified クレーム。false のときは Email を
	// 「無い」ものとして扱う（同一性は Subject だけで決める）。発行者が未検証のメール
	// アドレスでのサインアップを許す設定だと、検証していない相手が他人のアドレスを
	// 名乗って先取りできてしまうため（そのアドレスは users.email の一意索引に載るので、
	// 本当の持ち主が以後登録できなくなる）。
	EmailVerified bool
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

	sub := in.Subject
	if sub == "" {
		return nil, errors.New("id_token missing sub")
	}

	// email はここで 1 度だけ正規形へ畳み、以後の照会・比較・保存すべてでこの値を使う。
	// 生の claim 値のまま保存すると、DB の一意索引・byte 一致検索（畳まない）と
	// 同一性の定義がずれ、同じアドレスの行が複数作れてしまう。
	//
	// 検証していないアドレスは畳む前に「無い」ものとして扱う（UpsertUserFromIDTokenInput.
	// EmailVerified の doc 参照）。同一性は Subject だけで決まるので、これで作成・照会の
	// どちらも壊れない（uq_users_email_active は email が空文字の行を対象外にしている）。
	email := ""
	if in.EmailVerified {
		email = domain.NormalizeEmail(in.Email)
	}
	oidcName := in.Name

	existing, findErr := u.users.FindByOidcSubject(ctx, sub)
	if findErr != nil {
		return nil, fmt.Errorf(
			"find user by oidc subject: %w",
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
		// 検証済みのアドレスを、それまで持っていなかった相手へ後から付ける。
		// サインアップ時点では未検証で email を持てなかった相手が、後日
		// （発行者側で）確認リンクを踏んでから改めてログインしてきた場合の経路。
		// 既に別のアクティブユーザーがそのアドレスを使っていれば ErrEmailTaken が返るが、
		// ログイン自体は成立させる（identity の自己修復と同じ非致命扱い）。
		if email != "" && existing.Email == "" {
			if err := u.users.UpdateEmail(ctx, existing.ID, email); err != nil {
				if errors.Is(err, repository.ErrEmailTaken) {
					slog.WarnContext(ctx, "backfill verified email skipped: already used by another active user (non-fatal)", "userID", existing.ID)
				} else {
					slog.WarnContext(ctx, "backfill verified email failed (non-fatal)", "userID", existing.ID, "err", err)
				}
			} else {
				existing.Email = email
			}
		}
		// user_oidc_identities への冪等な保険。FindByOidcSubject は identity を突き合わせ条件に
		// するため通常この時点で identity は既に存在するが、provider ごとの張り直しを冪等に保証して
		// おく（失敗してもログイン自体は成立しているため致命扱いにしない）。
		if err := u.oidcIdentities.EnsureIdentity(ctx, existing.ID, domain.OidcProviderDefault, sub); err != nil {
			slog.WarnContext(ctx, "ensure oidc identity failed (self-heal, non-fatal)", "userID", existing.ID, "err", err)
		}
		return existing, nil
	}

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
		return u.oidcIdentities.EnsureIdentity(ctx, user.ID, domain.OidcProviderDefault, sub)
	}); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			// 同じ email で同時にサインアップが競合した（同一人物の二重送信など）。
			// 別の sub で先に確定しているだけなので、呼び出し元が区別できるよう
			// ErrEmailTaken をそのまま返す。ログには生の subject / email を書かない
			// （ログの保管先は DB より読める人が広いことがある）。
			slog.WarnContext(ctx, "signup rejected: email already taken by a concurrent signup")
			return nil, repository.ErrEmailTaken
		}
		return nil, fmt.Errorf("create user with oidc identity: %w", err)
	}

	// 生の subject / email ではなく、確定した内部 user.ID だけを記録する
	// （この関数の「既存ユーザー」側の各ログが既に同じ扱い）。
	slog.InfoContext(ctx, "self signup: created a new user", "userID", user.ID)

	return user, nil
}
