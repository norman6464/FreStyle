package domain_test

import (
	"strings"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_雛形名の検証 は domain.ValidateTemplateName の境界を固定する。
//
// 変異確認: len(trimmed) > PageTemplateNameMaxBytes の `>` を `>=` に変えると
// 「100 バイトちょうどは OK」のケースが緑のまま落ちなくなる — このテストがそれを捕まえる。
func Test_雛形名の検証(t *testing.T) {
	t.Run("空白のみはErrInvalidTemplateName", func(t *testing.T) {
		got, err := domain.ValidateTemplateName("   \n\t ")
		require.ErrorIs(t, err, domain.ErrInvalidTemplateName)
		assert.Empty(t, got)
	})

	t.Run("空文字はErrInvalidTemplateName", func(t *testing.T) {
		got, err := domain.ValidateTemplateName("")
		require.ErrorIs(t, err, domain.ErrInvalidTemplateName)
		assert.Empty(t, got)
	})

	t.Run("前後の空白はTrimSpaceされる", func(t *testing.T) {
		got, err := domain.ValidateTemplateName("  議事録  ")
		require.NoError(t, err)
		assert.Equal(t, "議事録", got)
	})

	t.Run("100バイトちょうどはOK", func(t *testing.T) {
		name := strings.Repeat("a", domain.PageTemplateNameMaxBytes)
		got, err := domain.ValidateTemplateName(name)
		require.NoError(t, err)
		assert.Equal(t, name, got)
		assert.Len(t, got, domain.PageTemplateNameMaxBytes)
	})

	t.Run("101バイトはErrInvalidTemplateName", func(t *testing.T) {
		name := strings.Repeat("a", domain.PageTemplateNameMaxBytes+1)
		got, err := domain.ValidateTemplateName(name)
		require.ErrorIs(t, err, domain.ErrInvalidTemplateName)
		assert.Empty(t, got)
	})

	t.Run("TrimSpace後に100バイトを超えるかで判定する", func(t *testing.T) {
		// TrimSpace 前は 100 + 空白ぶんで超えているが、trim すればちょうど 100 バイトなので OK。
		name := "  " + strings.Repeat("a", domain.PageTemplateNameMaxBytes) + "  "
		got, err := domain.ValidateTemplateName(name)
		require.NoError(t, err)
		assert.Len(t, got, domain.PageTemplateNameMaxBytes)
	})
}
