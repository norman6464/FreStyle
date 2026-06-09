import { test, expect, type Page } from '@playwright/test';

/**
 * ローカルビルド + API モックによる「主要ユーザーフロー」E2E。
 *
 * authenticated.spec.ts（認証ガード/画面到達）を補完し、ユーザーが実際にたどる導線
 * （一覧 → 詳細で中身が描画される）を検証する。本番 Cognito / DB には触れない。
 */

// 指定 role の認証済みユーザーとして /api/v2/** をモックする（authenticated.spec と同方針）。
async function mockAuthed(
  page: Page,
  overrides: Record<string, unknown> = {},
  role: 'trainee' | 'company_admin' | 'super_admin' = 'trainee'
) {
  await page.route('**/api/v2/**', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  );
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

const course = {
  id: 1,
  companyId: 1,
  createdByUserId: 1,
  title: 'E2E 学習コース',
  description: 'Playwright によるフロー検証用コース',
  sortOrder: 10,
  isPublished: true,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const material = {
  id: 11,
  courseId: 1,
  title: 'E2E 教材タイトル',
  content: '# E2E 教材タイトル\n\n本文サンプル。',
  sortOrder: 10,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const exercise = {
  id: 1,
  slug: 'e2e-fizzbuzz',
  language: 'go',
  orderIndex: 1,
  category: 'basic',
  title: 'E2E FizzBuzz 演習',
  description: '1 から N まで出力する',
  starterCode: 'package main',
  hintText: '',
  expectedOutput: '',
  mode: 'execute' as const,
  explanation: '',
  difficulty: 1,
  isPublished: true,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  // 一覧 API(MasterExerciseWithStatus)で必要。詳細 API では余分だが無害。
  status: '' as const,
  stats: { totalSubmissions: 0, solvedUsers: 0 },
};

test.describe('コース学習フロー', () => {
  test('コース一覧 → カードを開く → 詳細で教材が描画される', async ({ page }) => {
    await mockAuthed(page, {
      '**/api/v2/courses': [course],
      '**/api/v2/courses/1': course,
      '**/api/v2/courses/1/materials': [material],
    });

    await page.goto('/courses');
    await expect(page.getByText('E2E 学習コース')).toBeVisible();

    // カード（div の onClick で navigate）を開く。
    await page.getByText('E2E 学習コース').click();

    await expect(page).toHaveURL(/\/courses\/1/);
    // 詳細はサイドバーに教材一覧があり、選択すると本文(h1)に表示される。
    await page.getByRole('button', { name: /E2E 教材タイトル/ }).click();
    // 本文ヘッダ h1 と markdown 本文内 h1 の 2 つが出るため first を取る。
    await expect(
      page.getByRole('heading', { name: 'E2E 教材タイトル', level: 1 }).first()
    ).toBeVisible();
  });
});

test.describe('演習フロー', () => {
  test('演習一覧 → リンクで詳細に遷移し演習が描画される', async ({ page }) => {
    // 一覧は ?language=php クエリ付きで叩くため glob は ** で末尾も許容する。
    // 詳細/submissions の専用パターンは後勝ちで優先される。
    await mockAuthed(page, {
      '**/api/v2/exercises**': [exercise],
      '**/api/v2/exercises/e2e-fizzbuzz': { exercise, examples: [] },
      '**/api/v2/exercises/e2e-fizzbuzz/submissions': [],
    });

    await page.goto('/code-editor');
    const link = page.getByRole('link', { name: /E2E FizzBuzz 演習/ });
    await expect(link).toBeVisible();

    await link.click();

    await expect(page).toHaveURL(/\/code-editor\/e2e-fizzbuzz/);
    await expect(page.getByText('E2E FizzBuzz 演習').first()).toBeVisible();
  });
});

test.describe('ログイン画面', () => {
  test('未認証で /login を開くとメール/パスワードのログインフォームが表示される', async ({ page }) => {
    // すべての API を 401 にして未認証状態にする。
    await page.route('**/api/v2/**', (route) =>
      route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: '{"error":"unauthorized"}',
      })
    );

    await page.goto('/login');

    await expect(page).toHaveURL(/\/login/);
    // メール/パスワードフォーム + Google(Hosted UI) の 2 経路。
    await expect(page.getByRole('form', { name: 'ログインフォーム' })).toBeVisible();
    await expect(page.getByLabel('メールアドレス')).toBeVisible();
    await expect(page.getByLabel('パスワード')).toBeVisible();
    await expect(page.getByRole('button', { name: /Google/ })).toBeVisible();
  });
});
