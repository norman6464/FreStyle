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
		TargetWorkspace: domain.WorkspaceRefOf(invWsB),
		Email:           "np@example.com", Role: domain.RoleTrainee, Name: "山田",
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
	if repo.created == nil || repo.created.Role != domain.RoleTrainee ||
		repo.created.WorkspaceID == nil || *repo.created.WorkspaceID != invWsB {
		t.Errorf("invitation row wrong: %+v", repo.created)
	}
	if creator.gotEmail != "np@example.com" || creator.gotName != "山田" {
		t.Errorf("creator args: %q / %q", creator.gotEmail, creator.gotName)
	}
}

// 招待の email は正規形（domain.NormalizeEmail）で保存し、Cognito ユーザーも同じ値で作る。
// 生のまま残すと、ログイン時の招待ゲートは正規形の OIDC メールで引くため突き合わせられず、
// 招待したはずの相手が「招待なし」として拒否される。
func Test_初期パスワード招待_emailを正規形に畳んで保存する(t *testing.T) {
	repo := &stubAdminInvRepo{}
	creator := &stubTempCreator{pw: "Temp-1!"}
	uc := NewCreateTemporaryPasswordInvitationUseCase(repo, creator)

	out, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		TargetWorkspace: domain.WorkspaceRefOf(invWsB),
		Email:           "  NP@Example.com\t", Role: domain.RoleTrainee, Name: "山田",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Invitation.Email != "np@example.com" {
		t.Errorf("invitation email = %q, want %q", out.Invitation.Email, "np@example.com")
	}
	if creator.gotEmail != "np@example.com" {
		t.Errorf("cognito email = %q, want %q", creator.gotEmail, "np@example.com")
	}
}

func Test_初期パスワード招待_cognito未構成はErrUnavailable(t *testing.T) {
	uc := NewCreateTemporaryPasswordInvitationUseCase(&stubAdminInvRepo{}, nil)
	_, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		TargetWorkspace: domain.WorkspaceRefOf(invWsA),
		Email:           "a@b", Role: domain.RoleTrainee,
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
		TargetWorkspace: domain.WorkspaceRefOf(invWsA),
		Email:           "dup@b", Role: domain.RoleTrainee,
	})
	if !errors.Is(err, ErrInvitationUserAlreadyExists) {
		t.Fatalf("ErrInvitationUserAlreadyExists（usecase 語彙）を期待したが: %v", err)
	}
	// 重要: Cognito 失敗時に招待行を作ってはいけない（孤児行→テナント横断のワークスペース付け替えを防ぐ）。
	if repo.created != nil {
		t.Fatalf("Cognito 失敗時に招待行が作られている（孤児行の脆弱性）: %+v", repo.created)
	}
	if creator.calls != 1 {
		t.Errorf("Cognito 呼び出し回数 = %d, want 1（招待行より先に呼ぶ）", creator.calls)
	}
}

// TargetWorkspace 未設定はエラーで、Cognito にも触れない。逆順（Cognito 先）にすると、
// 招待行を持たない Cognito ユーザーだけが残り、誰も辿れない孤立ユーザーになる。
func Test_初期パスワード招待_TargetWorkspace未設定はCognitoを呼ばない(t *testing.T) {
	repo := &stubAdminInvRepo{}
	creator := &stubTempCreator{pw: "Temp-1!"}
	uc := NewCreateTemporaryPasswordInvitationUseCase(repo, creator)

	_, err := uc.Execute(context.Background(), CreateAdminInvitationInput{Email: "a@b", Role: domain.RoleTrainee})
	if err == nil {
		t.Fatal("TargetWorkspace 未設定はエラーであるべき")
	}
	if creator.calls != 0 {
		t.Errorf("Cognito 呼び出し回数 = %d, want 0", creator.calls)
	}
	if repo.created != nil {
		t.Errorf("招待行を作ってはいけない: %+v", repo.created)
	}
}

func Test_初期パスワード招待_Cognito成功後のDB失敗はエラーで招待行を返さない(t *testing.T) {
	repo := &stubAdminInvRepo{createErr: errors.New("db down")}
	creator := &stubTempCreator{pw: "Temp-1!"}
	uc := NewCreateTemporaryPasswordInvitationUseCase(repo, creator)

	out, err := uc.Execute(context.Background(), CreateAdminInvitationInput{
		TargetWorkspace: domain.WorkspaceRefOf(invWsA),
		Email:           "np@example.com", Role: domain.RoleTrainee,
	})
	if err == nil {
		t.Fatal("DB 失敗はエラーであるべき")
	}
	if out != nil {
		t.Errorf("失敗時に部分成功を返してはいけない: %+v", out)
	}
	// Cognito は呼ばれている（先に実行）が、招待行は作られない（fail closed）。
	if creator.calls != 1 {
		t.Errorf("Cognito calls = %d, want 1", creator.calls)
	}
}
