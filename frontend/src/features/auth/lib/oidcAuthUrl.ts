/**
 * 発行者（OpenID Connect の IdP）のログイン画面へ送る認可 URL を組み立てる。
 *
 * 組み立てと同時に、戻ってきたときに確かめるための値を 3 つ作って手元に置く。
 * 作る側と確かめる側が離れていると片方だけ実装されがちなので、保存はここで、
 * 取り出しはコールバック側で、と対にしてある（`consumeAuthFlowState`）。
 *
 * - **state**: 戻ってきた応答が、自分が始めた認可の応答かを確かめる。
 *   見ないと、攻撃者が自分の認可コードを他人のブラウザに踏ませて、
 *   被害者を攻撃者のアカウントでログインさせられる。
 * - **nonce**: 同じことを id_token の中身でも確かめる（照合はバックエンド）。
 * - **code_verifier**: PKCE。認可コードが盗まれても、それだけでは
 *   トークンに交換できないようにする。
 */

import type { ConfiguredAuth } from './authConfig';

const STORAGE_KEY = 'oidc.authFlow';

export type AuthFlowState = {
  state: string;
  nonce: string;
  codeVerifier: string;
};

/**
 * 認可 URL を作り、戻りで使う値を sessionStorage に置く。
 *
 * 設定は**引数で受け取る**。環境から自分で読むと、欠けているときに
 * `new URL(undefined)` で落ちるコードが型検査を通ってしまう。
 * `ConfiguredAuth` しか受け取らないので、設定が揃っていない状態で
 * ここを呼ぶコードは書けない（`readAuthConfig` の合併で絞られる）。
 *
 * provider を渡すと特定の IdP（例: Google）へ直行する。
 * screenHint に 'signup' を渡すと登録画面へ直行する。
 */
export async function buildAuthorizeUrl(
  cfg: ConfiguredAuth,
  provider?: string,
  screenHint?: 'signup' | 'signin',
): Promise<string> {
  const flow: AuthFlowState = {
    state: randomString(32),
    nonce: randomString(32),
    codeVerifier: randomString(64),
  };
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(flow));

  const url = new URL(cfg.authorizeUri);
  url.searchParams.set('client_id', cfg.clientId);
  url.searchParams.set('redirect_uri', cfg.redirectUri);
  url.searchParams.set('response_type', 'code');
  url.searchParams.set('scope', cfg.scope);
  url.searchParams.set('state', flow.state);
  url.searchParams.set('nonce', flow.nonce);
  url.searchParams.set('code_challenge', await sha256Base64Url(flow.codeVerifier));
  // 平文（plain）も規格上は許されるが、それでは盗聴された時点で意味が無い。
  url.searchParams.set('code_challenge_method', 'S256');
  if (provider) {
    url.searchParams.set('identity_provider', provider);
  }
  if (screenHint) {
    url.searchParams.set('prompt', screenHint === 'signup' ? 'create' : 'login');
  }
  return url.toString();
}

/**
 * 保存しておいた値を取り出して消す（使い切り）。
 *
 * 残しておくと、同じ値で 2 回目の交換が試せてしまう。
 * 見つからない場合は null（＝この端末で始めていない認可の戻り）。
 */
export function consumeAuthFlowState(): AuthFlowState | null {
  const raw = sessionStorage.getItem(STORAGE_KEY);
  sessionStorage.removeItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as Partial<AuthFlowState>;
    if (!parsed.state || !parsed.nonce || !parsed.codeVerifier) return null;
    return parsed as AuthFlowState;
  } catch {
    return null;
  }
}

/**
 * 暗号論的に安全な乱数から URL に載せられる文字列を作る。
 *
 * Math.random は予測できるので使わない。予測できると state も検証値も
 * 攻撃者が先回りして作れることになり、対策そのものが無効になる。
 */
function randomString(byteLength: number): string {
  const bytes = new Uint8Array(byteLength);
  crypto.getRandomValues(bytes);
  return base64UrlEncode(bytes);
}

/** PKCE の要約（code_challenge）を作る。 */
async function sha256Base64Url(value: string): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(value));
  return base64UrlEncode(new Uint8Array(digest));
}

/** URL に載せられる base64（+ / = を置き換え・除去）。 */
function base64UrlEncode(bytes: Uint8Array): string {
  let binary = '';
  for (const b of bytes) {
    binary += String.fromCharCode(b);
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
