package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeInvWorkspaceRepo は WorkspaceActivationReader の最小 fake。
// 受諾画面に出す招待元の表示名だけを答える。
type fakeInvWorkspaceRepo struct {
	byID map[string]*domain.Workspace
}

func (r *fakeInvWorkspaceRepo) FindWorkspaceByID(_ context.Context, workspaceID string) (*domain.Workspace, error) {
	if w, ok := r.byID[workspaceID]; ok {
		return w, nil
	}
	return nil, repository.ErrWorkspaceNotFound
}

// newPublicInvitationRouter は本番と同じ経路（PublicInvitationHandler.Validate）でルータを組む。
func newPublicInvitationRouter(inv *fakeInvitationRepo, workspaces *fakeInvWorkspaceRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPublicInvitationHandler(usecase.NewValidateInvitationTokenUseCase(inv, workspaces))
	r.GET("/invitations/accept/:token", h.Validate)
	return r
}

// Test_招待検証API_workspaceIdの有無 は HTTP レスポンスの JSON シリアライズを固定する。
// usecase 単体テストは ValidatedInvitation.WorkspaceID の値までしか見ないため、
// omitempty 契約が実際の JSON でも守られることをここで検証する。
func Test_招待検証API_workspaceIdの有無(t *testing.T) {
	wsID := "0198a000-0000-7000-8000-000000000001"
	cases := []struct {
		name        string
		inv         *domain.AdminInvitation
		wantHasWSID bool
	}{
		{
			name: "workspace_id 設定時はレスポンスに含む",
			inv: &domain.AdminInvitation{
				ID: 1, Role: domain.RoleTrainee, Name: "山田",
				Status: domain.InvitationStatusPending, WorkspaceID: &wsID,
			},
			wantHasWSID: true,
		},
		{
			name: "workspace_id 未設定時は省略する",
			inv: &domain.AdminInvitation{
				ID: 2, Role: domain.RoleTrainee, Name: "鈴木",
				Status: domain.InvitationStatusPending,
			},
			wantHasWSID: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			invRepo := &fakeInvitationRepo{pendingByToken: map[string]*domain.AdminInvitation{"tok": tc.inv}}
			// 招待先が決まっている場合、その行は FK があるので必ず引ける。引けないと
			// 不整合として 500 になるため、ここでは実在する状態にしておく。
			r := newPublicInvitationRouter(invRepo, &fakeInvWorkspaceRepo{byID: map[string]*domain.Workspace{
				wsID: {ID: wsID, Slug: "acme", Name: "アクメ ワークスペース", IsActive: true},
			}})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/invitations/accept/tok", nil))

			if w.Code != http.StatusOK {
				t.Fatalf("want 200, got %d (body=%s)", w.Code, w.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			_, has := body["workspaceId"]
			if has != tc.wantHasWSID {
				t.Fatalf("workspaceId present = %v, want %v (body=%v)", has, tc.wantHasWSID, body)
			}
		})
	}
}

// 招待が指すワークスペースが引けないのは不整合。名前を空にして 200 で通すと、
// どこに招かれたのか分からないまま受諾させることになるので、500 で止める。
func Test_招待検証API_招待先のワークスペースが引けなければ500(t *testing.T) {
	wsID := "0198a000-0000-7000-8000-000000000001"
	invRepo := &fakeInvitationRepo{pendingByToken: map[string]*domain.AdminInvitation{
		"tok": {
			ID: 1, Role: domain.RoleTrainee, Name: "山田",
			Status: domain.InvitationStatusPending, WorkspaceID: &wsID,
		},
	}}
	r := newPublicInvitationRouter(invRepo, &fakeInvWorkspaceRepo{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/invitations/accept/tok", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d (body=%s)", w.Code, w.Body.String())
	}
}

// 招待された人が「どこに招かれたのか」を判断する唯一の手掛かりなので、
// ワークスペースの表示名がそのまま workspaceName で返ることを固定する。
func Test_招待検証API_招待元のワークスペース名を返す(t *testing.T) {
	wsID := "0198a000-0000-7000-8000-000000000001"
	invRepo := &fakeInvitationRepo{pendingByToken: map[string]*domain.AdminInvitation{
		"tok": {
			ID: 1, Role: domain.RoleTrainee, Name: "山田",
			Status: domain.InvitationStatusPending, WorkspaceID: &wsID,
		},
	}}
	workspaces := &fakeInvWorkspaceRepo{byID: map[string]*domain.Workspace{
		wsID: {ID: wsID, Slug: "acme", Name: "アクメ ワークスペース", IsActive: true},
	}}
	r := newPublicInvitationRouter(invRepo, workspaces)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/invitations/accept/tok", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (body=%s)", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["workspaceName"] != "アクメ ワークスペース" {
		t.Fatalf("workspaceName = %v (body=%v)", body["workspaceName"], body)
	}
}

// Test_招待検証API_未知tokenは404 は既存の usecase 側テストと重複しない範囲で、
// HTTP レベルの契約（無効 token はメタ情報を漏らさず 404）を固定する。
func Test_招待検証API_未知tokenは404(t *testing.T) {
	r := newPublicInvitationRouter(&fakeInvitationRepo{}, &fakeInvWorkspaceRepo{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/invitations/accept/no-such-token", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d (body=%s)", w.Code, w.Body.String())
	}
}
