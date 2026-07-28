// Stryker (ミューテーションテスト) 設定。
// カバレッジが緑でも「実際にはバグを検出しないテスト」を、実装を機械的に書き換えて炙り出す。
// 運用: PR は light(重要な純ロジックの小サブセット) を非ブロッキング、nightly は STRYKER_FULL=1 で広め。
// 閾値 break は当面 null(gating しない)。数値が安定してから別チケットで設定する。
const full = process.env.STRYKER_FULL === '1';

/** @type {import('@stryker-mutator/api/core').PartialStrykerOptions} */
export default {
  testRunner: 'vitest',
  vitest: { configFile: 'vitest.config.js' },
  coverageAnalysis: 'perTest',
  reporters: ['clear-text', 'html', 'json'],
  htmlReporter: { fileName: 'reports/mutation/index.html' },
  jsonReporter: { fileName: 'reports/mutation/mutation.json' },
  concurrency: 4,
  timeoutMS: 60000,
  // PR light: 純粋ロジック(entities/lib, shared/lib, page の lib)に絞って高速に回す。
  // nightly full: テスト・生成物・エントリを除いた src 全体。
  mutate: full
    ? [
        'src/**/*.{ts,tsx}',
        '!src/**/*.{test,spec}.{ts,tsx}',
        '!src/**/__tests__/**',
        '!src/**/*.d.ts',
        '!src/generated/**',
        '!src/app/index.tsx',
        '!src/test/**',
      ]
    : [
        'src/entities/**/lib/**/*.ts',
        'src/shared/lib/**/*.ts',
        'src/pages/**/lib/**/*.ts',
        '!src/**/*.{test,spec}.{ts,tsx}',
        '!src/**/__tests__/**',
      ],
  // 表示上の目安(色分け)のみ。break を指定しないので CI を fail させない。
  thresholds: { high: 80, low: 60, break: null },
};
