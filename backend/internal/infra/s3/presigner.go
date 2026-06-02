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

// PresignPut は指定 key への PutObject presigned URL を返す。
// contentType は presign に焼き込まれるため PUT 時のヘッダと完全一致が必要（不一致だと SignatureDoesNotMatch）。
func (p *Presigner) PresignPut(ctx context.Context, key, contentType string) (string, time.Duration, error) {
	if key == "" {
		return "", 0, fmt.Errorf("s3: key is required")
	}
	req, err := p.client.PresignPutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, awss3.WithPresignExpires(p.ttl))
	if err != nil {
		return "", 0, fmt.Errorf("s3: presign put: %w", err)
	}
	return req.URL, p.ttl, nil
}
