package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestOnboardingHandler_Complete_Unauthorized は current user 無しで 401 を返すことを検証する。
// uid==0 のガードで usecase 到達前に返るため usecase は nil で安全。
func Test_オンボーディングハンドラ_完了_未認証(t *testing.T) {
	h := &OnboardingHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	h.Complete(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}
