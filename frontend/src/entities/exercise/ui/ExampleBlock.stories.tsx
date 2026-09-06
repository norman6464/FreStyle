import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import ExampleBlock from './ExampleBlock';
import type { MasterExerciseExample } from '../model/types';

/**
 * 演習の「入力される値」と「期待する出力」を 1 組ぶん見せる。
 *
 * 組が 2 つ以上あるときだけ見出しに番号が付く。1 つしかないのに「入力 1」と書くと、
 * 「2 はどこ？」と探させてしまうため。
 *
 * 入力が空の演習には、標準入力の決まり（末尾に改行が 1 つ入る）を添える。
 * ここを知らないまま詰まる人が多い。
 */
const meta = {
  title: 'entities/exercise/ExampleBlock',
  component: ExampleBlock,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="max-w-xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ExampleBlock>;

export default meta;
type Story = StoryObj<typeof meta>;

const example = (over: Partial<MasterExerciseExample> = {}): MasterExerciseExample => ({
  id: 1,
  exerciseId: 10,
  orderIndex: 1,
  inputText: '3\n1 2 3',
  expectedOutput: '6',
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
  ...over,
});

/** 組が 1 つだけのとき。番号は付かない。 */
export const 一組だけ: Story = {
  args: { index: 1, total: 1, example: example() },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('入力される値')).toBeVisible();
    // 1 組しかないので番号は出さない。
    await expect(canvas.queryByText('入力される値 1')).toBeNull();
  },
};

/** 組が複数あるとき。番号が付く。 */
export const 番号つき: Story = {
  args: { index: 2, total: 3, example: example({ inputText: '5\n1 2 3 4 5', expectedOutput: '15' }) },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('入力される値 2')).toBeVisible();
  },
};

/** 入力が無い演習。標準入力の決まりを添える。 */
export const 入力なし: Story = {
  args: { index: 1, total: 1, example: example({ inputText: '', expectedOutput: 'こんにちは' }) },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('ありません。')).toBeVisible();
  },
};

/** 長い入出力。折り返して収まる。 */
export const 長い入出力: Story = {
  args: {
    index: 1,
    total: 1,
    example: example({
      inputText: Array.from({ length: 8 }, (_, i) => `行 ${i + 1}: ${'あ'.repeat(30)}`).join('\n'),
      expectedOutput: Array.from({ length: 8 }, (_, i) => `結果 ${i + 1}`).join('\n'),
    }),
  },
};
