package repository

import "context"

// KbImagePresigner はナレッジのページ画像用 presigned URL を発行する。
//
// 認可（閲覧・編集権限、key とページの結びつき）は呼び出し側（usecase）が済ませている前提で、
// ここは署名の発行だけを担う。これは新しい「表」ではなくオブジェクトストレージという
// 外部サービスへの port なので、KnowledgeBaseRepository（pages / blocks への port）には含めない。
type KbImagePresigner interface {
	// PresignUpload は指定 key への PUT 用 presigned URL を返す。
	PresignUpload(ctx context.Context, key, contentType string, size int64) (url string, expiresIn int, err error)
	// PresignDownload は指定 key からの GET（ダウンロード）用 presigned URL を返す。
	PresignDownload(ctx context.Context, key string) (url string, expiresIn int, err error)
}
