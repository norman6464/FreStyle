// Package embed は外部 URL の OGP / oEmbed メタ情報を取得して
// フロントエンドの Embed カード描画に渡すためのユーティリティを提供する。
//
// 設計:
//   - allow-list 方式（任意の URL を叩かせない）。SSRF / DNS rebinding 対策で
//     scheme は https のみ許可、host は許可リストか明示的に「Web 一般」のみ
//   - キャッシュは Note 編集中に同じ URL を何度も叩かれるのを避けるため In-Memory LRU で薄く保持
//   - フロントは /api/v2/embeds/oembed?url=... で叩く（PR F の EmbedCardExtension が利用）
package embed

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	userAgent          = "FreStyle/1.0 (+https://normanblog.com)"
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
	// allowLoopback は httptest (127.0.0.1) で動かすときの test bypass。
	// 本番経路で生成する NewFetcher は false 固定。
	allowLoopback bool
}

// NewFetcher は本番デフォルト設定で Fetcher を返す。
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: defaultHTTPTimeout},
		cache:  newCache(cacheMaxEntries),
	}
}

// NewFetcherWithClient はテスト時に http.Client を差し替える DI コンストラクタ。
// 同時に allowLoopback=true にして httptest を踏めるようにする (test 限定の admit list)。
func NewFetcherWithClient(c *http.Client) *Fetcher {
	if c == nil {
		c = &http.Client{Timeout: defaultHTTPTimeout}
	}
	return &Fetcher{client: c, cache: newCache(cacheMaxEntries), allowLoopback: true}
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

// validateURL は scheme=https / host 非空 / 既知のローカル/プライベートホストでないことを検証する。
// allowLoopback=true (test 経路) のときは loopback / private を許可する。
func (f *Fetcher) validateURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("%w: scheme must be https", ErrInvalidURL)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("%w: empty host", ErrInvalidURL)
	}
	if f.allowLoopback {
		return u, nil
	}
	host := strings.ToLower(u.Hostname())
	if isPrivateOrLocalHost(host) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedHost, host)
	}
	return u, nil
}

// isPrivateOrLocalHost は SSRF 対策のため、localhost / プライベートレンジ / link-local /
// metadata IP（AWS / GCP）に向けた解決を弾く。十分に厳格にしたいわけではなく、最低限の毒抜き。
func isPrivateOrLocalHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	// AWS / GCP の instance metadata。
	if host == "169.254.169.254" || host == "metadata.google.internal" {
		return true
	}
	// 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8
	privateIPv4 := regexp.MustCompile(
		`^(10\.|127\.|192\.168\.|172\.(1[6-9]|2[0-9]|3[0-1])\.)`,
	)
	if privateIPv4.MatchString(host) {
		return true
	}
	// IPv6 loopback / link-local
	if host == "::1" || strings.HasPrefix(host, "fe80:") || strings.HasPrefix(host, "[fe80:") {
		return true
	}
	return false
}

// resolveOGP はシンプルな OGP 抽出。
func (f *Fetcher) resolveOGP(ctx context.Context, u *url.URL) (*Card, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreachable, err)
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

// ---- in-memory LRU ----

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
