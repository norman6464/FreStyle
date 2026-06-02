package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestEmbedHandler_Resolve_MissingURL は url クエリ欠落で 400 を返すことを検証する。
// fetcher 到達前のガードのため fetcher は nil で安全。
func TestEmbedHandler_Resolve_MissingURL(t *testing.T) {
	h := &EmbedHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/embeds/oembed", nil)
	h.Resolve(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
