package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
)

type stubTempCreator struct {
	pw       string
	err      error
	gotEmail string
	gotName  string
	calls    int
}

func (s *stubTempCreator) CreateWithTemporaryPassword(_ context.Context, email, name string) (string, error) {
	s.calls++
	s.gotEmail, s.gotName = email, name
	return s.pw, s.err
}

func Test_初期パスワード招待_成功で行と一時パスワードを返す(t *testing.T) {
	repo := &stubAdminInvRepo{}
	creator := &stubTempCreator{pw: "Temp-1!"}
	uc := NewCreateTemporaryPasswordInvitationUseCase(repo, creator)

	out, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		CompanyID: 42, Email: "np@example.com", Role: domain.RoleTrainee, Name: "山田",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.TemporaryPassword != "Temp-1!" {
		t.Errorf("temp password = %q", out.TemporaryPassword)
	}
	if out.Invitation == nil || out.Invitation.Status != domain.InvitationStatusPending {
		t.Errorf("invitation not pending: %+v", out.Invitation)
	}
	if repo.created == nil || repo.created.Role != domain.RoleTrainee || repo.created.CompanyID != 42 {
		t.Errorf("invitation row wrong: %+v", repo.created)
	}
	if creator.gotEmail != "np@example.com" || creator.gotName != "山田" {
		t.Errorf("creator args: %q / %q", creator.gotEmail, creator.gotName)
	}
}

func Test_初期パスワード招待_cognito未構成はErrUnavailable(t *testing.T) {
	uc := NewCreateTemporaryPasswordInvitationUseCase(&stubAdminInvRepo{}, nil)
	_, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		CompanyID: 1, Email: "a@b", Role: domain.RoleTrainee,
	})
	if !errors.Is(err, ErrTemporaryPasswordUnavailable) {
		t.Fatalf("ErrTemporaryPasswordUnavailable を期待したが: %v", err)
	}
}

func Test_初期パスワード招待_既存ユーザーエラーは伝播(t *testing.T) {
	repo := &stubAdminInvRepo{}
	creator := &stubTempCreator{err: cognito.ErrUserAlreadyExists}
	uc := NewCreateTemporaryPasswordInvitationUseCase(repo, creator)

	_, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		CompanyID: 1, Email: "dup@b", Role: domain.RoleTrainee,
	})
	if !errors.Is(err, cognito.ErrUserAlreadyExists) {
		t.Fatalf("ErrUserAlreadyExists の伝播を期待したが: %v", err)
	}
}

func Test_初期パスワード招待_必須項目チェック(t *testing.T) {
	uc := NewCreateTemporaryPasswordInvitationUseCase(&stubAdminInvRepo{}, &stubTempCreator{})
	_, err := uc.Execute(context.Background(), CreateAdminInvitationInput{Email: "a@b", Role: domain.RoleTrainee})
	if err == nil {
		t.Fatal("companyID=0 はエラーであるべき")
	}
}
