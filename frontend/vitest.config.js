import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig({
  plugins: [react()],
  // vite.config.js / tsconfig.json と同じ '@' → src のエイリアス（FRESTYLE-155）。
  // ここが無いと、テストだけが絶対パスを解決できず一斉に落ちる。
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test/setup.ts',
    // Playwright E2E は別 runner（npm run e2e）で実行する。
    // Vitest が e2e/*.spec.ts を拾うと @playwright/test の test.describe が
    // 「configuration から呼ばれた」扱いになって落ちるので明示的に除外する。
    exclude: ['node_modules', 'dist', 'e2e/**', '**/playwright-report/**'],
    coverage: {
      provider: 'v8',
      // 閾値ゲート: 下回ると `vitest run --coverage` が非ゼロ終了し CI を fail させる。
      // 現状 lines 88.6 / statements 87.6 / functions 86.1 / branches 83.4 を基準に、
      // 揺らぎ分のマージンを引いた floor。カバレッジ向上に合わせて適宜引き上げる。
      thresholds: {
        lines: 85,
        statements: 85,
        functions: 80,
        branches: 78,
      },
    },
  },
});
