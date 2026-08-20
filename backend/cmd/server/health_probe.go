package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// healthProbeArg は自己ヘルスチェックモードを起動する引数。
// distroless イメージにはシェルも curl も無く、compose / ECS の healthcheck から叩ける
// コマンドがサーバのバイナリ自身しかないため、自分で自分の /health を叩けるようにする。
const healthProbeArg = "-healthcheck"

// isHealthProbe は引数が自己ヘルスチェックモードの起動かを判定する。
func isHealthProbe(args []string) bool {
	return len(args) > 1 && args[1] == healthProbeArg
}

// healthProbeURL は自分自身のヘルスチェック URL を組み立てる。待ち受けポートは
// config と同じ PORT 環境変数に従う（既定 8080）。
func healthProbeURL() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return "http://127.0.0.1:" + port + "/api/v2/health"
}

// probeHealth は url を叩き、2xx なら nil を返す。DB 断のときは health handler が 503 を
// 返すため、ここで非 nil になる（＝コンテナは unhealthy と判定される）。
func probeHealth(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("health check returned %d", resp.StatusCode)
	}
	return nil
}

// runHealthProbe は自己ヘルスチェックを実行し、プロセスの終了コードを返す。
func runHealthProbe() int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := probeHealth(ctx, healthProbeURL()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
