package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// RealClientIP は本番（Cloud Run）でも攻撃者に詐称されないクライアント IP を返す。
//
// # なぜ gin の c.ClientIP() を使わないか
//
// gin の既定の c.ClientIP() は SetTrustedProxies を呼ばない限り「全 IP を信頼する」
// 設定のまま動き、その場合 X-Forwarded-For の**先頭**を読む。先頭はリクエストを送る側が
// 自由に書ける値なので、ヘッダを詐称すれば IP 単位の流量制限の鍵を要求ごとに変えられる。
//
// gin の SetTrustedProxies は「既知の CIDR の踏み台を、末尾から辿って信頼できる区間を
// 読み飛ばす」設計を前提にしている。Cloud Run のフロントエンドが実際にどの IP から
// コンテナへ接続してくるかは公開・固定されていないため、この「信頼する踏み台の CIDR を
// 列挙する」前提に乗らない。
//
// # Cloud Run 特有の前提
//
// この API の唯一の入口は Cloud Run で、コンテナへの直接接続経路は無い（Cloud Run
// 自身のフロントエンドを必ず経由する）。Cloud Run は要求を転送する際、送られてきた
// X-Forwarded-For をそのまま残した上で、**自分が実際に接続を受けた相手の IP を必ず
// 末尾に追記する**（書き換えも検証もしない）。つまり先頭以降のどの要素も詐称され得るが、
// 末尾の 1 要素だけは Cloud Run 自身が観測した接続元で、要求元がすり替えられない。
//
// ヘッダが無い経路（ローカル開発など、手前にリバースプロキシが無い場合）は
// c.Request.RemoteAddr（TCP の直接の接続元）を使う。
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
