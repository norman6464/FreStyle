package middleware

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AuditEntry は監査ログに記録する 1 操作の情報（middleware が組み立てて渡す）。
// usecase 型に依存させないため middleware 側で定義し、記録処理は func で注入する。
type AuditEntry struct {
	ActorID    uint64
	ActorEmail string
	ActorRole  string
	Action     string
	TargetID   uint64
}

// AuditLog は、付与したルートの本処理が成功（HTTP < 400）したときに監査ログを 1 件記録する middleware。
// 記録は best-effort（record 内のエラーで本処理を壊さない）。読み取りや失敗レスポンスは記録しない。
// CurrentUser middleware より後段で使う（actor を context から取得するため）。
func AuditLog(record func(ctx context.Context, e AuditEntry)) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 失敗（4xx/5xx）は記録しない。成功した変更操作だけ残す。
		if c.Writer.Status() >= 400 {
			return
		}
		actor := CurrentUserFromContext(c)
		if actor == nil {
			return
		}
		record(c.Request.Context(), AuditEntry{
			ActorID:    actor.ID,
			ActorEmail: actor.Email,
			ActorRole:  actor.Role,
			Action:     c.Request.Method + " " + c.FullPath(),
			TargetID:   targetIDFromParams(c),
		})
	}
}

// targetIDFromParams は操作対象 ID を path パラメータから推定する（取得できなければ 0）。
func targetIDFromParams(c *gin.Context) uint64 {
	for _, name := range []string{"id", "userId"} {
		if v := c.Param(name); v != "" {
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				return n
			}
		}
	}
	return 0
}
