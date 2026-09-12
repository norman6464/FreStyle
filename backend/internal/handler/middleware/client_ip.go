package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// RealClientIP は本番（Cloud Run）でも攻撃者に詐称されないクライアント IP を返す。
//
// gin 既定の c.ClientIP() は SetTrustedProxies を呼ばない限り X-Forwarded-For の先頭を読むが、
// 先頭はリクエスト送信側が自由に書ける値なので、詐称すれば IP 単位の流量制限の鍵を要求ごと
// に変えられる。SetTrustedProxies は「既知の CIDR の踏み台を末尾から辿る」設計が前提だが、
// Cloud Run のフロントエンドがどの IP からコンテナへ接続するかは公開・固定されておらず、
// その前提に乗らない。
//
// この API の唯一の入口は Cloud Run で、コンテナへの直接接続経路は無い。Cloud Run は要求を
// 転送する際、送られてきた X-Forwarded-For をそのまま残した上で、自分が実際に接続を受けた
// 相手の IP を必ず末尾に追記する（書き換えも検証もしない）。つまり先頭以降のどの要素も
// 詐称され得るが、末尾の 1 要素だけは Cloud Run 自身が観測した接続元で、すり替えられない。
//
// ヘッダが無い経路（ローカル開発など）は c.Request.RemoteAddr（TCP の直接の接続元）を使う。
func RealClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		last := strings.TrimSpace(parts[len(parts)-1])
		if ip := net.ParseIP(last); ip != nil {
			return ip.String()
		}
	}
	if host, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return host
	}
	return c.Request.RemoteAddr
}
