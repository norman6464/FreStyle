/**
 * 認可の設定を「欠けうるもの」として 1 か所で読む。
 *
 * 設定はビルド時にバンドルへ焼き込まれる。焼き込まれなければ**無い**。
 * 「必ずある」と信じて読むと、欠けた設定で認可 URL を組み立てるコードが
 * 素通りし、押して初めて例外になる（利用者から見ると「押しても何も起きない」）。
 *
 * そこで戻り値を判別可能な合併にして、**設定が揃っている枝でしか
 * 認可 URL を作れない**ようにする。揃っていない枝には値そのものが無いので、
 * 組み立てを書こうとすると型検査で落ちる。
 *
 * 読み取りをこの 1 か所に閉じてあるのは、将来ここだけを差し替えられるようにするため。
 * 環境ごとに作り直すのをやめて 1 つの成果物を使い回す段になったら、
 * 入力元を `import.meta.env` から配信された設定へ変えれば、呼ぶ側は変わらない。
 */

/** 認可に必要な設定がすべて揃っている状態。 */
export type ConfiguredAuth = {
  readonly status: 'configured';
  readonly authorizeUri: string;
  readonly clientId: string;
  readonly redirectUri: string;
  readonly scope: string;
};

/** 設定が欠けている状態。何が欠けているかを持つ（表示ではなく記録・検査のため）。 */
export type UnconfiguredAuth = {
  readonly status: 'unconfigured';
  readonly missing: readonly string[];
};

export type AuthConfig = ConfiguredAuth | UnconfiguredAuth;

/**
 * 既定の範囲。
 *
 * `openid` だけだと id_token に名前もメールも載らず、`offline_access` が無いと
 * 更新用のトークンがそもそも発行されない（切れた瞬間に全員ログイン画面へ飛ぶ）。
 */
export const DEFAULT_SCOPE = 'openid profile email offline_access';

/**
 * 絶対 http(s) URL として解釈できるか。
 *
 * **空でないことだけを見てはいけない。** 値が「有る」ことと「使える」ことは別で、
 * `authorize` のような相対文字列や打ち間違えた値は空チェックを素通りし、
 * 発行者の画面へ飛んで初めて分かる。
 */
function isHttpUrl(value: string | undefined): value is string {
  if (!value) return false;
  try {
    const url = new URL(value);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

/**
 * 焼き込まれた設定を読んで、揃っているかを判定する。
 *
 * 欠けている名前を集めて返すのは、`missing` を検査と記録に使うため。
 * 人が読む文には出さない（環境変数の名前は利用者に意味が無い）。
 */
export function readAuthConfig(): AuthConfig {
  const env = import.meta.env;
  const authorizeUri = env.VITE_OIDC_AUTHORIZE_URI;
  const clientId = env.VITE_OIDC_CLIENT_ID;
  const redirectUri = env.VITE_OIDC_REDIRECT_URI;

  const missing: string[] = [];
  if (!isHttpUrl(authorizeUri)) missing.push('VITE_OIDC_AUTHORIZE_URI');
  if (!clientId) missing.push('VITE_OIDC_CLIENT_ID');
  if (!isHttpUrl(redirectUri)) missing.push('VITE_OIDC_REDIRECT_URI');

  // `missing.length === 0` で分岐しても型は絞れない（配列の長さは値の有無を語らない）。
  // 条件そのものをもう一度書くことで、`configured` の枝では 3 つとも string になる。
  if (isHttpUrl(authorizeUri) && clientId && isHttpUrl(redirectUri)) {
    return {
      status: 'configured',
      authorizeUri,
      clientId,
      redirectUri,
      scope: env.VITE_OIDC_SCOPE || DEFAULT_SCOPE,
    };
  }

  return { status: 'unconfigured', missing };
}
