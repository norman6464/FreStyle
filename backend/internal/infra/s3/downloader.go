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

// Downloader はサーバ側で S3 オブジェクトの実体を GetObject で取得する
// （presigned PUT 済みの添付を Bedrock に渡す前にバイト列で読む用途）。
type Downloader struct {
	client *awss3.Client
	bucket string
	// allowedKeyPrefix は読み出し許容 prefix（空ならチェックなし）。他用途の prefix 読み出しを防ぐ。
	allowedKeyPrefix string
	// maxBytes は 1 オブジェクトの最大読み込みサイズ（OOM 防止）。
	maxBytes int64
}

// defaultMaxDownloadBytes は 1 オブジェクトの上限（Bedrock 上限より余裕を持たせた 10MB）。
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

// Download は指定 key のオブジェクトをメモリに一括ロードして返す。
// prefix 不一致と maxBytes 超過は error で弾く（超過判定は +1 byte 読みで行う）。
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
