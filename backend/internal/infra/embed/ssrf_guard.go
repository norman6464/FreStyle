package embed

import (
	"context"
	"fmt"
	"net"
)

// resolverFunc は「ホスト名 → IP 群」を引く関数の形。本番は defaultResolve
// （実 DNS）を使うが、DNS リバインディングの検証はテストから偽の解決結果を
// 注入できるよう関数として切り出してある（本物の DNS に依存すると、外部の
// ドメインを乗っ取らない限り「解決結果が変わる」状況をテストで再現できない）。
type resolverFunc func(ctx context.Context, host string) ([]net.IP, error)

// defaultResolve は本番で使う実 DNS 解決。
func defaultResolve(ctx context.Context, host string) ([]net.IP, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(addrs))
	for i, a := range addrs {
		ips[i] = a.IP
	}
	return ips, nil
}

// isSafeIP は SSRF 対策として「サーバから見て外部のホストに向いた IP」かを判定する。
// ホスト名の文字列照合（localhost・metadata.google.internal 等のハードコード）ではなく
// 解決後の IP そのものを見るので、任意のドメインを private / link-local / metadata の
// アドレスへ向けた変種（DNS リバインディングを含む）も同じ 1 か所で弾ける。
func isSafeIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() {
		return false
	}
	// CGNAT（100.64.0.0/10）。IsPrivate は RFC1918 (10/8, 172.16/12, 192.168/16) しか
	// 見ないので、キャリア級 NAT のレンジは別に見る必要がある。
	// ip.To4() は IPv4 射影 IPv6（::ffff:a.b.c.d）も展開して返すため、その変種も拾う。
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return false
		}
	}
	return true
}

// resolveSafeIP は host を解決し、isSafeIP を満たす最初の IP を返す。
// 満たす IP が 1 つも無ければ ErrUnsupportedHost、解決自体に失敗すれば ErrUnreachable。
func resolveSafeIP(ctx context.Context, resolve resolverFunc, host string) (net.IP, error) {
	ips, err := resolve(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve %s: %w", ErrUnreachable, host, err)
	}
	for _, ip := range ips {
		if isSafeIP(ip) {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("%w: %s has no public address", ErrUnsupportedHost, host)
}

// dialContextFunc は net.Dialer.DialContext と同じ形（テストから差し替え可能にするための型）。
type dialContextFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// safeDialContext は http.Transport.DialContext に差し込む関数を組み立てる。
//
// http.Transport は新しい接続を張るたびに（最初の要求はもちろん、リダイレクトで
// ホストが変わった場合の各ホップでも）DialContext を呼ぶ。ここで resolve → isSafeIP →
// その IP へ直接 dial、という順に固定することで:
//
//   - リダイレクト先の再検証: ホップごとに新しい接続が要るため、ここが自動的に
//     ホップごとの検査になる（CheckRedirect 側で改めて IP を検査する必要が無い）
//   - 検査と接続の TOCTOU（DNS リバインディング）の防止: 検査に使った IP を
//     そのまま dial する（addr の文字列を渡し直して dialer 自身に再解決させない）。
//     再解決を挟むと、検査に使った瞬間と接続する瞬間で異なる IP が返る攻撃
//     （最初は無害な IP を返し、検査が通った直後に社内アドレスへ TTL を切り替える）が
//     成立してしまう
func safeDialContext(resolve resolverFunc, dial dialContextFunc) dialContextFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ip, err := resolveSafeIP(ctx, resolve, host)
		if err != nil {
			return nil, err
		}
		return dial(ctx, network, net.JoinHostPort(ip.String(), port))
	}
}
