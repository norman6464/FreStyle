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
 * - **nonce**: 同じことを id_token の中身でも確かめる。backend はもうこの
 *   コード交換に立ち会わない（Bearer の検証だけを行う）ため、照合は
 *   id_token を受け取った側＝フロント自身が `verifyIdTokenNonce` で行う。
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

/** 発行者のトークン応答から、この先必要な3値だけを取り出したもの。 */
export type DexTokenResponse = {
  idToken: string;
  /** 発行者が offline_access を許さない、または返さない構成では null。 */
  refreshToken: string | null;
  expiresInSeconds: number;
};

/**
 * トークン交換に失敗したことを表す。
 *
 * `Error` をそのまま投げず専用の型にしているのは、呼び出し側（login-callback）が
 * 「発行者が拒んだ（コードの使い回し等）」と「ネットワーク自体が繋がらない」を
 * 区別せずに同じ案内文言へまとめてよいためで、型を分ける意味がここでは無い。
 * それでも `instanceof Error` で拾えるようにしておく。
 */
export class DexTokenExchangeError extends Error {}

/**
 * 認可コードを ID トークンに交換する。
 *
 * backend はもう認可コードの交換を仲介しない（Bearer の ID トークン検証だけを行う）。
 * Dex は公開クライアント（`public: true`、client_secret を持たない）で PKCE を必須にしている
 * ため、ブラウザから直接 `cfg.tokenUri` を叩いても安全（code_verifier を知っているのが
 * このブラウザだけであることが交換の正当性を保証する）。
 *
 * Dex 側に `web.allowedOrigins` を設定していないと、ここは CORS で失敗する
 * （`docker/idp/config.yaml` 参照）。
 */
export async function exchangeCodeForToken(
  cfg: ConfiguredAuth,
  code: string,
  codeVerifier: string,
): Promise<DexTokenResponse> {
  return requestToken(cfg.tokenUri, {
    grant_type: 'authorization_code',
    code,
    redirect_uri: cfg.redirectUri,
    client_id: cfg.clientId,
    code_verifier: codeVerifier,
  });
}

/**
 * refresh_token で ID トークンを更新する。
 *
 * ID トークン自体には更新の手段が無く（署名済みで書き換えられない）、
 * 発行者に対して refresh_token を渡し新しい ID トークンを発行してもらう。
 */
export async function refreshDexToken(cfg: ConfiguredAuth, refreshToken: string): Promise<DexTokenResponse> {
  return requestToken(cfg.tokenUri, {
    grant_type: 'refresh_token',
    refresh_token: refreshToken,
    client_id: cfg.clientId,
  });
}

async function requestToken(tokenUri: string, body: Record<string, string>): Promise<DexTokenResponse> {
  let response: Response;
  try {
    response = await fetch(tokenUri, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams(body),
    });
  } catch (err) {
    throw new DexTokenExchangeError('発行者に接続できませんでした。', { cause: err });
  }

  if (!response.ok) {
    // 本文にエラーの詳細（invalid_grant 等）が入ることがあるが、利用者に出す文言は
    // 一律にする。ここで理由を出し分けると、コードの使い回し・期限切れ・設定ミスの
    // どれなのかを外部から探る手がかりを与えてしまう。
    throw new DexTokenExchangeError(`token endpoint returned ${response.status}`);
  }

  let data: { id_token?: string; refresh_token?: string; expires_in?: number };
  try {
    data = await response.json();
  } catch (err) {
    throw new DexTokenExchangeError('発行者の応答を解釈できませんでした。', { cause: err });
  }

  if (!data.id_token) {
    throw new DexTokenExchangeError('発行者の応答に id_token がありません。');
  }

  return {
    idToken: data.id_token,
    refreshToken: data.refresh_token ?? null,
    expiresInSeconds: data.expires_in ?? 3600,
  };
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

/**
 * id_token（JWT）の nonce クレームが、この端末が認可要求に載せた値と一致するかを確かめる。
 *
 * 署名検証はしない（id_token 自体は Dex からブラウザへ直接届いており、経路は
 * `exchangeCodeForToken` が HTTPS/CORS で守っている。署名検証は backend が
 * Bearer として受け取った時点で行う）。ここで見るのは「戻ってきた id_token が
 * “自分がさっき始めた認可要求” に対応する応答か」だけ——なりすまし・古い
 * トークンの再利用に対する防御で、署名の真正性とは別の観点。
 */
export function verifyIdTokenNonce(idToken: string, expectedNonce: string): boolean {
  const payloadSegment = idToken.split('.')[1];
  if (!payloadSegment) return false;
  try {
    const payload = JSON.parse(decodeBase64UrlToUtf8(payloadSegment)) as { nonce?: unknown };
    return payload.nonce === expectedNonce;
  } catch {
    return false;
  }
}

/** base64url セグメントを UTF-8 文字列へ戻す（JWT の header/payload デコード用）。 */
function decodeBase64UrlToUtf8(segment: string): string {
  const base64 = segment.replace(/-/g, '+').replace(/_/g, '/');
  const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=');
  const binary = atob(padded);
  const bytes = Uint8Array.from(binary, (c) => c.charCodeAt(0));
  return new TextDecoder().decode(bytes);
}
