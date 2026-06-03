package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader は採番したリクエスト ID を載せるレスポンスヘッダ名。
// 障害調査時にユーザーから受け取ったこの ID でログを一意に追える。
const RequestIDHeader = "X-Request-ID"

// ContextKeyRequestID は c.Get で request_id を取り出すキー。
const ContextKeyRequestID = "requestID"

// RequestLogger は 1 リクエストごとに request_id を採番し、完了時に
// method / path / status / latency_ms / client_ip / user_id を構造化ログ(slog)で出力する。
// status に応じてレベルを変える(5xx=Error / 4xx=Warn / それ以外=Info)ので、
// アラートやダッシュボードでエラー率を拾いやすい。
//
// skipPaths はアクセスログを出さないルートパターン(ヘルスチェック等のノイズ抑制)。
// gin の c.FullPath()(例: /api/v2/health)で一致判定する。
func RequestLogger(skipPaths ...string) gin.HandlerFunc {
	skip := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skip[p] = struct{}{}
	}
	return func(c *gin.Context) {
		// 既存の X-Request-ID を尊重し(フロント/プロキシ採番)、無ければ生成する。
		rid := c.GetHeader(RequestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(ContextKeyRequestID, rid)
		c.Header(RequestIDHeader, rid)

		start := time.Now()
		c.Next()

		if _, ok := skip[c.FullPath()]; ok {
			return
		}

		attrs := []any{
			slog.String("request_id", rid),
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
		}
		if uid := CurrentUserIDOrZero(c); uid != 0 {
			attrs = append(attrs, slog.Uint64("user_id", uid))
		}

		switch status := c.Writer.Status(); {
		case status >= 500:
			slog.Error("request", attrs...)
		case status >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	}
}
