import { describe, it, expect, beforeEach, vi } from 'vitest';
import { readAuthConfig, DEFAULT_SCOPE } from '../authConfig';

const complete = {
  VITE_OIDC_AUTHORIZE_URI: 'https://issuer.test/oauth/v2/authorize',
  VITE_OIDC_CLIENT_ID: 'test-client-id',
  VITE_OIDC_REDIRECT_URI: 'https://app.test/login/callback',
  VITE_OIDC_SCOPE: 'openid profile email offline_access',
};

function stub(overrides: Record<string, string> = {}) {
  for (const [k, v] of Object.entries({ ...complete, ...overrides })) {
    vi.stubEnv(k, v);
  }
}

beforeEach(() => {
  vi.unstubAllEnvs();
});

describe('readAuthConfig', () => {
  it('揃っていれば configured を返す', () => {
    stub();
    const config = readAuthConfig();
    expect(config).toEqual({
      status: 'configured',
      authorizeUri: complete.VITE_OIDC_AUTHORIZE_URI,
      clientId: complete.VITE_OIDC_CLIENT_ID,
      redirectUri: complete.VITE_OIDC_REDIRECT_URI,
      scope: complete.VITE_OIDC_SCOPE,
    });
  });

  // scope に offline_access が無いと、発行者は更新用のトークンをそもそも出さない。
  // その状態だと、アクセストークンが切れた瞬間に全員がログイン画面へ飛ぶ。
  it('scope が空なら既定に profile / email / offline_access を含める', () => {
    stub({ VITE_OIDC_SCOPE: '' });
    const config = readAuthConfig();
    expect(config.status).toBe('configured');
    expect(DEFAULT_SCOPE.split(' ')).toEqual(
      expect.arrayContaining(['openid', 'profile', 'email', 'offline_access']),
    );
    if (config.status === 'configured') {
      expect(config.scope).toBe(DEFAULT_SCOPE);
    }
  });

  it.each([
    ['VITE_OIDC_AUTHORIZE_URI'],
    ['VITE_OIDC_CLIENT_ID'],
    ['VITE_OIDC_REDIRECT_URI'],
  ])('%s が空なら unconfigured で、その名前を挙げる', (key) => {
    stub({ [key]: '' });
    const config = readAuthConfig();
    expect(config.status).toBe('unconfigured');
    if (config.status === 'unconfigured') {
      expect(config.missing).toEqual([key]);
    }
  });

  it('複数欠けていれば全部挙げる', () => {
    stub({ VITE_OIDC_AUTHORIZE_URI: '', VITE_OIDC_CLIENT_ID: '' });
    const config = readAuthConfig();
    expect(config.status).toBe('unconfigured');
    if (config.status === 'unconfigured') {
      expect(config.missing).toEqual(['VITE_OIDC_AUTHORIZE_URI', 'VITE_OIDC_CLIENT_ID']);
    }
  });

  // **値が「有る」ことと「使える」ことは別。**
  // 空チェックだけだと、相対文字列や打ち間違えた値が素通りして、
  // 発行者の画面へ飛んで初めて（あるいは new URL が例外を投げて初めて）分かる。
  it.each([
    ['authorize', '相対パス'],
    ['ttps://issuer.test/authorize', '綴り違い'],
    ['javascript:alert(1)', 'http(s) ではない仕組み'],
    ['   ', '空白だけ'],
  ])('authorizeUri が %s（%s）なら unconfigured', (value) => {
    stub({ VITE_OIDC_AUTHORIZE_URI: value });
    const config = readAuthConfig();
    expect(config.status).toBe('unconfigured');
    if (config.status === 'unconfigured') {
      expect(config.missing).toContain('VITE_OIDC_AUTHORIZE_URI');
    }
  });

  it('redirectUri が絶対 http(s) URL でなければ unconfigured', () => {
    stub({ VITE_OIDC_REDIRECT_URI: '/login/callback' });
    const config = readAuthConfig();
    expect(config.status).toBe('unconfigured');
    if (config.status === 'unconfigured') {
      expect(config.missing).toContain('VITE_OIDC_REDIRECT_URI');
    }
  });
});
