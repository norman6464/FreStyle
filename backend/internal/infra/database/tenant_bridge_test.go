package database

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestWorkspaceSlugFor は自動採番した slug が workspaces の制約
// （グローバル一意・1..64 文字・URL に出せる文字）を満たすことを固定する。
func TestWorkspaceSlugFor(t *testing.T) {
	id := uuid.MustParse("0198f3c1-2b4d-7c5e-8a9b-0c1d2e3f4a5b")
	slug := workspaceSlugFor(id)

	require.Equal(t, "ws-0198f3c12b4d7c5e8a9b0c1d2e3f4a5b", slug)
	require.LessOrEqual(t, len(slug), 64)
	require.Regexp(t, `^ws-[0-9a-f]{32}$`, slug)

	// 別 ID なら別 slug（一意性の根拠は UUID そのもの）。
	require.NotEqual(t, slug, workspaceSlugFor(uuid.MustParse("0198f3c1-2b4d-7c5e-8a9b-0c1d2e3f4a5c")))
}

// TestTruncateRunes は表示名の切り詰めが文字数基準であることを固定する。
// varchar(200) の 200 は文字数なので、バイト数で切ると日本語の会社名で長さがずれる。
func TestTruncateRunes(t *testing.T) {
	require.Equal(t, "あいう", truncateRunes("あいう", 200))
	require.Equal(t, "あい", truncateRunes("あいうえお", 2))
	require.Equal(t, "", truncateRunes("あ", 0))
}
