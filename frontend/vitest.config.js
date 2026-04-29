import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test/setup.ts',
    // Playwright E2E は別 runner（npm run e2e）で実行する。
    // Vitest が e2e/*.spec.ts を拾うと @playwright/test の test.describe が
    // 「configuration から呼ばれた」扱いになって落ちるので明示的に除外する。
    exclude: ['node_modules', 'dist', 'e2e/**', '**/playwright-report/**'],
  },
});
