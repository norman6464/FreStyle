import js from '@eslint/js';
import globals from 'globals';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import { defineConfig, globalIgnores } from 'eslint/config';
import tseslint from 'typescript-eslint';

/*
 * FSD の層間依存ルール（FRESTYLE-154 / 155）。
 *
 * 公式仕様:
 *   「Slice 内のモジュールは、厳密に下の層にある Slice しか import できない」
 *   app > pages > widgets > features > entities > shared
 *   （processes は公式で非推奨のため採用しない）
 *
 * 例外: app と shared のあいだは相互に import してよい。
 *
 * 移行中（Phase 1〜6）は **警告** に留める。旧構造と新構造が混在する期間に
 * エラーにすると CI が常時赤になり、移行そのものが進まなくなるため。
 * 旧ディレクトリを撤去する Phase 7 で 'error' へ昇格する。
 */
const FSD_LAYERS = ['app', 'pages', 'widgets', 'features', 'entities', 'shared'];

/** その層が import してはいけない層（＝自分と同じか上の層）のパターンを作る。 */
function forbiddenLayersFor(layer) {
  const index = FSD_LAYERS.indexOf(layer);
  // app と shared は相互参照可（公式の例外）なので、shared からの app 参照は禁止しない。
  const upperOrSame = FSD_LAYERS.slice(0, index + 1).filter((l) => {
    if (layer === 'shared' && l === 'app') return false;
    return true;
  });
  return upperOrSame.flatMap((l) => [`@/${l}/*`, `@/${l}`]);
}

const fsdBoundaryConfigs = FSD_LAYERS.map((layer) => ({
  files: [`src/${layer}/**/*.{ts,tsx}`],
  rules: {
    'no-restricted-imports': [
      'warn',
      {
        patterns: [
          {
            group: forbiddenLayersFor(layer),
            message:
              `FSD 違反: ${layer} 層は自分と同じか上の層を import できません（下向きの一方通行）。` +
              ' 共通化したいものは下の層へ降ろすか、上の層で組み合わせてください。',
          },
        ],
      },
    ],
  },
}));

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{js,jsx,ts,tsx}'],
    extends: [
      js.configs.recommended,
      ...tseslint.configs.recommended,
      reactHooks.configs['recommended-latest'],
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
      parserOptions: {
        ecmaVersion: 'latest',
        ecmaFeatures: { jsx: true },
        sourceType: 'module',
      },
    },
    rules: {
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
    },
  },
  ...fsdBoundaryConfigs,
]);
