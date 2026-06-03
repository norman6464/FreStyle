// Package logging はアプリ全体の構造化ログ(slog / JSON)を初期化する。
// JSON 出力にすることで CloudWatch Logs Insights や将来の APM(NewRelic 等)で
// フィールド(request_id / status / latency_ms / user_id 等)を集計・検索できる。
package logging

import (
	"log/slog"
	"os"
)

// Setup はアプリ全体の slog を JSON ハンドラで初期化し、slog.Default に設定する。
// レベルは本番(production / prod)は Info、それ以外は Debug。
func Setup(env string) {
	level := slog.LevelInfo
	if env != "production" && env != "prod" {
		level = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
