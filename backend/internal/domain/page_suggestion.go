package domain

import (
	"errors"
	"time"
)

// PageSuggestionStatus は提案の解決状態。
type PageSuggestionStatus string

const (
	// PageSuggestionStatusOpen はまだ採用も却下もされていない状態。
	PageSuggestionStatusOpen PageSuggestionStatus = "open"
	// PageSuggestionStatusAccepted は採用され、本文へ反映済みの状態。
	PageSuggestionStatusAccepted PageSuggestionStatus = "accepted"
	// PageSuggestionStatusRejected は却下され、本文には反映しないと決まった状態。
	PageSuggestionStatusRejected PageSuggestionStatus = "rejected"
)

// ErrPageSuggestionNotFound は対象の提案が存在しない（無い ID・別ページ・別テナントのもの）
// ときに返す。
var ErrPageSuggestionNotFound = errors.New("page suggestion not found")

// ErrPageSuggestionAlreadyResolved は既に採用・却下済みの提案を解決しようとしたときに返す
// （2 人が同時に採用・却下を叩いても片方しか成功しない、の判定結果）。
var ErrPageSuggestionAlreadyResolved = errors.New("page suggestion already resolved")

// ErrPageSuggestionStale は提案の採用時に返す — 提案した時点（BaseSeq）より後に、
// 別の編集でそのページの版が進んでいる状態。差分は BaseSeq の内容を基準に計算されているため、
// そのまま採用すると提案作成後の編集を黙って巻き戻してしまう。採用者は最新の本文を見直してから
// 却下するか、提案自体を作り直してもらう必要がある。
var ErrPageSuggestionStale = errors.New("page suggestion is stale: the page changed after the suggestion was created")

// PageSuggestion は commenter（閲覧+コメントはできるが編集はできない役割）が本文を書き換えた
// ときに、blocks を直接更新する代わりに積む 1 件。editor 以上が採用すれば通常の保存経路
// （ReplacePageBlocksUseCase）へ渡り本文へ反映される。差分の表示は BaseSeq が指す
// page_versions.doc と Doc を突き合わせて行う（差分形式そのものはここに持ち込まない）。
type PageSuggestion struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	PageID      string `json:"pageId"`
	// BaseSeq は提案した時点のそのページの最新版（page_versions.seq）。まだ版が 1 つも無い
	// ページへの提案は nil のまま。
	BaseSeq *int64 `json:"baseSeq,omitempty"`
	// Doc は提案後の本文全体（ProseMirror doc）。API へは handler の response 型で
	// json.RawMessage に変換して出す（domain.PageVersion.Doc と同じ方針）。
	Doc    string               `json:"-"`
	Status PageSuggestionStatus `json:"status"`
	// AuthorUserID は提案した人（users.id）。API へは著者名を解決した応答形で出す
	// （domain.PageVersion.AuthorUserID と同じ役割分担）。
	AuthorUserID uint64    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	// ResolvedAt / ResolvedByUserID は採用・却下されるまで両方 nil。
	ResolvedAt *time.Time `json:"resolvedAt,omitempty"`
	// ResolvedByUserID も AuthorUserID と同じく、API へは解決者名を解決した応答形で出す。
	ResolvedByUserID *uint64 `json:"-"`
}
