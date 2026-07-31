import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import axios from 'axios';
import apiClient from '../axios';

/**
 * 401 レスポンス時の挙動（トークンリフレッシュ → 失敗時のログイン画面遷移）の検証。
 *
 * 公開ページの認証確認（skipAuthRedirect）で遷移してしまうと、LP の訪問者や
 * 検索エンジンのクローラがログイン画面に飛ばされる（FRESTYLE-225 の本番回帰）。
 */

// 401 を返すアダプタ。引数の config は interceptor が受け取る config と同一なので、
// skipAuthRedirect の伝播もそのまま検証できる。
const reject401 = (config: unknown) =>
  Promise.reject({ isAxiosError: true, response: { status: 401 }, config });

let hrefSetter: ReturnType<typeof vi.fn>;
let originalLocation: Location;

beforeEach(() => {
  originalLocation = window.location;
  hrefSetter = vi.fn();
  Object.defineProperty(window, 'location', {
    configurable: true,
    value: {
      set href(value: string) {
        hrefSetter(value);
      },
      get href() {
        return 'http://localhost/';
      },
    },
  });
  // リフレッシュは常に失敗させる（未ログイン状態の再現）。
  vi.spyOn(axios, 'post').mockRejectedValue(new Error('refresh failed'));
});

afterEach(() => {
  Object.defineProperty(window, 'location', {
    configurable: true,
    value: originalLocation,
  });
  vi.restoreAllMocks();
});

describe('apiClient の 401 ハンドリング', () => {
  it('通常の呼び出しはリフレッシュ失敗でログイン画面へ遷移する', async () => {
    await expect(apiClient.get('/protected', { adapter: reject401 })).rejects.toThrow();
    expect(hrefSetter).toHaveBeenCalledWith('/login');
  });

  it('skipAuthRedirect の呼び出しはリフレッシュ失敗でも遷移しない（公開ページ用）', async () => {
    await expect(
      apiClient.get('/auth/me', { adapter: reject401, skipAuthRedirect: true }),
    ).rejects.toThrow();
    expect(hrefSetter).not.toHaveBeenCalled();
  });
});
