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

/**
 * toCspOrigin は API のベース URL を CSP に書ける origin へ正規化する。
 *
 * path / query / 空白 / セミコロンをそのまま CSP へ流すと、ディレクティブを増やされる等の
 * 事故になりうるため origin だけを取り出す。解釈できない値や http(s) 以外は、
 * 黙って無視せず vite の設定読み込み時に落として設定ミスに気づけるようにする。
 */
export function toCspOrigin(apiBaseUrl) {
  if (!apiBaseUrl) return null;

  let url;
  try {
    url = new URL(apiBaseUrl);
  } catch {
    throw new Error(`VITE_API_BASE_URL を URL として解釈できません: ${apiBaseUrl}`);
  }
  if (url.protocol !== 'http:' && url.protocol !== 'https:') {
    throw new Error(`VITE_API_BASE_URL は http / https のみ対応します: ${apiBaseUrl}`);
  }
  return url.origin;
}

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

/** devCsp は開発サーバ限定で CSP に API のオリジンを足す vite プラグインを返す。 */
export function devCsp(apiBaseUrl) {
  // 設定ミスは plugin 生成時（= vite 設定読み込み時）に落とす。
  const origin = toCspOrigin(apiBaseUrl);

  return {
    name: 'frestyle-dev-csp',
    // 'serve' 限定。npm run build の成果物には一切影響しない。
    apply: 'serve',
    transformIndexHtml(html) {
      return addConnectSrc(html, [origin]);
    },
  };
}
