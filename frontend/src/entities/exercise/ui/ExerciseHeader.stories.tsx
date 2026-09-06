import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import ExerciseHeader from './ExerciseHeader';
import type { MasterExercise } from '../model/types';

/**
 * 演習の詳細画面のいちばん上（戻るリンク・題名・言語・難易度・採点の札）。
 *
 * ふつうの演習と Q&A 形式の演習で同じ見出しを使う。以前は同じ見た目を 2 か所に書いていて、
 * 片方だけ直る事故が起きたため 1 つにまとめてある。
 *
 * 採点の札は、提出したあとにだけ出る。
 */
const meta = {
  title: 'entities/exercise/ExerciseHeader',
  component: ExerciseHeader,
  parameters: { layout: 'padded' },
  decorators: [
    withRouter,
    (Story) => (
      <div className="max-w-2xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ExerciseHeader>;

export default meta;
type Story = StoryObj<typeof meta>;

const exercise = (over: Partial<MasterExercise> = {}): MasterExercise => ({
  id: 10,
  slug: 'go-slice-append',
  language: 'go',
  orderIndex: 12,
  category: '基礎',
  title: 'スライスに要素を足す',
  description: 'append を使って、渡された数値を末尾に足してください。',
  starterCode: 'package main',
  hintText: 'append は新しいスライスを返します。',
  expectedOutput: '[1 2 3 4]',
  mode: 'execute',
  explanation: '',
  difficulty: 2,
  isPublished: true,
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
  ...over,
});

/** まだ提出していないとき。採点の札は出ない。 */
export const 提出前: Story = {
  args: { exercise: exercise(), submitResult: null },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('heading', { level: 1 })).toHaveTextContent('スライスに要素を足す');
    await expect(canvas.queryByText('全テストケース合格')).toBeNull();
  },
};

/** 通ったあと。 */
export const 合格したあと: Story = {
  args: {
    exercise: exercise(),
    submitResult: { submissionId: 1, isCorrect: true, results: [] },
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('全テストケース合格')).toBeVisible();
  },
};

/** 通らなかったあと。 */
export const 不合格だったあと: Story = {
  args: {
    exercise: exercise(),
    submitResult: { submissionId: 1, isCorrect: false, results: [] },
  },
};

/** 難易度は星で表す（1〜5 に収める）。 */
export const 難易度いろいろ: Story = {
  args: { exercise: exercise(), submitResult: null },
  render: () => (
    <div className="space-y-8">
      {[1, 3, 5].map((difficulty) => (
        // header は、記事の外に置くと「ページの帯」として数えられる。並べて比べるための
        // 見本では 1 ページに複数出てしまうので、それぞれを記事の中に入れて帯にしない。
        <article key={difficulty}>
          <ExerciseHeader
            exercise={exercise({ difficulty, title: `難易度 ${difficulty} の演習` })}
            submitResult={null}
          />
        </article>
      ))}
    </div>
  ),
};

/** 題名が長いとき。折り返して、札は右端に留まる。 */
export const 長い題名: Story = {
  args: {
    exercise: exercise({
      title: '標準入力から受け取った複数行の数値を集計して、条件に合うものだけを昇順で出力する',
    }),
    submitResult: { submissionId: 1, isCorrect: true, results: [] },
  },
};
