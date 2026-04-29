import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright E2E config
 *
 * - 既定 baseURL は production (https://normanblog.com)。
 *   ローカルや Preview 環境を狙うときは PLAYWRIGHT_BASE_URL を上書きする。
 * - CI では retries=2 + workers=2、ローカルは retries=0 で速く回す。
 * - 失敗時は trace + screenshot + video を artifact として残す。
 *
 * 参考:
 *   npx playwright test                       # 全 E2E
 *   npx playwright test e2e/smoke.spec.ts     # 単一 spec
 *   PLAYWRIGHT_BASE_URL=http://localhost:5173 npx playwright test
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 2 : undefined,
  reporter: process.env.CI
    ? [['html', { open: 'never' }], ['github']]
    : [['html', { open: 'never' }], ['list']],
  timeout: 30_000,
  expect: { timeout: 10_000 },
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'https://normanblog.com',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    extraHTTPHeaders: {
      // Bot 判定避けに UA 文字列を素直に送る（Playwright 既定でも入るが明示）。
      'User-Agent': 'Mozilla/5.0 (FreStyle E2E / Playwright)',
    },
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
