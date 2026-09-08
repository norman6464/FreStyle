// Package gcs は Cloud Storage への PUT presigned URL 発行と GET presigned URL による
// ダウンロードを担う Infra 層。
//
// 署名は秘密鍵ファイルを使わない。IAM Credentials API の signBlob RPC で行う
// （Cloud Run のランタイムサービスアカウントに roles/iam.serviceAccountTokenCreator を
// 自分自身に対して付与し、roles/storage.objectAdmin を対象バケットに付与しておくことが
// 前提。インフラ側で適用済み）。V4 署名の期限は 10 分。
package gcs

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
	credentials "cloud.google.com/go/iam/credentials/apiv1"
	"cloud.google.com/go/iam/credentials/apiv1/credentialspb"
	"cloud.google.com/go/storage"
)

const defaultPresignTTL = 10 * time.Minute

// signedURLIssuer は storage.BucketHandle.SignedURL のうち Presigner が使う部分だけを
// 切り出した interface。テストで実 GCS に繋がずに URL 組み立てロジックだけを検証するため。
type signedURLIssuer interface {
	SignedURL(object string, opts *storage.SignedURLOptions) (string, error)
}

// Presigner は GCS の V4 signed URL（PUT / GET）を発行する。
type Presigner struct {
	bucket    signedURLIssuer
	closeFn   func() error
	accessID  string
	signBytes func(ctx context.Context, b []byte) ([]byte, error)
	ttl       time.Duration
}

// NewPresigner は Cloud Run にアタッチされたサービスアカウント（ADC）を使って Presigner を
// 組み立てる。秘密鍵ファイルは一切使わない。
func NewPresigner(ctx context.Context, bucketName string) (*Presigner, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("gcs: bucket name is required")
	}

	// ランタイムサービスアカウントのメールアドレスをメタデータサーバーから取得する
	// （ハードコードしない。signBlob の呼び出し先「自分自身」を特定するために要る）。
	saEmail, err := metadata.EmailWithContext(ctx, "default")
	if err != nil {
		return nil, fmt.Errorf("gcs: get service account email from metadata: %w", err)
	}

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcs: new storage client: %w", err)
	}

	iamClient, err := credentials.NewIamCredentialsClient(ctx)
	if err != nil {
		_ = storageClient.Close()
		return nil, fmt.Errorf("gcs: new iam credentials client: %w", err)
	}

	// "-" はプロジェクトの代わりに置く固定のワイルドカード。プロジェクト ID を
	// 入れると invalid になる（IAM Credentials API の SignBlobRequest.Name の仕様）。
	saResource := "projects/-/serviceAccounts/" + saEmail

	return &Presigner{
		bucket: storageClient.Bucket(bucketName),
		closeFn: func() error {
			err1 := iamClient.Close()
			err2 := storageClient.Close()
			if err1 != nil {
				return err1
			}
			return err2
		},
		accessID: saEmail,
		signBytes: func(ctx context.Context, b []byte) ([]byte, error) {
			resp, err := iamClient.SignBlob(ctx, &credentialspb.SignBlobRequest{
				Name:    saResource,
				Payload: b,
			})
			if err != nil {
				return nil, fmt.Errorf("gcs: iam signBlob: %w", err)
			}
			return resp.SignedBlob, nil
		},
		ttl: defaultPresignTTL,
	}, nil
}

// Close は IamCredentialsClient / storage.Client が保持する gRPC コネクションを解放する。
// AWS SDK v2 ベースだった旧 infra/s3.Presigner と異なり、こちらは明示的な Close が要る。
func (p *Presigner) Close() error {
	if p.closeFn == nil {
		return nil
	}
	return p.closeFn()
}

// PresignPut は指定 key への PUT アップロード用 V4 signed URL を返す。
// contentType は署名に焼き込まれるため PUT 時のヘッダと完全一致が必要
// （不一致だと GCS 側で署名不一致エラーになる。旧 S3 実装と同じ制約）。
//
// contentLength は 0 より大きいときだけ Content-Length を署名対象ヘッダとして焼き込む。
// **GCS の V4 signed URL には S3 の content-length-range のような「範囲」制約を表す
// 仕組みが無い**（x-goog-content-length-range は POST Policy V4 専用の条件で、
// signed URL には適用されない）。ここでは Content-Length ヘッダそのものを署名対象に
// 含めることで「その値と完全一致しない PUT は拒否される」という厳密一致の制約をかけている
// （範囲ではなく一致）。旧 S3 実装（ContentLength を signed request に含める）と
// 意味的に同じで、呼び出し側は事前に検証済みの実サイズをそのまま渡す前提。
// 0 は「サイズを制約しない」呼び出し側（profile 画像等）をそのまま動かすための値。
func (p *Presigner) PresignPut(ctx context.Context, key, contentType string, contentLength int64) (string, time.Duration, error) {
	if key == "" {
		return "", 0, fmt.Errorf("gcs: key is required")
	}

	opts := &storage.SignedURLOptions{
		GoogleAccessID: p.accessID,
		SignBytes: func(b []byte) ([]byte, error) {
			return p.signBytes(ctx, b)
		},
		Scheme:      storage.SigningSchemeV4,
		Method:      http.MethodPut,
		Expires:     time.Now().Add(p.ttl),
		ContentType: contentType,
	}
	if contentLength > 0 {
		opts.Headers = append(opts.Headers, fmt.Sprintf("Content-Length:%d", contentLength))
	}

	url, err := p.bucket.SignedURL(key, opts)
	if err != nil {
		return "", 0, fmt.Errorf("gcs: presign put: %w", err)
	}
	return url, p.ttl, nil
}

// PresignGet は指定 key からの GET ダウンロード用 V4 signed URL を返す。
func (p *Presigner) PresignGet(ctx context.Context, key string) (string, time.Duration, error) {
	if key == "" {
		return "", 0, fmt.Errorf("gcs: key is required")
	}

	opts := &storage.SignedURLOptions{
		GoogleAccessID: p.accessID,
		SignBytes: func(b []byte) ([]byte, error) {
			return p.signBytes(ctx, b)
		},
		Scheme:  storage.SigningSchemeV4,
		Method:  http.MethodGet,
		Expires: time.Now().Add(p.ttl),
	}

	url, err := p.bucket.SignedURL(key, opts)
	if err != nil {
		return "", 0, fmt.Errorf("gcs: presign get: %w", err)
	}
	return url, p.ttl, nil
}
