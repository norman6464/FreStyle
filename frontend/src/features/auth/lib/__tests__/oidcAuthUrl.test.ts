import { describe, it, expect, beforeEach } from 'vitest';
import { buildAuthorizeUrl, consumeAuthFlowState } from '../oidcAuthUrl';
import type { ConfiguredAuth } from '../authConfig';

// 設定は環境から読まず引数で渡す。stubEnv でお膳立てする必要が無くなった分、
// 「何を渡したらどうなるか」がテストの中だけで完結する。
const auth: ConfiguredAuth = {
  status: 'configured',
  authorizeUri: 'http://localhost:8081/oauth/v2/authorize',
  clientId: 'test-client-id',
  redirectUri: 'http://localhost:5173/login/callback',
  scope: 'openid profile email offline_access',
};

beforeEach(() => {
  sessionStorage.clear();
});

describe('buildAuthorizeUrl', () => {
  it('発行者の認可エンドポイントへ向ける', async () => {
    const url = new URL(await buildAuthorizeUrl(auth));
    expect(`${url.origin}${url.pathname}`).toBe(auth.authorizeUri);
  });

  it('client_id / redirect_uri / response_type / scope を載せる', async () => {
    const url = new URL(await buildAuthorizeUrl(auth));
    expect(url.searchParams.get('client_id')).toBe(auth.clientId);
    expect(url.searchParams.get('redirect_uri')).toBe(auth.redirectUri);
    expect(url.searchParams.get('response_type')).toBe('code');
    expect(url.searchParams.get('scope')).toBe(auth.scope);
  });

  // PKCE。要約だけを認可要求に載せ、元の値は手元に置く。
  it('S256 の code_challenge を載せ、検証値は手元に置く', async () => {
    const url = new URL(await buildAuthorizeUrl(auth));
    expect(url.searchParams.get('code_challenge_method')).toBe('S256');
    const challenge = url.searchParams.get('code_challenge');
    expect(challenge).toBeTruthy();
    // 要約そのものは URL に出るが、元の検証値は出てはいけない。
    const raw = sessionStorage.getItem('oidc.authFlow');
    expect(raw).toBeTruthy();
    const flow = JSON.parse(raw as string);
    expect(flow.codeVerifier).toBeTruthy();
    expect(url.toString()).not.toContain(flow.codeVerifier);
  });

  // **要約と検証値が対応していること。**
  // ここを見ないと、challenge に無関係な乱数を入れていても上のテストは通る。
  // 対応が崩れていると、発行者が交換の時点で必ず弾く（ログインが最後まで通らない）。
  it('code_challenge は手元の検証値の S256 要約になっている', async () => {
    const url = new URL(await buildAuthorizeUrl(auth));
    const flow = JSON.parse(sessionStorage.getItem('oidc.authFlow') as string);

    const digest = await crypto.subtle.digest(
      'SHA-256',
      new TextEncoder().encode(flow.codeVerifier),
    );
    const expected = btoa(String.fromCharCode(...new Uint8Array(digest)))
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=+$/, '');

    expect(url.searchParams.get('code_challenge')).toBe(expected);
  });

  it('state と nonce を載せ、同じ値を手元に置く', async () => {
    const url = new URL(await buildAuthorizeUrl(auth));
    const flow = JSON.parse(sessionStorage.getItem('oidc.authFlow') as string);
    expect(url.searchParams.get('state')).toBe(flow.state);
    expect(url.searchParams.get('nonce')).toBe(flow.nonce);
  });

  // 毎回同じ値だと、一度盗まれたものを使い回せる。
  it('呼ぶたびに違う値を作る', async () => {
    await buildAuthorizeUrl(auth);
    const first = JSON.parse(sessionStorage.getItem('oidc.authFlow') as string);
    await buildAuthorizeUrl(auth);
    const second = JSON.parse(sessionStorage.getItem('oidc.authFlow') as string);
    expect(second.state).not.toBe(first.state);
    expect(second.nonce).not.toBe(first.nonce);
    expect(second.codeVerifier).not.toBe(first.codeVerifier);
  });

  it('provider を渡すとその IdP へ直行する', async () => {
    const url = new URL(await buildAuthorizeUrl(auth, 'Google'));
    expect(url.searchParams.get('identity_provider')).toBe('Google');
  });

  it('provider を渡さなければ identity_provider を付けない', async () => {
    const url = new URL(await buildAuthorizeUrl(auth));
    expect(url.searchParams.has('identity_provider')).toBe(false);
  });

  it('signup を渡すと登録画面へ誘導する', async () => {
    const url = new URL(await buildAuthorizeUrl(auth, undefined, 'signup'));
    expect(url.searchParams.get('prompt')).toBe('create');
  });
});

describe('consumeAuthFlowState', () => {
  it('置いた値を取り出す', async () => {
    await buildAuthorizeUrl(auth);
    const flow = consumeAuthFlowState();
    expect(flow?.state).toBeTruthy();
    expect(flow?.nonce).toBeTruthy();
    expect(flow?.codeVerifier).toBeTruthy();
  });

  // 使い切りにする。残すと同じ値で 2 回目の交換が試せる。
  it('取り出したら消える', async () => {
    await buildAuthorizeUrl(auth);
    consumeAuthFlowState();
    expect(consumeAuthFlowState()).toBeNull();
    expect(sessionStorage.getItem('oidc.authFlow')).toBeNull();
  });

  it('置かれていなければ null', () => {
    expect(consumeAuthFlowState()).toBeNull();
  });

  // 壊れた値を「一致した」と読まない。
  it('壊れた値なら null', () => {
    sessionStorage.setItem('oidc.authFlow', '{ not json');
    expect(consumeAuthFlowState()).toBeNull();
  });

  it('値が欠けていれば null', () => {
    sessionStorage.setItem('oidc.authFlow', JSON.stringify({ state: 's' }));
    expect(consumeAuthFlowState()).toBeNull();
  });
});
