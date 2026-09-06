import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import { withApi, withToast, routerWithParam } from '../../../../.storybook/decorators';
import ExerciseDetailPage from './ExerciseDetailPage';

/**
 * 演習 1 問の画面。問題文・入出力の例・コードを書くところ・提出。
 *
 * どの問題かは URL（`/code-editor/:language/:slug`）から取る。見本でも同じ形の道に嵌めて描く。
 *
 * 出題の形は 3 つあり、この画面がその場で切り替える。ふつうの実行、コマンドを書き取る形、
 * 見た目を作る形。**画面を分けていない**のは、戻る・進むの道筋を 3 通りに増やさないため。
 */
const meta = {
  title: 'pages/exercise-detail/ExerciseDetailPage',
  component: ExerciseDetailPage,
  parameters: { layout: 'fullscreen' },
  decorators: [withToast],
} satisfies Meta<typeof ExerciseDetailPage>;

export default meta;
type Story = StoryObj<typeof meta>;

const exercise = {
  id: 10,
  slug: 'go-slice-append',
  language: 'go',
  orderIndex: 12,
  category: '基礎',
  title: 'スライスに要素を足す',
  description: '標準入力から受け取った数値を末尾に足して、結果を出力してください。',
  starterCode: 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("ここから")\n}',
  hintText: 'append は新しいスライスを返します。',
  expectedOutput: '[1 2 3 4]',
  mode: 'execute' as const,
  explanation: '',
  difficulty: 2,
  isPublished: true,
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
};

const examples = [
  {
    id: 1,
    exerciseId: 10,
    orderIndex: 1,
    inputText: '4',
    expectedOutput: '[1 2 3 4]',
    createdAt: '2026-09-01T00:00:00Z',
    updatedAt: '2026-09-01T00:00:00Z',
  },
];

const submission = (id: number, isCorrect: boolean) => ({
  id,
  userId: 1,
  exerciseKind: 'master' as const,
  exerciseId: 10,
  submittedCode: 'package main',
  stdout: '',
  stderr: '',
  exitCode: isCorrect ? 0 : 1,
  isCorrect,
  submittedAt: '2026-09-06T09:41:00Z',
});

// 突き合わせは前から順。提出履歴は詳細の宛先を含むので先に書く。
const api = (over: Record<string, unknown> = {}) => ({
  '/exercises/go-slice-append/submissions': [],
  '/exercises/go-slice-append': { exercise, examples },
  '/code/warmup': {},
  ...over,
});

const at = routerWithParam('/code-editor/:language/:slug', '/code-editor/go/go-slice-append');

/** はじめて開いたとき。 */
export const 既定: Story = {
  decorators: [at, withApi(api())],
  play: async ({ canvasElement }) => {
    await waitFor(
      async () => {
        await expect(
          within(canvasElement).getByRole('heading', { name: 'スライスに要素を足す' }),
        ).toBeVisible();
      },
      { timeout: 5000 },
    );
  },
};

/** 過去に提出したことがあるとき。履歴が並ぶ。 */
export const 提出履歴つき: Story = {
  decorators: [
    at,
    withApi(
      api({
        '/exercises/go-slice-append/submissions': [
          submission(3, true),
          submission(2, false),
          submission(1, false),
        ],
      }),
    ),
  ],
};

/** コマンドを書き取る形の問題。 */
export const 書き取り形式: Story = {
  decorators: [
    at,
    withApi(
      api({
        '/exercises/go-slice-append': {
          exercise: {
            ...exercise,
            mode: 'qa',
            title: '停止中のコンテナも含めて一覧を出す',
            description: '停止しているものも含め、すべてのコンテナを一覧するコマンドを書いてください。',
            expectedOutput: 'docker ps -a',
            explanation: '`-a` を付けると、停止しているものも含めて全部出ます。',
            language: 'docker',
          },
          examples: [],
        },
      }),
    ),
  ],
};

/** 問題が取れなかったとき。 */
export const 取得に失敗: Story = {
  decorators: [at, withApi({})],
};
