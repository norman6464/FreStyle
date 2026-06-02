import { test, expect } from '@playwright/test';

/**
 * スモーク E2E: 認証前でも検証可能な「サイトが生きているか」を網羅的に確認する。
 *
 * - 公開 SPA がレンダリングできる
 * - CloudFront 経由のセキュリティヘッダーが正しく配信されている
 * - API ヘルスチェックが 200 を返す
 * - 認証必須エンドポイントが Cookie 無しで 401 を返す
 *
 * 認証付きフロー (ログイン後の管理画面 / ノート CRUD 等) は Cognito Hosted UI に
 * 依存するため、別 spec で `storageState` 経由の事前認証パターンを使う想定。
 */

const API_BASE = 'https://api.normanblog.com';

test.describe('FreStyle smoke', () => {
  test('SPA がロードされ FreStyle ロゴ / ログイン誘導が見える', async ({ page }) => {
    // networkidle は SPA のヘルスポーリング等で「無通信」に到達せず timeout して flake るため使わない。
    // domcontentloaded で遷移し、描画要素（ログインフォーム）の出現を明示的に待つ。
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page).toHaveTitle(/FreStyle/);
    // 未認証アクセスは /login にリダイレクトされ、ログインフォームが描画される。
    // 本番のコールドロード + 認証チェックを見込んで余裕を持った timeout で待つ。
    await expect(
      page.getByRole('form', { name: 'ログインフォーム' })
    ).toBeVisible({ timeout: 20_000 });
  });

  test('CloudFront セキュリティヘッダーが配信される', async ({ request }) => {
    const res = await request.get('/');
    const headers = res.headers();
    expect(headers['strict-transport-security']).toMatch(/max-age=\d+/);
    expect(headers['x-frame-options']).toBe('DENY');
    expect(headers['x-content-type-options']).toBe('nosniff');
    expect(headers['referrer-policy']).toBe('strict-origin-when-cross-origin');
    expect(headers['permissions-policy']).toContain('camera=()');
  });

  test('CSP meta タグが index.html に含まれている', async ({ request }) => {
    const res = await request.get('/');
    const html = await res.text();
    expect(html).toMatch(/<meta http-equiv="Content-Security-Policy"/);
    expect(html).toContain("script-src 'self'");
    expect(html).toContain('upgrade-insecure-requests');
  });

  test('API /api/v2/health は 200 を返す', async ({ request }) => {
    const res = await request.get(`${API_BASE}/api/v2/health`);
    expect(res.status()).toBe(200);
  });

  test('認証必須エンドポイントは Cookie 無で 401 を返す', async ({ request }) => {
    // PR-A で `/api/v2/score-goals` 系は撤去済 → 認可ガードの代表として
    //   /auth/me / /notes / /profile/me / /notifications / /ai-chat/sessions を確認する。
    for (const path of [
      '/api/v2/auth/me',
      '/api/v2/notes',
      '/api/v2/profile/me',
      '/api/v2/ai-chat/sessions',
      '/api/v2/notifications',
    ]) {
      const res = await request.get(`${API_BASE}${path}`);
      expect.soft(res.status(), `${path} should be 401`).toBe(401);
    }
  });

  test('SockJS フォールバック路は廃止済（404 / 401）', async ({ request }) => {
    // PR #1557 で sockJSInfo handler を廃止したので /info 系は 404 になる。
    const ws = await request.get(`${API_BASE}/api/v2/ws/ai-chat/info`);
    expect([401, 404]).toContain(ws.status());
  });
});
