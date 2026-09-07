package persistence

import (
	"context"
	"fmt"
	"time"
)

// s3Presigner は infra/s3.Presigner と同等の minimal interface（persistence が infra/s3 に
// 直接依存しないよう依存方向を反転する）。4 つの presigner（rich-text 画像 / profile 画像 /
// KB ページ画像 / KB カバー）が共通で使う。
type s3Presigner interface {
	PresignPut(ctx context.Context, key, contentType string, contentLength int64) (url string, ttl time.Duration, err error)
	PresignGet(ctx context.Context, key string) (url string, ttl time.Duration, err error)
}

// stubPresigner は presigner 共通の test / dev 用 stub（本番では infra/s3.NewPresigner が s3Presigner を満たす）。
type stubPresigner struct{ bucket string }

func (s *stubPresigner) PresignPut(_ context.Context, key, _ string, _ int64) (string, time.Duration, error) {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?X-Amz-Stub=1", s.bucket, key), 10 * time.Minute, nil
}

func (s *stubPresigner) PresignGet(_ context.Context, key string) (string, time.Duration, error) {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?X-Amz-Stub=1&X-Amz-Mode=get", s.bucket, key), 10 * time.Minute, nil
}
