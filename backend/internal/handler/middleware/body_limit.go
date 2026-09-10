package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxRequestBody はリクエスト本文の読み込みサイズに上限を課す middleware を返す。
//
// c.ShouldBindJSON 等は内部で本文をまるごとメモリに読む。http.MaxBytesReader で
// 包まない限り読み込み量に上限が無く、認証済みの 1 要求が巨大な本文を送るだけで
// プロセスのメモリを食いつぶせる（並行させれば他テナントの API も巻き込む）。
//
// 個別のハンドラが自分の本文の性質に合わせてより厳しい上限（limitTicketBody /
// limitKnowledgeBaseBody）を重ねて呼ぶのは構わない — http.MaxBytesReader は
// 呼ぶたびにその時点の残り読み込み量を上書きするので、後から呼んだ方（より厳しい方）が
// 効く。ここでの狙いは「個別に呼び忘れたハンドラ」を無防備にしないことで、
// 新しいハンドラを足すたびにこの呼び出しを思い出す必要が無いよう、全ルートの入口
// （NewRouter）1 箇所に置く。
func MaxRequestBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
