import { test, expect, type Page } from '@playwright/test';

/**
 * ローカルビルド + API モックによる「主要ユーザーフロー」E2E。
 *
 * authenticated.spec.ts（認証ガード/画面到達）を補完し、ユーザーが実際にたどる導線
 * （一覧 → 詳細で中身が描画される）を検証する。本番 Cognito / DB には触れない。
 */

// 認証済みユーザーとして /api/v2/** をモックする（authenticated.spec と同方針）。
async function mockAuthed(page: Page, overrides: Record<string, unknown> = {}) {
  await page.route('**/api/v2/**', (route) =>
    route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  );
  await page.route('**/api/v2/auth/me', (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ id: 1, email: 'e2e@example.com', name: 'E2E ユーザー' }),
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

test.describe('演習フロー', () => {
  // コード学習は「言語選択カード → その言語の問題一覧 → 問題」の 3 段（FRESTYLE-152）。
  test('言語選択 → 問題一覧 → リンクで詳細に遷移し演習が描画される', async ({ page }) => {
    // 一覧は ?language=go クエリ付きで叩くため glob は ** で末尾も許容する。
    // summary / 詳細 / submissions の専用パターンは後勝ちで優先される。
    // 一覧はスクロール型ページネーション化（ExercisePage = { items, hasNext, offset, limit }）
    // されたので、配列直返しではなくページオブジェクトでモックする。
    await mockAuthed(page, {
      '**/api/v2/exercises**': { items: [exercise], hasNext: false, offset: 0, limit: 20 },
      '**/api/v2/exercises/summary': [{ language: 'go', total: 1, solved: 0 }],
      '**/api/v2/exercises/e2e-fizzbuzz': { exercise, examples: [] },
      '**/api/v2/exercises/e2e-fizzbuzz/submissions': [],
    });

    // 入口は言語選択カード。
    await page.goto('/code-editor');
    const languageCard = page.getByRole('link', { name: /Go の問題一覧へ/ });
    await expect(languageCard).toBeVisible();

    await languageCard.click();

    // その言語の問題一覧へ。
    await expect(page).toHaveURL(/\/code-editor\/lang\/go/);
    const link = page.getByRole('link', { name: /E2E FizzBuzz 演習/ });
    await expect(link).toBeVisible();

    await link.click();

    await expect(page).toHaveURL(/\/code-editor\/e2e-fizzbuzz/);
    await expect(page.getByText('E2E FizzBuzz 演習').first()).toBeVisible();
  });
});

test.describe('ログイン画面', () => {
  test('未認証で /login を開くと発行者へ送るボタンが出る', async ({ page }) => {
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
    // 発行者のログイン画面へ送るボタンと、IdP 直行の 2 経路。
    await expect(page.getByRole('button', { name: 'ログインする' })).toBeVisible();
    await expect(page.getByRole('button', { name: /Google/ })).toBeVisible();
  });

  // パスワードを受け取るのは発行者のログイン画面の役目。アプリが受け取ると、
  // 二要素・ロックアウト・パスワードの強さといった発行者側の守りを
  // 素通りする経路を自分で開くことになる。
  test('ログイン画面にメールとパスワードの入力欄が無い', async ({ page }) => {
    await page.route('**/api/v2/**', (route) =>
      route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: '{"error":"unauthorized"}',
      })
    );

    await page.goto('/login');

    await expect(page.getByRole('button', { name: 'ログインする' })).toBeVisible();
    await expect(page.getByLabel('メールアドレス')).toHaveCount(0);
    await expect(page.getByLabel('パスワード', { exact: true })).toHaveCount(0);
  });
});
