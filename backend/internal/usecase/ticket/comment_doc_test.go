package ticket_test

import (
	"encoding/json"
	"strconv"
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

	t.Run("大量のmentionノードでも上限件数で打ち切る", func(t *testing.T) {
		// 上限（50）を大きく超える数の異なる userId を並べる。1 本の発言から拾う id 数が
		// 頭打ちにならないと、宛先解決の問い合わせがその分だけ膨らむ（notify 側の懸念）。
		const nodeCount = 5000
		nodes := make([]map[string]any, 0, nodeCount)
		for i := 1; i <= nodeCount; i++ {
			nodes = append(nodes, map[string]any{
				"type":  "mention",
				"attrs": map[string]any{"userId": strconv.Itoa(i)},
			})
		}
		body, err := json.Marshal(nodes)
		assert.NoError(t, err)

		got := ticket.ExtractTicketCommentMentions(body)
		assert.Len(t, got, 50, "上限を超えた分は切り捨てる")
		assert.Equal(t, uint64(1), got[0], "先頭から順に拾う")
		assert.Equal(t, uint64(50), got[49])
	})
}
