package ticket

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

// チケット本文（ProseMirror の doc）を歩く処理。ページ側 page_usecase.go と同じ考え方だが、
// チケットは本文をブロック分解せず 1 本の doc を jsonb で持つため、木を直接歩く素朴な実装にする。

// ticketPageRefNodeType / ticketTicketRefNodeType は本文中でページ / 他チケットを指す
// ProseMirror ノードの type 名（pageRef は既存 kb と共通、ticketRef はチケット機能の新設で同じ形に揃える）。
const (
	ticketPageRefNodeType   = "pageRef"
	ticketTicketRefNodeType = "ticketRef"
)

// ExtractDocRefs は本文中の pageRef / ticketRef ノードから参照先 ID を文書順・重複なしで集める
// （DB を読まない純粋な木の走査なので domain ではなくここに置く）。不正な UUID・空文字は無視し
// 保存は落とさない（実在確認は repository.Replace*Links が別途行う）。
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

// StripDocRefTitles は保存前の doc から pageRef / ticketRef の title を取り除く。title は
// 読み出し時に解決する派生値で保存してはいけない — 保存すると、閲覧できない読み手にまで
// 保存者が見えていた題名が漏れる経路になる（page 側の StripPageRefTitles と同じ理由）。
// 参照が無ければ入力をそのまま返す。
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

// BuildPlainText は検索用の派生値（tickets.plain_text）を doc から作る。text ノードの内容だけを
// 集め、pageRef / ticketRef の属性（id・title）は含めない — 見えないページの題名が検索から
// 漏れる経路を作らないため。壊れた JSON は空文字を返す（保存自体は別の検証が守る）。
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

func attrString(node map[string]any, key string) string {
	attrs, ok := node["attrs"].(map[string]any)
	if !ok {
		return ""
	}
	s, _ := attrs[key].(string)
	return s
}

// canonicalRefID は参照 ID を UUID の正規形へ寄せる（page_usecase.go の canonicalPageRefID と同じ役割）。
func canonicalRefID(id string) (string, bool) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
}

// DocsEqual は 2 つの doc（jsonb の生バイト列）が意味的に同じ木かを比べる。バイト列のままでは
// 比較しない — StripDocRefTitles は変更が無ければ入力をそのまま返し、変更があれば
// json.Marshal で作り直す（キー順がアルファベット順に正規化される）ため、意味は同じでも
// バイト列が食い違うことがある。UpdateTicketUseCase の変更検知の土台になるため木として
// 比較する。壊れた JSON はどちらも false（保存前の検証は別で行う）。
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
