package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// settingsCompanyRepo は CompanyRepository の最小スタブ。
type settingsCompanyRepo struct {
	company *domain.Company
	updated *bool
}

func (s *settingsCompanyRepo) ListAll(_ context.Context) ([]domain.Company, error) { return nil, nil }

func (s *settingsCompanyRepo) FindByID(_ context.Context, _ uint64) (*domain.Company, error) {
	return s.company, nil
}

func (s *settingsCompanyRepo) UpdateActive(context.Context, uint64, bool) error { return nil }

func (s *settingsCompanyRepo) UpdateAiChatEnabled(_ context.Context, _ uint64, enabled bool) error {
	s.updated = &enabled
	return nil
}

// newCompanySettingsTestRouter は context に actor を注入してから company settings ルートを通す。
func newCompanySettingsTestRouter(repo *settingsCompanyRepo, actor *domain.User) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewCompanySettingsHandler(
		usecase.NewGetCompanyAiChatSettingUseCase(repo),
		usecase.NewUpdateCompanyAiChatSettingUseCase(repo),
	)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if actor != nil {
			c.Set(middleware.ContextKeyCurrentUser, actor)
		}
		c.Next()
	})
	r.GET("/company/settings", h.Get)
	r.PUT("/company/settings", h.Update)
	return r
}

func Test_会社設定ハンドラ_取得_管理者(t *testing.T) {
	repo := &settingsCompanyRepo{company: &domain.Company{ID: 1, AiChatEnabledForTrainees: false}}
	actor := &domain.User{Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(1)}
	r := newCompanySettingsTestRouter(repo, actor)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/company/settings", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]bool
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["aiChatEnabledForTrainees"] != false {
		t.Errorf("aiChatEnabledForTrainees = %v, want false", body["aiChatEnabledForTrainees"])
	}
}

func Test_会社設定ハンドラ_取得_traineeは禁止(t *testing.T) {
	repo := &settingsCompanyRepo{company: &domain.Company{ID: 1}}
	actor := &domain.User{Role: domain.RoleTrainee, CompanyID: u64ptr(1)}
	r := newCompanySettingsTestRouter(repo, actor)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/company/settings", nil))
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func Test_会社設定ハンドラ_更新_管理者(t *testing.T) {
	repo := &settingsCompanyRepo{company: &domain.Company{ID: 1, AiChatEnabledForTrainees: true}}
	actor := &domain.User{Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(1)}
	r := newCompanySettingsTestRouter(repo, actor)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/company/settings", strings.NewReader(`{"aiChatEnabledForTrainees":false}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if repo.updated == nil || *repo.updated != false {
		t.Errorf("UpdateAiChatEnabled が false で呼ばれていない: %v", repo.updated)
	}
}

func Test_会社設定ハンドラ_更新_ボディ欠落(t *testing.T) {
	repo := &settingsCompanyRepo{company: &domain.Company{ID: 1}}
	actor := &domain.User{Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(1)}
	r := newCompanySettingsTestRouter(repo, actor)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/company/settings", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func u64ptr(v uint64) *uint64 { return &v }
