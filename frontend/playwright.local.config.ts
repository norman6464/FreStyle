import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E config（ローカルビルド + API モック）
 *
 * 本番 (playwright.config.ts) とは別系統。`vite preview` で配信したビルド済み SPA に対し、
 * Playwright の route 機能で `/api/v2/**` をモックして「認証付き導線・主要画面」を検証する。
 * 本番の認証基盤 / 本番 DB に一切触れないため、CI で安全に毎回回せる。
 *
 * 重要: ビルドは VITE_API_BASE_URL='' （同一オリジン相対 /api/v2/*）で行う。index.html の
 * CSP connect-src 'self' に収め、Playwright route がモックを差し込めるようにするため。
 * cross-origin のダミーホストにすると CSP でブロックされ route に到達しない。
 *
 *   pnpm run e2e:local           # build 済み前提（CI は build → preview → test）
 */
export default defineConfig({
  testDir: './e2e/local',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  // 11 テストなので複数 worker で安全に並列化できる（route モックはページ / コンテキスト
  // 単位に閉じており、worker をまたいで共有する状態は無い）。CI 実測: workers 1 で 7.3s、
  // 4 で 3.2〜4.1s。ジョブ全体(1〜1.5分)の大半は checkout / install / build の
  // セットアップ時間で、テスト本体を増やしても短縮できるのはこの数秒だけ
  // （ジョブを machine 単位で分割する shard 化は、この規模だとセットアップの重複コストが
  // 上回り逆に遅くなるため見送っている）。
  workers: process.env.CI ? 4 : undefined,
  reporter: process.env.CI
    ? [['html', { open: 'never' }], ['github']]
    : [['list']],
  timeout: 30_000,
  expect: { timeout: 10_000 },
  use: {
    baseURL: 'http://localhost:4173',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  // ビルド済み dist/ を vite preview で配信する（SPA history fallback 込み）。
  webServer: {
    // pnpm は `--` を引数として素通しするため、付けると vite が `--` を受け取って
    // 後ろの --port が効かない（npm とは違う）。区切りは書かない。
    command: 'pnpm run preview --port 4173 --strictPort',
    url: 'http://localhost:4173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
