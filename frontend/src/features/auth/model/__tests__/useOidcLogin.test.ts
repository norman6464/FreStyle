import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useOidcLogin } from '../useOidcLogin';
import { logger } from '@/shared/lib/logger';

vi.mock('@/shared/lib/logger', () => ({
  logger: { error: vi.fn(), warn: vi.fn(), info: vi.fn(), debug: vi.fn() },
}));

const DEX = {
  VITE_OIDC_AUTHORIZE_URI: 'http://localhost:5556/dex/auth',
  VITE_OIDC_TOKEN_URI: 'http://localhost:5556/dex/token',
  VITE_OIDC_CLIENT_ID: 'frestyle-local',
  VITE_OIDC_REDIRECT_URI: 'http://localhost:5173/login/callback',
};

const FIREBASE = {
  VITE_FIREBASE_API_KEY: 'test-api-key',
  VITE_FIREBASE_AUTH_DOMAIN: 'frestyle-507912.firebaseapp.com',
  VITE_FIREBASE_PROJECT_ID: 'frestyle-507912',
};

function stub(vars: Record<string, string>) {
  for (const [k, v] of Object.entries(vars)) vi.stubEnv(k, v);
}

// 手元の .env が読み込まれているので、何も stub しないと「欠けている」状態を作れない。
// 見たい状態を毎回明示的に組み立てる。
function clearAll() {
  for (const k of [...Object.keys(DEX), ...Object.keys(FIREBASE)]) vi.stubEnv(k, '');
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.unstubAllEnvs();
});

describe('useOidcLogin', () => {
  it('Dex の設定が揃っていれば start を持つ', () => {
    clearAll();
    stub(DEX);
    const { result } = renderHook(() => useOidcLogin());
    expect(result.current.available).toBe(true);
  });

  it('Dex の設定が欠けていれば start を持たず、欠けている名前を返す', () => {
    clearAll();
    const { result } = renderHook(() => useOidcLogin());
    expect(result.current.available).toBe(false);
    if (!result.current.available) {
      expect(result.current.missing).toContain('VITE_OIDC_CLIENT_ID');
    }
  });

  // 本番は GCIP の設定だけを焼き込む。Dex 向けの設定が無いのは正常な状態なので、
  // それを異常として記録してはいけない。記録すると「ログインは動いているのに
  // 認可の設定が揃っていないという記録が全ページ表示ごとに残る」ことになり、
  // 後から調べる人を誤らせる。
  it('Firebase が設定済みなら、Dex 未設定でもエラーを記録しない', () => {
    clearAll();
    stub(FIREBASE);
    renderHook(() => useOidcLogin());
    expect(logger.error).not.toHaveBeenCalled();
  });

  it('どの発行者も設定されていないときだけエラーを記録する', () => {
    clearAll();
    renderHook(() => useOidcLogin());
    expect(logger.error).toHaveBeenCalledTimes(1);
    expect(vi.mocked(logger.error).mock.calls[0][0]).toContain('どの発行者の設定も揃っていない');
  });
});
