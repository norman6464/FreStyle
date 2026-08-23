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
  category: 'database',
  language: 'postgresql',
  sortOrder: 10,
  isPublished: true,
  materialCount: 1,
  completedCount: 0,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const material = {
  id: 11,
  courseId: 1,
  title: 'E2E 教材タイトル',
  // 本文はリッチ本文（doc / tiptap JSON）が正本。先頭 h1 はタイトルの二重表示防止で
  // 本文からは除去される（stripLeadingDocTitle）。
  doc: {
    type: 'doc',
    content: [
      { type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: 'E2E 教材タイトル' }] },
      { type: 'paragraph', content: [{ type: 'text', text: '本文サンプル。' }] },
    ],
  },
  revision: 1,
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
      // 一覧はメタデータのみ、 本文は選択時に GET /teaching-materials/:id で取得する。
      '**/api/v2/courses/1/materials': [material],
      '**/api/v2/teaching-materials/11': material,
    });

    // コースは「学習領域の選択 → その領域の一覧 → 詳細」の 3 段(FRESTYLE-177)。
    await page.goto('/courses');
    await page.getByRole('link', { name: /データベース のコース一覧へ/ }).click();
    await expect(page).toHaveURL(/\/courses\/category\/database/);
    await expect(page.getByText('E2E 学習コース')).toBeVisible();

    // カード（div の onClick で navigate）を開く。
    await page.getByText('E2E 学習コース').click();

    await expect(page).toHaveURL(/\/courses\/1/);
    // 詳細はサイドバーに教材一覧があり、選択すると本文が tiptap(ProseMirror)で描画される。
    await page.getByRole('button', { name: /E2E 教材タイトル/ }).click();
    // ヘッダの h1 はタイトル。本文内の重複 h1 は除去されるため 1 つだけ。
    await expect(page.getByRole('heading', { name: 'E2E 教材タイトル', level: 1 })).toBeVisible();
    // 本文は tiptap の読み取り専用描画(.ProseMirror)に出る。
    await expect(page.locator('.ProseMirror').getByText('本文サンプル。')).toBeVisible();
  });
});

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
    // 「パスワードを表示」トグルボタンと衝突するので完全一致で入力欄だけを取る。
    await expect(page.getByLabel('パスワード', { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: /Google/ })).toBeVisible();
  });
});
