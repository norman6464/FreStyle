package persistence

import (
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func at(sec int) time.Time {
	return time.Date(2026, 7, 25, 12, 0, sec, 0, time.UTC)
}

func Test_会話履歴の整列_SKのUUID順で崩れた履歴を時系列に戻す(t *testing.T) {
	// DynamoDB Query は SK(ランダム UUID)順で返すため、時系列がシャッフルされた状態を再現する。
	msgs := []domain.AiChatMessage{
		{MessageID: "b", Role: domain.AiChatRoleUser, Content: "2つ目の質問", CreatedAt: at(30)},
		{MessageID: "a", Role: domain.AiChatRoleAssistant, Content: "1つ目の応答", CreatedAt: at(20)},
		{MessageID: "c", Role: domain.AiChatRoleUser, Content: "1つ目の質問", CreatedAt: at(10)},
		{MessageID: "d", Role: domain.AiChatRoleAssistant, Content: "2つ目の応答", CreatedAt: at(40)},
	}

	sortAiChatMessages(msgs)

	got := make([]string, len(msgs))
	for i, m := range msgs {
		got[i] = m.Content
	}
	assert.Equal(t, []string{"1つ目の質問", "1つ目の応答", "2つ目の質問", "2つ目の応答"}, got)
}

func Test_会話履歴の整列_同一秒内はuserが先(t *testing.T) {
	// created_at は秒精度のため、短い応答では user と assistant が同一秒になりうる。
	msgs := []domain.AiChatMessage{
		{MessageID: "x", Role: domain.AiChatRoleAssistant, Content: "応答", CreatedAt: at(5)},
		{MessageID: "y", Role: domain.AiChatRoleUser, Content: "質問", CreatedAt: at(5)},
	}

	sortAiChatMessages(msgs)

	assert.Equal(t, "質問", msgs[0].Content)
	assert.Equal(t, "応答", msgs[1].Content)
}

func Test_会話履歴の整列_同一秒同一ロールはmessageIdで安定(t *testing.T) {
	msgs := []domain.AiChatMessage{
		{MessageID: "zz", Role: domain.AiChatRoleUser, Content: "後", CreatedAt: at(5)},
		{MessageID: "aa", Role: domain.AiChatRoleUser, Content: "先", CreatedAt: at(5)},
	}

	sortAiChatMessages(msgs)

	assert.Equal(t, "先", msgs[0].Content)
	assert.Equal(t, "後", msgs[1].Content)
}
