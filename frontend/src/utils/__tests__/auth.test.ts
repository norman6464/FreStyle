import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getCognitoAuthUrl } from '../auth';

const mockEnv = {
  VITE_COGNITO_DOMAIN: 'test.auth.ap-northeast-1.amazoncognito.com',
  VITE_CLIENT_ID: 'test-client-id',
  VITE_REDIRECT_URI: 'http://localhost:5173/callback',
  VITE_RESPONSE_TYPE: 'code',
  VITE_SCOPE: 'openid email profile',
};

beforeEach(() => {
  vi.stubEnv('VITE_COGNITO_DOMAIN', mockEnv.VITE_COGNITO_DOMAIN);
  vi.stubEnv('VITE_CLIENT_ID', mockEnv.VITE_CLIENT_ID);
  vi.stubEnv('VITE_REDIRECT_URI', mockEnv.VITE_REDIRECT_URI);
  vi.stubEnv('VITE_RESPONSE_TYPE', mockEnv.VITE_RESPONSE_TYPE);
  vi.stubEnv('VITE_SCOPE', mockEnv.VITE_SCOPE);
});

describe('getCognitoAuthUrl', () => {
  it('正しいCognitoドメインのURLを生成する', () => {
    const url = getCognitoAuthUrl('Google');
    expect(url).toContain(`https://${mockEnv.VITE_COGNITO_DOMAIN}/oauth2/authorize`);
  });

  it('client_idパラメータが設定される', () => {
    const url = new URL(getCognitoAuthUrl('Google'));
    expect(url.searchParams.get('client_id')).toBe(mockEnv.VITE_CLIENT_ID);
  });

  it('redirect_uriパラメータが設定される', () => {
    const url = new URL(getCognitoAuthUrl('Google'));
    expect(url.searchParams.get('redirect_uri')).toBe(mockEnv.VITE_REDIRECT_URI);
  });

  it('response_typeパラメータが設定される', () => {
    const url = new URL(getCognitoAuthUrl('Google'));
    expect(url.searchParams.get('response_type')).toBe(mockEnv.VITE_RESPONSE_TYPE);
  });

  it('scopeパラメータが設定される', () => {
    const url = new URL(getCognitoAuthUrl('Google'));
    expect(url.searchParams.get('scope')).toBe(mockEnv.VITE_SCOPE);
  });

  it('identity_providerにprovider引数が設定される', () => {
    const url = new URL(getCognitoAuthUrl('Google'));
    expect(url.searchParams.get('identity_provider')).toBe('Google');
  });

  it('stateパラメータがランダムな16進文字列として生成される', () => {
    const url = new URL(getCognitoAuthUrl('Google'));
    const state = url.searchParams.get('state');
    expect(state).toBeTruthy();
    expect(state).toMatch(/^[0-9a-f]+$/);
    expect(state!.length).toBe(64); // 32 bytes * 2 hex chars
  });

  it('呼び出すたびに異なるstateが生成される', () => {
    const url1 = new URL(getCognitoAuthUrl('Google'));
    const url2 = new URL(getCognitoAuthUrl('Google'));
    expect(url1.searchParams.get('state')).not.toBe(url2.searchParams.get('state'));
  });

  it('異なるプロバイダを渡せる', () => {
    const url = new URL(getCognitoAuthUrl('LoginWithAmazon'));
    expect(url.searchParams.get('identity_provider')).toBe('LoginWithAmazon');
  });
});
