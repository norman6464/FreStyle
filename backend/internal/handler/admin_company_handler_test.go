package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// fakeCompanyRepo は repository.CompanyRepository の最小 fake。
type fakeCompanyRepo struct {
	rows []domain.Company
	err  error
}

func (f *fakeCompanyRepo) ListAll(context.Context) ([]domain.Company, error) { return f.rows, f.err }

func (f *fakeCompanyRepo) FindByID(context.Context, uint64) (*domain.Company, error) {
	return nil, f.err
}

func newAdminCompanyHandler(repo *fakeCompanyRepo) *AdminCompanyHandler {
	return NewAdminCompanyHandler(usecase.NewListCompaniesUseCase(repo))
}

func TestAdminCompanyHandler_List_OK(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/companies", nil)
	newAdminCompanyHandler(&fakeCompanyRepo{rows: []domain.Company{{ID: 1, Name: "Co"}}}).List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestAdminCompanyHandler_List_Error(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/companies", nil)
	newAdminCompanyHandler(&fakeCompanyRepo{err: context.DeadlineExceeded}).List(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}
