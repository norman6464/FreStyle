package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// captureLogs は RequestLogger の出力(slog/JSON)を buffer に差し替えて回収する。
func captureLogs(t *testing.T, run func(r *gin.Engine)) []map[string]any {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	defer slog.SetDefault(prev)

	r := gin.New()
	run(r)

	var out []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(line, &m); err != nil {
			t.Fatalf("ログが JSON でない: %v (%s)", err, line)
		}
		out = append(out, m)
	}
	return out
}

func TestRequestLogger_EmitsStructuredFields(t *testing.T) {
	logs := captureLogs(t, func(r *gin.Engine) {
		r.Use(RequestLogger())
		r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// レスポンスに request_id ヘッダが付与される。
		if w.Header().Get(RequestIDHeader) == "" {
			t.Fatalf("%s ヘッダが無い", RequestIDHeader)
		}
	})

	if len(logs) != 1 {
		t.Fatalf("ログは 1 件のはず, got %d", len(logs))
	}
	got := logs[0]
	if got["msg"] != "request" {
		t.Errorf("msg=request のはず, got %v", got["msg"])
	}
	if got["method"] != http.MethodGet || got["path"] != "/ping" {
		t.Errorf("method/path が不正: %v %v", got["method"], got["path"])
	}
	if got["status"] != float64(http.StatusOK) {
		t.Errorf("status=200 のはず, got %v", got["status"])
	}
	if _, ok := got["request_id"]; !ok {
		t.Errorf("request_id フィールドが無い")
	}
	if _, ok := got["latency_ms"]; !ok {
		t.Errorf("latency_ms フィールドが無い")
	}
}

func TestRequestLogger_SkipPaths(t *testing.T) {
	logs := captureLogs(t, func(r *gin.Engine) {
		r.Use(RequestLogger("/api/v2/health"))
		r.GET("/api/v2/health", func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/api/v2/health", nil)
		r.ServeHTTP(httptest.NewRecorder(), req)
	})
	if len(logs) != 0 {
		t.Fatalf("skip 対象はログを出さないはず, got %d", len(logs))
	}
}

func TestRequestLogger_5xxIsErrorLevel(t *testing.T) {
	logs := captureLogs(t, func(r *gin.Engine) {
		r.Use(RequestLogger())
		r.GET("/boom", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })
		req := httptest.NewRequest(http.MethodGet, "/boom", nil)
		r.ServeHTTP(httptest.NewRecorder(), req)
	})
	if len(logs) != 1 || logs[0]["level"] != "ERROR" {
		t.Fatalf("5xx は ERROR レベルのはず, got %v", logs)
	}
}

func TestRequestLogger_HonorsIncomingRequestID(t *testing.T) {
	const incoming = "abc-123"
	logs := captureLogs(t, func(r *gin.Engine) {
		r.Use(RequestLogger())
		r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set(RequestIDHeader, incoming)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Header().Get(RequestIDHeader) != incoming {
			t.Fatalf("既存の request_id を引き継ぐはず, got %s", w.Header().Get(RequestIDHeader))
		}
	})
	if logs[0]["request_id"] != incoming {
		t.Fatalf("ログの request_id が引き継がれていない: %v", logs[0]["request_id"])
	}
}
