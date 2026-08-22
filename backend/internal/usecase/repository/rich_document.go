package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrRichDocumentNotFound は対象文書が存在しない（または論理削除済み）ときに返す。
var ErrRichDocumentNotFound = errors.New("rich document not found")

// ErrRichDocumentConflict は楽観ロックの版番号が一致せず更新できなかったときに返す。
// 呼び出し側はこれを 409 Conflict に対応づける。
var ErrRichDocumentConflict = errors.New("rich document revision conflict")

// ErrRichDocumentInvalidData は doc / title が DB の型（jsonb / text）に格納できない値
// （U+0000 や不正サロゲート等）だったときに返す。呼び出し側は 400 に対応づける。
var ErrRichDocumentInvalidData = errors.New("rich document invalid data")

// RichDocumentRepository は rich_documents テーブルへのアクセスを提供する。
type RichDocumentRepository interface {
	// Create は新規文書を作成する。ID 未設定なら UUID を採番して doc.ID に反映する。
	Create(ctx context.Context, doc *domain.RichDocument) error
	// FindByID は ID で 1 件引く（論理削除は除外）。無ければ ErrRichDocumentNotFound。
	FindByID(ctx context.Context, id string) (*domain.RichDocument, error)
	// UpdateWithRevision は楽観ロック付き更新。expectedRevision が現在値と一致する行だけを
	// 更新し revision を +1 する。行が無ければ ErrRichDocumentNotFound、版不一致なら
	// ErrRichDocumentConflict を返す。成功時は doc に更新後の値（revision など）を反映する。
	UpdateWithRevision(ctx context.Context, doc *domain.RichDocument, expectedRevision int) error
	// SoftDelete は owner を条件に論理削除する（他人の文書は消せない）。
	// 対象が無ければ ErrRichDocumentNotFound。
	SoftDelete(ctx context.Context, id string, ownerID uint64) error
}
