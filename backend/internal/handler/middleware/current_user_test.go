package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// stubUsers は UserRepository の最小 stub。FindByCognitoSub だけ返す。
type stubUsers struct{ user *domain.User }

func (s *stubUsers) FindByCognitoSub(context.Context, string) (*domain.User, error) {
	return s.user, nil
}

func (s *stubUsers) FindByID(context.Context, uint64) (*domain.User, error) { return s.user, nil }

func (s *stubUsers) ListByRole(context.Context, domain.RoleName) ([]domain.User, error) {
	return nil, nil
}

func (s *stubUsers) ListByWorkspaceID(context.Context, string) ([]domain.User, error) {
	return nil, nil
}

func (s *stubUsers) CreateWithOidcIdentity(context.Context, *domain.User, string, string) error {
	return nil
}

func (s *stubUsers) CreateFirstSuperAdminWithOidcIdentity(
	context.Context, *domain.User, string, string,
) (bool, error) {
	return true, nil
}
func (s *stubUsers) EnsureOidcIdentity(context.Context, uint64, string, string) error { return nil }
func (s *stubUsers) FindActiveByEmail(context.Context, string) (*domain.User, error)  { return nil, nil }

func (s *stubUsers) CognitoSubjectByUserID(context.Context, uint64) (string, error) { return "", nil }
func (s *stubUsers) UpdateName(context.Context, uint64, string) error               { return nil }
func (s *stubUsers) UpdateRole(context.Context, uint64, domain.RoleName) error      { return nil }
func (s *stubUsers) UpdateWorkspaceID(context.Context, uint64, *string) error       { return nil }
func (s *stubUsers) UpdateActive(context.Context, uint64, bool) error               { return nil }
func (s *stubUsers) SoftDelete(context.Context, uint64) error                       { return nil }

// stubWorkspaces は WorkspaceActivationReader の最小 stub。workspace / err を返し、
// 問い合わせに使われた workspace_id を記録する。
type stubWorkspaces struct {
	workspace *domain.Workspace
	err       error

	gotWorkspaceID string
}

func (s *stubWorkspaces) FindWorkspaceByID(_ context.Context, workspaceID string) (*domain.Workspace, error) {
	s.gotWorkspaceID = workspaceID
	return s.workspace, s.err
}

// currentUserResult は CurrentUser を通したリクエストの結果。
type currentUserResult struct {
	rec *httptest.ResponseRecorder
	// reached は CurrentUser の後ろの handler まで届いたか。遮断が「abort フラグを
	// 立てただけ」で後続を止め損ねていないかは、これを見ないと分からない。
	reached bool
	// user は後続の handler から見えた currentUser。
	user *domain.User
}

// runCurrentUser は本物のルーターに CurrentUser を載せてリクエストを 1 本通す。
// gin.CreateTestContext で middleware を直に呼ぶと chain が無いため、AbortWithStatusJSON が
// 後続を止めることを確かめられない（止め損ねても通ってしまう）。
func runCurrentUser(t *testing.T, users *stubUsers, workspaces *stubWorkspaces) currentUserResult {
	t.Helper()
	got := currentUserResult{rec: httptest.NewRecorder()}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(ContextKeyCognitoSub, "sub-123") })
	r.Use(CurrentUser(users, workspaces))
	r.GET("/", func(c *gin.Context) {
		got.reached = true
		got.user = CurrentUserFromContext(c)
		c.Status(http.StatusOK)
	})
	r.ServeHTTP(got.rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return got
}

func strPtr(v string) *string { return &v }

func Test_カレントユーザー_停止中のワークスペースを遮断(t *testing.T) {
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, WorkspaceID: strPtr("ws-7")}}
	workspaces := &stubWorkspaces{workspace: &domain.Workspace{ID: "ws-7", IsActive: false}}

	got := runCurrentUser(t, users, workspaces)

	if workspaces.gotWorkspaceID != "ws-7" {
		t.Fatalf("ワークスペースはユーザーの workspace_id で引くべき: got %q", workspaces.gotWorkspaceID)
	}
	if got.rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", got.rec.Code)
	}
	if got.reached {
		t.Fatal("停止中ワークスペースのリクエストを後続へ通してはならない")
	}
}

func Test_カレントユーザー_有効なワークスペースは許可(t *testing.T) {
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, WorkspaceID: strPtr("ws-7")}}
	workspaces := &stubWorkspaces{workspace: &domain.Workspace{ID: "ws-7", IsActive: true}}

	got := runCurrentUser(t, users, workspaces)

	if !got.reached {
		t.Fatal("有効なワークスペースのリクエストは後続へ通すべき")
	}
	if got.user == nil {
		t.Fatal("currentUser が context にセットされるべき")
	}
}

func Test_カレントユーザー_未所属ユーザーはワークスペースを引かずに許可(t *testing.T) {
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleSuperAdmin, IsActive: true, WorkspaceID: nil}}
	workspaces := &stubWorkspaces{err: repository.ErrWorkspaceNotFound}

	got := runCurrentUser(t, users, workspaces)

	if workspaces.gotWorkspaceID != "" {
		t.Fatalf("未所属ユーザーでワークスペースを引くべきではない: got %q", workspaces.gotWorkspaceID)
	}
	if !got.reached {
		t.Fatal("未所属ユーザーは後続へ通すべき")
	}
}

func Test_カレントユーザー_所属先の行が無ければ遮断(t *testing.T) {
	// users.workspace_id には FK が張ってあるので、所属先の行は必ず存在するはず。
	// 無いのはデータ不整合であって「停止されていない」ことの証拠ではないので、
	// 素通りさせずに弾く（素通りにすると FK が外れた瞬間に遮断が効かなくなる）。
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, WorkspaceID: strPtr("ws-99")}}
	workspaces := &stubWorkspaces{err: repository.ErrWorkspaceNotFound}

	got := runCurrentUser(t, users, workspaces)

	if got.rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", got.rec.Code)
	}
	if got.reached {
		t.Fatal("所属先の行が無いリクエストを後続へ通してはならない")
	}
}

func Test_カレントユーザー_無効なユーザーを遮断(t *testing.T) {
	// IsActive=false のユーザーはワークスペースが有効でも弾く（即時に利用不可）。
	users := &stubUsers{user: &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: false, WorkspaceID: strPtr("ws-7")}}
	workspaces := &stubWorkspaces{workspace: &domain.Workspace{ID: "ws-7", IsActive: true}}

	got := runCurrentUser(t, users, workspaces)

	if got.rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", got.rec.Code)
	}
	if got.reached {
		t.Fatal("無効ユーザーのリクエストを後続へ通してはならない")
	}
}
