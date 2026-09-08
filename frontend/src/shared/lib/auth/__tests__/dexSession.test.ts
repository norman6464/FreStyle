import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  saveDexSession,
  clearDexSession,
  getValidDexIdToken,
  hasDexSession,
} from '../dexSession';
import { refreshDexToken, DexTokenExchangeError } from '../oidcAuthUrl';
import type { ConfiguredAuth } from '../authConfig';
import { createMockStorage } from '@/test/mockStorage';

vi.mock('../oidcAuthUrl', async () => {
  const actual = await vi.importActual<typeof import('../oidcAuthUrl')>('../oidcAuthUrl');
  return { ...actual, refreshDexToken: vi.fn() };
});

const auth: ConfiguredAuth = {
  status: 'configured',
  authorizeUri: 'http://localhost:8081/oauth/v2/authorize',
  tokenUri: 'http://localhost:8081/oauth/v2/token',
  clientId: 'test-client-id',
  redirectUri: 'http://localhost:5173/login/callback',
  scope: 'openid profile email offline_access',
};

beforeEach(() => {
  // FreStyle では jsdom 環境の localStorage を都度スタブする方針に統一している
  // （src/test/mockStorage.ts 参照）。
  vi.stubGlobal('localStorage', createMockStorage());
  vi.clearAllMocks();
  vi.useFakeTimers();
  vi.setSystemTime(new Date('2026-01-01T00:00:00Z'));
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe('dexSession', () => {
  it('保存していなければ hasDexSession は false、getValidDexIdToken は null', async () => {
    expect(hasDexSession()).toBe(false);
    expect(await getValidDexIdToken(auth)).toBeNull();
  });

  it('保存直後は hasDexSession が true になる', () => {
    saveDexSession('id-token-1', 'refresh-1', 3600);
    expect(hasDexSession()).toBe(true);
  });

  it('期限内なら保存した ID トークンをそのまま返す（更新は呼ばない）', async () => {
    saveDexSession('id-token-1', 'refresh-1', 3600);
    const token = await getValidDexIdToken(auth);
    expect(token).toBe('id-token-1');
    expect(refreshDexToken).not.toHaveBeenCalled();
  });

  it('clearDexSession の後は無い扱いになる', () => {
    saveDexSession('id-token-1', 'refresh-1', 3600);
    clearDexSession();
    expect(hasDexSession()).toBe(false);
  });

  // 期限ギリギリのタイミングで使うと、リクエストの途中で切れて 401 になりうる。
  // 使う前に余裕を持って更新しておく。
  it('期限まで残り1分を切っていたら、使う前に更新する', async () => {
    saveDexSession('old-id-token', 'refresh-1', 60); // 60秒後に切れる
    vi.mocked(refreshDexToken).mockResolvedValue({
      idToken: 'new-id-token',
      refreshToken: 'new-refresh',
      expiresInSeconds: 3600,
    });

    vi.advanceTimersByTime(5_000); // 残り55秒 = leeway(60秒)を切っている

    const token = await getValidDexIdToken(auth);

    expect(refreshDexToken).toHaveBeenCalledWith(auth, 'refresh-1');
    expect(token).toBe('new-id-token');
  });

  it('更新後の値を保存し直す（次回は再度更新しなくてよい）', async () => {
    saveDexSession('old-id-token', 'refresh-1', 60);
    vi.mocked(refreshDexToken).mockResolvedValue({
      idToken: 'new-id-token',
      refreshToken: 'new-refresh',
      expiresInSeconds: 3600,
    });
    vi.advanceTimersByTime(5_000);
    await getValidDexIdToken(auth);

    vi.mocked(refreshDexToken).mockClear();
    const token = await getValidDexIdToken(auth);

    expect(refreshDexToken).not.toHaveBeenCalled();
    expect(token).toBe('new-id-token');
  });

  it('refresh_token を保存していなければ、期限切れ時は更新を試みず null（保存も消す）', async () => {
    saveDexSession('old-id-token', null, 60);
    vi.advanceTimersByTime(5_000);

    const token = await getValidDexIdToken(auth);

    expect(refreshDexToken).not.toHaveBeenCalled();
    expect(token).toBeNull();
    expect(hasDexSession()).toBe(false);
  });

  // 中途半端に古い値を残すと、次回も同じ失敗する更新を繰り返すだけになる。
  it('更新に失敗したら保存を消して null を返す', async () => {
    saveDexSession('old-id-token', 'refresh-1', 60);
    vi.mocked(refreshDexToken).mockRejectedValue(new DexTokenExchangeError('invalid_grant'));
    vi.advanceTimersByTime(5_000);

    const token = await getValidDexIdToken(auth);

    expect(token).toBeNull();
    expect(hasDexSession()).toBe(false);
  });

  // backend が 401 を返す理由は期限切れだけとは限らない（時計のずれ・失効等）。
  // 期限内に見えても、呼び出し側が明示的に force すれば更新を試みる。
  it('forceRefresh:true なら期限内でも更新する', async () => {
    saveDexSession('old-id-token', 'refresh-1', 3600); // 十分先まで有効
    vi.mocked(refreshDexToken).mockResolvedValue({
      idToken: 'new-id-token',
      refreshToken: 'new-refresh',
      expiresInSeconds: 3600,
    });

    const token = await getValidDexIdToken(auth, true);

    expect(refreshDexToken).toHaveBeenCalledWith(auth, 'refresh-1');
    expect(token).toBe('new-id-token');
  });

  it('壊れた値が入っていたら無い扱いにする', async () => {
    localStorage.setItem('dex.session', '{ not json');
    expect(await getValidDexIdToken(auth)).toBeNull();
    expect(hasDexSession()).toBe(false);
  });
});
