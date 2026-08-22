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
  test('「新しいノート」で POST /documents が呼ばれ一覧に追加される', async ({ page }) => {
    await mockAuthenticated(page);

    let postCalled = false;
    // /notes はリッチ文書（/api/v2/documents）ベース。GET 一覧（既存 1 件で空状態を回避）・
    // POST 作成・GET 個別取得（作成後に選択されエディタが本文を読む）を出し分ける。
    // mockAuthenticated の後に登録するためこの handler が優先される。
    const createdDoc = {
      id: '99999999-9999-9999-9999-999999999999',
      ownerId: 7,
      kind: 'note',
      title: '無題',
      isPublic: false,
      schemaVersion: 1,
      revision: 1,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
      doc: { type: 'doc', content: [{ type: 'paragraph' }] },
    };
    await page.route('**/api/v2/documents**', (route) => {
      const method = route.request().method();
      const isById = /\/documents\/[^/?]+/.test(route.request().url());
      if (method === 'POST') {
        postCalled = true;
        return route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify(createdDoc),
        });
      }
      if (isById && method === 'GET') {
        // 作成後に選択された文書の本文取得（doc 込み）。
        return route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(createdDoc),
        });
      }
      // GET 一覧（doc 本体なしの軽量サマリ）。
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          {
            id: '11111111-1111-1111-1111-111111111111',
            ownerId: 7,
            kind: 'note',
            title: '既存ノート',
            isPublic: false,
            schemaVersion: 1,
            revision: 1,
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
