package s3

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// Downloader はサーバ側で S3 オブジェクトの実体を取得する。
//
// 用途: ユーザーが presigned PUT で S3 にアップロードした画像 / ドキュメントを
// バックエンドが Bedrock Converse に渡す前に GetObject でバイト列として取り出す。
// presigner.go と分けているのは「presigned URL は短期 / 一方通行（PUT）専用」と
// 「server-side GetObject は IAM 権限経由 / 任意のサイズ」で責務が違うため。
type Downloader struct {
	client *awss3.Client
	bucket string
	// allowedKeyPrefix は読み出し許容 prefix。空のときは prefix チェックをスキップする。
	// AI チャット添付では `ai-chat/` を渡して、誤って他用途の prefix を読まれないようにする。
	allowedKeyPrefix string
	// maxBytes は 1 オブジェクトあたりの最大読み込みサイズ（バイト）。
	// 任意 S3 オブジェクトを読まれた場合の OOM 防止用。
	maxBytes int64
}

// 既定の最大読み込みサイズ。Bedrock の image 上限 (5MB) と document 上限 (4.5MB) より
// 余裕をもたせて 10MB を 1 オブジェクトの上限とする。実際のサイズ検証は usecase / handler 側で
// 行うが、Downloader でも belt-and-suspenders として強制する。
const defaultMaxDownloadBytes int64 = 10 * 1024 * 1024

// NewDownloader は本番用 (ECS Task Role 経由) で Downloader を組み立てる。
// allowedKeyPrefix は読み出し許容 prefix（空文字は無制限。AI チャット用途では "ai-chat/" 推奨）。
func NewDownloader(ctx context.Context, region, bucket, allowedKeyPrefix string) (*Downloader, error) {
	if bucket == "" {
		return nil, fmt.Errorf("s3: bucket name is required")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("s3: load aws config: %w", err)
	}
	return &Downloader{
		client:           awss3.NewFromConfig(cfg),
		bucket:           bucket,
		allowedKeyPrefix: allowedKeyPrefix,
		maxBytes:         defaultMaxDownloadBytes,
	}, nil
}

// Download は指定 key のオブジェクトをメモリに読み込んで返す。
// Bedrock の image / document ブロックがバイト列を要求するため、ストリーミングではなく
// 一括ロードで返す。
//
// 防御:
//   - allowedKeyPrefix を持つ場合は prefix 不一致を error で弾く
//   - maxBytes を超える object は error にしてメモリを使い切らない
//     （io.LimitReader で +1 byte 余分に読んで超過判定する）
func (d *Downloader) Download(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("s3: key is required")
	}
	if d.allowedKeyPrefix != "" && !strings.HasPrefix(key, d.allowedKeyPrefix) {
		return nil, fmt.Errorf("s3: key %q outside allowed prefix %q", key, d.allowedKeyPrefix)
	}
	out, err := d.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3: get object %q: %w", key, err)
	}
	defer out.Body.Close()

	limit := d.maxBytes
	if limit <= 0 {
		limit = defaultMaxDownloadBytes
	}
	data, err := io.ReadAll(io.LimitReader(out.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("s3: read object %q: %w", key, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("s3: object %q exceeds %d bytes", key, limit)
	}
	return data, nil
}
