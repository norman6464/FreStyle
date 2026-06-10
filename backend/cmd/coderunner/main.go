// Command coderunner は学習者コード（php / go / bash）をサンドボックス実行する
// 軽量 HTTP サーバ。backend 本体（cmd/server）から HTTP 越しに呼ばれるサイドカーとして動く。
//
// backend 本体イメージを distroless（go/php を含まない）まで絞れるよう、go/php ランタイムは
// この runner イメージ側にだけ同梱する。runner には DB / Cognito / AWS の機密 env を注入しない。
//
// エンドポイント:
//
//	POST /run     { code, language, stdin } -> { stdout, stderr, exitCode }
//	POST /warmup  { language }              -> { ready }
//	GET  /healthz                           -> 200 OK
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/sandbox"
)

func main() {
	port := os.Getenv("CODE_RUNNER_PORT")
	if port == "" {
		port = "9000"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: newMux(sandbox.NewRunner()),
		// Slowloris 対策にヘッダ読み取りの上限だけ設定する。実行（go run のコンパイル）は
		// 長くなりうるため body 読み取り後の処理時間に WriteTimeout は掛けない。
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("code-runner starting", "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("code-runner stopped", "error", err)
		os.Exit(1)
	}
}

// newMux は runner を HTTP に載せた ServeMux を返す（テストで httptest から使う）。
func newMux(runner *sandbox.Runner) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /run", func(w http.ResponseWriter, r *http.Request) {
		var in domain.CodeExecutionInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		out, err := runner.Run(r.Context(), in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, out)
	})

	mux.HandleFunc("POST /warmup", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Language string `json:"language"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if err := runner.Warmup(r.Context(), req.Language); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]bool{"ready": true})
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
