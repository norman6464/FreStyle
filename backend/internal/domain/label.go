package domain

import (
	"errors"
	"time"
)

// Label はスペースごとのラベル（名前 + 色）。段 4・設計 Ⅵ。
//
// 同名は空白・大文字小文字違いも含めてスペース内で作れない（uq_labels_space_name の
// 部分一意。DB 側は既にトリム済みの name しか受け付けない — 呼び出し側が保存前に
// strings.TrimSpace を通す分担は ticket_statuses/ticket_types と同じ）。
type Label struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"-"`
	SpaceID     string    `json:"spaceId"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MaxLabelNameLen は名前の列幅（character varying(64)）。
const MaxLabelNameLen = 64

var (
	ErrInvalidLabelName  = errors.New("domain: invalid label name")
	ErrInvalidLabelColor = errors.New("domain: invalid label color")
)
