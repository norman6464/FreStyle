package ticket

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

// チケット本文（ProseMirror の doc、tickets.doc）を歩く処理。ページ側（page_usecase.go の
// pageRefCollector・StripPageRefTitles）と同じ考え方だが、ページは本文をブロック単位に
// 分解して持つのに対しチケットは 1 本の doc をそのまま jsonb で持つ（設計 Ⅳ-E）ため、
// ここでは分解されていない JSON の木を直接歩く素朴な実装にしている。

// ticketPageRefNodeType / ticketTicketRefNodeType は本文中でページ / 他チケットを指す
// ProseMirror ノードの type 名。pageRef は既存のナレッジ機能と同じノード（attrs.pageId）。
// ticketRef はチケット機能で新設するノードで、同じ形に揃える（attrs.ticketId）。
const (
	ticketPageRefNodeType   = "pageRef"
	ticketTicketRefNodeType = "ticketRef"
)

// ExtractDocRefs は本文中の pageRef / ticketRef ノードから参照先 ID を、文書順・
// 重複なしで集める（domain の役目ではなく、DB を読まない純粋な木の走査なのでここに置く）。
// 不正な UUID・空文字は黙って無視する（保存を落とす理由にはしない。実在確認は
// repository.ReplaceTicketPageLinks/ReplaceTicketTicketLinks が別途行う）。
func ExtractDocRefs(doc []byte) (pageIDs, ticketIDs []string, err error) {
	var root any
	if err := json.Unmarshal(doc, &root); err != nil {
		return nil, nil, err
	}
	seenPage := map[string]struct{}{}
	seenTicket := map[string]struct{}{}
	var walk func(node any)
	walk = func(node any) {
		switch v := node.(type) {
		case map[string]any:
			switch v["type"] {
			case ticketPageRefNodeType:
				if id, ok := canonicalRefID(attrString(v, "pageId")); ok {
					if _, dup := seenPage[id]; !dup {
						seenPage[id] = struct{}{}
						pageIDs = append(pageIDs, id)
					}
				}
			case ticketTicketRefNodeType:
				if id, ok := canonicalRefID(attrString(v, "ticketId")); ok {
					if _, dup := seenTicket[id]; !dup {
						seenTicket[id] = struct{}{}
						ticketIDs = append(ticketIDs, id)
					}
				}
			}
			walk(v["content"])
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(root)
	return pageIDs, ticketIDs, nil
}

// StripDocRefTitles は保存前の doc から pageRef / ticketRef の title を取り除く。
// title は読み手ごとに読み出し時へ解決する派生値で、保存してはいけない（page 側の
// StripPageRefTitles と同じ理由。閲覧できない読み手の画面にまで、保存した人が見えていた
// 題名がそのまま漏れる経路になる）。参照が無ければ入力をそのまま返す。
func StripDocRefTitles(doc []byte) ([]byte, error) {
	var root any
	if err := json.Unmarshal(doc, &root); err != nil {
		return nil, err
	}
	if !stripDocRefTitlesNode(root) {
		return doc, nil
	}
	return json.Marshal(root)
}

func stripDocRefTitlesNode(node any) bool {
	changed := false
	switch v := node.(type) {
	case map[string]any:
		if v["type"] == ticketPageRefNodeType || v["type"] == ticketTicketRefNodeType {
			if attrs, ok := v["attrs"].(map[string]any); ok {
				if _, has := attrs["title"]; has && attrs["title"] != nil {
					attrs["title"] = nil
					changed = true
				}
			}
		}
		if stripDocRefTitlesNode(v["content"]) {
			changed = true
		}
	case []any:
		for _, child := range v {
			if stripDocRefTitlesNode(child) {
				changed = true
			}
		}
	}
	return changed
}

// BuildPlainText は検索用の派生値（tickets.plain_text）を doc から作る。text ノードの
// 内容だけを集め、pageRef / ticketRef の属性（id・title）は一切含めない
// （設計 Ⅳ-E: 検索は plain_text の ILIKE。参照先の題名や id が検索にヒットする理由に
// なってはいけない — 見えないページの題名が検索から漏れる経路を作らないため）。
// 壊れた JSON は空文字を返す（検索に載らないだけで、保存自体は別の検証が守る）。
func BuildPlainText(doc []byte) string {
	var root any
	if err := json.Unmarshal(doc, &root); err != nil {
		return ""
	}
	var parts []string
	var walk func(node any)
	walk = func(node any) {
		switch v := node.(type) {
		case map[string]any:
			if v["type"] == "text" {
				if s, ok := v["text"].(string); ok && s != "" {
					parts = append(parts, s)
				}
			}
			walk(v["content"])
		case []any:
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(root)
	return strings.Join(parts, "\n")
}

// attrString は node["attrs"][key] を文字列として読む（無ければ空文字）。
func attrString(node map[string]any, key string) string {
	attrs, ok := node["attrs"].(map[string]any)
	if !ok {
		return ""
	}
	s, _ := attrs[key].(string)
	return s
}

// canonicalRefID は参照 ID を UUID の正規形へ寄せる（page_usecase.go の
// canonicalPageRefID と同じ役割。ticketRef にも同じ規則を使う）。
func canonicalRefID(id string) (string, bool) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
}

// DocsEqual は 2 つの doc（jsonb の生バイト列）が意味的に同じ木かを比べる。
//
// バイト列そのままの比較はしない。StripDocRefTitles は「変更が無ければ入力をそのまま
// 返す・変更があれば json.Marshal で作り直す」という節約をしていて、Go の
// encoding/json は map のキーを常にアルファベット順で出力する。そのため、同じ内容の
// doc でも「一度も剥がされていない（元の入力バイト列のまま）」ものと「剥がされて
// 作り直された（キー順が正規化された）」ものとでは、意味は同じでもバイト列が違う
// ことがある。UpdateTicketUseCase が「本文が変わったか」を判定する土台になるため、
// 誤検知（変えていないのに履歴が積まれる）を避けてここで木として比較する。
// 壊れた JSON はどちらも false（違うものとして扱う。保存前の検証は別で行う）。
func DocsEqual(a, b []byte) bool {
	var av, bv any
	if err := json.Unmarshal(a, &av); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		return false
	}
	return reflect.DeepEqual(av, bv)
}
