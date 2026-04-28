package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
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

func TestProfileResolveUserID_MeKeyword(t *testing.T) {
	h := &ProfileHandler{}
	uid, err := h.resolveUserID(makeCtx(7, "me"))
	if err != nil || uid != 7 {
		t.Fatalf("'me' should resolve to current user; got uid=%d err=%v", uid, err)
	}
}

func TestProfileResolveUserID_EmptyParam(t *testing.T) {
	h := &ProfileHandler{}
	uid, err := h.resolveUserID(makeCtx(7, ""))
	if err != nil || uid != 7 {
		t.Fatalf("empty param should resolve to current user; got uid=%d err=%v", uid, err)
	}
}

func TestProfileResolveUserID_MatchingNumeric(t *testing.T) {
	h := &ProfileHandler{}
	uid, err := h.resolveUserID(makeCtx(7, "7"))
	if err != nil || uid != 7 {
		t.Fatalf("matching numeric should pass; got uid=%d err=%v", uid, err)
	}
}

func TestProfileResolveUserID_MismatchNumericIsForbidden(t *testing.T) {
	h := &ProfileHandler{}
	if _, err := h.resolveUserID(makeCtx(7, "99")); err != errProfileForbidden {
		t.Fatalf("mismatch numeric should be forbidden; got %v", err)
	}
}

func TestProfileResolveUserID_NoCurrentUserIsUnauthorized(t *testing.T) {
	h := &ProfileHandler{}
	if _, err := h.resolveUserID(makeCtx(0, "me")); err != errProfileUnauthorized {
		t.Fatalf("no current user should be unauthorized; got %v", err)
	}
}
