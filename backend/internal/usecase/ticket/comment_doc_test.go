package ticket_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
	"github.com/stretchr/testify/assert"
)

func Test_発言のメンション抽出(t *testing.T) {
	t.Run("mentionノードのuserIdを文書順・重複なしで集める", func(t *testing.T) {
		body := `[
			{"type":"text","text":"こんにちは "},
			{"type":"mention","attrs":{"userId":"7"}},
			{"type":"text","text":" と "},
			{"type":"mention","attrs":{"userId":"3"}},
			{"type":"mention","attrs":{"userId":"7"}}
		]`
		got := ticket.ExtractTicketCommentMentions([]byte(body))
		assert.Equal(t, []uint64{7, 3}, got)
	})

	t.Run("mentionが無ければ空", func(t *testing.T) {
		body := `[{"type":"text","text":"ただの発言"}]`
		assert.Empty(t, ticket.ExtractTicketCommentMentions([]byte(body)))
	})

	t.Run("不正なuserId_数値でない_0以下_は無視する", func(t *testing.T) {
		body := `[
			{"type":"mention","attrs":{"userId":"abc"}},
			{"type":"mention","attrs":{"userId":"0"}},
			{"type":"mention","attrs":{"userId":"-1"}},
			{"type":"mention","attrs":{"userId":"5"}}
		]`
		assert.Equal(t, []uint64{5}, ticket.ExtractTicketCommentMentions([]byte(body)))
	})

	t.Run("壊れたJSONは空", func(t *testing.T) {
		assert.Empty(t, ticket.ExtractTicketCommentMentions([]byte(`not json`)))
	})
}
