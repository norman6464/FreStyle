package ticket

import (
	"encoding/json"
	"strconv"
)

// ticketCommentMentionNodeType は発言本文中で人を指す ProseMirror インラインノードの type 名。
// attrs.userId は users.id（bigint）の 10 進文字列を持つ（担当・ウォッチャーと違い、
// 発言者は必ず実在の人なので principals への複合参照は要らない）。
const ticketCommentMentionNodeType = "mention"

// ExtractTicketCommentMentions は発言本文（ProseMirror インラインノードの配列。
// domain.ValidateCommentBody と同じ形）から名指しされた userId を、本文順・重複なしで
// 集める（doc.go の ExtractDocRefs と同じ考え方の素朴な木の走査）。
//
// 不正な値（数値でない・0 以下）は黙って無視する（保存を落とす理由にはしない。宛先が
// 本当にこのワークスペースの一員かは呼び出し側の usecase が別途確認する）。
func ExtractTicketCommentMentions(body []byte) []uint64 {
	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		return nil
	}
	seen := map[uint64]struct{}{}
	var ids []uint64
	var walk func(node any)
	walk = func(node any) {
		switch v := node.(type) {
		case map[string]any:
			if v["type"] == ticketCommentMentionNodeType {
				if id, ok := parseMentionUserID(attrString(v, "userId")); ok {
					if _, dup := seen[id]; !dup {
						seen[id] = struct{}{}
						ids = append(ids, id)
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
	for _, item := range items {
		var node any
		if err := json.Unmarshal(item, &node); err != nil {
			continue
		}
		walk(node)
	}
	return ids
}

func parseMentionUserID(raw string) (uint64, bool) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}
