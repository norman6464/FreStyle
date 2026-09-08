import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import apiClient from '../axios';
import { getCurrentIdToken } from '@/shared/lib/auth/currentIdToken';

/**
 * Bearer ヘッダの自動付与と、401 レスポンス時の挙動（トークン更新 →
 * 失敗時のログイン画面遷移）の検証。
 *
 * 公開ページの認証確認（skipAuthRedirect）で遷移してしまうと、LP の訪問者や
 * 検索エンジンのクローラがログイン画面に飛ばされる（本番回帰の再発防止）。
 */

vi.mock('@/shared/lib/auth/currentIdToken');

// 200 を返すだけのアダプタ。config は interceptor が受け取るものと同一なので、
// リクエストに実際に付いたヘッダを検証できる。
const resolve200 = (config: { headers?: unknown }) =>
  Promise.resolve({ data: {}, status: 200, statusText: 'OK', headers: {}, config });

// 401 を返すアダプタ。
const reject401 = (config: unknown) =>
  Promise.reject({ isAxiosError: true, response: { status: 401 }, config });

let hrefSetter: ReturnType<typeof vi.fn>;
let originalLocation: Location;

beforeEach(() => {
  vi.mocked(getCurrentIdToken).mockReset();
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
});

afterEach(() => {
  Object.defineProperty(window, 'location', {
    configurable: true,
    value: originalLocation,
  });
  vi.restoreAllMocks();
});

describe('apiClient のリクエスト（Bearer ヘッダの付与）', () => {
  it('サインインしていればAuthorization: Bearerを付ける', async () => {
    vi.mocked(getCurrentIdToken).mockResolvedValue('id-token-1');
    let captured: string | undefined;
    await apiClient.get('/whoami', {
      adapter: (config) => {
        captured = config.headers?.get?.('Authorization') as string | undefined;
        return resolve200(config);
      },
    });
    expect(captured).toBe('Bearer id-token-1');
  });

  it('サインインしていなければAuthorizationヘッダを付けない', async () => {
    vi.mocked(getCurrentIdToken).mockResolvedValue(null);
    let captured: string | undefined;
    await apiClient.get('/public', {
      adapter: (config) => {
        captured = config.headers?.get?.('Authorization') as string | undefined;
        return resolve200(config);
      },
    });
    expect(captured).toBeUndefined();
  });
});

describe('apiClient の 401 ハンドリング', () => {
  it('更新後のトークンが取れれば1回だけ再試行して成功させる', async () => {
    // 実際の getCurrentIdToken は forceRefresh:true で更新した結果を裏で覚えていて、
    // 直後の（forceRefresh 無しの）呼び出しでも新しい値を返す（Firebase SDK のキャッシュ /
    // Dex セッションの保存し直しのどちらでも同じ）。1回目の呼び出しだけ古い値、
    // それ以降はすべて新しい値、という形でその挙動を模する。
    let calls = 0;
    vi.mocked(getCurrentIdToken).mockImplementation(async () => {
      calls += 1;
      return calls === 1 ? 'expired-token' : 'fresh-token';
    });

    let attempt = 0;
    const capturedHeaders: (string | undefined)[] = [];
    const adapter = (config: { headers?: { get?: (name: string) => unknown } }) => {
      attempt += 1;
      capturedHeaders.push(config.headers?.get?.('Authorization') as string | undefined);
      if (attempt === 1) return reject401(config);
      return resolve200(config as never);
    };

    const response = await apiClient.get('/protected', { adapter });

    expect(response.status).toBe(200);
    expect(attempt).toBe(2);
    // forceRefresh は 1 回だけ呼ばれる（強制更新の呼び出し）。
    expect(getCurrentIdToken).toHaveBeenCalledWith(true);
    // 2回目のリクエストは更新後のトークンを積んでいる。
    expect(capturedHeaders[1]).toBe('Bearer fresh-token');
  });

  it('更新に失敗（トークンが取れない）したらログイン画面へ遷移する', async () => {
    vi.mocked(getCurrentIdToken).mockResolvedValueOnce('expired-token').mockResolvedValueOnce(null);

    await expect(apiClient.get('/protected', { adapter: reject401 })).rejects.toThrow();
    expect(hrefSetter).toHaveBeenCalledWith('/login');
  });

  it('skipAuthRedirect の呼び出しは更新失敗でも遷移しない（公開ページ用）', async () => {
    vi.mocked(getCurrentIdToken).mockResolvedValueOnce(null).mockResolvedValueOnce(null);

    await expect(
      apiClient.get('/auth/me', { adapter: reject401, skipAuthRedirect: true } as never),
    ).rejects.toThrow();
    expect(hrefSetter).not.toHaveBeenCalled();
  });
});
