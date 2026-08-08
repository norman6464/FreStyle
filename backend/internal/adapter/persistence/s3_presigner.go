package persistence

import (
	"context"
	"fmt"
	"time"
)

// noContentLengthConstraint は PresignPut に Content-Length を署名させないことを表す。
// サイズ検証を持たない呼び出し元（AI チャット添付 / プロフィール画像）が渡す。
const noContentLengthConstraint int64 = 0

// unsetUserID は current user を解決できなかったことを表す番兵値。3 つの presigner が共通で使う。
const unsetUserID uint64 = 0

// s3Presigner は infra/s3.Presigner と同等の minimal interface（persistence が infra/s3 に
// 直接依存しないよう依存方向を反転する）。3 つの presigner が共通で使う。
//
// sizeBytes > 0 のとき Content-Length が署名対象に入り、超過アップロードを S3 側で拒否させる。
// 上限値は呼び出し元が決める（ノート画像は 5MB = 5242880 byte。AI チャット添付とプロフィール画像は
// noContentLengthConstraint を渡すため、この経路では上限を署名しない）。
type s3Presigner interface {
	PresignPut(ctx context.Context, key, contentType string, sizeBytes int64) (url string, ttl time.Duration, err error)
}

// stubPresignTTL は stub が返す有効期限。本番の infra/s3.defaultPresignTTL と同値に揃えている。
const stubPresignTTL = 10 * time.Minute

// stubPresignURLFormat は stub が返す URL の組み立て（bucket, key の順）。実在しないダミー URL。
const stubPresignURLFormat = "https://%s.s3.amazonaws.com/%s?X-Amz-Stub=1"

// stubPresigner は presigner 共通の test / dev 用 stub（本番では infra/s3.NewPresigner が s3Presigner を満たす）。
type stubPresigner struct{ bucket string }

func (s *stubPresigner) PresignPut(_ context.Context, key, _ string, _ int64) (string, time.Duration, error) {
	return fmt.Sprintf(stubPresignURLFormat, s.bucket, key), stubPresignTTL, nil
}
