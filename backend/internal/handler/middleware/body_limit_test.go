package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_MaxRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func(maxBytes int64) *gin.Engine {
		r := gin.New()
		r.Use(MaxRequestBody(maxBytes))
		r.POST("/x", func(c *gin.Context) {
			b := new(bytes.Buffer)
			if _, err := b.ReadFrom(c.Request.Body); err != nil {
				c.String(http.StatusRequestEntityTooLarge, "too large")
				return
			}
			c.String(http.StatusOK, "%d", b.Len())
		})
		return r
	}

	t.Run("上限以下の本文は読み切れる", func(t *testing.T) {
		r := newRouter(10)
		req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader([]byte("12345")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK || w.Body.String() != "5" {
			t.Fatalf("got status=%d body=%q", w.Code, w.Body.String())
		}
	})

	t.Run("上限を超える本文は読み込み時にエラーになる", func(t *testing.T) {
		r := newRouter(5)
		req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader([]byte("123456789012")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("上限超過は読み込みエラーになるはず: status=%d body=%q", w.Code, w.Body.String())
		}
	})
}
