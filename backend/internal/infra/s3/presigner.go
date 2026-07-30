// Package s3 は AWS S3 への PutObject presigned URL 発行と GetObject ダウンロードを担う Infra 層。
// 認証は default chain、presigned URL の期限は 10 分。
package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

const defaultPresignTTL = 10 * time.Minute

// Presigner は PutObject presigned URL を生成する。
type Presigner struct {
	client *awss3.PresignClient
	bucket string
	ttl    time.Duration
}

// NewPresigner は本番用 (ECS Task Role 経由) で Presigner を組み立てる。
func NewPresigner(ctx context.Context, region, bucket string) (*Presigner, error) {
	if bucket == "" {
		return nil, fmt.Errorf("s3: bucket name is required")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("s3: load aws config: %w", err)
	}
	return &Presigner{
		client: awss3.NewPresignClient(awss3.NewFromConfig(cfg)),
		bucket: bucket,
		ttl:    defaultPresignTTL,
	}, nil
}

// PresignPut は指定 key への PutObject presigned URL を返す。contentType と、sizeBytes > 0 の場合は
// Content-Length が署名に焼き込まれるため PUT 時のヘッダと完全一致が必要で、異なる種別やサイズで
// 送ると S3 が SignatureDoesNotMatch で拒否する（sizeBytes <= 0 なら Content-Length は署名しない）。
func (p *Presigner) PresignPut(ctx context.Context, key, contentType string, sizeBytes int64) (string, time.Duration, error) {
	if key == "" {
		return "", 0, fmt.Errorf("s3: key is required")
	}
	in := &awss3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}
	if sizeBytes > 0 {
		in.ContentLength = aws.Int64(sizeBytes)
	}
	req, err := p.client.PresignPutObject(ctx, in, awss3.WithPresignExpires(p.ttl))
	if err != nil {
		return "", 0, fmt.Errorf("s3: presign put: %w", err)
	}
	return req.URL, p.ttl, nil
}
