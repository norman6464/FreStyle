package domain

import (
	"time"
	"unicode"
	"unicode/utf8"
)

// Page はナレッジの 1 ページ。ページ同士は ParentID で木構造をなす（無限入れ子）。
//
// 兄弟の並び順は整数の連番ではなく分数インデックス（internal/pkg/fracindex が採番する文字列キー）で
// 持つ。1 行動かすたびに後続を振り直す UPDATE を避けるため。順序の比較は「同じ親の中」でのみ意味を持つ。
// DB 側は position 列を COLLATE "C" に固定し、Go のバイト比較と ORDER BY を一致させる。
type Page struct {
	ID string `json:"id"`
	// WorkspaceID はテナント境界。space / 親ページとの複合 FK に使い、テナント越えの親子を DB が弾く。
	WorkspaceID string `json:"workspaceId"`
	// SpaceID は所属スペース。(workspace_id, space_id) の複合 FK で spaces を参照する。
	SpaceID string `json:"spaceId"`
	// ParentID は親ページ。NULL はスペース直下（ルート）を意味する。
	ParentID *string `json:"parentId,omitempty"`
	// Position は兄弟内の並び順を表す分数インデックス（fracindex.Between で採番する）。
	Position string `json:"position"`
	// Title は一覧・検索・パンくずに使うページ名。
	Title string `json:"title"`
	// CreatedByUserID は作成者（users.id）。
	CreatedByUserID uint64 `json:"createdByUserId"`
	// ArchivedAt はアーカイブ日時。NULL が現役。物理削除ではなくアーカイブで隠す運用のため、
	// 一意制約（同じ親の中で position が重複しない）はアーカイブ済みを除外した部分ユニークで張る。
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	// Icon はページの顔（絵文字のみ、いまのところ）。NULL は「付けていない」。
	Icon *PageIcon `json:"icon,omitempty"`
	// Cover はページ頭部のカバー画像。設定 API は 1b で追加するため、いまは
	// 常に nil（列との往復だけを domain 側に用意しておく）。
	Cover *PageCover `json:"cover,omitempty"`
	// LastEditedByUserID は最終編集者（users.id）。NULL は「作成後まだ誰も本文を
	// 保存していない」。本文の保存経路（ReplacePageBlocksUseCase）だけが書く。
	LastEditedByUserID *uint64 `json:"lastEditedByUserId,omitempty"`
}

// PageIconType はページアイコンの種類。いまのところ絵文字だけを許す
// （アップロード画像・外部 URL は要件に無く、whitelist を広げるのは要る時でよい）。
type PageIconType string

// PageIconTypeEmoji は唯一許可される種類。
const PageIconTypeEmoji PageIconType = "emoji"

// PageIcon はページの顔。保存形は pages.icon の jsonb
// （例: {"type":"emoji","value":"📘"}）。
type PageIcon struct {
	Type  PageIconType `json:"type"`
	Value string       `json:"value"`
}

// pageIconValueMaxBytes / pageIconValueMaxRunes は Value に許す上限。
// 絵文字 1 つは UTF-8 で最大 4 byte、ZWJ で繋いだ複合絵文字（家族・肌色修飾）でも
// 数個の code point に収まるため、この桁で「見た目 1 つ」を大きく外れる値は弾ける。
const (
	pageIconValueMaxBytes = 64
	pageIconValueMaxRunes = 16
)

// Valid は保存してよい形かを返す。
//
// 見るのは「種類の whitelist・空でないこと・正しい UTF-8・長さの上限・空白と制御文字を
// 含まないこと」まで。**「厳密に 1 grapheme か」は見ない** — 判定には
// unicode/text segmentation の依存が要り、見た目の一意性はピッカー側の選択肢
// （厳選した絵文字の格子 + countGraphemes による自由入力の絞り込み）が担うため、
// ここでは持たない。
func (i PageIcon) Valid() bool {
	if i.Type != PageIconTypeEmoji {
		return false
	}
	if i.Value == "" || len(i.Value) > pageIconValueMaxBytes {
		return false
	}
	if !utf8.ValidString(i.Value) {
		return false
	}
	if utf8.RuneCountInString(i.Value) > pageIconValueMaxRunes {
		return false
	}
	for _, r := range i.Value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// PageCoverType はページカバーの種類。いまのところアップロード画像（S3 key）のみ。
type PageCoverType string

// PageCoverTypeFile はアップロード画像を指す種類（保存形は S3 key）。
const PageCoverTypeFile PageCoverType = "file"

// PageCover はページ頭部のカバー画像。保存形は pages.cover の jsonb。
// 設定・解除の API は段 1b で追加する。ここでは列との往復（読み出し）だけを担う。
type PageCover struct {
	Type PageCoverType `json:"type"`
	Key  string        `json:"key"`
}
