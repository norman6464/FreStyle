package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type upsertUserRepoSpy struct {
	stubUserRepo
	created *domain.User

	roleUpdateUserID    uint64
	roleUpdateValue     string
	companyUpdateUserID uint64
	companyUpdateValue  uint64
}

type upsertInvitationRepoSpy struct {
	pending         *domain.AdminInvitation
	pendingByToken  *domain.AdminInvitation
	tokenFindErr    error
	emailFindErr    error
	tokenFindCalled bool
	emailFindCalled bool
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
	s.updatedID = id
	s.updatedStatus = status
	return nil
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
	uc := NewUpsertUserFromIDTokenUseCase(users, invitations)

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

func (s *upsertUserRepoSpy) Create(_ context.Context, user *domain.User) error {
	copied := *user
	s.created = &copied
	return nil
}

func (s *upsertUserRepoSpy) UpdateRole(
	_ context.Context,
	userID uint64,
	role string,
) error {
	s.roleUpdateUserID = userID
	s.roleUpdateValue = role
	return nil
}

func (s *upsertUserRepoSpy) UpdateCompanyID(
	_ context.Context,
	userID uint64,
	companyID uint64,
) error {
	s.companyUpdateUserID = userID
	s.companyUpdateValue = companyID
	return nil
}

func Test_UpsertUserFromIDToken_既存TraineeをCompanyAdminへ昇格して会社に紐付ける(t *testing.T) {
	users := &upsertUserRepoSpy{
		stubUserRepo: stubUserRepo{
			user: &domain.User{
				ID:         7,
				CognitoSub: "existing-sub",
				Email:      "existing@example.com",
				Name:       "Existing User",
				Role:       domain.RoleTrainee,
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
	uc := NewUpsertUserFromIDTokenUseCase(users, invitations)

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
	uc := NewUpsertUserFromIDTokenUseCase(users, nil)

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

func Test_UpsertUserFromIDToken_CognitoAdminは招待なしでもSuperAdminとして作成する(t *testing.T) {
	users := &upsertUserRepoSpy{}
	uc := NewUpsertUserFromIDTokenUseCase(users, nil)

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
	if !allowed {
		t.Fatal("Cognito adminは招待なしでも許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleSuperAdmin {
		t.Fatalf("role = %q, want %q", users.created.Role, domain.RoleSuperAdmin)
	}
}

func Test_UpsertUserFromIDToken_検索エラーを返す(t *testing.T) {
	tokenFindErr := errors.New("token lookup failed")
	emailFindErr := errors.New("email lookup failed")
	userFindErr := errors.New("user lookup failed")

	tests := []struct {
		name        string
		users       *stubUserRepo
		invitations *upsertInvitationRepoSpy
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
			uc := NewUpsertUserFromIDTokenUseCase(
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
	uc := NewUpsertUserFromIDTokenUseCase(users, invitations)

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
