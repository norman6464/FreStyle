import { test, expect } from '@playwright/test';

/**
 * 公開ページ（未ログインの訪問者・検索エンジンのクローラ向け）の E2E。
 *
 * FRESTYLE-225 の本番回帰の再発防止:
 * 公開トップは「ログイン済みならダッシュボードへ送る」ために /auth/me を呼ぶが、
 * 未ログインの 401 でトークンリフレッシュが走り、その失敗で axios interceptor が
 * /login へ強制遷移していた。結果、公開 LP が誰にも（Googlebot にも）見えなくなっていた。
 */

test.describe('公開トップ（未ログイン）', () => {
  test('未ログインでもログイン画面へ飛ばされず LP を表示し続ける', async ({ page }) => {
    // 認証系を含むすべての API を 401 にする（未ログイン状態の再現）。
    await page.route('**/api/v2/**', (route) =>
      route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: '{"error":"unauthorized"}',
      })
    );
    // ヘルスチェックだけは 200。ここを 401 にするとバックエンド障害と判定され
    // メンテナンス画面になり、「未ログイン訪問者」の検証にならないため
    // （後から登録した route が優先される）。
    await page.route('**/api/v2/health', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{"status":"ok"}' })
    );

    await page.goto('/');

    // ヒーロー見出しが出ていること（LP が描画されている）。
    await expect(
      page.getByRole('heading', { level: 1, name: /新卒ITエンジニア向け研修プラットフォーム/ })
    ).toBeVisible();

    // リダイレクトは非同期に起きるため、猶予を置いてから URL を確認する。
    await page.waitForTimeout(2000);
    await expect(page).not.toHaveURL(/\/login/);
    await expect(
      page.getByRole('heading', { level: 1, name: /新卒ITエンジニア向け研修プラットフォーム/ })
    ).toBeVisible();
  });

  test('ログイン済みならダッシュボードへ送られる', async ({ page }) => {
    await page.route('**/api/v2/**', (route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
    );
    await page.route('**/api/v2/auth/me', (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ isAdmin: false, role: 'trainee' }),
      })
    );

    await page.goto('/');

    await expect(page).toHaveURL(/\/dashboard/);
  });
});
