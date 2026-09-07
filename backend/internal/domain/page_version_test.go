package domain_test

import (
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_版のメモ検証 は domain.ValidateVersionNote の境界を固定する。
//
// 変異確認: len(trimmed) > PageVersionNoteMaxBytes の `>` を `>=` に変えると
// 「500 バイトちょうどは OK」のケースが緑のまま落ちなくなる — このテストがそれを捕まえる。
func Test_版のメモ検証(t *testing.T) {
	t.Run("nilはそのままnilを返しOK", func(t *testing.T) {
		got, err := domain.ValidateVersionNote(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("空文字はnilに正規化されOK", func(t *testing.T) {
		empty := ""
		got, err := domain.ValidateVersionNote(&empty)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("空白のみはnilに正規化されOK", func(t *testing.T) {
		blank := "   \n\t "
		got, err := domain.ValidateVersionNote(&blank)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("前後の空白はTrimSpaceされる", func(t *testing.T) {
		padded := "  メモ  "
		got, err := domain.ValidateVersionNote(&padded)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "メモ", *got)
	})

	t.Run("500バイトちょうどはOK", func(t *testing.T) {
		note := strings.Repeat("a", domain.PageVersionNoteMaxBytes)
		got, err := domain.ValidateVersionNote(&note)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, note, *got)
		assert.Len(t, *got, domain.PageVersionNoteMaxBytes)
	})

	t.Run("501バイトはErrInvalidPageVersionNote", func(t *testing.T) {
		note := strings.Repeat("a", domain.PageVersionNoteMaxBytes+1)
		got, err := domain.ValidateVersionNote(&note)
		require.ErrorIs(t, err, domain.ErrInvalidPageVersionNote)
		assert.Nil(t, got)
	})

	t.Run("TrimSpace後に500バイトを超えるかで判定する", func(t *testing.T) {
		// TrimSpace 前は 500 + 空白ぶんで超えているが、trim すればちょうど 500 バイトなので OK。
		note := "  " + strings.Repeat("a", domain.PageVersionNoteMaxBytes) + "  "
		got, err := domain.ValidateVersionNote(&note)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Len(t, *got, domain.PageVersionNoteMaxBytes)
	})
}
