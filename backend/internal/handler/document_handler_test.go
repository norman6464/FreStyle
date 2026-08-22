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

const (
	testDocUUID = "31400a07-297e-8057-884b-c05dbdf9fa53"
	testDocBody = `{"type":"doc","content":[{"type":"paragraph"}]}`
)

// fakeDocRepo は repository.RichDocumentRepository の最小 fake。
type fakeDocRepo struct {
	getDoc    *domain.RichDocument
	getErr    error
	createErr error
	updateErr error
	updateDoc *domain.RichDocument
	deleteErr error
}

func (f *fakeDocRepo) Create(_ context.Context, doc *domain.RichDocument) error {
	if f.createErr != nil {
		return f.createErr
	}
	if doc.ID == "" {
		doc.ID = testDocUUID
	}
	return nil
}

func (f *fakeDocRepo) FindByID(_ context.Context, _ string) (*domain.RichDocument, error) {
	return f.getDoc, f.getErr
}

func (f *fakeDocRepo) UpdateWithRevision(_ context.Context, doc *domain.RichDocument, _ int) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if f.updateDoc != nil {
		*doc = *f.updateDoc
	}
	return nil
}
func (f *fakeDocRepo) SoftDelete(_ context.Context, _ string, _ uint64) error { return f.deleteErr }

func newDocHandler(repo repository.RichDocumentRepository) *DocumentHandler {
	return NewDocumentHandler(
		usecase.NewGetRichDocumentUseCase(repo),
		usecase.NewCreateRichDocumentUseCase(repo),
		usecase.NewUpdateRichDocumentUseCase(repo),
		usecase.NewDeleteRichDocumentUseCase(repo),
	)
}

func docCtx(method, body string, uid uint64, idVal string) (*httptest.ResponseRecorder, *gin.Context) {
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

func Test_文書ハンドラ_作成(t *testing.T) {
	t.Run("未認証は401", func(t *testing.T) {
		w, c := docCtx(http.MethodPost, `{}`, 0, "")
		newDocHandler(&fakeDocRepo{}).Create(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系は201", func(t *testing.T) {
		body := `{"kind":"note","title":"メモ","doc":` + testDocBody + `}`
		w, c := docCtx(http.MethodPost, body, 7, "")
		newDocHandler(&fakeDocRepo{}).Create(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
	})
	t.Run("不正なdocは400", func(t *testing.T) {
		body := `{"kind":"note","title":"メモ","doc":{"type":"paragraph"}}`
		w, c := docCtx(http.MethodPost, body, 7, "")
		newDocHandler(&fakeDocRepo{}).Create(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("kind欠落は400(binding)", func(t *testing.T) {
		body := `{"title":"メモ","doc":` + testDocBody + `}`
		w, c := docCtx(http.MethodPost, body, 7, "")
		newDocHandler(&fakeDocRepo{}).Create(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func Test_文書ハンドラ_取得(t *testing.T) {
	t.Run("所有者は200", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Doc: testDocBody, Kind: domain.DocumentKindNote}}
		w, c := docCtx(http.MethodGet, "", 7, testDocUUID)
		newDocHandler(repo).Get(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"doc":{`) {
			t.Fatalf("doc should be embedded as JSON object, got %s", w.Body.String())
		}
	})
	t.Run("他人の非公開は404", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, IsPublic: false, Doc: testDocBody}}
		w, c := docCtx(http.MethodGet, "", 99, testDocUUID)
		newDocHandler(repo).Get(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
	t.Run("不正なIDは400", func(t *testing.T) {
		w, c := docCtx(http.MethodGet, "", 7, "not-a-uuid")
		newDocHandler(&fakeDocRepo{}).Get(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("ダッシュ無しUUIDも受理", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Doc: testDocBody}}
		w, c := docCtx(http.MethodGet, "", 7, "31400a07297e8057884bc05dbdf9fa53")
		newDocHandler(repo).Get(c)
		if w.Code != http.StatusOK {
			t.Fatalf("dashless UUID should be accepted, got %d", w.Code)
		}
	})
}

func Test_文書ハンドラ_更新(t *testing.T) {
	body := `{"title":"new","doc":` + testDocBody + `,"revision":3}`
	t.Run("所有者は200", func(t *testing.T) {
		repo := &fakeDocRepo{
			getDoc:    &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Revision: 3},
			updateDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Title: "new", Doc: testDocBody, Revision: 4},
		}
		w, c := docCtx(http.MethodPut, body, 7, testDocUUID)
		newDocHandler(repo).Update(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
	})
	t.Run("他人は存在を漏らさず404", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Revision: 3}}
		w, c := docCtx(http.MethodPut, body, 99, testDocUUID)
		newDocHandler(repo).Update(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
	t.Run("版不一致は409", func(t *testing.T) {
		repo := &fakeDocRepo{
			getDoc:    &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Revision: 5},
			updateErr: repository.ErrRichDocumentConflict,
		}
		w, c := docCtx(http.MethodPut, body, 7, testDocUUID)
		newDocHandler(repo).Update(c)
		if w.Code != http.StatusConflict {
			t.Fatalf("want 409, got %d", w.Code)
		}
	})
	t.Run("存在しないは404", func(t *testing.T) {
		repo := &fakeDocRepo{getErr: repository.ErrRichDocumentNotFound}
		w, c := docCtx(http.MethodPut, body, 7, testDocUUID)
		newDocHandler(repo).Update(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
	t.Run("revision欠落は400", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7}}
		w, c := docCtx(http.MethodPut, `{"title":"x","doc":`+testDocBody+`}`, 7, testDocUUID)
		newDocHandler(repo).Update(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func Test_文書ハンドラ_削除(t *testing.T) {
	t.Run("成功は204", func(t *testing.T) {
		w, c := docCtx(http.MethodDelete, "", 7, testDocUUID)
		newDocHandler(&fakeDocRepo{}).Delete(c)
		// c.Status() は body を書かないと直接呼び出しでは flush されないため明示的に確定する。
		c.Writer.WriteHeaderNow()
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d", w.Code)
		}
	})
	t.Run("存在しない(他人)は404", func(t *testing.T) {
		repo := &fakeDocRepo{deleteErr: repository.ErrRichDocumentNotFound}
		w, c := docCtx(http.MethodDelete, "", 7, testDocUUID)
		newDocHandler(repo).Delete(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
}
