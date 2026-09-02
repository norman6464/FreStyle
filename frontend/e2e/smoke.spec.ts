import { test, expect } from '@playwright/test';

/**
 * スモーク E2E: 本番の配信が生きているかを、認証の前に確かめる。
 *
 * - 公開 SPA がレンダリングできる
 * - 配信のセキュリティヘッダーが正しく載っている
 *
 * **API を叩く分は落とした。** 本番の API を動かしていた実行環境ごと畳んだので、
 * api.frestyle.jp は名前解決すらできない。存在しない相手に対する検査を残すと、
 * すべての PR で赤いままになり、やがて誰も見なくなる。
 * API を出し直したときに、この spec へ戻す。
 *
 * 認証付きの導線は e2e/local/ 側で、API をモックして確かめている。
 */

test.describe('FreStyle smoke', () => {
  test('SPA がロードされ FreStyle ブランドが見える', async ({ page }) => {
    // networkidle は SPA のヘルスポーリング等で「無通信」に到達せず timeout して flake るため使わない。
    // domcontentloaded で遷移し、描画要素の出現を明示的に待つ。
    await page.goto('/', { waitUntil: 'domcontentloaded' });
    await expect(page).toHaveTitle(/FreStyle/);
    // "/" は未ログインだとログイン画面へ送られる（公開ランディングは廃止した）。
    // その公開ヘッダー(PublicHeader)に FreStyle ブランドが出ることを、
    // 本番のコールドロードを見込んだ余裕のある timeout で待つ。
    await expect(
      page.getByRole('link', { name: 'FreStyle ホーム' })
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

});
