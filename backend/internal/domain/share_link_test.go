package domain_test

import (
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_共有リンク_使えるかの判定(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	cases := []struct {
		name   string
		link   domain.ShareLink
		usable bool
	}{
		{"無期限・未失効なら使える", domain.ShareLink{}, true},
		{"期限内なら使える", domain.ShareLink{ExpiresAt: &future}, true},
		{"期限を過ぎたら使えない", domain.ShareLink{ExpiresAt: &past}, false},
		{"期限ちょうどは使えない", domain.ShareLink{ExpiresAt: &now}, false},
		{"失効済みは使えない", domain.ShareLink{RevokedAt: &past}, false},
		{"失効済みは期限内でも使えない", domain.ShareLink{RevokedAt: &past, ExpiresAt: &future}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.usable, tc.link.Usable(now))
		})
	}
}

func Test_共有リンク_パスワードの要否(t *testing.T) {
	hash := "$2a$10$dummy"
	assert.False(t, domain.ShareLink{}.RequiresPassword())
	assert.True(t, domain.ShareLink{PasswordHash: &hash}.RequiresPassword())
}
