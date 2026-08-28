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

    // "/" は公開 LP に変わったため、保護ルート(ダッシュボード)で検証する。
    await page.goto('/dashboard');

    await expect(page).not.toHaveURL(/\/login/);
  });
});

test.describe('主要画面（認証済み）', () => {
  test('コース一覧でモックしたコースの学習領域カードが描画される', async ({ page }) => {
    // /courses は「学習領域の選択カード」になった(FRESTYLE-177)。
    // モックコースのカテゴリ(database)の領域カードが出ることを確認する。
    const course = {
      id: 1,
      companyId: 1,
      createdByUserId: 1,
      title: 'E2E モックコース',
      description: 'Playwright によるモックコース',
      category: 'database',
      language: 'postgresql',
      sortOrder: 10,
      isPublished: true,
      materialCount: 0,
      completedCount: 0,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    };
    await mockAuthenticated(page, { '**/api/v2/courses': [course] });

    await page.goto('/courses');

    await expect(page).not.toHaveURL(/\/login/);
    await expect(page.getByRole('link', { name: /データベース のコース一覧へ/ })).toBeVisible();
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

  test('super_admin が trainee 向けパス（/code-editor）を開くと企業一覧へリダイレクトされる', async ({
    page,
  }) => {
    // Protected: role === 'super_admin' かつ trainee 向けパス → /admin/companies。
    await mockAuthenticated(page, {}, 'super_admin');
    await page.goto('/code-editor');
    await expect(page).toHaveURL(/\/admin\/companies/);
  });

  test('super_admin でも /notes は開ける（旧ナレッジを統合した共有の面）', async ({ page }) => {
    // 運用の手順や決めごとを書き残すのは、むしろ管理する側の仕事になる。
    await mockAuthenticated(page, {}, 'super_admin');
    await page.goto('/notes');
    await expect(page).toHaveURL(/\/notes/);
    await expect(page).not.toHaveURL(/\/admin\/companies/);
  });
});

test.describe('ノート作成導線（POST モック）', () => {
  test('所属が無ければワークスペース作成フォームが出て、名前だけで作れる', async ({ page }) => {
    await mockAuthenticated(page);

    let postBody: unknown = null;
    // /notes はナレッジ基盤（/api/v2/kb/…）ベース。所属一覧が空 → 作成フォーム表示 →
    // POST /kb/workspaces（名前のみ。slug はサーバーが自動採番）という導線を検証する。
    // mockAuthenticated の後に登録するためこの handler が優先される。
    await page.route('**/api/v2/kb/workspaces', (route) => {
      if (route.request().method() === 'POST') {
        postBody = route.request().postDataJSON();
        return route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            slug: 'w-3f2a9c',
            name: '開発チーム',
            createdAt: '2026-01-01T00:00:00Z',
          }),
        });
      }
      // GET 所属一覧: 空（まだどこにも所属していない）。
      return route.fulfill({ status: 200, contentType: 'application/json', body: '[]' });
    });

    await page.goto('/notes');
    await expect(page).toHaveURL(/\/notes/);
    // 行き止まりにしない: 作成フォームが出る。
    await expect(page.getByText(/まだワークスペースがありません/)).toBeVisible();

    await page.getByLabel('ワークスペースの名前').fill('開発チーム');
    await page.getByRole('button', { name: 'ワークスペースを作る' }).click();

    // 名前だけが送られる（URL に出る短い名前はサーバーが自動採番する）。
    await expect
      .poll(() => postBody)
      .toEqual({ name: '開発チーム' });
  });
});
