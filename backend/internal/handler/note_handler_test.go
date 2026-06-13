package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeNoteRepo は repository.NoteRepository の最小 fake。
type fakeNoteRepo struct {
	rows []domain.Note
	one  *domain.Note
	err  error
}

func (f *fakeNoteRepo) ListByUserID(context.Context, uint64) ([]domain.Note, error) {
	return f.rows, f.err
}

func (f *fakeNoteRepo) FindByID(context.Context, uint64) (*domain.Note, error) {
	return f.one, f.err
}

func (f *fakeNoteRepo) Create(_ context.Context, n *domain.Note) error {
	if f.err == nil {
		n.ID = 100
	}
	return f.err
}
func (f *fakeNoteRepo) Update(context.Context, *domain.Note) error   { return f.err }
func (f *fakeNoteRepo) Delete(context.Context, uint64, uint64) error { return f.err }

func newNoteHandler(repo repository.NoteRepository) *NoteHandler {
	return NewNoteHandler(
		usecase.NewListNotesByUserIDUseCase(repo),
		usecase.NewCreateNoteUseCase(repo),
		usecase.NewUpdateNoteUseCase(repo),
		usecase.NewDeleteNoteUseCase(repo),
	)
}

// noteCtx は ContextKeyCurrentUserID（CurrentUserIDOrZero 用）+ JSON body + id param を持つ context。
func noteCtx(method, body string, uid uint64, idVal string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if uid != 0 {
		c.Set(middleware.ContextKeyCurrentUserID, uid)
	}
	if idVal != "" {
		c.Params = gin.Params{{Key: "id", Value: idVal}}
	}
	return w, c
}

func Test_ノートハンドラ_一覧(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := noteCtx(http.MethodGet, "", 0, "")
		newNoteHandler(&fakeNoteRepo{}).List(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := noteCtx(http.MethodGet, "", 7, "")
		newNoteHandler(&fakeNoteRepo{rows: []domain.Note{{ID: 1, Title: "n"}}}).List(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("リポジトリエラー → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodGet, "", 7, "")
		newNoteHandler(&fakeNoteRepo{err: context.DeadlineExceeded}).List(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func Test_ノートハンドラ_作成(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"title":"X"}`, 0, "")
		newNoteHandler(&fakeNoteRepo{}).Create(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("title 欠落 → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{}`, 7, "")
		newNoteHandler(&fakeNoteRepo{}).Create(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系 → 201", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"title":"X"}`, 7, "")
		newNoteHandler(&fakeNoteRepo{}).Create(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d", w.Code)
		}
	})
}

func Test_ノートハンドラ_更新(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := noteCtx(http.MethodPut, `{"title":"X"}`, 0, "5")
		newNoteHandler(&fakeNoteRepo{}).Update(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("title 欠落 → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPut, `{}`, 7, "5")
		newNoteHandler(&fakeNoteRepo{}).Update(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系 → 200", func(t *testing.T) {
		w, c := noteCtx(http.MethodPut, `{"title":"X"}`, 7, "5")
		newNoteHandler(&fakeNoteRepo{one: &domain.Note{ID: 5, UserID: 7}}).Update(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
}

func Test_ノートハンドラ_削除(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := noteCtx(http.MethodDelete, "", 0, "5")
		newNoteHandler(&fakeNoteRepo{}).Delete(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系 → 204", func(t *testing.T) {
		_, c := noteCtx(http.MethodDelete, "", 7, "5")
		newNoteHandler(&fakeNoteRepo{}).Delete(c)
		if c.Writer.Status() != http.StatusNoContent {
			t.Fatalf("want 204, got %d", c.Writer.Status())
		}
	})
	t.Run("リポジトリエラー → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodDelete, "", 7, "5")
		newNoteHandler(&fakeNoteRepo{err: context.DeadlineExceeded}).Delete(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}
