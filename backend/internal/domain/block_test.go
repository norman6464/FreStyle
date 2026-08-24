package domain

import "testing"

func Test_BlockType_Valid(t *testing.T) {
	cases := []struct {
		name string
		t    BlockType
		want bool
	}{
		{"paragraph", BlockTypeParagraph, true},
		{"heading", BlockTypeHeading, true},
		{"codeBlock", BlockTypeCodeBlock, true},
		{"tableCell", BlockTypeTableCell, true},
		{"taskItem", BlockTypeTaskItem, true},
		{"空文字", BlockType(""), false},
		{"インラインノード(text)はブロックにしない", BlockType("text"), false},
		{"doc はページそのものなのでブロックにしない", BlockType("doc"), false},
		{"未知ノード", BlockType("weird"), false},
		{"大文字違い", BlockType("Paragraph"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.t.Valid(); got != tc.want {
				t.Fatalf("Valid() = %v, want %v", got, tc.want)
			}
		})
	}
}

// Test_ValidBlockTypes_重複なし は一覧の取りこぼし・二重登録を防ぐ。
func Test_ValidBlockTypes_重複なし(t *testing.T) {
	seen := make(map[BlockType]struct{}, len(ValidBlockTypes))
	for _, v := range ValidBlockTypes {
		if v == "" {
			t.Fatal("ValidBlockTypes に空文字が含まれています")
		}
		if _, dup := seen[v]; dup {
			t.Fatalf("ValidBlockTypes に %q が重複しています", v)
		}
		seen[v] = struct{}{}
	}
}
