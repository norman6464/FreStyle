package persistence

import (
	"context"
	"fmt"
	"time"
)

// s3Presigner は infra/s3.Presigner と同等の interface を minimal に切り出した型。
// persistence パッケージ が infra/s3 パッケージ に 直接 依存 しない よう dep direction
// を 反転 (依存性 逆転 原則)。 profile_image / note_image / ai_chat_attachment の
// 3 つ の presigner が 共通 で 使う。
type s3Presigner interface {
	PresignPut(ctx context.Context, key, contentType string) (url string, ttl time.Duration, err error)
}

// stubPresigner は profile_image / note_image / ai_chat_attachment の test / dev 用
// に 共有 さ れる stub 実装。 本番 経路 で は infra/s3.NewPresigner が s3Presigner を
// 満たす。
type stubPresigner struct{ bucket string }

func (s *stubPresigner) PresignPut(_ context.Context, key, _ string) (string, time.Duration, error) {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?X-Amz-Stub=1", s.bucket, key), 10 * time.Minute, nil
}
