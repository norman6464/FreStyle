package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type fakeHealthRepo struct{ err error }

func (f fakeHealthRepo) PingDB(context.Context) error { return f.err }

func runHealthGet(t *testing.T, repoErr error) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHealthHandler(usecase.NewCheckHealthUseCase(fakeHealthRepo{err: repoErr}))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	h.Get(c)
	return w
}

func TestHealthHandler_Get_Up(t *testing.T) {
	w := runHealthGet(t, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var body domain.Health
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Status != domain.StatusUp {
		t.Errorf("want status UP, got %q", body.Status)
	}
}

func TestHealthHandler_Get_Down(t *testing.T) {
	w := runHealthGet(t, errors.New("db unreachable"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", w.Code)
	}
	var body domain.Health
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Status != domain.StatusDown {
		t.Errorf("want status DOWN, got %q", body.Status)
	}
}
