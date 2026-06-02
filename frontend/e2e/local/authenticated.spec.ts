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

// 指定 role の認証済みユーザーとして /api/v2/** をモックする。
// 個別エンドポイントを上書きできるよう overrides を受け取る。
async function mockAuthenticated(
  page: Page,
  overrides: Record<string, unknown> = {},
  role: 'trainee' | 'company_admin' | 'super_admin' = 'trainee'
) {
  // 既定: 未指定の API は空配列で 200（リスト/オブジェクトどちらの消費側も undefined 安全）。
  await page.route('**/api/v2/**', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  );
  // 認証確認: 指定 role で認証済みにする。
  await page.route('**/api/v2/auth/me', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ isAdmin: role !== 'trainee', role }),
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

test.describe('認証済み導線（trainee）', () => {
  test('ノート画面はログインに飛ばされず描画される', async ({ page }) => {
    await mockAuthenticated(page);
    await page.goto('/notes');
    await expect(page).not.toHaveURL(/\/login/);
    await expect(page).toHaveURL(/\/notes/);
  });

  test('AI チャット画面はログインに飛ばされず描画される', async ({ page }) => {
    await mockAuthenticated(page);
    await page.goto('/chat/ask-ai');
    await expect(page).not.toHaveURL(/\/login/);
    await expect(page).toHaveURL(/\/chat\/ask-ai/);
  });

  test('学習レポート画面はログインに飛ばされず描画される', async ({ page }) => {
    await mockAuthenticated(page);
    await page.goto('/reports');
    await expect(page).not.toHaveURL(/\/login/);
    await expect(page).toHaveURL(/\/reports/);
  });
});

test.describe('認証済み導線（super_admin）', () => {
  test('super_admin が企業一覧を開ける', async ({ page }) => {
    await mockAuthenticated(page, {}, 'super_admin');
    await page.goto('/admin/companies');
    await expect(page).not.toHaveURL(/\/login/);
    await expect(page).toHaveURL(/\/admin\/companies/);
  });

  test('super_admin が trainee 向けパス（/notes）を開くと企業一覧へリダイレクトされる', async ({
    page,
  }) => {
    // Protected: role === 'super_admin' かつ trainee 向けパス → /admin/companies。
    await mockAuthenticated(page, {}, 'super_admin');
    await page.goto('/notes');
    await expect(page).toHaveURL(/\/admin\/companies/);
  });
});

test.describe('ノート作成導線（POST モック）', () => {
  test('「新しいノート」で POST /notes が呼ばれ一覧に追加される', async ({ page }) => {
    await mockAuthenticated(page);

    let postCalled = false;
    // GET /notes（既存 1 件で空状態を回避）と POST /notes（作成）を method で出し分ける。
    // mockAuthenticated の後に登録するため /notes ではこの handler が優先される。
    await page.route('**/api/v2/notes', (route) => {
      if (route.request().method() === 'POST') {
        postCalled = true;
        return route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            id: 999,
            userId: 7,
            title: '無題',
            content: '',
            isPublic: false,
            isPinned: false,
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
          }),
        });
      }
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: 1,
            userId: 7,
            title: '既存ノート',
            content: '',
            isPublic: false,
            isPinned: false,
            createdAt: '2026-01-01T00:00:00Z',
            updatedAt: '2026-01-01T00:00:00Z',
          },
        ]),
      });
    });

    await page.goto('/notes');
    await expect(page).toHaveURL(/\/notes/);
    // 既存ノートが一覧に出ている（GET モックが効いている）。
    await expect(page.getByText('既存ノート').filter({ visible: true }).first()).toBeVisible();

    await page
      .getByRole('button', { name: '新しいノート' })
      .filter({ visible: true })
      .first()
      .click();

    // POST が呼ばれ、作成された「無題」ノートが一覧に追加される。
    await expect(page.getByText('無題').filter({ visible: true }).first()).toBeVisible();
    expect(postCalled).toBe(true);
  });
});
