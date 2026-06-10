// Package coderunner は、別コンテナ（サイドカー）で動く code-runner（cmd/coderunner）への
// HTTP クライアントを提供する。backend 本体イメージから go/php ランタイムを外せるよう、
// コード実行を HTTP 越しに runner へ委譲する（usecase.CodeRunner を満たす）。
package coderunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// Client は code-runner への HTTP クライアント。
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient は runner の baseURL（例 http://127.0.0.1:9000）を受け取りクライアントを返す。
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		// go run のコンパイルを含むため全体タイムアウトは長め。runner 側も言語別に上限を持つ。
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// Run は runner の POST /run にコードを渡して実行結果を得る。
func (c *Client) Run(ctx context.Context, in domain.CodeExecutionInput) (*domain.CodeExecutionResult, error) {
	var out domain.CodeExecutionResult
	if err := c.postJSON(ctx, "/run", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type warmupRequest struct {
	Language string `json:"language"`
}

// Warmup は runner の POST /warmup に言語を渡し実行環境を温める。
func (c *Client) Warmup(ctx context.Context, language string) error {
	return c.postJSON(ctx, "/warmup", warmupRequest{Language: language}, nil)
}

// postJSON は body を JSON で POST し、out が非 nil ならレスポンスを out にデコードする。
func (c *Client) postJSON(ctx context.Context, path string, body, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("リクエストのエンコードに失敗: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("リクエスト生成に失敗: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("code-runner への接続に失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("code-runner が異常応答: status=%d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("code-runner 応答のデコードに失敗: %w", err)
	}
	return nil
}
