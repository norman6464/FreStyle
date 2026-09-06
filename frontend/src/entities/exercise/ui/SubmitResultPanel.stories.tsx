import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import SubmitResultPanel from './SubmitResultPanel';
import type { ExerciseTestCaseResult } from '../model/types';

/**
 * 提出したあとの、テストケースごとの採点結果。
 *
 * 最初は 1 行ずつ畳んである。開くと、その回の入力・期待した出力・実際の出力が出る。
 * 全部を開いたまま出すと、通ったケースの中身で画面が埋まり、**落ちた 1 件**が埋もれる。
 */
const meta = {
  title: 'entities/exercise/SubmitResultPanel',
  component: SubmitResultPanel,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="max-w-2xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SubmitResultPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

const testCase = (over: Partial<ExerciseTestCaseResult> = {}): ExerciseTestCaseResult => ({
  orderIndex: 1,
  input: '3\n1 2 3',
  expectedOutput: '6',
  actualOutput: '6',
  stderr: '',
  passed: true,
  ...over,
});

/** 全部通ったとき。 */
export const 全部合格: Story = {
  args: {
    results: [testCase(), testCase({ orderIndex: 2, input: '2\n5 5', expectedOutput: '10', actualOutput: '10' })],
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/2\/2 合格/)).toBeVisible();
  },
};

/** 一部が落ちたとき。落ちた行が目に付くように色を変える。 */
export const 一部が不合格: Story = {
  args: {
    results: [
      testCase(),
      testCase({
        orderIndex: 2,
        input: '0',
        expectedOutput: '0',
        actualOutput: '',
        passed: false,
      }),
      testCase({ orderIndex: 3 }),
    ],
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/2\/3 合格/)).toBeVisible();
  },
};

/** 行を開いて中身を見る。 */
export const 開いて中身を見る: Story = {
  args: {
    results: [
      testCase({
        orderIndex: 1,
        expectedOutput: '6',
        actualOutput: '5',
        passed: false,
      }),
    ],
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByText('テストケース 1'));
    await expect(canvas.getByText('期待出力')).toBeVisible();
    await expect(canvas.getByText('実際の出力')).toBeVisible();
  },
};

/** 実行時にエラーが出たとき。stderr も添える。 */
export const エラーつき: Story = {
  args: {
    results: [
      testCase({
        orderIndex: 1,
        actualOutput: '',
        stderr: 'panic: runtime error: index out of range [3] with length 3',
        passed: false,
      }),
    ],
  },
};
