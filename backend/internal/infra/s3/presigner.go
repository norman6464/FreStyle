// Package s3 は AWS S3 への PutObject presigned URL 発行を担当する Infra 層。
//
// 設計:
//   - aws-sdk-go-v2 の s3.PresignClient を使う
//   - 認証は default chain (ECS Task Role / EC2 Instance Role / 環境変数 / ~/.aws/credentials)
//   - 期限は 10 分 (presigned URL は短期で十分)
//   - bucket / region / 任意の CDN base URL は config から渡す
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
// クライアント (フロントの axios) は HTTP PUT で同 URL に Content-Type ヘッダ付き
// で body を送れば S3 が直接受け取る。
//
// contentType は presign に焼き込まれるため、PUT 時のヘッダと **完全一致** する必要がある
// （食い違うと S3 が SignatureDoesNotMatch を返す）。
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
