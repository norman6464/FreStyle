package domain

import (
	"errors"
	"strings"
	"time"
)

// ErrPageVersionNotFound は対象の版が存在しない（無い seq・別ページ・別テナントのもの）ときに返す。
var ErrPageVersionNotFound = errors.New("page version not found")

// ErrInvalidPageVersionNote は版に添えるメモが保存してよい形でないときに返す
// （PageVersionNoteMaxBytes バイトを超えるもの）。
var ErrInvalidPageVersionNote = errors.New("invalid page version note")

// PageVersionNoteMaxBytes は版のメモに許す最大バイト数（TrimSpace 後）。
const PageVersionNoteMaxBytes = 500

// PageVersion はページ本文（doc）の 1 時点のスナップショット。page_snapshots が
// 「1 ページ 1 行の読み取りキャッシュ」なのに対し、こちらは「複数行が積み上がる履歴」。
// PK = (page_id, seq)。同じページの中で seq は 1 始まりの通し番号。
type PageVersion struct {
	PageID string `json:"pageId"`
	Seq    int64  `json:"seq"`
	// Doc は tiptap の getJSON() 相当（type='doc' の ProseMirror ドキュメント）。API へは
	// handler の response 型で json.RawMessage に変換して出す。
	Doc string `json:"-"`
	// AuthorUserID はこの版を作った人（users.id）。API へは著者名を解決した応答形で出す。
	AuthorUserID uint64 `json:"-"`
	// Note は版に添えた任意のメモ。空文字・空白のみは保存しない（ValidateVersionNote が
	// nil へ正規化する）。
	Note *string `json:"note,omitempty"`
	// CreatedAt はこの版を作った時刻。10 分規則・30 日掃除の判定にも使う。
	CreatedAt time.Time `json:"createdAt"`
}

// ValidateVersionNote は版に添えるメモが保存してよい形かを検証し、保存用に正規化した値を返す。
//
//   - nil はそのまま nil を返す（メモ無し）
//   - 空文字・空白のみは nil へ正規化して返す（「メモを入力しないで送る」操作をエラーにしない）
//   - TrimSpace 後 PageVersionNoteMaxBytes バイトを超えるものは ErrInvalidPageVersionNote
func ValidateVersionNote(note *string) (*string, error) {
	if note == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*note)
	if trimmed == "" {
		return nil, nil
	}
	if len(trimmed) > PageVersionNoteMaxBytes {
		return nil, ErrInvalidPageVersionNote
	}
	return &trimmed, nil
}
