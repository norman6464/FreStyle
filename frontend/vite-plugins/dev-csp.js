/*
 * 開発サーバ（vite dev）のときだけ index.html の CSP connect-src に開発用オリジンを足す。
 *
 * index.html の CSP は本番向けに `connect-src 'self' https://api.frestyle.jp ...` を許可している。
 * ローカル開発は SPA が localhost:5173、API が localhost:8080 と別オリジンになるため、
 * そのままでは health / auth の fetch が CSP で遮断され、画面がメンテナンス表示のまま進まない。
 * 本番ビルドの CSP は変えたくないので、dev サーバのときだけ配信する HTML を書き換える。
 */

// index.html 冒頭の解説コメントにも connect-src の記述があるため、
// meta タグの content 属性だけを対象にする（コメント側を書き換えないため）。
const CSP_META = /(<meta\s+http-equiv="Content-Security-Policy"\s+content=")([^"]*)(")/i;

/** addConnectSrc は CSP meta の connect-src に origins を追記した HTML を返す。 */
export function addConnectSrc(html, origins) {
  const allowed = origins.filter(Boolean);
  if (allowed.length === 0) return html;

  return html.replace(
    CSP_META,
    (_match, head, content, tail) =>
      head +
      content.replace(/connect-src[^;]*/, (directive) => `${directive} ${allowed.join(' ')}`) +
      tail,
  );
}

/** devCsp は開発サーバ限定で CSP を緩める vite プラグインを返す。 */
export function devCsp(origins) {
  return {
    name: 'frestyle-dev-csp',
    // 'serve' 限定。npm run build の成果物には一切影響しない。
    apply: 'serve',
    transformIndexHtml(html) {
      return addConnectSrc(html, origins);
    },
  };
}
