// Package embed は外部 URL の OGP / oEmbed メタ情報を取得し、Embed カード描画用に返す。
// SSRF 対策で https のみ許可、結果は In-Memory LRU で薄くキャッシュする。
package embed

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	defaultHTTPTimeout = 6 * time.Second
	maxBodyBytes       = 512 * 1024 // OGP 抽出に必要なのは <head> の一部のみ
	cacheTTL           = 30 * time.Minute
	cacheMaxEntries    = 256
	userAgent          = "FreStyle/1.0 (+https://frestyle.dev)"
)

// Card はフロントエンドに返す統一カード DTO。
// OGP / oEmbed どちらの経路で取得しても同形式で返す。
type Card struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
	SiteName    string `json:"siteName,omitempty"`
	// Provider は "ogp" / "youtube" / "github" など、どの戦略で解決したかを示す。
	Provider string `json:"provider,omitempty"`
}

// Fetcher は URL → Card を解決する。
type Fetcher struct {
	client *http.Client
	cache  *cache
}

// maxRedirects は追うリダイレクトの最大ホップ数。CheckRedirect を独自に設定すると
// net/http の既定（10 ホップ）が効かなくなるため、同じ値をここで明示する。
const maxRedirects = 10

// NewFetcher は本番デフォルト設定で Fetcher を返す。
//
// Transport.DialContext を safeDialContext に差し替えることで、最初の接続だけでなく
// リダイレクトで新しく張る接続も含め、すべての接続が「解決した IP が外部向けか」の
// 検査を通る（safeDialContext の doc 参照）。CheckRedirect は IP の再検査までは
// 担わず、スキームの検査（https のみ）だけをホップごとにやり直す。
func NewFetcher() *Fetcher {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{Timeout: defaultHTTPTimeout}
	transport.DialContext = safeDialContext(defaultResolve, dialer.DialContext)
	f := &Fetcher{cache: newCache(cacheMaxEntries)}
	f.client = &http.Client{
		Timeout:       defaultHTTPTimeout,
		Transport:     transport,
		CheckRedirect: f.checkRedirect,
	}
	return f
}

// NewFetcherWithClient はテスト用。http.Client を丸ごと差し替える
// （safeDialContext / checkRedirect は適用されない。テストは httptest サーバの
// Transport をそのまま使うため、これらの本番専用の防御には元々乗らない経路）。
func NewFetcherWithClient(c *http.Client) *Fetcher {
	if c == nil {
		c = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return &Fetcher{client: c, cache: newCache(cacheMaxEntries)}
}

// checkRedirect はリダイレクト追跡のホップごとに呼ばれる。IP の安全性そのものは
// safeDialContext がホップごとの新規接続で必ず検査するので、ここでは
// スキーム（https のみ）とホップ数だけを見る。
func (f *Fetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return fmt.Errorf("%w: stopped after %d redirects", ErrUnreachable, maxRedirects)
	}
	return validateScheme(req.URL)
}

var (
	// ErrInvalidURL は URL parse 不能 / scheme が https でない / host が空 等。
	ErrInvalidURL = errors.New("embed: invalid url")
	// ErrUnreachable は HTTP 通信の失敗（DNS / TLS / timeout）。
	ErrUnreachable = errors.New("embed: target unreachable")
	// ErrUnsupportedHost は SSRF 防御で localhost / private IP に向けた要求を弾いたとき。
	ErrUnsupportedHost = errors.New("embed: unsupported host")
)

// Resolve は与えた URL を Card に解決する。キャッシュヒット時は HTTP を踏まない。
func (f *Fetcher) Resolve(ctx context.Context, raw string) (*Card, error) {
	u, err := f.validateURL(raw)
	if err != nil {
		return nil, err
	}
	key := u.String()
	if c, ok := f.cache.get(key); ok {
		return c, nil
	}
	card, err := f.resolveOGP(ctx, u)
	if err != nil {
		return nil, err
	}
	f.cache.set(key, card)
	return card, nil
}

// validateURL は URL をパースし、scheme=https / host 非空 を検証する。
//
// 「private / local なホストでないか」はここでは見ない。文字列の照合（旧実装）は
// ホスト名にしか効かず、公開ドメインを private / metadata の IP へ向ける変種
// （DNS リバインディングを含む）を素通りさせてしまう。その検査は実際に接続する
// 瞬間の IP に対して行うべきなので、safeDialContext（ssrf_guard.go）へ寄せてある。
func (f *Fetcher) validateURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}
	if err := validateScheme(u); err != nil {
		return nil, err
	}
	return u, nil
}

// validateScheme は https のみを許可する。初回の URL・リダイレクト先の双方から
// 呼ぶ共通ロジック（f.checkRedirect 参照）。
func validateScheme(u *url.URL) error {
	if u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be https", ErrInvalidURL)
	}
	if u.Host == "" {
		return fmt.Errorf("%w: empty host", ErrInvalidURL)
	}
	return nil
}

// resolveOGP はシンプルな OGP 抽出。
func (f *Fetcher) resolveOGP(ctx context.Context, u *url.URL) (*Card, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := f.client.Do(req)
	if err != nil {
		// err はここでさらに ErrUnreachable として包むが、safeDialContext /
		// checkRedirect が返した ErrUnsupportedHost・ErrInvalidURL は err の中に
		// （net/http が挟む *url.Error / *net.OpError 越しでも）残ったままなので、
		// errors.Is で拾える（embed_handler.go の switch は ErrUnsupportedHost /
		// ErrInvalidURL を ErrUnreachable より先に判定している。両方に一致する
		// エラーでも、より具体的な方の分岐が先に選ばれる）。
		return nil, fmt.Errorf("%w: %w", ErrUnreachable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%w: status=%d", ErrUnreachable, resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxBodyBytes)
	body, _ := io.ReadAll(limited)
	html := string(body)

	card := &Card{URL: u.String(), Provider: "ogp"}
	card.Title = firstMeta(html, []string{`property="og:title"`, `name="twitter:title"`})
	card.Description = firstMeta(html, []string{`property="og:description"`, `name="twitter:description"`, `name="description"`})
	card.ImageURL = firstMeta(html, []string{`property="og:image"`, `name="twitter:image"`})
	card.SiteName = firstMeta(html, []string{`property="og:site_name"`})
	if card.Title == "" {
		card.Title = extractTitleTag(html)
	}
	if card.Title == "" {
		card.Title = u.Host
	}
	return card, nil
}

// firstMeta は <meta property|name="..." content="..."> を順番に試して最初に見つかった content を返す。
// HTML パーサーは入れず、正規表現で十分（OGP は基本的に <head> 内の単純なメタタグ）。
func firstMeta(html string, selectors []string) string {
	for _, sel := range selectors {
		// <meta {sel} content="..."> または <meta content="..." {sel}>
		patterns := []string{
			fmt.Sprintf(`<meta\s+[^>]*?%s[^>]*?content="([^"]*)"`, regexp.QuoteMeta(sel)),
			fmt.Sprintf(`<meta\s+[^>]*?content="([^"]*)"[^>]*?%s`, regexp.QuoteMeta(sel)),
		}
		for _, p := range patterns {
			re := regexp.MustCompile(p)
			m := re.FindStringSubmatch(html)
			if len(m) > 1 && m[1] != "" {
				return strings.TrimSpace(m[1])
			}
		}
	}
	return ""
}

// extractTitleTag は <title>...</title> を最初の 1 件だけ拾う。
func extractTitleTag(html string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	m := re.FindStringSubmatch(html)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

type cacheEntry struct {
	card    *Card
	expires time.Time
}

type cache struct {
	mu      sync.Mutex
	max     int
	entries map[string]cacheEntry
}

func newCache(max int) *cache {
	return &cache{max: max, entries: make(map[string]cacheEntry)}
}

func (c *cache) get(k string) (*Card, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[k]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expires) {
		delete(c.entries, k)
		return nil, false
	}
	return e.card, true
}

func (c *cache) set(k string, v *Card) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.max {
		// 単純な oldest 1 件削除（厳密 LRU でなく first-iteration なので acceptable）。
		for key := range c.entries {
			delete(c.entries, key)
			break
		}
	}
	c.entries[k] = cacheEntry{card: v, expires: time.Now().Add(cacheTTL)}
}
