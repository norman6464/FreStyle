package usecase

import (
	"context"
	"errors"
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
	companyUpdateCalls    int
	companyUpdateErr      error

	roleUpdateUserID    uint64
	roleUpdateValue     domain.RoleName
	companyUpdateUserID uint64
	companyUpdateValue  uint64

	// ブートストラップ判定（既存 super_admin の有無）の制御。
	superAdmins     []domain.User
	listByRoleErr   error
	listByRoleCalls int
	listByRoleValue domain.RoleName
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
	updateCalls     int
	updateErr       error
	updatedID       uint64
	updatedStatus   string
}

func (s *upsertInvitationRepoSpy) ListAll(
	_ context.Context,
) ([]domain.AdminInvitation, error) {
	return nil, nil
}

func (s *upsertInvitationRepoSpy) ListByCompanyID(
	_ context.Context,
	_ uint64,
) ([]domain.AdminInvitation, error) {
	return nil, nil
}

func (s *upsertInvitationRepoSpy) FindPendingByEmail(
	_ context.Context,
	_ string,
) (*domain.AdminInvitation, error) {
	s.emailFindCalled = true
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
	return NewUpsertUserFromIDTokenUseCase(users, invitations, "")
}

// newUpsertUserFromIDTokenUseCaseWithBootstrap はブートストラップ用アドレスを設定した usecase を返す。
func newUpsertUserFromIDTokenUseCaseWithBootstrap(
	users repository.UserRepository,
	invitations repository.AdminInvitationRepository,
	bootstrapEmail string,
) *UpsertUserFromIDTokenUseCase {
	return NewUpsertUserFromIDTokenUseCase(users, invitations, bootstrapEmail)
}

func Test_UpsertUserFromIDToken_招待のRoleとCompanyを適用してAcceptedにする(t *testing.T) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:        10,
			Role:      domain.RoleCompanyAdmin,
			CompanyID: 42,
			Name:      "Invited User",
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
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
	if !allowed {
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
	if users.created.CompanyID == nil || *users.created.CompanyID != 42 {
		t.Fatalf("companyID = %v, want 42", users.created.CompanyID)
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

func (s *upsertUserRepoSpy) UpdateCompanyID(
	_ context.Context,
	userID uint64,
	companyID uint64,
) error {
	s.companyUpdateCalls++
	if s.companyUpdateErr != nil {
		return s.companyUpdateErr
	}

	s.companyUpdateUserID = userID
	s.companyUpdateValue = companyID
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
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:        20,
			Role:      domain.RoleCompanyAdmin,
			CompanyID: 42,
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-sub",
			Email:      "existing@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
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
	if users.companyUpdateUserID != 7 {
		t.Fatalf("company更新対象ID = %d, want 7", users.companyUpdateUserID)
	}
	if users.companyUpdateValue != 42 {
		t.Fatalf("更新companyID = %d, want 42", users.companyUpdateValue)
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

func Test_UpsertUserFromIDToken_招待も管理者権限もない新規ユーザーを拒否(t *testing.T) {
	users := &stubUserRepo{}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, nil)

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "new-sub",
			Email:      "new@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("招待も管理者権限もない新規ユーザーは拒否されるべき")
	}
}

// Cognito の admin グループに属しているだけでは招待統制を迂回できない（グループ名 1 つで
// 会社をまたぐ super_admin が作れてしまう穴を塞ぐ）。
func Test_UpsertUserFromIDToken_CognitoAdminでも招待が無ければ新規作成を拒否(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, nil)

	allowed, err := uc.Execute(
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
	if allowed {
		t.Fatal("招待の無い新規ユーザーは Cognito admin グループでも拒否されるべき")
	}
	if users.created != nil {
		t.Fatalf("ユーザーを作成してはいけない: %+v", users.created)
	}
}

// ブートストラップ: 明示したアドレス + Cognito admin グループ + super_admin が 0 人のときだけ、
// 招待なしで最初の運営管理者を作れる。
func Test_UpsertUserFromIDToken_ブートストラップ指定アドレスは招待なしでSuperAdminを作る(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, "  Ops@Example.com ")

	allowed, err := uc.Execute(
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
	if !allowed {
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

// ブートストラップは「最初の 1 人」限定。super_admin が既に居れば経路は閉じる。
func Test_UpsertUserFromIDToken_ブートストラップはSuperAdmin在籍時に閉じる(t *testing.T) {
	users := &upsertUserRepoSpy{
		superAdmins: []domain.User{{ID: 1, Role: domain.RoleSuperAdmin}},
	}
	uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, "ops@example.com")

	allowed, err := uc.Execute(
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
	if allowed {
		t.Fatal("既に super_admin が居るならブートストラップは効いてはいけない")
	}
	if users.created != nil {
		t.Fatalf("ユーザーを作成してはいけない: %+v", users.created)
	}
}

// ブートストラップは指定した 1 アドレスだけ。別アドレスや admin グループ非所属には効かない。
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

			allowed, err := uc.Execute(
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
			if allowed {
				t.Fatal("ブートストラップ対象外は拒否されるべき")
			}
			if users.created != nil {
				t.Fatalf("ユーザーを作成してはいけない: %+v", users.created)
			}
		})
	}
}

// 既存 super_admin を数えられないときは「居ないこと」を確認できていないので免除しない。
func Test_UpsertUserFromIDToken_ブートストラップ判定の照会失敗はエラーで止める(t *testing.T) {
	listErr := errors.New("db down")
	users := &upsertUserRepoSpy{listByRoleErr: listErr}
	uc := newUpsertUserFromIDTokenUseCaseWithBootstrap(users, nil, "ops@example.com")

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub:     "admin-sub",
			Email:          "ops@example.com",
			IsCognitoAdmin: true,
		},
	)
	if allowed {
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

			allowed, err := uc.Execute(
				context.Background(),
				tc.input,
			)

			if allowed {
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
			ID:        10,
			Role:      domain.RoleTrainee,
			CompanyID: 100,
			Name:      "Token User",
			Status:    domain.InvitationStatusPending,
		},
		pending: &domain.AdminInvitation{
			ID:        20,
			Role:      domain.RoleCompanyAdmin,
			CompanyID: 200,
			Name:      "Email User",
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
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
	if !allowed {
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
	if users.created.CompanyID == nil || *users.created.CompanyID != 100 {
		t.Fatalf("companyID = %v, want 100", users.created.CompanyID)
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

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "",
			Email:      "user@example.com",
		},
	)

	if allowed {
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
					ID:        30,
					Role:      invitationRole,
					CompanyID: 42,
					Status:    domain.InvitationStatusPending,
				},
			}
			uc := newUpsertUserFromIDTokenUseCaseForTest(
				users,
				invitations,
			)

			allowed, err := uc.Execute(
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
			if !allowed {
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
			ID:        40,
			Role:      domain.RoleSuperAdmin,
			CompanyID: 42,
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "unsupported-role",
			Email:      "user@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
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

func Test_UpsertUserFromIDToken_招待のCompanyIDが0なら未所属にする(
	t *testing.T,
) {
	users := &upsertUserRepoSpy{}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:        50,
			Role:      domain.RoleTrainee,
			CompanyID: 0,
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "no-company",
			Email:      "no-company@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("有効な招待があるユーザーは許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.CompanyID != nil {
		t.Fatalf(
			"companyID = %v, want nil",
			users.created.CompanyID,
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
			ID:        60,
			Role:      domain.RoleTrainee,
			CompanyID: 42,
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-user",
			Email:      "existing@example.com",
			Name:       "OIDC User",
		},
	)

	if allowed {
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
			ID:        60,
			Role:      domain.RoleCompanyAdmin,
			CompanyID: 42,
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-user",
			Email:      "existing@example.com",
		},
	)

	if allowed {
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

func Test_UpsertUserFromIDToken_会社更新に失敗する(t *testing.T) {
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
		companyUpdateErr: mutationErr,
	}
	invitations := &upsertInvitationRepoSpy{
		pending: &domain.AdminInvitation{
			ID:        60,
			Role:      domain.RoleTrainee,
			CompanyID: 42,
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-user",
			Email:      "existing@example.com",
		},
	)

	if allowed {
		t.Fatal("会社更新失敗時にユーザーを許可してはいけない")
	}
	if !errors.Is(err, mutationErr) {
		t.Fatalf("error = %v, want wrapped %v", err, mutationErr)
	}
	if users.companyUpdateCalls != 1 {
		t.Fatalf(
			"UpdateCompanyID calls = %d, want 1",
			users.companyUpdateCalls,
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
			ID:        60,
			Role:      domain.RoleTrainee,
			CompanyID: 42,
			Status:    domain.InvitationStatusPending,
		},
		updateErr: mutationErr,
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "existing-user",
			Email:      "existing@example.com",
		},
	)

	if allowed {
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
			ID:        70,
			Role:      domain.RoleCompanyAdmin,
			CompanyID: 42,
			Status:    domain.InvitationStatusPending,
		},
	}
	uc := newUpsertUserFromIDTokenUseCaseForTest(users, invitations)

	allowed, err := uc.Execute(
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
	if !allowed {
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
	if users.companyUpdateCalls != 0 {
		t.Fatalf(
			"UpdateCompanyID calls = %d, want 0",
			users.companyUpdateCalls,
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
					ID:        70,
					Role:      domain.RoleTrainee,
					CompanyID: 42,
					Status:    domain.InvitationStatusPending,
				},
			}
			tc.configure(users, invitations)

			uc := newUpsertUserFromIDTokenUseCaseForTest(
				users,
				invitations,
			)

			allowed, err := uc.Execute(
				context.Background(),
				UpsertUserFromIDTokenInput{
					CognitoSub: "new-user-error",
					Email:      "new-user@example.com",
				},
			)

			if allowed {
				t.Fatal("トランザクション失敗時に許可してはいけない")
			}
			if !errors.Is(err, mutationErr) {
				t.Fatalf(
					"error = %v, want wrapped %v",
					err,
					mutationErr,
				)
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

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "new-sub-1",
			Email:      "new@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
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

	allowed, err := uc.Execute(
		context.Background(),
		UpsertUserFromIDTokenInput{
			CognitoSub: "old-sub",
			Email:      "e@example.com",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
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
