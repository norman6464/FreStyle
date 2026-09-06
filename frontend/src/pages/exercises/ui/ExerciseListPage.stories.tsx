import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import { withApi, withToast, routerWithParam } from '../../../../.storybook/decorators';
import ExerciseListPage from './ExerciseListPage';

/**
 * 選んだ言語の演習一覧。
 *
 * どの言語かは URL（`/code-editor/:language`）から取る。そのため見本でも、素のルータでは
 * なく**同じ形の道**に嵌めて描く（嵌めないと言語がずっと空になり、一覧が出ない）。
 *
 * 各行に、解いたかどうかと、何人が解いたかが出る。
 */
const meta = {
  title: 'pages/exercises/ExerciseListPage',
  component: ExerciseListPage,
  parameters: { layout: 'fullscreen' },
  decorators: [
    routerWithParam('/code-editor/:language', '/code-editor/go'),
    withToast,
    (Story) => (
      <div className="bg-surface p-6">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ExerciseListPage>;

export default meta;
type Story = StoryObj<typeof meta>;

const row = (id: number, title: string, status: '' | 'solved' | 'in_progress') => ({
  id,
  slug: `go-${id}`,
  language: 'go',
  orderIndex: id,
  category: '基礎',
  title,
  difficulty: (id % 5) + 1,
  mode: 'execute' as const,
  isPublished: true,
  status,
  stats: { totalSubmissions: id * 7, solvedUsers: id * 3 },
});

/** 一覧が出ているとき。 */
export const 既定: Story = {
  decorators: [
    withApi({
      '/exercises': {
        items: [
          row(1, 'はじめての出力', 'solved'),
          row(2, '変数と型', 'in_progress'),
          row(3, 'スライスに要素を足す', ''),
          row(4, 'マップを数える', ''),
        ],
        hasNext: false,
        offset: 0,
        limit: 20,
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('スライスに要素を足す')).toBeVisible();
    });
  },
};

/** その言語の演習がまだ無いとき。 */
export const 空: Story = {
  decorators: [
    withApi({ '/exercises': { items: [], hasNext: false, offset: 0, limit: 20 } }),
  ],
};

/** 取れなかったとき。 */
export const 取得に失敗: Story = {
  decorators: [withApi({})],
};
