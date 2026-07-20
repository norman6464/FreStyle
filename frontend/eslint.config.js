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

/*
 * app と shared は Slice を持たない層（公式仕様）。
 * 「App と Shared は互いに自由に import してよく、層内の Segment 同士も同様」とされているため、
 * この 2 層については自分自身と相手を禁止対象から外す。
 */
const SLICELESS_LAYERS = ['app', 'shared'];

/** その層が import してはいけない層（＝自分と同じか上の層）のパターンを作る。 */
function forbiddenLayersFor(layer) {
  const index = FSD_LAYERS.indexOf(layer);
  const upperOrSame = FSD_LAYERS.slice(0, index + 1).filter((l) => {
    // app / shared は相互参照可 + 層内の Segment 同士も可（公式の例外）。
    if (SLICELESS_LAYERS.includes(layer) && SLICELESS_LAYERS.includes(l)) return false;
    return true;
  });
  return upperOrSame.flatMap((l) => [`@/${l}/*`, `@/${l}`]);
}

/*
 * entities 層の「同一レイヤー参照」だけは `@x` 記法を例外として許可する（公式仕様）。
 *
 * 例: `entities/user` が `entities/course/@x/user` を参照する。
 * 参照される側が「誰に何を見せるか」を明示的に宣言する形になるため、
 * 無制限な Slice 間 import とは区別して許可する（出す内容は最小限に保つこと）。
 *
 * なぜ group ではなく regex なのか:
 * `group` は gitignore 記法で評価されるため「親ディレクトリが除外されていると
 * 子を再包含できない」という制約があり、`@/entities/*` を禁止したうえで
 * `!@/entities/*'/@x/*` で戻すことができない（実際に効かなかった）。
 * 否定先読みを書ける regex なら 1 つのパターンで表現できる。
 */
const ENTITIES_SAME_LAYER_REGEX = '^@/entities(?:/(?![^/]+/@x/).*)?$';

/*
 * テストは層構造の対象外にする。
 * テストは対象を描画するために上位層の Provider でラップすることが正当にあり
 * （例: pages のテストが app の ToastProvider を使う）、これを違反として扱うと
 * 実装の依存グラフとは無関係なノイズになるため。
 */
const TEST_FILE_PATTERNS = ['**/__tests__/**', '**/*.test.{ts,tsx}', '**/*.spec.{ts,tsx}'];

const BOUNDARY_MESSAGE = (layer) =>
  `FSD 違反: ${layer} 層は自分と同じか上の層を import できません（下向きの一方通行）。` +
  ' 共通化したいものは下の層へ降ろすか、上の層で組み合わせてください。';

const fsdBoundaryConfigs = FSD_LAYERS
  // app は最上位なので禁止対象が空になる。ESLint は空の group を受け付けないため設定自体を出さない。
  .filter((layer) => forbiddenLayersFor(layer).length > 0)
  .map((layer) => ({
    files: [`src/${layer}/**/*.{ts,tsx}`],
    ignores: TEST_FILE_PATTERNS,
    rules: {
      'no-restricted-imports': [
        'warn',
        {
          patterns:
            layer === 'entities'
              ? [
                  // 上位層（pages / widgets / features）は無条件に禁止。
                  {
                    group: forbiddenLayersFor(layer).filter((p) => !p.startsWith('@/entities')),
                    message: BOUNDARY_MESSAGE(layer),
                  },
                  // 同一レイヤー（entity 同士）は `@x` 経由のみ許可。
                  {
                    regex: ENTITIES_SAME_LAYER_REGEX,
                    message:
                      'FSD 違反: entity 同士は直接 import できません。' +
                      ' どうしても参照し合う場合は `@x` 記法（entities/<相手>/@x/<自分>）で明示してください。' +
                      ' 自分の Slice 内は相対パスで参照します。',
                  },
                ]
              : [{ group: forbiddenLayersFor(layer), message: BOUNDARY_MESSAGE(layer) }],
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
