import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E config（ローカルビルド + API モック）
 *
 * 本番 (playwright.config.ts) とは別系統。`vite preview` で配信したビルド済み SPA に対し、
 * Playwright の route 機能で `/api/v2/**` をモックして「認証付き導線・主要画面」を検証する。
 * 本番 Cognito / 本番 DB に一切触れないため、CI で安全に毎回回せる。
 *
 * 重要: ビルドは VITE_API_BASE_URL='' （同一オリジン相対 /api/v2/*）で行う。index.html の
 * CSP connect-src 'self' に収め、Playwright route がモックを差し込めるようにするため。
 * cross-origin のダミーホストにすると CSP でブロックされ route に到達しない。
 *
 *   npm run e2e:local            # build 済み前提（CI は build → preview → test）
 */
export default defineConfig({
  testDir: './e2e/local',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
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
    command: 'npm run preview -- --port 4173 --strictPort',
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
