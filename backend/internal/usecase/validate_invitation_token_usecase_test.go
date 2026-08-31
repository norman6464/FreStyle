package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// validateWsA は受諾画面の検証で使うワークスペース ID。
const validateWsA = "0198a000-0000-7000-8000-0000000000d1"

// stubWorkspaces は招待の workspace_id からワークスペースを引くだけの
// WorkspaceActivationReader スタブ。受諾画面のワークスペース名はこの引き直しで付く。
type stubWorkspaces struct {
	workspaces []domain.Workspace
	err        error
}

func (s *stubWorkspaces) FindWorkspaceByID(_ context.Context, workspaceID string) (*domain.Workspace, error) {
	if s.err != nil {
		return nil, s.err
	}
	for i := range s.workspaces {
		if s.workspaces[i].ID == workspaceID {
			return &s.workspaces[i], nil
		}
	}
	return nil, repository.ErrWorkspaceNotFound
}

// stubAdminInvRepoWithToken は ValidateInvitationTokenUseCase 専用の stub。
// stubAdminInvRepo (admin_invitation_usecase_test.go) は他テストで FindPendingByToken を
// nil 固定で返してしまうため、このテストでは別の stub を持つ。
type stubAdminInvRepoWithToken struct {
	stubAdminInvRepo
	pendingByToken map[string]*domain.AdminInvitation
}

func (s *stubAdminInvRepoWithToken) FindPendingByToken(_ context.Context, token string) (*domain.AdminInvitation, error) {
	if s.err != nil {
		return nil, s.err
	}
	if v, ok := s.pendingByToken[token]; ok {
		return v, nil
	}
	return nil, nil
}

func Test_招待token検証_空tokenはnil(t *testing.T) {
	uc := NewValidateInvitationTokenUseCase(&stubAdminInvRepoWithToken{}, &stubWorkspaces{})
	got, err := uc.Execute(context.Background(), "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != nil {
		t.Fatalf("empty token must return nil, got %+v", got)
	}
}

func Test_招待token検証_見つからなければnil(t *testing.T) {
	uc := NewValidateInvitationTokenUseCase(&stubAdminInvRepoWithToken{}, &stubWorkspaces{})
	got, err := uc.Execute(context.Background(), "missing-token")
	if err != nil || got != nil {
		t.Fatalf("missing token must return (nil, nil), got=%+v err=%v", got, err)
	}
}

func Test_招待token検証_正常系_ワークスペース名を付与(t *testing.T) {
	wsID := "0198a000-0000-7000-8000-000000000001"
	repo := &stubAdminInvRepoWithToken{
		pendingByToken: map[string]*domain.AdminInvitation{
			"abc-123": {
				ID: 9, Email: "u@example.com",
				Role: domain.RoleCompanyAdmin, Name: "山田", WorkspaceID: &wsID,
			},
		},
	}
	workspaces := &stubWorkspaces{
		workspaces: []domain.Workspace{
			{ID: wsID, Slug: "frestyle", Name: "株式会社FreStyle"},
		},
	}
	uc := NewValidateInvitationTokenUseCase(repo, workspaces)

	got, err := uc.Execute(context.Background(), "abc-123")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got == nil {
		// return を明示するのは、静的解析が t.Fatalf の非復帰を追えない環境でも
		// 以降の参照が到達不能だと分かるようにするため。
		t.Fatalf("expected non-nil result")
		return
	}
	if got.Role != domain.RoleCompanyAdmin {
		t.Errorf("Role = %q, want company_admin", got.Role)
	}
	if got.Name != "山田" {
		t.Errorf("Name = %q, want 山田", got.Name)
	}
	if got.WorkspaceName != "株式会社FreStyle" {
		t.Errorf("WorkspaceName = %q, want 株式会社FreStyle", got.WorkspaceName)
	}
	// 招待行の workspace_id をそのまま返す。サブクエリで引き直さない。
	if got.WorkspaceID == nil || *got.WorkspaceID != wsID {
		t.Errorf("WorkspaceID = %v, want %q", got.WorkspaceID, wsID)
	}
}

// 招待の workspace_id が未設定（バックフィル未到達等）でもエラーにはせず nil のまま返す。
func Test_招待token検証_workspace未設定はnilのまま(t *testing.T) {
	repo := &stubAdminInvRepoWithToken{
		pendingByToken: map[string]*domain.AdminInvitation{
			"t": {ID: 9, Role: domain.RoleTrainee},
		},
	}
	uc := NewValidateInvitationTokenUseCase(repo, &stubWorkspaces{
		workspaces: []domain.Workspace{{ID: validateWsA, Name: "x"}},
	})

	got, err := uc.Execute(context.Background(), "t")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got == nil || got.WorkspaceID != nil {
		t.Errorf("WorkspaceID = %v, want nil", got)
	}
}

func Test_招待token検証_ワークスペースが引けなければエラー(t *testing.T) {
	// invitations.workspace_id には FK があるので、招待先の行は必ず在るはず。
	// 引けないのは不整合であって「名前が無い招待」ではない。名前を空にして 200 で
	// 通すと、どこに招かれたのか分からないまま受諾させることになる。
	cases := []struct {
		name       string
		workspaces *stubWorkspaces
	}{
		{"参照が失敗する", &stubWorkspaces{err: errors.New("db down")}},
		{"招待先の行が無い", &stubWorkspaces{err: repository.ErrWorkspaceNotFound}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := &stubAdminInvRepoWithToken{
				pendingByToken: map[string]*domain.AdminInvitation{
					"t": {ID: 9, Email: "u@example.com", Role: domain.RoleTrainee, WorkspaceID: strPtr(validateWsA)},
				},
			}
			uc := NewValidateInvitationTokenUseCase(repo, c.workspaces)

			got, err := uc.Execute(context.Background(), "t")
			if err == nil {
				t.Fatalf("エラーを返すべき: got %+v", got)
			}
			if got != nil {
				t.Errorf("エラー時は結果を返さない: got %+v", got)
			}
		})
	}
}

func Test_招待token検証_未知のroleを正規化(t *testing.T) {
	repo := &stubAdminInvRepoWithToken{
		pendingByToken: map[string]*domain.AdminInvitation{
			"t": {ID: 9, Role: "garbage_role", WorkspaceID: strPtr(validateWsA)},
		},
	}
	uc := NewValidateInvitationTokenUseCase(repo, &stubWorkspaces{
		workspaces: []domain.Workspace{{ID: validateWsA, Name: "x"}},
	})
	got, _ := uc.Execute(context.Background(), "t")
	if got == nil || got.Role != domain.RoleTrainee {
		t.Errorf("unknown role must fallback to trainee, got %+v", got)
	}
}
