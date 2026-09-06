import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import ResultBadge from './ResultBadge';

/**
 * 採点の結果を 1 つの札で示す。
 *
 * 「合格」は**全部のテストケースを通ったとき**だけ。1 つでも落ちたら不合格にする。
 * 部分点を出すと「だいたい合っている」で止まってしまい、直しきる動機が消えるため。
 */
const meta = {
  title: 'entities/exercise/ResultBadge',
  component: ResultBadge,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof ResultBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 全部通ったとき。 */
export const 合格: Story = {
  args: { isCorrect: true },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('全テストケース合格')).toBeVisible();
  },
};

/** 1 つでも落ちたとき。 */
export const 不合格: Story = {
  args: { isCorrect: false },
};

/** 並べて比べる。 */
export const 並べたところ: Story = {
  args: { isCorrect: true },
  render: () => (
    <div className="flex items-center gap-3">
      <ResultBadge isCorrect />
      <ResultBadge isCorrect={false} />
    </div>
  ),
};
