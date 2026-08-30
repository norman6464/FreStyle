package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// newPublicInvitationRouter は本番と同じ経路（PublicInvitationHandler.Validate）でルータを組む。
func newPublicInvitationRouter(inv *fakeInvitationRepo, companies *fakeCompanyRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPublicInvitationHandler(usecase.NewValidateInvitationTokenUseCase(inv, companies))
	r.GET("/invitations/accept/:token", h.Validate)
	return r
}

// Test_招待検証API_workspaceIdの有無 は HTTP レスポンスの JSON シリアライズを固定する。
// usecase 単体テストは ValidatedInvitation.WorkspaceID の値までしか見ないため、
// companyId と同じ omitempty 契約が実際の JSON でも守られることをここで検証する。
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
				ID: 1, CompanyID: 1, Role: domain.RoleTrainee, Name: "山田",
				Status: domain.InvitationStatusPending, WorkspaceID: &wsID,
			},
			wantHasWSID: true,
		},
		{
			name: "workspace_id 未設定時は省略する",
			inv: &domain.AdminInvitation{
				ID: 2, CompanyID: 1, Role: domain.RoleTrainee, Name: "鈴木",
				Status: domain.InvitationStatusPending,
			},
			wantHasWSID: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			invRepo := &fakeInvitationRepo{pendingByToken: map[string]*domain.AdminInvitation{"tok": tc.inv}}
			r := newPublicInvitationRouter(invRepo, &fakeCompanyRepo{})

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

// Test_招待検証API_未知tokenは404 は既存の usecase 側テストと重複しない範囲で、
// HTTP レベルの契約（無効 token はメタ情報を漏らさず 404）を固定する。
func Test_招待検証API_未知tokenは404(t *testing.T) {
	r := newPublicInvitationRouter(&fakeInvitationRepo{}, &fakeCompanyRepo{})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/invitations/accept/no-such-token", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d (body=%s)", w.Code, w.Body.String())
	}
}
