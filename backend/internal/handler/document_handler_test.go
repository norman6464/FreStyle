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
	listDocs  []domain.RichDocument
	listErr   error
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

func (f *fakeDocRepo) ListByOwner(_ context.Context, _ uint64, _ domain.DocumentKind) ([]domain.RichDocument, error) {
	return f.listDocs, f.listErr
}

func newDocHandler(repo repository.RichDocumentRepository) *DocumentHandler {
	return NewDocumentHandler(
		usecase.NewGetRichDocumentUseCase(repo),
		usecase.NewCreateRichDocumentUseCase(repo),
		usecase.NewUpdateRichDocumentUseCase(repo),
		usecase.NewDeleteRichDocumentUseCase(repo),
		usecase.NewListRichDocumentsUseCase(repo),
	)
}

// newDocRouter は本番と同じ path/method でルートを張ったテスト用ルータを返す。
// handler を直接呼ぶと c.Status(204) が body 無しでフラッシュされない gin の挙動を避け、
// 本番同様 ServeHTTP 経由で HTTP メソッド・パス・:id 抽出・レスポンス確定まで検証する。
// uid==0 のときは current user middleware を挟まず未認証を再現する。
func newDocRouter(repo repository.RichDocumentRepository, uid uint64) *gin.Engine {
	return newDocRouterWithWorkspace(repo, uid, "")
}

// newDocRouterWithWorkspace は newDocRouter に加えて workspaceID（閲覧側の境界判定に使う）も設定する。
// workspaceID=="" は「ワークスペース未所属」を表す。
func newDocRouterWithWorkspace(repo repository.RichDocumentRepository, uid uint64, workspaceID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := newDocHandler(repo)
	r := gin.New()
	if uid != 0 {
		var wid *string
		if workspaceID != "" {
			v := workspaceID
			wid = &v
		}
		r.Use(func(c *gin.Context) {
			c.Set(middleware.ContextKeyCurrentUserID, uid)
			c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: uid, WorkspaceID: wid, Role: domain.RoleTrainee})
			c.Next()
		})
	}
	r.GET("/documents", h.List)
	r.POST("/documents", h.Create)
	r.GET("/documents/:id", h.Get)
	r.PUT("/documents/:id", h.Update)
	r.DELETE("/documents/:id", h.Delete)
	return r
}

func doDocReq(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	r.ServeHTTP(w, req)
	return w
}

func Test_文書ハンドラ_作成(t *testing.T) {
	t.Run("未認証は401", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 0)
		w := doDocReq(r, http.MethodPost, "/documents", `{}`)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系は201", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 7)
		body := `{"kind":"note","title":"メモ","doc":` + testDocBody + `}`
		w := doDocReq(r, http.MethodPost, "/documents", body)
		if w.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
		}
	})
	t.Run("不正なdocは400", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 7)
		body := `{"kind":"note","title":"メモ","doc":{"type":"paragraph"}}`
		w := doDocReq(r, http.MethodPost, "/documents", body)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("kind欠落は400(binding)", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 7)
		body := `{"title":"メモ","doc":` + testDocBody + `}`
		w := doDocReq(r, http.MethodPost, "/documents", body)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func Test_文書ハンドラ_取得(t *testing.T) {
	t.Run("所有者は200", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Doc: testDocBody, Kind: domain.DocumentKindNote}}
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodGet, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"doc":{`) {
			t.Fatalf("doc should be embedded as JSON object, got %s", w.Body.String())
		}
	})
	t.Run("他人の非公開は404", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, IsPublic: false, Doc: testDocBody}}
		r := newDocRouter(repo, 99)
		w := doDocReq(r, http.MethodGet, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
	t.Run("同一ワークスペースの他人は公開を読める(200)", func(t *testing.T) {
		wsA := "0198a000-0000-7000-8000-0000000000e1"
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, WorkspaceID: &wsA, IsPublic: true, Doc: testDocBody}}
		r := newDocRouterWithWorkspace(repo, 99, wsA)
		w := doDocReq(r, http.MethodGet, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
	})
	t.Run("別ワークスペースの他人は公開でも404", func(t *testing.T) {
		wsA := "0198a000-0000-7000-8000-0000000000e1"
		wsB := "0198a000-0000-7000-8000-0000000000e2"
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, WorkspaceID: &wsA, IsPublic: true, Doc: testDocBody}}
		r := newDocRouterWithWorkspace(repo, 99, wsB)
		w := doDocReq(r, http.MethodGet, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})
	t.Run("ワークスペース不明(NULL)の公開は他人から404", func(t *testing.T) {
		wsA := "0198a000-0000-7000-8000-0000000000e1"
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, IsPublic: true, Doc: testDocBody}}
		r := newDocRouterWithWorkspace(repo, 99, wsA)
		w := doDocReq(r, http.MethodGet, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
		}
	})
	t.Run("所有者はワークスペースが食い違っても200", func(t *testing.T) {
		wsA := "0198a000-0000-7000-8000-0000000000e1"
		wsB := "0198a000-0000-7000-8000-0000000000e2"
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, WorkspaceID: &wsB, IsPublic: true, Doc: testDocBody}}
		r := newDocRouterWithWorkspace(repo, 7, wsA)
		w := doDocReq(r, http.MethodGet, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
	})
	t.Run("未認証は401", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 0)
		w := doDocReq(r, http.MethodGet, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("不正なIDは400", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 7)
		w := doDocReq(r, http.MethodGet, "/documents/not-a-uuid", "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("ダッシュ無しUUIDも受理", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Doc: testDocBody}}
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodGet, "/documents/31400a07297e8057884bc05dbdf9fa53", "")
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
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodPut, "/documents/"+testDocUUID, body)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
	})
	t.Run("他人は存在を漏らさず404", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Revision: 3}}
		r := newDocRouter(repo, 99)
		w := doDocReq(r, http.MethodPut, "/documents/"+testDocUUID, body)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
	t.Run("負のrevisionは400", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Revision: 3}}
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodPut, "/documents/"+testDocUUID, `{"title":"x","doc":`+testDocBody+`,"revision":-1}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("版不一致は409", func(t *testing.T) {
		repo := &fakeDocRepo{
			getDoc:    &domain.RichDocument{ID: testDocUUID, OwnerID: 7, Revision: 5},
			updateErr: repository.ErrRichDocumentConflict,
		}
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodPut, "/documents/"+testDocUUID, body)
		if w.Code != http.StatusConflict {
			t.Fatalf("want 409, got %d", w.Code)
		}
	})
	t.Run("存在しないは404", func(t *testing.T) {
		repo := &fakeDocRepo{getErr: repository.ErrRichDocumentNotFound}
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodPut, "/documents/"+testDocUUID, body)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
	t.Run("revision欠落は400", func(t *testing.T) {
		repo := &fakeDocRepo{getDoc: &domain.RichDocument{ID: testDocUUID, OwnerID: 7}}
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodPut, "/documents/"+testDocUUID, `{"title":"x","doc":`+testDocBody+`}`)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func Test_文書ハンドラ_削除(t *testing.T) {
	t.Run("成功は204", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 7)
		w := doDocReq(r, http.MethodDelete, "/documents/"+testDocUUID, "")
		// ServeHTTP 経由なので c.Status(204) は自動でフラッシュされる（手動 WriteHeaderNow は不要）。
		if w.Code != http.StatusNoContent {
			t.Fatalf("want 204, got %d", w.Code)
		}
	})
	t.Run("存在しない(他人)は404", func(t *testing.T) {
		repo := &fakeDocRepo{deleteErr: repository.ErrRichDocumentNotFound}
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodDelete, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404, got %d", w.Code)
		}
	})
	t.Run("未認証は401", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 0)
		w := doDocReq(r, http.MethodDelete, "/documents/"+testDocUUID, "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
}

func Test_文書ハンドラ_一覧(t *testing.T) {
	t.Run("所有者は200で配列を返し doc本体を含まない", func(t *testing.T) {
		repo := &fakeDocRepo{listDocs: []domain.RichDocument{
			{ID: testDocUUID, OwnerID: 7, Kind: domain.DocumentKindNote, Title: "メモA", Revision: 2, Doc: testDocBody},
			{ID: "aaaaaaaa-297e-8057-884b-c05dbdf9fa53", OwnerID: 7, Kind: domain.DocumentKindNote, Title: "メモB", Revision: 1, Doc: testDocBody},
		}}
		r := newDocRouter(repo, 7)
		w := doDocReq(r, http.MethodGet, "/documents", "")
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if !strings.Contains(body, "メモA") || !strings.Contains(body, "メモB") {
			t.Fatalf("titles missing: %s", body)
		}
		// 一覧サマリは doc 本体を含めない。
		if strings.Contains(body, `"doc"`) {
			t.Fatalf("summary should not contain doc body: %s", body)
		}
	})
	t.Run("空でも200で空配列", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{listDocs: nil}, 7)
		w := doDocReq(r, http.MethodGet, "/documents", "")
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if strings.TrimSpace(w.Body.String()) != "[]" {
			t.Fatalf("want empty array, got %s", w.Body.String())
		}
	})
	t.Run("未認証は401", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 0)
		w := doDocReq(r, http.MethodGet, "/documents", "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("不正なkindは400", func(t *testing.T) {
		r := newDocRouter(&fakeDocRepo{}, 7)
		w := doDocReq(r, http.MethodGet, "/documents?kind=weird", "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}
