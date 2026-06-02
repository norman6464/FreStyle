import { test, expect, type Page } from '@playwright/test';

/**
 * ローカルビルド + API モックによる「認証付き導線・主要画面」E2E。
 *
 * 本番 Cognito / DB に触れず、`/api/v2/**` を Playwright route でモックして
 * 認証ガード (AuthInitializer → Protected) と主要画面の描画を検証する。
 *
 * 認証は `GET /auth/me` のレスポンスで制御する:
 *   - 401 を返す → 未認証扱い → /login へリダイレクト
 *   - 200 + { role } を返す → 認証済み → AppShell + ページ描画
 */

// 認証済みユーザー（trainee）として /api/v2/** をモックする。
// 個別エンドポイントを上書きできるよう overrides を受け取る。
async function mockAuthenticated(
  page: Page,
  overrides: Record<string, unknown> = {}
) {
  // 既定: 未指定の API は空配列で 200（リスト/オブジェクトどちらの消費側も undefined 安全）。
  await page.route('**/api/v2/**', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  );
  // 認証確認: trainee として認証済みにする。
  await page.route('**/api/v2/auth/me', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ isAdmin: false, role: 'trainee' }),
    })
  );
  for (const [pattern, body] of Object.entries(overrides)) {
    await page.route(pattern, (route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(body),
      })
    );
  }
}

test.describe('認証ガード', () => {
  test('未認証で保護ルートを開くと /login にリダイレクトされる', async ({ page }) => {
    // すべての API を 401 にする → getCurrentUser 401 → refresh 401 → /login。
    await page.route('**/api/v2/**', (route) =>
      route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: '{"error":"unauthorized"}',
      })
    );

    await page.goto('/courses');

    await expect(page).toHaveURL(/\/login/);
  });

  test('認証済みなら保護ルートはログインに飛ばされない', async ({ page }) => {
    await mockAuthenticated(page);

    await page.goto('/');

    await expect(page).not.toHaveURL(/\/login/);
  });
});

test.describe('主要画面（認証済み）', () => {
  test('コース一覧でモックしたコースが描画される', async ({ page }) => {
    const course = {
      id: 1,
      companyId: 1,
      createdByUserId: 1,
      title: 'E2E モックコース',
      description: 'Playwright によるモックコース',
      sortOrder: 10,
      isPublished: true,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    };
    await mockAuthenticated(page, { '**/api/v2/courses': [course] });

    await page.goto('/courses');

    await expect(page).not.toHaveURL(/\/login/);
    await expect(page.getByText('E2E モックコース')).toBeVisible();
  });
});
