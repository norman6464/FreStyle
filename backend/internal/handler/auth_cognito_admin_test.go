package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// captureSlogLines は slog の既定出力を buffer に差し替えて JSON ログを回収する。
func captureSlogLines(t *testing.T, run func()) []map[string]any {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	defer slog.SetDefault(prev)

	run()

	var out []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(line, &m); err != nil {
			t.Fatalf("ログが JSON でない: %v (%s)", err, line)
		}
		out = append(out, m)
	}
	return out
}

// findLog は msg が一致する最初のログを返す。
func findLog(logs []map[string]any, msg string) map[string]any {
	for _, l := range logs {
		if l["msg"] == msg {
			return l
		}
	}
	return nil
}

// newMeCtx は /auth/me 用に sub / groups を積んだテスト context を返す。
func newMeCtx(sub string, groups []string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v2/auth/me", nil)
	c.Set(middleware.ContextKeyCognitoSub, sub)
	c.Set(middleware.ContextKeyCognitoGroups, groups)
	return c, rec
}

func newMeHandler(users *fakeUserRepo) *AuthHandler {
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	h.getCurrentUser = usecase.NewGetCurrentUserUseCase(users)
	return h
}

// Cognito の admin グループに属しているだけでは招待統制を迂回できない。自己サインアップ
// 自体は許可されるが、super_admin へは昇格しない。
func Test_IDトークンからユーザー登録_Cognito_adminでも招待なしの新規はSuperAdminへ昇格しない(t *testing.T) {
	users := &fakeUserRepo{}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})

	idToken := makeIDToken(t, map[string]any{
		"sub":            "new-admin",
		"email":          "attacker@example.com",
		"cognito:groups": []string{"admin"},
	})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("招待の無い新規ユーザーも自己サインアップできるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role == domain.RoleSuperAdmin {
		t.Fatalf("admin グループだけで昇格してはいけない: %+v", users.created)
	}
}

// ブートストラップ用に指定した 1 アドレスだけは、招待なしで最初の運営管理者になれる。
func Test_IDトークンからユーザー登録_ブートストラップアドレスは招待なしで作成(t *testing.T) {
	users := &fakeUserRepo{}
	h := newTestAuthHandlerWithBootstrap(users, &fakeInvitationRepo{}, "ops@example.com")

	idToken := makeIDToken(t, map[string]any{
		"sub":            "first-admin",
		"email":          "ops@example.com",
		"cognito:groups": []string{"admin"},
	})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("ブートストラップ指定アドレスは許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role != domain.RoleSuperAdmin {
		t.Fatalf("role = %q, want %q", users.created.Role, domain.RoleSuperAdmin)
	}
}

// ブートストラップは「最初の 1 人」限定。運営管理者が既に居れば効かず、自己サインアップは
// 通っても super_admin へは昇格しない。
func Test_IDトークンからユーザー登録_ブートストラップは運営管理者在籍時に閉じる(t *testing.T) {
	users := &fakeUserRepo{superAdmins: []domain.User{{ID: 1, Role: domain.RoleSuperAdmin}}}
	h := newTestAuthHandlerWithBootstrap(users, &fakeInvitationRepo{}, "ops@example.com")

	idToken := makeIDToken(t, map[string]any{
		"sub":            "second-admin",
		"email":          "ops@example.com",
		"cognito:groups": []string{"admin"},
	})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("bootstrap が閉じても自己サインアップは許可されるべき")
	}
	if users.created == nil {
		t.Fatal("ユーザーが作成されていない")
	}
	if users.created.Role == domain.RoleSuperAdmin {
		t.Fatalf("運営管理者が既に居るなら昇格してはいけない: %+v", users.created)
	}
}

// 招待があれば従来どおり作成できる（ブートストラップ未設定でも影響しない）。
func Test_IDトークンからユーザー登録_招待があるCognito_adminは従来どおり作成(t *testing.T) {
	users := &fakeUserRepo{}
	invs := &fakeInvitationRepo{
		pendingByEmail: map[string]*domain.AdminInvitation{
			"ops@example.com": {ID: 3, Role: domain.RoleCompanyAdmin, CompanyID: 1},
		},
	}
	h := newTestAuthHandler(users, invs)

	idToken := makeIDToken(t, map[string]any{
		"sub":            "invited-admin",
		"email":          "ops@example.com",
		"cognito:groups": []string{"admin"},
	})

	if !upsertAllowed(h, newGinCtx(), idToken, "") {
		t.Fatal("招待がある admin グループユーザーは許可されるべき")
	}
	if users.created == nil || users.created.Role != domain.RoleSuperAdmin {
		t.Fatalf("created = %+v, want role %q", users.created, domain.RoleSuperAdmin)
	}
}

// 招待なしの自己サインアップは握り潰さずログに残す（誰がどのアドレスで作られたかを
// 運用が追えるようにする）。
func Test_IDトークンからユーザー登録_招待なしのサインアップをログに残す(t *testing.T) {
	users := &fakeUserRepo{}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	idToken := makeIDToken(t, map[string]any{
		"sub":            "new-admin",
		"email":          "attacker@example.com",
		"cognito:groups": []string{"admin"},
	})

	logs := captureSlogLines(t, func() {
		if !upsertAllowed(h, newGinCtx(), idToken, "") {
			t.Fatal("許可されるべき")
		}
	})

	got := findLog(logs, "self signup: creating a new user without invitation")
	if got == nil {
		t.Fatalf("サインアップのログが出ていない: %+v", logs)
	}
	if got["email"] != "attacker@example.com" {
		t.Fatalf("ログの内容が足りない: %+v", got)
	}
}

// /auth/me: Cognito admin グループのユーザーは DB role を super_admin に同期する（昇格のみ）。
func Test_現在ユーザー取得_Cognito_adminをSuperAdminへ同期(t *testing.T) {
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 5, Email: "u@example.com", Role: domain.RoleTrainee},
	}}
	h := newMeHandler(users)
	c, rec := newMeCtx("u1", []string{"admin"})

	h.Me(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if users.updateRoleID != 5 || users.updateRoleVal != domain.RoleSuperAdmin {
		t.Fatalf("UpdateRole(%d, %q) されていない", users.updateRoleID, users.updateRoleVal)
	}
}

// 昇格したらレスポンスの role も昇格後の値にする。
// 捨てると、初回だけ「isAdmin=true / role=trainee」という起き得ない組み合わせを返してしまい、
// role を見て出し分けている画面が初回ログインのときだけ違う挙動になる。
func Test_現在ユーザー取得_昇格後のroleをレスポンスに反映する(t *testing.T) {
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 5, Email: "u@example.com", Role: domain.RoleTrainee},
	}}
	h := newMeHandler(users)
	c, rec := newMeCtx("u1", []string{"admin"})

	h.Me(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("レスポンスが JSON でない: %v (%s)", err, rec.Body.String())
	}
	if body["role"] != string(domain.RoleSuperAdmin) {
		t.Errorf("role = %v, want %q", body["role"], domain.RoleSuperAdmin)
	}
	if body["isAdmin"] != true {
		t.Errorf("isAdmin = %v, want true", body["isAdmin"])
	}
}

// 逆に、昇格できなかったときは role を勝手に書き換えない（DB の実態より強い権限を返さない）。
func Test_現在ユーザー取得_昇格に失敗したらroleは元のまま(t *testing.T) {
	users := &fakeUserRepo{
		existingBySub: map[string]*domain.User{
			"u1": {ID: 5, Email: "u@example.com", Role: domain.RoleTrainee},
		},
		updateRoleErr: errors.New(`unknown role "super_admin"`),
	}
	h := newMeHandler(users)
	c, rec := newMeCtx("u1", []string{"admin"})

	h.Me(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("レスポンスが JSON でない: %v (%s)", err, rec.Body.String())
	}
	if body["role"] != string(domain.RoleTrainee) {
		t.Errorf("role = %v, want %q", body["role"], domain.RoleTrainee)
	}
}

// 昇格に失敗しても /auth/me は返すが、失敗は必ずログに残す（握り潰さない）。
// 握り潰すと「UI は管理者・API は 403」の壊れた状態が無言で続く。
func Test_現在ユーザー取得_ロール同期の失敗をログに残す(t *testing.T) {
	updateErr := errors.New(`unknown role "super_admin"`)
	users := &fakeUserRepo{
		existingBySub: map[string]*domain.User{
			"u1": {ID: 5, Email: "u@example.com", Role: domain.RoleTrainee},
		},
		updateRoleErr: updateErr,
	}
	h := newMeHandler(users)
	c, rec := newMeCtx("u1", []string{"admin"})

	logs := captureSlogLines(t, func() { h.Me(c) })

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	got := findLog(logs, "cognito admin role sync failed")
	if got == nil {
		t.Fatalf("同期失敗のログが出ていない: %+v", logs)
	}
	if got["cognitoSub"] != "u1" {
		t.Errorf("cognitoSub = %v, want u1", got["cognitoSub"])
	}
	if errText, _ := got["err"].(string); !strings.Contains(errText, updateErr.Error()) {
		t.Errorf("err = %v, want 原因を含む", got["err"])
	}
}

// 管理者グループでないユーザーはロールに触らない（権限を与えも奪いもしない）。
func Test_現在ユーザー取得_非管理者はロールを触らない(t *testing.T) {
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 5, Email: "u@example.com", Role: domain.RoleTrainee},
	}}
	h := newMeHandler(users)
	c, rec := newMeCtx("u1", []string{"other"})

	h.Me(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if users.updateRoleCalls != 0 {
		t.Fatalf("UpdateRole が %d 回呼ばれた", users.updateRoleCalls)
	}
}

// companyId と同じく、workspaceId も所属していればレスポンスに含む。
func Test_現在ユーザー取得_所属していればcompanyIdとworkspaceIdを含む(t *testing.T) {
	wsID := "0198a000-0000-7000-8000-000000000001"
	cid := uint64(7)
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 5, Email: "u@example.com", Role: domain.RoleTrainee, CompanyID: &cid, WorkspaceID: &wsID},
	}}
	h := newMeHandler(users)
	c, rec := newMeCtx("u1", nil)

	h.Me(c)

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["companyId"] != float64(cid) {
		t.Fatalf("companyId = %v, want %d", body["companyId"], cid)
	}
	if body["workspaceId"] != wsID {
		t.Fatalf("workspaceId = %v, want %q", body["workspaceId"], wsID)
	}
}

// 未所属（company_id/workspace_id が NULL）のユーザーは、どちらのフィールドも省略する。
func Test_現在ユーザー取得_未所属はcompanyIdとworkspaceIdを省略する(t *testing.T) {
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"u1": {ID: 5, Email: "u@example.com", Role: domain.RoleSuperAdmin},
	}}
	h := newMeHandler(users)
	c, rec := newMeCtx("u1", nil)

	h.Me(c)

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := body["companyId"]; ok {
		t.Fatalf("未所属なのに companyId が含まれている: %v", body["companyId"])
	}
	if _, ok := body["workspaceId"]; ok {
		t.Fatalf("未所属なのに workspaceId が含まれている: %v", body["workspaceId"])
	}
}

// access_token 経路（id_token に groups が無い federated ユーザー）でも同期する。
func Test_アクセストークンからロール同期_SuperAdminへ昇格(t *testing.T) {
	users := &fakeUserRepo{existingBySub: map[string]*domain.User{
		"fed-1": {ID: 8, Email: "f@example.com", Role: domain.RoleTrainee},
	}}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	token := makeIDToken(t, map[string]any{
		"sub":            "fed-1",
		"cognito:groups": []string{"admin"},
	})

	h.syncRoleFromAccessToken(newGinCtx(), token)

	if users.updateRoleID != 8 || users.updateRoleVal != domain.RoleSuperAdmin {
		t.Fatalf("UpdateRole(%d, %q) されていない", users.updateRoleID, users.updateRoleVal)
	}
}

// DB 障害と「ユーザーが居ない」を同じ無反応に畳まない。障害はログに残す。
func Test_アクセストークンからロール同期_検索失敗をログに残す(t *testing.T) {
	findErr := errors.New("db down")
	users := &fakeUserRepo{findErr: findErr}
	h := newTestAuthHandler(users, &fakeInvitationRepo{})
	token := makeIDToken(t, map[string]any{
		"sub":            "fed-1",
		"cognito:groups": []string{"admin"},
	})

	logs := captureSlogLines(t, func() { h.syncRoleFromAccessToken(newGinCtx(), token) })

	got := findLog(logs, "cognito admin role sync failed")
	if got == nil {
		t.Fatalf("検索失敗のログが出ていない: %+v", logs)
	}
	if errText, _ := got["err"].(string); !strings.Contains(errText, findErr.Error()) {
		t.Errorf("err = %v, want 原因を含む", got["err"])
	}
}
