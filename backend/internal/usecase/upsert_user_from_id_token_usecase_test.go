package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

type upsertUserRepoSpy struct {
	stubUserRepo
	created         *domain.User
	createdProvider string
	createdSubject  string

	findByCognitoSubCalls int
	createCalls           int
	createErr             error
	nameUpdateCalls       int
	nameUpdateErr         error
	ensureIdentityCalls   int
	ensuredUserID         uint64
	ensuredProvider       string
	ensuredSubject        string
	roleUpdateCalls       int
	roleUpdateErr         error
	workspaceUpdateCalls  int
	workspaceUpdateErr    error

	roleUpdateUserID      uint64
	roleUpdateValue       domain.RoleName
	workspaceUpdateUserID uint64
	workspaceUpdateValue  *string

	// ブートストラップ判定（既存 super_admin の有無）の制御。
	createFirstSuperAdminCalls int
	superAdmins                []domain.User
	listByRoleErr              error
	listByRoleCalls            int
	listByRoleValue            domain.RoleName
}

func (s *upsertUserRepoSpy) ListByRole(
	_ context.Context,
	role domain.RoleName,
) ([]domain.User, error) {
	s.listByRoleCalls++
	s.listByRoleValue = role
	if s.listByRoleErr != nil {
		return nil, s.listByRoleErr
	}
	return s.superAdmins, nil
}

func (s *upsertUserRepoSpy) FindByCognitoSub(
	ctx context.Context,
	sub string,
) (*domain.User, error) {
	s.findByCognitoSubCalls++
	return s.stubUserRepo.FindByCognitoSub(ctx, sub)
}

type upsertInvitationRepoSpy struct {
	pending         *domain.AdminInvitation
	pendingByToken  *domain.AdminInvitation
	tokenFindErr    error
	emailFindErr    error
	tokenFindCalled bool
	emailFindCalled bool
	emailFindArg    string
	updateCalls     int
	updateErr       error
	updatedID       uint64
	updatedStatus   string
}

type fakeUserInvitationTransactionRunner struct {
	users       repository.UserWithOidcIdentityCreator
	invitations repository.InvitationStatusUpdater
	rollback    func()
}

func (r *fakeUserInvitationTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(
		users repository.UserWithOidcIdentityCreator,
		invitations repository.InvitationStatusUpdater,
	) error,
) error {
	err := fn(r.users, r.invitations)
	if err != nil && r.rollback != nil {
		r.rollback()
	}
	return err
}

// upsertWsA / upsertWsB は招待の workspace_id 比較を固定するための 2 つのワークスペース ID。
const (
	upsertWsA = "0198a000-0000-7000-8000-0000000000a1"
	upsertWsB = "0198a000-0000-7000-8000-0000000000a2"
)

func (s *upsertInvitationRepoSpy) ListAll(
	_ context.Context,
) ([]domain.AdminInvitation, error) {
	return nil, nil
}

func (s *upsertInvitationRepoSpy) ListByWorkspaceID(
	_ context.Context,
	_ string,
) ([]domain.AdminInvitation, error) {
	return nil, nil
}

func (s *upsertInvitationRepoSpy) FindPendingByEmail(
	_ context.Context,
	email string,
) (*domain.AdminInvitation, error) {
	s.emailFindCalled = true
	s.emailFindArg = email
	return s.pending, s.emailFindErr
}

func (s *upsertInvitationRepoSpy) FindPendingByToken(
	_ context.Context,
	_ string,
) (*domain.AdminInvitation, error) {
	s.tokenFindCalled = true
	return s.pendingByToken, s.tokenFindErr
}

func (s *upsertInvitationRepoSpy) FindByID(
	_ context.Context,
	_ uint64,
) (*domain.AdminInvitation, error) {
	return nil, nil
}

func (s *upsertInvitationRepoSpy) Create(
	_ context.Context,
	_ *domain.AdminInvitation,
) error {
	return nil
}

func (s *upsertInvitationRepoSpy) UpdateStatus(
	_ context.Context,
	id uint64,
	status string,
) error {
	s.updateCalls++
	s.updatedID = id
	s.updatedStatus = status
	return s.updateErr
}

func (s *upsertUserRepoSpy) EnsureOidcIdentity(
	_ context.Context,
	userID uint64,
	provider, subject string,
) error {
	s.ensureIdentityCalls++
	s.ensuredUserID = userID
	s.ensuredProvider = provider
	s.ensuredSubject = subject
	return nil
}

func newUpsertUserFromIDTokenUseCaseForTest(
	users repository.UserRepository,
	invitations repository.AdminInvitationRepository,
) *UpsertUserFromIDTokenUseCase {
	// ブートストラップ免除なし（本番の既定）。
	return NewUpsertUserFromIDTokenUseCase(
		users,
		invitations,
		"",
		&fakeUserInvitationTransactionRunner{
			users:       users,
			invitations: invitations,
		},
	)
}

// newUpsertUserFromIDTokenUseCaseWithBootstrap はブートストラップ用アドレスを設定した usecase を返す。
func newUpsertUserFromIDTokenUseCaseWithBootstrap(
	users repository.UserRepository,
	invitations repository.AdminInvitationRepository,
	bootstrapEmail string,
) *UpsertUserFromIDTokenUseCase {
	return NewUpsertUserFromIDTokenUseCase(
		users,
		invitations,
		bootstrapEmail,
		&fakeUserInvitationTransactionRunner{
			users:       users,
			invitations: invitations,
		},
	)
}

func Test_UpsertUserFromIDToken_招待のRoleとワークスペースを適用してAcceptedにする(t *testing.T) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:          10,
			Role:        domain.RoleCompanyAdmin,
			WorkspaceID: strPtr(upsertWsA),
			Name:        "Invited User",
			Status:      domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "invited-sub",
			Email:      "invited@example.com",
			Name:       "OIDC User",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("有効な招待があるユーザーは許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleCompanyAdmin {
		t.Fatalf(
			"role = %q, want %q",
			users.created.Role,
			domain.RoleCompanyAdmin,
		)
	}
	if users.created.WorkspaceID == nil || *users.created.WorkspaceID != upsertWsA {
		t.Fatalf("workspaceID = %v, want %q", users.created.WorkspaceID, upsertWsA)
	}
	if users.created.Name != "Invited User" {
		t.Fatalf("name = %q, want %q", users.created.Name, "Invited User")
	}
	if invitations.updatedID != 10 {
		t.Fatalf("updated invitation ID = %d, want 10", invitations.updatedID)
	}
	if invitations.updatedStatus != domain.InvitationStatusAccepted {
		t.Fatalf(
			"status = %q, want %q",
			invitations.updatedStatus,
			domain.InvitationStatusAccepted,
		)
	}
}

func (s *upsertUserRepoSpy) CreateWithOidcIdentity(
	_ context.Context,
	user *domain.User,
	provider, subject string,
) error {
	s.createCalls++
	if s.createErr != nil {
		return s.createErr
	}

	copied := *user
	s.created = &copied
	s.createdProvider = provider
	s.createdSubject = subject
	return nil
}

// CreateFirstSuperAdminWithOidcIdentity は本物の repository と同じく「渡された role が
// super_admin でなければ拒否し、super_admin が 0 人のときだけ作る」振る舞いを模す
// （判定は spy が持つ superAdmins を見る）。
func (s *upsertUserRepoSpy) CreateFirstSuperAdminWithOidcIdentity(
	ctx context.Context,
	user *domain.User,
	provider, subject string,
) (bool, error) {
	s.createFirstSuperAdminCalls++
	// 本物（persistence.userRepository）はこの経路を super_admin 専用として先頭で弾く。
	// ダブルがこの事前条件を持たないと、免除の条件が壊れて trainee がこの経路に流れても
	// テストは緑のままになり、本番でだけ「最初の運営管理者を作れない」状態を出荷してしまう。
	if user.Role != domain.RoleSuperAdmin {
		return false, fmt.Errorf("最初の運営管理者の作成に role %q が渡されました（super_admin 専用の経路です）", user.Role)
	}
	if len(s.superAdmins) > 0 {
		return false, nil
	}
	if err := s.CreateWithOidcIdentity(ctx, user, provider, subject); err != nil {
		return false, err
	}
	s.superAdmins = append(s.superAdmins, *user)
	return true, nil
}

func (s *upsertUserRepoSpy) UpdateName(
	_ context.Context,
	_ uint64,
	_ string,
) error {
	s.nameUpdateCalls++
	return s.nameUpdateErr
}

func (s *upsertUserRepoSpy) UpdateRole(
	_ context.Context,
	userID uint64,
	role domain.RoleName,
) error {
	s.roleUpdateCalls++
	if s.roleUpdateErr != nil {
		return s.roleUpdateErr
	}

	s.roleUpdateUserID = userID
	s.roleUpdateValue = role
	return nil
}

func (s *upsertUserRepoSpy) UpdateWorkspaceID(
	_ context.Context,
	userID uint64,
	workspaceID *string,
) error {
	s.workspaceUpdateCalls++
	if s.workspaceUpdateErr != nil {
		return s.workspaceUpdateErr
	}

	s.workspaceUpdateUserID = userID
	s.workspaceUpdateValue = workspaceID
	return nil
}

func Test_UpsertUserFromIDToken_既存TraineeをCompanyAdminへ昇格して会社に紐付ける(t *testing.T) {
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{
				ID:    7,
				Email: "existing@example.com",
				Name:  "Existing User",
				Role:  domain.RoleTrainee,
			},
		},
	}
	invWorkspaceID := "0198a000-0000-7000-8000-0000000000ab"
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:          20,
			Role:        domain.RoleCompanyAdmin,
			WorkspaceID: &invWorkspaceID,
			Status:      domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-sub",
			Email:      "existing@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("既存ユーザーは許可されるべき")
	}
	if users.created != nil {
		t.Fatal("既存ユーザーを新規作成してはいけない")
	}
	if users.roleUpdateUserID != 7 {
		t.Fatalf("role更新対象ID = %d, want 7", users.roleUpdateUserID)
	}
	if users.roleUpdateValue != domain.RoleCompanyAdmin {
		t.Fatalf(
			"更新role = %q, want %q",
			users.roleUpdateValue,
			domain.RoleCompanyAdmin,
		)
	}
	if users.workspaceUpdateUserID != 7 {
		t.Fatalf("ワークスペース更新対象ID = %d, want 7", users.workspaceUpdateUserID)
	}
	if users.workspaceUpdateValue == nil || *users.workspaceUpdateValue != invWorkspaceID {
		t.Fatalf(
			"更新workspaceID = %v, want %q（招待の workspace_id をサブクエリで引き直さずそのまま渡す）",
			users.workspaceUpdateValue, invWorkspaceID,
		)
	}
	if invitations.updatedID != 20 {
		t.Fatalf("更新された招待ID = %d, want 20", invitations.updatedID)
	}
	if invitations.updatedStatus != domain.InvitationStatusAccepted {
		t.Fatalf(
			"招待status = %q, want %q",
			invitations.updatedStatus,
			domain.InvitationStatusAccepted,
		)
	}
}

// 招待ゲートは撤去済み（個人サインアップ）。招待も管理者権限も無い新規ユーザーは
// 拒否されず、所属ワークスペース無しで作られる。
func Test_UpsertUserFromIDToken_招待も管理者権限もない新規ユーザーは自己サインアップできる(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, nil)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "new-sub",
			Email:      "new@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("招待の無い新規ユーザーも自己サインアップできるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleTrainee {
		t.Fatalf("role = %q, want %q", users.created.Role, domain.RoleTrainee)
	}
	if users.created.WorkspaceID != nil {
		t.Fatalf("workspaceID = %v, want nil", users.created.WorkspaceID)
	}
}

// Cognito の admin グループに属しているだけでは招待統制を迂回できない（グループ名 1 つで
// 会社をまたぐ super_admin が作れてしまう穴を塞ぐ）。自己サインアップ自体は許すが、
// 招待も bootstrap も無ければ super_admin へは昇格させない。
func Test_UpsertUserFromIDToken_CognitoAdminでも招待が無ければSuperAdminへ昇格しない(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, nil)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:     "admin-sub",
			Email:          "admin@example.com",
			Name:           "Admin User",
			IsCognitoAdmin: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("招待の無い新規ユーザーも自己サインアップできるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleTrainee {
		t.Fatalf("role = %q, want %q（Cognito admin グループだけで昇格してはいけない）", users.created.Role, domain.RoleTrainee)
	}
}

// ブートストラップ: 明示したアドレス + Cognito admin グループ + super_admin が 0 人のときだけ、
// 招待なしで最初の運営管理者を作れる。
func Test_UpsertUserFromIDToken_ブートストラップ指定アドレスは招待なしでSuperAdminを作る(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, "  Ops@Example.com ")

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:     "admin-sub",
			Email:          "ops@example.com",
			Name:           "Ops User",
			IsCognitoAdmin: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("ブートストラップ指定アドレスは許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleSuperAdmin {
		t.Fatalf("role = %q, want %q", users.created.Role, domain.RoleSuperAdmin)
	}
	if users.listByRoleValue != domain.RoleSuperAdmin {
		t.Fatalf("既存 super_admin の有無を確認していない: %q", users.listByRoleValue)
	}
}

// ブートストラップは「最初の 1 人」限定。super_admin が既に居れば bootstrap は閉じ、
// 通常の自己サインアップ（既定ロール）へ流れる。
func Test_UpsertUserFromIDToken_ブートストラップはSuperAdmin在籍時に閉じる(t *testing.T) {
	users := &upsertUserRepoSpy{
		superAdmins: []domain.User{{ID: 1, Role: domain.RoleSuperAdmin}},
	}
	uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, "ops@example.com")

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:     "admin-sub",
			Email:          "ops@example.com",
			IsCognitoAdmin: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("bootstrap が閉じても自己サインアップは許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleTrainee {
		t.Fatalf("role = %q, want %q（bootstrap 対象外なら昇格しない）", users.created.Role, domain.RoleTrainee)
	}
	// 既に居ると分かっている以上、bootstrap の作成の試行まで進まない
	// （作成側の再判定は同時実行のための最後の砦で、事前の照会の代わりではない）。
	if users.createFirstSuperAdminCalls != 0 {
		t.Fatalf("bootstrap 作成を試みてはいけない: %d 回呼ばれた", users.createFirstSuperAdminCalls)
	}
}

// ブートストラップは指定した 1 アドレスだけ。別アドレスや admin グループ非所属には
// 効かず、既定ロールで自己サインアップとして作られる。
func Test_UpsertUserFromIDToken_ブートストラップは指定アドレス以外に効かない(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		isCognitoAdmin bool
	}{
		{name: "別アドレス", email: "other@example.com", isCognitoAdmin: true},
		{name: "adminグループ非所属", email: "ops@example.com", isCognitoAdmin: false},
		{name: "メール無し", email: "", isCognitoAdmin: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &upsertUserRepoSpy{}
			uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, "ops@example.com")

			user, err := uc.Execute(
				context.Background(),
				UpsertUserFromIDTokenInput{
					CognitoSub:     "sub",
					Email:          tt.email,
					IsCognitoAdmin: tt.isCognitoAdmin,
				},
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil {
				t.Fatal("bootstrap 対象外でも自己サインアップは許可されるべき")
			}
			if users.created == nil {
				t.Fatal("ユーザーが作成されていない")
			}
			if users.created.Role != domain.RoleTrainee {
				t.Fatalf("role = %q, want %q（bootstrap 対象外なら昇格しない）", users.created.Role, domain.RoleTrainee)
			}
		})
	}
}

// 既存 super_admin を数えられないときは「居ないこと」を確認できていないので免除しない。
func Test_UpsertUserFromIDToken_ブートストラップ判定の照会失敗はエラーで止める(t *testing.T) {
	listErr := errors.New("db down")
	users := &upsertUserRepoSpy{listByRoleErr: listErr}
	uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, "ops@example.com")

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:     "admin-sub",
			Email:          "ops@example.com",
			IsCognitoAdmin: true,
		},
	)
	if user != nil {
		t.Fatal("照会に失敗したら許可してはいけない")
	}
	if !errors.Is(err, listErr) {
		t.Fatalf("err = %v, want wrapped %v", err, listErr)
	}
	if users.created != nil {
		t.Fatalf("ユーザーを作成してはいけない: %+v", users.created)
	}
}

func Test_UpsertUserFromIDToken_検索エラーを返す(t *testing.T) {
	tokenFindErr := errors.New("token lookup failed")
	emailFindErr := errors.New("email lookup failed")
	userFindErr := errors.New("user lookup failed")

	tests := []struct {
		name        string
		users       *stubUserRepo
		invitations repository.AdminInvitationRepository
		input       UpsertUserFromIDTokenInput
		wantErr     error
		wantMessage string
	}{
		{
			name:  "トークンによる招待検索が失敗する",
			users: &stubUserRepo{},
			invitations: &upsertInvitationRepoSpy{
				tokenFindErr: tokenFindErr,
			},
			input: UpsertUserFromIDTokenInput{
				CognitoSub:      "token-error-sub",
				InvitationToken: "invitation-token",
			},
			wantErr:     tokenFindErr,
			wantMessage: "find pending invitation by token",
		},
		{
			name:  "メールによる招待検索が失敗する",
			users: &stubUserRepo{},
			invitations: &upsertInvitationRepoSpy{
				emailFindErr: emailFindErr,
			},
			input: UpsertUserFromIDTokenInput{
				CognitoSub: "email-error-sub",
				Email:      "user@example.com",
			},
			wantErr:     emailFindErr,
			wantMessage: "find pending invitation by email",
		},
		{
			name: "ユーザー検索が失敗する",
			users: &stubUserRepo{
				err: userFindErr,
			},
			input: UpsertUserFromIDTokenInput{
				CognitoSub: "user-error-sub",
			},
			wantErr:     userFindErr,
			wantMessage: "find user by cognito sub",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := newUpsertUserFromIDTokenUseCaseForTest(
				tc.users,
				tc.invitations,
			)

			user, err := uc.Execute(
				context.Background(),
				tc.input,
			)

			if user != nil {
				t.Fatal("検索エラー時に許可してはいけない")
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want wrapped %v", err, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantMessage) {
				t.Fatalf(
					"error = %q, want message containing %q",
					err.Error(),
					tc.wantMessage,
				)
			}
		})
	}
}

func Test_UpsertUserFromIDToken_招待トークンをメールより優先する(t *testing.T) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{
		pendingByToken: &domain.AdminInvitation{
			ID:          10,
			Role:        domain.RoleTrainee,
			WorkspaceID: strPtr(upsertWsA),
			Name:        "Token User",
			Status:      domain.InvitationStatusPending,
		},
		pending: &domain.AdminInvitation{
			ID:          20,
			Role:        domain.RoleCompanyAdmin,
			WorkspaceID: strPtr(upsertWsB),
			Name:        "Email User",
			Status:      domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:      "invited-sub",
			Email:           "invited@example.com",
			InvitationToken: "invitation-token",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("有効なトークン招待があるユーザーは許可されるべき")
	}
	if !invitations.tokenFindCalled {
		t.Fatal("トークン検索が呼ばれていない")
	}
	if invitations.emailFindCalled {
		t.Fatal("トークンで招待が見つかった場合はメール検索を呼んではいけない")
	}
	if invitations.updatedStatus != domain.InvitationStatusAccepted {
		t.Fatalf(
			"invitation status = %q, want %q",
			invitations.updatedStatus,
			domain.InvitationStatusAccepted,
		)
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleTrainee {
		t.Fatalf(
			"role = %q, want %q",
			users.created.Role,
			domain.RoleTrainee,
		)
	}
	if users.created.WorkspaceID == nil || *users.created.WorkspaceID != upsertWsA {
		t.Fatalf("workspaceID = %v, want %q", users.created.WorkspaceID, upsertWsA)
	}
	if users.created.Name != "Token User" {
		t.Fatalf("name = %q, want %q", users.created.Name, "Token User")
	}
	if invitations.updatedID != 10 {
		t.Fatalf("accepted invitation ID = %d, want 10", invitations.updatedID)
	}
}

func Test_UpsertUserFromIDToken_CognitoSubが空なら処理しない(t *testing.T) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "",
			Email:      "user@example.com",
		},
	)

	if user != nil {
		t.Fatal("CognitoSubが空のユーザーを許可してはいけない")
	}
	if err == nil {
		t.Fatal("CognitoSubが空の場合はエラーを返すべき")
	}
	if !strings.Contains(err.Error(), "id_token missing sub") {
		t.Fatalf(
			"error = %q, want message containing %q",
			err.Error(),
			"id_token missing sub",
		)
	}
	if users.findByCognitoSubCalls != 0 {
		t.Fatalf(
			"FindByCognitoSub calls = %d, want 0",
			users.findByCognitoSubCalls,
		)
	}
	if invitations.tokenFindCalled || invitations.emailFindCalled {
		t.Fatal("CognitoSubが空の場合は招待を検索してはいけない")
	}
	if users.createCalls != 0 {
		t.Fatalf("Create calls = %d, want 0", users.createCalls)
	}
}

func Test_UpsertUserFromIDToken_Cognito管理者は招待Roleで降格しない(
	t *testing.T,
) {
	invitationRoles := []domain.RoleName{
		domain.RoleTrainee,
		domain.RoleCompanyAdmin,
	}

	for _, invitationRole := range invitationRoles {
		t.Run(string(invitationRole), func(t *testing.T) {
			users := &upsertUserRepoSpy{}
			invitations := &upsertInvitationRepoSpy{
				pending: &domain.AdminInvitation{
					ID:     30,
					Role:   invitationRole,
					Status: domain.InvitationStatusPending,
				},
			}
			uc := newUpsertUserFromIDTokenUseCaseForTest(
				users,
				invitations,
			)

			user, err := uc.Execute(
				context.Background(),
				UpsertUserFromIDTokenInput{
					CognitoSub:     "admin-with-invitation",
					Email:          "admin@example.com",
					IsCognitoAdmin: true,
				},
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil {
				t.Fatal("Cognito管理者は許可されるべき")
			}
			if users.created == nil {
				t.Fatal("ユーザーが作成されていない")
			}
			if users.created.Role != domain.RoleSuperAdmin {
				t.Fatalf(
					"role = %q, want %q",
					users.created.Role,
					domain.RoleSuperAdmin,
				)
			}
			if invitations.updateCalls != 1 {
				t.Fatalf(
					"UpdateStatus calls = %d, want 1",
					invitations.updateCalls,
				)
			}
			if invitations.updatedStatus !=
				domain.InvitationStatusAccepted {
				t.Fatalf(
					"status = %q, want %q",
					invitations.updatedStatus,
					domain.InvitationStatusAccepted,
				)
			}
		})
	}
}

func Test_UpsertUserFromIDToken_未対応の招待Roleは適用しない(
	t *testing.T,
) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:     40,
			Role:   domain.RoleSuperAdmin,
			Status: domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "unsupported-role",
			Email:      "user@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("有効な招待があるユーザーは許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleTrainee {
		t.Fatalf(
			"role = %q, want safe default %q",
			users.created.Role,
			domain.RoleTrainee,
		)
	}
	if invitations.updatedStatus != domain.InvitationStatusAccepted {
		t.Fatalf(
			"status = %q, want %q",
			invitations.updatedStatus,
			domain.InvitationStatusAccepted,
		)
	}
}

func Test_UpsertUserFromIDToken_招待のワークスペースが未設定なら未所属にする(
	t *testing.T,
) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:          50,
			Role:        domain.RoleTrainee,
			WorkspaceID: nil,
			Status:      domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "no-company",
			Email:      "no-company@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("有効な招待があるユーザーは許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.WorkspaceID != nil {
		t.Fatalf(
			"workspaceID = %v, want nil",
			users.created.WorkspaceID,
		)
	}
}

func Test_UpsertUserFromIDToken_名前補完の更新に失敗する(t *testing.T) {
	mutationErr := errors.New("mutation failed")
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{
				ID:    7,
				Email: "existing@example.com",
				Name:  "existing@example.com",
				Role:  domain.RoleTrainee,
			},
		},
		nameUpdateErr: mutationErr,
	}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:     60,
			Role:   domain.RoleTrainee,
			Status: domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-user",
			Email:      "existing@example.com",
			Name:       "OIDC User",
		},
	)

	if user != nil {
		t.Fatal("名前補完の更新失敗時にユーザーを許可してはいけない")
	}
	if !errors.Is(err, mutationErr) {
		t.Fatalf("error = %v, want wrapped %v", err, mutationErr)
	}
	if users.nameUpdateCalls != 1 {
		t.Fatalf("UpdateName calls = %d, want 1", users.nameUpdateCalls)
	}
	if invitations.updateCalls != 0 {
		t.Fatalf("UpdateStatus calls = %d, want 0", invitations.updateCalls)
	}
}

func Test_UpsertUserFromIDToken_ロール更新に失敗する(t *testing.T) {
	mutationErr := errors.New("mutation failed")
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{
				ID:    7,
				Email: "existing@example.com",
				Name:  "Existing User",
				Role:  domain.RoleTrainee,
			},
		},
		roleUpdateErr: mutationErr,
	}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:     60,
			Role:   domain.RoleCompanyAdmin,
			Status: domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-user",
			Email:      "existing@example.com",
		},
	)

	if user != nil {
		t.Fatal("ロール更新失敗時にユーザーを許可してはいけない")
	}
	if !errors.Is(err, mutationErr) {
		t.Fatalf("error = %v, want wrapped %v", err, mutationErr)
	}
	if users.roleUpdateCalls != 1 {
		t.Fatalf("UpdateRole calls = %d, want 1", users.roleUpdateCalls)
	}
	if invitations.updateCalls != 0 {
		t.Fatalf("UpdateStatus calls = %d, want 0", invitations.updateCalls)
	}
}

func Test_UpsertUserFromIDToken_ワークスペース更新に失敗する(t *testing.T) {
	mutationErr := errors.New("mutation failed")
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{
				ID:    7,
				Email: "existing@example.com",
				Name:  "Existing User",
				Role:  domain.RoleTrainee,
			},
		},
		workspaceUpdateErr: mutationErr,
	}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:          60,
			Role:        domain.RoleTrainee,
			WorkspaceID: strPtr(upsertWsA),
			Status:      domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-user",
			Email:      "existing@example.com",
		},
	)

	if user != nil {
		t.Fatal("ワークスペース更新失敗時にユーザーを許可してはいけない")
	}
	if !errors.Is(err, mutationErr) {
		t.Fatalf("error = %v, want wrapped %v", err, mutationErr)
	}
	if users.workspaceUpdateCalls != 1 {
		t.Fatalf(
			"UpdateWorkspaceID calls = %d, want 1",
			users.workspaceUpdateCalls,
		)
	}
	if invitations.updateCalls != 0 {
		t.Fatalf("UpdateStatus calls = %d, want 0", invitations.updateCalls)
	}
}

func Test_UpsertUserFromIDToken_招待ステータス更新に失敗する(t *testing.T) {
	mutationErr := errors.New("mutation failed")
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{
				ID:    7,
				Email: "existing@example.com",
				Name:  "Existing User",
				Role:  domain.RoleTrainee,
			},
		},
	}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:     60,
			Role:   domain.RoleTrainee,
			Status: domain.InvitationStatusPending,
		},
		updateErr: mutationErr,
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-user",
			Email:      "existing@example.com",
		},
	)

	if user != nil {
		t.Fatal("招待ステータス更新失敗時にユーザーを許可してはいけない")
	}
	if !errors.Is(err, mutationErr) {
		t.Fatalf("error = %v, want wrapped %v", err, mutationErr)
	}
	if invitations.updateCalls != 1 {
		t.Fatalf("UpdateStatus calls = %d, want 1", invitations.updateCalls)
	}
}

func Test_UpsertUserFromIDToken_既存Cognito管理者は招待を適用しない(
	t *testing.T,
) {
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{
				ID:    70,
				Email: "admin@example.com",
				Name:  "Existing Admin",
				Role:  domain.RoleTrainee,
			},
		},
	}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:          70,
			Role:        domain.RoleCompanyAdmin,
			WorkspaceID: strPtr(upsertWsA),
			Status:      domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:     "existing-admin",
			Email:          "admin@example.com",
			IsCognitoAdmin: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("Cognito管理者は許可されるべき")
	}
	if users.roleUpdateCalls != 1 ||
		users.roleUpdateValue != domain.RoleSuperAdmin {
		t.Fatalf(
			"UpdateRole calls = %d, role = %q",
			users.roleUpdateCalls,
			users.roleUpdateValue,
		)
	}
	if users.workspaceUpdateCalls != 0 {
		t.Fatalf(
			"UpdateWorkspaceID calls = %d, want 0",
			users.workspaceUpdateCalls,
		)
	}
	if invitations.updateCalls != 0 {
		t.Fatalf(
			"UpdateStatus calls = %d, want 0",
			invitations.updateCalls,
		)
	}
}

func Test_UpsertUserFromIDToken_新規ユーザーのトランザクションエラーを返す(
	t *testing.T,
) {
	mutationErr := errors.New("new user mutation failed")

	tests := []struct {
		name                  string
		configure             func(*upsertUserRepoSpy, *upsertInvitationRepoSpy)
		wantCreateCalls       int
		wantUpdateStatusCalls int
	}{
		{
			name: "ユーザー作成に失敗する",
			configure: func(
				users *upsertUserRepoSpy,
				_ *upsertInvitationRepoSpy,
			) {
				users.createErr = mutationErr
			},
			wantCreateCalls:       1,
			wantUpdateStatusCalls: 0,
		},
		{
			name: "ユーザー作成後の招待更新に失敗する",
			configure: func(
				_ *upsertUserRepoSpy,
				invitations *upsertInvitationRepoSpy,
			) {
				invitations.updateErr = mutationErr
			},
			wantCreateCalls:       1,
			wantUpdateStatusCalls: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			users := &upsertUserRepoSpy{}
			invitations := &upsertInvitationRepoSpy{
				pending: &domain.AdminInvitation{
					ID:     70,
					Role:   domain.RoleTrainee,
					Status: domain.InvitationStatusPending,
				},
			}
			tc.configure(users, invitations)

			runner := &fakeUserInvitationTransactionRunner{
				users:       users,
				invitations: invitations,
				rollback: func() {
					users.created = nil
				},
			}
			uc := NewUpsertUserFromIDTokenUseCase(
				users,
				invitations,
				"",
				runner,
			)

			user, err := uc.Execute(
				context.Background(),
				UpsertUserFromIDTokenInput{
					CognitoSub: "new-user-error",
					Email:      "new-user@example.com",
				},
			)

			if user != nil {
				t.Fatal("トランザクション失敗時に許可してはいけない")
			}
			if !errors.Is(err, mutationErr) {
				t.Fatalf(
					"error = %v, want wrapped %v",
					err,
					mutationErr,
				)
			}
			if users.created != nil {
				t.Fatal("処理失敗時にユーザーを残してはいけない")
			}
			if users.createCalls != tc.wantCreateCalls {
				t.Fatalf(
					"Create calls = %d, want %d",
					users.createCalls,
					tc.wantCreateCalls,
				)
			}
			if invitations.updateCalls !=
				tc.wantUpdateStatusCalls {
				t.Fatalf(
					"UpdateStatus calls = %d, want %d",
					invitations.updateCalls,
					tc.wantUpdateStatusCalls,
				)
			}
		})
	}
}

// Test_UpsertUserFromIDToken_同じemailでの同時サインアップはErrEmailTakenを返す は、
// bootstrap 競合負け（nil, nil）と区別できるよう、repository.ErrEmailTaken を
// そのまま呼び出し元へ返すことを固定する（呼び出し元の 403/409 の出し分けが前提にする契約）。
func Test_UpsertUserFromIDToken_同じemailでの同時サインアップはErrEmailTakenを返す(t *testing.T) {
	users := &upsertUserRepoSpy{createErr: repository.ErrEmailTaken}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:     70,
			Role:   domain.RoleTrainee,
			Status: domain.InvitationStatusPending,
		},
	}
	runner := &fakeUserInvitationTransactionRunner{users: users, invitations: invitations}
	uc := NewUpsertUserFromIDTokenUseCase(users, invitations, "", runner)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "race-sub",
			Email:      "race@example.com",
		},
	)

	if user != nil {
		t.Fatal("email 衝突時にユーザーを返してはいけない")
	}
	if !errors.Is(err, repository.ErrEmailTaken) {
		t.Fatalf("error = %v, want wrapped %v", err, repository.ErrEmailTaken)
	}
}

func Test_UpsertUserFromIDToken_新規作成でOIDCidentityを対で作る(t *testing.T) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:     11,
			Role:   domain.RoleTrainee,
			Status: domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "new-sub-1",
			Email:      "new@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("招待ありの新規ユーザーは許可されるべき")
	}
	// 新規ユーザーは users 行と identity を CreateWithOidcIdentity で不可分に作る。
	if users.createCalls != 1 {
		t.Fatalf("CreateWithOidcIdentity calls = %d, want 1", users.createCalls)
	}
	if users.createdProvider != domain.OidcProviderCognito {
		t.Fatalf("provider = %q, want %q", users.createdProvider, domain.OidcProviderCognito)
	}
	if users.createdSubject != "new-sub-1" {
		t.Fatalf("subject = %q, want %q", users.createdSubject, "new-sub-1")
	}
}

func Test_UpsertUserFromIDToken_既存ユーザーでもidentityをセルフヒールする(t *testing.T) {
	existing := &domain.User{ID: 77, Email: "e@example.com", Role: domain.RoleTrainee}
	users := &upsertUserRepoSpy{stubUserRepo: stubUserRepo{user: existing}}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, &upsertInvitationRepoSpy{})

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "old-sub",
			Email:      "e@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("既存ユーザーは許可されるべき")
	}
	if users.ensureIdentityCalls != 1 {
		t.Fatalf("EnsureOidcIdentity calls = %d, want 1（セルフヒールされていない）", users.ensureIdentityCalls)
	}
	if users.ensuredUserID != 77 {
		t.Fatalf("ensured userID = %d, want 77", users.ensuredUserID)
	}
	if users.ensuredSubject != "old-sub" {
		t.Fatalf("subject = %q, want %q", users.ensuredSubject, "old-sub")
	}
}

// ブートストラップの冒頭ガードは「空文字どうしの一致」を止めている。
// env 未設定（免除アドレス無し）で、かつ id_token に email claim が無い（federated ID token では
// 実際に起こる）admin グループのユーザーが来ても、空文字が一致して免除が成立してはいけない。
// 既存 super_admin の照会まで進まないことも併せて固定する（判定は照会の前に閉じている）。
func Test_UpsertUserFromIDToken_ブートストラップはenv未設定かつメール空では成立しない(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, nil) // 免除アドレス未設定（本番の既定）

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:     "admin-sub",
			Email:          "",
			IsCognitoAdmin: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("self signup must be allowed even when bootstrap does not apply")
	}
	if users.created == nil {
		t.Fatal("user was not created")
	}
	if users.created.Role != domain.RoleTrainee {
		t.Fatalf("role = %q, want %q (must not be promoted)", users.created.Role, domain.RoleTrainee)
	}
	if users.listByRoleCalls != 0 {
		t.Fatalf("must not query existing super admins when no bootstrap address is set: %d", users.listByRoleCalls)
	}
}

// 免除の突き合わせは正規形どうしの一致で行う。
// 大小文字・前後空白の違いは同じアドレスとして通し、逆に「畳めば同じだがバイトが違う」
// 文字（U+017F など strings.EqualFold が畳む文字）は別アドレスとして扱う。
func Test_UpsertUserFromIDToken_ブートストラップの突き合わせは正規形で行う(t *testing.T) {
	tests := []struct {
		name           string
		bootstrapEmail string
		claimEmail     string
		wantSuperAdmin bool
		wantSavedEmail string
	}{
		{
			name:           "大小文字と前後空白の違いは同じアドレス",
			bootstrapEmail: "ops@example.com",
			claimEmail:     "  OPS@Example.com ",
			wantSuperAdmin: true,
			wantSavedEmail: "ops@example.com",
		},
		{
			// strings.EqualFold は U+017F(ſ) を 's' に畳むため、この 2 つを同一と見なしていた。
			// 正規形（小文字化）では畳まれないので別アドレスとして扱い、昇格させない。
			name:           "単純フォールドでのみ一致する別アドレスは昇格させない",
			bootstrapEmail: "sops@example.com",
			claimEmail:     "\u017Fops@example.com",
			wantSuperAdmin: false,
			wantSavedEmail: "\u017Fops@example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &upsertUserRepoSpy{}
			uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, tt.bootstrapEmail)

			user, err := uc.Execute(
				context.Background(),
				UpsertUserFromIDTokenInput{
					CognitoSub:     "admin-sub",
					Email:          tt.claimEmail,
					IsCognitoAdmin: true,
				},
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil {
				t.Fatal("招待ゲート撤去後は常にユーザーが作られるべき")
			}
			if users.created == nil {
				t.Fatal("ユーザーが作成されていない")
			}
			wantRole := domain.RoleTrainee
			if tt.wantSuperAdmin {
				wantRole = domain.RoleSuperAdmin
			}
			if users.created.Role != wantRole {
				t.Fatalf("role = %q, want %q", users.created.Role, wantRole)
			}
			// 保存されるのは生の claim 値ではなく正規形。生値のまま保存すると、以後の
			// byte 一致検索・一意索引と食い違う。
			if users.created.Email != tt.wantSavedEmail {
				t.Fatalf("保存された email = %q, want %q", users.created.Email, tt.wantSavedEmail)
			}
		})
	}
}

// 判定と作成のあいだに別の運営管理者ができたら、作成側で弾いて拒否する
// （「最初の 1 人ができた瞬間に閉じる」を、判定した事実ではなく作成の可否で決める）。
func Test_UpsertUserFromIDToken_ブートストラップは作成側で閉じられたら拒否する(t *testing.T) {
	users := &upsertUserRepoSpyRaceLoser{}
	uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, "ops@example.com")

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:     "admin-sub",
			Email:          "ops@example.com",
			IsCognitoAdmin: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Fatal("作成側が「既に super_admin が居る」と答えたら許可してはいけない")
	}
	if users.created != nil {
		t.Fatalf("ユーザーを作成してはいけない: %+v", users.created)
	}
}

// upsertUserRepoSpyRaceLoser は「判定のあとで別の super_admin が確定した」状態の repository。
// 事前の照会（ListByRole）では 0 人に見えるが、作成側の再判定では既に居る。
type upsertUserRepoSpyRaceLoser struct{ upsertUserRepoSpy }

func (s *upsertUserRepoSpyRaceLoser) CreateFirstSuperAdminWithOidcIdentity(
	_ context.Context,
	_ *domain.User,
	_, _ string,
) (bool, error) {
	s.createFirstSuperAdminCalls++
	return false, nil
}

// 招待の照会にも正規形を渡す。生の claim 値で引くと、同じアドレスなのに招待が見つからない
// （= 招待を通したはずの人が拒否される）ずれが生まれる。
func Test_UpsertUserFromIDToken_招待の照会と保存に正規形のメールを使う(t *testing.T) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:    10,
			Email: "Member@Example.com",
			Role:  domain.RoleCompanyAdmin,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	user, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "member-sub",
			Email:      " Member@Example.com ",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("招待があるなら許可されるべき")
	}
	if invitations.emailFindArg != "member@example.com" {
		t.Fatalf("招待の照会に渡した email = %q, want %q", invitations.emailFindArg, "member@example.com")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Email != "member@example.com" {
		t.Fatalf("保存された email = %q, want %q", users.created.Email, "member@example.com")
	}
}
