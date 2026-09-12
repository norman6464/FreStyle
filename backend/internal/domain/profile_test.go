package domain_test

import (
	"testing"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_プロフィール_ステータス失効判定(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	cases := []struct {
		name      string
		expiresAt *time.Time
		wantEmoji string
		wantText  string
	}{
		{"無期限なら見える", nil, "🎉", "休暇中"},
		{"期限内なら見える", &future, "🎉", "休暇中"},
		{"期限を過ぎたら空", &past, "", ""},
		{"期限ちょうどは空", &now, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			emoji, text := domain.ClearExpiredStatus("🎉", "休暇中", tc.expiresAt, now)
			assert.Equal(t, tc.wantEmoji, emoji)
			assert.Equal(t, tc.wantText, text)
		})
	}
}

func Test_プロフィール_ステータス表示の組み立て(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	cases := []struct {
		name      string
		emoji     string
		text      string
		expiresAt *time.Time
		want      string
	}{
		{"絵文字とテキストを1つに結合する", "🎉", "休暇中", nil, "🎉 休暇中"},
		{"絵文字だけでも組み立つ", "🎉", "", nil, "🎉"},
		{"テキストだけでも組み立つ", "", "休暇中", nil, "休暇中"},
		{"どちらも空なら空文字", "", "", nil, ""},
		{"期限内なら見える", "🎉", "休暇中", &future, "🎉 休暇中"},
		{"失効していれば空文字", "🎉", "休暇中", &past, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.ComposeStatusDisplay(tc.emoji, tc.text, tc.expiresAt, now)
			assert.Equal(t, tc.want, got)
		})
	}
}
