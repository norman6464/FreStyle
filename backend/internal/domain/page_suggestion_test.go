package domain_test

import (
	"errors"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test_提案の状態定数はDBのCHECK制約と同じ文字列である は、
// ck_page_suggestions_status（status IN ('open','accepted','rejected')）と定数の食い違いを
// 型では検出できないため、文字列そのものを固定する。
func Test_提案の状態定数はDBのCHECK制約と同じ文字列である(t *testing.T) {
	assert.Equal(t, domain.PageSuggestionStatus("open"), domain.PageSuggestionStatusOpen)
	assert.Equal(t, domain.PageSuggestionStatus("accepted"), domain.PageSuggestionStatusAccepted)
	assert.Equal(t, domain.PageSuggestionStatus("rejected"), domain.PageSuggestionStatusRejected)
}

// Test_提案のセンチネルエラーは区別できる は ErrPageSuggestionNotFound /
// ErrPageSuggestionAlreadyResolved が別々のエラーとして errors.Is で判定できることを固定する
// （repository.Resolve がこの 2 つを使い分けるため、混同すると 404 と 409 を取り違える）。
func Test_提案のセンチネルエラーは区別できる(t *testing.T) {
	assert.NotErrorIs(t, domain.ErrPageSuggestionNotFound, domain.ErrPageSuggestionAlreadyResolved)
	assert.True(t, errors.Is(domain.ErrPageSuggestionNotFound, domain.ErrPageSuggestionNotFound))
	assert.True(t, errors.Is(domain.ErrPageSuggestionAlreadyResolved, domain.ErrPageSuggestionAlreadyResolved))
}
