package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// makeCtx は gin.Context を生成し、context に current user を埋め込んで返す。
func makeCtx(currentUserID uint64, paramUserID string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	if currentUserID != 0 {
		c.Set(middleware.ContextKeyCurrentUserID, currentUserID)
	}
	c.Params = gin.Params{{Key: "userId", Value: paramUserID}}
	return c
}

func Test_プロフィール_ユーザーID解決_meキーワード(t *testing.T) {
	h := &ProfileHandler{}
	uid, err := h.resolveUserID(makeCtx(7, "me"))
	if err != nil || uid != 7 {
		t.Fatalf("'me' should resolve to current user; got uid=%d err=%v", uid, err)
	}
}

func Test_プロフィール_ユーザーID解決_空パラメータ(t *testing.T) {
	h := &ProfileHandler{}
	uid, err := h.resolveUserID(makeCtx(7, ""))
	if err != nil || uid != 7 {
		t.Fatalf("empty param should resolve to current user; got uid=%d err=%v", uid, err)
	}
}

func Test_プロフィール_ユーザーID解決_一致する数値(t *testing.T) {
	h := &ProfileHandler{}
	uid, err := h.resolveUserID(makeCtx(7, "7"))
	if err != nil || uid != 7 {
		t.Fatalf("matching numeric should pass; got uid=%d err=%v", uid, err)
	}
}

func Test_プロフィール_ユーザーID解決_不一致の数値は禁止(t *testing.T) {
	h := &ProfileHandler{}
	if _, err := h.resolveUserID(makeCtx(7, "99")); !errors.Is(err, errProfileForbidden) {
		t.Fatalf("mismatch numeric should be forbidden; got %v", err)
	}
}

func Test_プロフィール_ユーザーID解決_カレントユーザーなしは未認証(t *testing.T) {
	h := &ProfileHandler{}
	if _, err := h.resolveUserID(makeCtx(0, "me")); !errors.Is(err, errProfileUnauthorized) {
		t.Fatalf("no current user should be unauthorized; got %v", err)
	}
}

// stubProfileUserRepo は Update 経路で使うメソッドだけ実装した UserRepository スタブ。
// 未実装メソッドは埋め込んだ nil interface 経由で panic する（呼ばれない前提の検知になる）。
type stubProfileUserRepo struct {
	repository.UserRepository
	updatedName   string
	updateCalled  bool
	foundUserName string
}

func (s *stubProfileUserRepo) UpdateName(_ context.Context, _ uint64, name string) error {
	s.updateCalled = true
	s.updatedName = name
	return nil
}

func (s *stubProfileUserRepo) FindByID(_ context.Context, id uint64) (*domain.User, error) {
	return &domain.User{ID: id, Name: s.foundUserName}, nil
}

// stubProfileRepo は ProfileRepository の in-memory スタブ。
type stubProfileRepo struct {
	saved *domain.Profile
}

func (s *stubProfileRepo) FindByUserID(_ context.Context, userID uint64) (*domain.Profile, error) {
	if s.saved != nil {
		return s.saved, nil
	}
	return &domain.Profile{UserID: userID}, nil
}

func (s *stubProfileRepo) Upsert(_ context.Context, p *domain.Profile) error {
	s.saved = p
	return nil
}

// doProfileUpdate は PUT /profile/me を httptest で実行し recorder と stub を返す。
func doProfileUpdate(t *testing.T, body string) (*httptest.ResponseRecorder, *stubProfileUserRepo, *stubProfileRepo) {
	t.Helper()
	users := &stubProfileUserRepo{foundUserName: "既存の名前"}
	profiles := &stubProfileRepo{}
	h := NewProfileHandler(
		usecase.NewGetProfileUseCase(profiles),
		usecase.NewUpdateProfileUseCase(profiles),
		users,
	)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyCurrentUserID, uint64(7))
	c.Params = gin.Params{{Key: "userId", Value: "me"}}
	c.Request = httptest.NewRequest("PUT", "/profile/me", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Update(c)
	return w, users, profiles
}

func Test_プロフィール更新_displayNameキーで氏名がUpdateNameに渡る(t *testing.T) {
	// フロント (UpdateProfileRequest) の実送信キーは displayName。
	// 旧タグ json:"name" ではここが常に空になり氏名が保存されなかった (FRESTYLE-198)。
	w, users, profiles := doProfileUpdate(t, `{"displayName":"河野拓真","bio":"自己紹介","status":"勤務中"}`)
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if !users.updateCalled || users.updatedName != "河野拓真" {
		t.Fatalf("UpdateName should be called with 河野拓真; called=%v name=%q", users.updateCalled, users.updatedName)
	}
	if profiles.saved == nil || profiles.saved.Bio != "自己紹介" || profiles.saved.StatusMessage != "勤務中" {
		t.Fatalf("bio/status should be upserted together; got %+v", profiles.saved)
	}
}

func Test_プロフィール更新_氏名省略時はUpdateNameを呼ばない(t *testing.T) {
	w, users, profiles := doProfileUpdate(t, `{"bio":"自己紹介のみ"}`)
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if users.updateCalled {
		t.Fatalf("UpdateName should not be called when displayName is omitted")
	}
	if profiles.saved == nil || profiles.saved.Bio != "自己紹介のみ" {
		t.Fatalf("bio should still be upserted; got %+v", profiles.saved)
	}
}

func Test_プロフィール表示_JSONの氏名キーはdisplayName(t *testing.T) {
	// フロントの Profile 型は displayName を読む。name で返すと氏名欄・ヘッダーが空になる。
	b, err := json.Marshal(domain.ProfileView{Name: "河野拓真"})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["displayName"] != "河野拓真" {
		t.Fatalf(`ProfileView JSON should expose displayName; got %v`, m)
	}
	if _, ok := m["name"]; ok {
		t.Fatalf("ProfileView JSON should not expose legacy key name; got %v", m)
	}
}
