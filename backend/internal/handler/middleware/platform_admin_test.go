package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// platformAdminUsers は運営権限の同期だけを見るための stub。
// 実際の persistence と同じく、更新後の読み直しでは実効役割（ResolveEffectiveRole）を反映する。
type platformAdminUsers struct {
	stubUsers
	stored    domain.RoleName
	flag      bool
	updateErr error
	findErr   error
	// reReadErr は「更新は通ったが、反映後の読み直しだけが落ちる」を作るためのもの。
	reReadErr   error
	findCalls   int
	updateCalls int
}

func (p *platformAdminUsers) FindByCognitoSub(context.Context, string) (*domain.User, error) {
	p.findCalls++
	if p.findErr != nil {
		return nil, p.findErr
	}
	if p.reReadErr != nil && p.updateCalls > 0 {
		return nil, p.reReadErr
	}
	return p.current(), nil
}

func (p *platformAdminUsers) UpdatePlatformAdmin(_ context.Context, _ uint64, v bool) error {
	p.updateCalls++
	if p.updateErr != nil {
		return p.updateErr
	}
	p.flag = v
	return nil
}

func (p *platformAdminUsers) current() *domain.User {
	return &domain.User{
		ID:              7,
		Role:            domain.ResolveEffectiveRole(p.stored, p.flag),
		IsActive:        true,
		IsPlatformAdmin: p.flag,
	}
}

// runSyncPlatformAdmin は CurrentUser 相当（sub と *domain.User を context へ）を済ませた状態で
// middleware を 1 リクエスト分だけ回し、status とハンドラまで届いた user を返す。
func runSyncPlatformAdmin(t *testing.T, users *platformAdminUsers, groups any, hasGroups bool) (int, *domain.User) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var got *domain.User
	r := gin.New()
	r.GET(
		"/x",
		func(c *gin.Context) {
			c.Set(ContextKeyCognitoSub, "sub-1")
			if hasGroups {
				c.Set(ContextKeyCognitoGroups, groups)
			}
			c.Set(ContextKeyCurrentUser, users.current())
			users.findCalls = 0 // CurrentUser 相当の読み出しは数えない
			c.Next()
		},
		SyncPlatformAdmin(usecase.NewSyncPlatformAdminUseCase(users), users),
		func(c *gin.Context) {
			got = CurrentUserFromContext(c)
			c.Status(http.StatusOK)
		},
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	return w.Code, got
}

func Test_運営権限の同期middleware_claimが無ければ触らない(t *testing.T) {
	users := &platformAdminUsers{stored: domain.RoleSuperAdmin, flag: true}
	code, got := runSyncPlatformAdmin(t, users, nil, false)

	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if users.updateCalls != 0 || users.findCalls != 0 {
		t.Fatalf("claim 欠落で DB を触ってはならない: update=%d find=%d", users.updateCalls, users.findCalls)
	}
	if got.Role != domain.RoleSuperAdmin {
		t.Fatalf("claim 欠落で降格してはならない: %v", got.Role)
	}
}

func Test_運営権限の同期middleware_一致していれば読み直さない(t *testing.T) {
	users := &platformAdminUsers{stored: domain.RoleSuperAdmin, flag: true}
	code, got := runSyncPlatformAdmin(t, users, []string{"admin"}, true)

	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if users.updateCalls != 0 || users.findCalls != 0 {
		t.Fatalf("値が同じなら DB に触らない: update=%d find=%d", users.updateCalls, users.findCalls)
	}
	if got.Role != domain.RoleSuperAdmin {
		t.Fatalf("expected super_admin, got %v", got.Role)
	}
}

// 認可（RequireAdmin / handler の役割検査）に渡す前に降格していることを見る。
// ここで差し替えないと、/auth/me を通らずに管理 API を叩いた退任者が古い値で認可される。
func Test_運営権限の同期middleware_グループから外れたら認可前に降格する(t *testing.T) {
	users := &platformAdminUsers{stored: domain.RoleSuperAdmin, flag: true}
	code, got := runSyncPlatformAdmin(t, users, []string{"users"}, true)

	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if users.updateCalls != 1 {
		t.Fatalf("expected 1 update, got %d", users.updateCalls)
	}
	if got.IsPlatformAdmin {
		t.Fatal("運営権限は剥奪されているべき")
	}
	if got.Role != domain.RoleTrainee {
		t.Fatalf("実効役割は最小権限へ倒れる: %v", got.Role)
	}
}

func Test_運営権限の同期middleware_グループに戻れば認可前に付与する(t *testing.T) {
	users := &platformAdminUsers{stored: domain.RoleSuperAdmin, flag: false}
	code, got := runSyncPlatformAdmin(t, users, []string{"admin"}, true)

	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if got.Role != domain.RoleSuperAdmin {
		t.Fatalf("expected super_admin, got %v", got.Role)
	}
}

func Test_運営権限の同期middleware_剥奪に失敗したら通さない(t *testing.T) {
	users := &platformAdminUsers{stored: domain.RoleSuperAdmin, flag: true, updateErr: errors.New("db down")}
	code, got := runSyncPlatformAdmin(t, users, []string{"users"}, true)

	if code != http.StatusInternalServerError {
		t.Fatalf("剥奪に失敗したら fail closed: got %d", code)
	}
	if got != nil {
		t.Fatal("handler まで到達してはならない")
	}
}

func Test_運営権限の同期middleware_付与に失敗しても妨げない(t *testing.T) {
	users := &platformAdminUsers{stored: domain.RoleSuperAdmin, flag: false, updateErr: errors.New("db down")}
	code, got := runSyncPlatformAdmin(t, users, []string{"admin"}, true)

	if code != http.StatusOK {
		t.Fatalf("付与の失敗は権限が上がらないだけなので素通しでよい: got %d", code)
	}
	if got.Role != domain.RoleTrainee {
		t.Fatalf("付与できていないので実効役割は最小権限のまま: %v", got.Role)
	}
}

func Test_運営権限の同期middleware_読み直しに失敗したら通さない(t *testing.T) {
	users := &platformAdminUsers{stored: domain.RoleSuperAdmin, flag: true, reReadErr: errors.New("db down")}
	code, got := runSyncPlatformAdmin(t, users, []string{"users"}, true)

	if code != http.StatusInternalServerError {
		t.Fatalf("反映後の事実を読めないまま通してはならない: got %d", code)
	}
	if got != nil {
		t.Fatal("handler まで到達してはならない")
	}
	if users.updateCalls != 1 {
		t.Fatalf("更新自体は通っている前提のケース: update=%d", users.updateCalls)
	}
}

func Test_運営権限の同期middleware_同期の失敗を握り潰さない(t *testing.T) {
	users := &platformAdminUsers{stored: domain.RoleSuperAdmin, flag: true, findErr: errors.New("db down")}
	code, _ := runSyncPlatformAdmin(t, users, []string{"users"}, true)

	if code != http.StatusInternalServerError {
		t.Fatalf("剥奪すべきなのに DB を読めない: fail closed であるべき got %d", code)
	}
}

func Test_運営権限の同期middleware_未配線なら素通しする(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", SyncPlatformAdmin(nil, nil), func(c *gin.Context) { c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
