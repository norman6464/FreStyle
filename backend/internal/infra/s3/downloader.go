package s3

import (
	"context"
	"fmt"
	"io"

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
}

// NewDownloader は本番用 (ECS Task Role 経由) で Downloader を組み立てる。
func NewDownloader(ctx context.Context, region, bucket string) (*Downloader, error) {
	if bucket == "" {
		return nil, fmt.Errorf("s3: bucket name is required")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("s3: load aws config: %w", err)
	}
	return &Downloader{
		client: awss3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

// Download は指定 key のオブジェクトをメモリに読み込んで返す。
// Bedrock の image / document ブロックがバイト列を要求するため、ストリーミングではなく
// 一括ロードで返す。サイズ上限は呼び出し側で事前に確認する前提。
func (d *Downloader) Download(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("s3: key is required")
	}
	out, err := d.client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3: get object %q: %w", key, err)
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}
