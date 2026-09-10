package domain

import (
	"errors"
	"strings"
	"time"
)

// TicketAttachment はチケットに添付したファイル 1 件のメタデータ（段 4・設計 Ⅵ）。
// 本体は Cloud Storage（Key で指す）。論理削除は持たない（削除は本当に消える）。
type TicketAttachment struct {
	ID               string    `json:"id"`
	WorkspaceID      string    `json:"-"`
	TicketID         string    `json:"ticketId"`
	Key              string    `json:"-"`
	Filename         string    `json:"filename"`
	ContentType      string    `json:"contentType"`
	SizeBytes        int64     `json:"sizeBytes"`
	UploadedByUserID uint64    `json:"uploadedByUserId"`
	CreatedAt        time.Time `json:"createdAt"`
}

// MaxAttachmentFilenameBytes はファイル名の上限（character varying(255) の列幅に合わせる）。
const MaxAttachmentFilenameBytes = 255

var ErrInvalidAttachmentFilename = errors.New("domain: invalid attachment filename")

// ValidateAttachmentFilename はファイル名として保存してよい形かを見る
// （空でない・トリム済み・上限バイト数以内。パスとして解釈しない — 実際の保存先は
// Key が指す別のランダムなオブジェクトキーなので、ファイル名は表示専用のメタデータ）。
func ValidateAttachmentFilename(name string) error {
	if name == "" || name != strings.TrimSpace(name) || len(name) > MaxAttachmentFilenameBytes {
		return ErrInvalidAttachmentFilename
	}
	return nil
}
