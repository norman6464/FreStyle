package handler

import (
	"net/http"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// newTMHandler は no-op fake repo で組んだ TeachingMaterialHandler を返す。
// 成功パスは List（materials.ListByCompany が空を返す）のみ、他は actorContext /
// param / bind のガード分岐を検証する（usecase の深い分岐は結合テスト側）。
func newTMHandler() *TeachingMaterialHandler {
	return NewTeachingMaterialHandler(usecase.NewTeachingMaterialUseCase(fakeMaterialRepo{}, &fakeCourseRepo{}))
}

func TestTeachingMaterialHandler_List(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, nil)
		newTMHandler().List(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", nil, superAdminCo())
		newTMHandler().List(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
}

func TestTeachingMaterialHandler_ListByCourse(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("1"), nil)
		newTMHandler().ListByCourse(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("不正な id → 400", func(t *testing.T) {
		w, c := ctxJSON(http.MethodGet, "", idParam("abc"), superAdminCo())
		newTMHandler().ListByCourse(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func TestTeachingMaterialHandler_Get_BadID(t *testing.T) {
	w, c := ctxJSON(http.MethodGet, "", idParam("abc"), superAdminCo())
	newTMHandler().Get(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestTeachingMaterialHandler_Create_BadJSON(t *testing.T) {
	w, c := ctxJSON(http.MethodPost, `{`, nil, superAdminCo())
	newTMHandler().Create(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestTeachingMaterialHandler_Update_BadID(t *testing.T) {
	w, c := ctxJSON(http.MethodPut, `{"title":"X"}`, idParam("abc"), superAdminCo())
	newTMHandler().Update(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestTeachingMaterialHandler_Delete_BadID(t *testing.T) {
	w, c := ctxJSON(http.MethodDelete, "", idParam("abc"), superAdminCo())
	newTMHandler().Delete(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
