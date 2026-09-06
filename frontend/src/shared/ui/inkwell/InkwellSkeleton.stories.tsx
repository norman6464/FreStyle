import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import InkwellSkeleton from './InkwellSkeleton';

/**
 * 中身が届くまでのあいだ置いておく、灰色の仮の形。
 *
 * ぐるぐる（スピナー）との違いは「**これから何が出るか**」が分かること。行が 3 本なら
 * 文が 3 行来る、と読めるので、届いたときに画面が飛び跳ねない。
 */
const meta = {
  title: 'shared/inkwell/InkwellSkeleton',
  component: InkwellSkeleton,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="w-80">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof InkwellSkeleton>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 文字の行。 */
export const 行: Story = {
  args: { variant: 'text' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('status')).toHaveAccessibleName('読み込み中');
  },
};

/** 四角（画像や図の場所）。 */
export const 四角: Story = {
  args: { variant: 'rect' },
};

/** 丸（人の絵の場所）。 */
export const 丸: Story = {
  args: { variant: 'circle' },
};

/** 大きさを指定する。 */
export const 大きさを指定: Story = {
  args: { variant: 'rect', width: 240, height: 80 },
};

/** 実際の使い方。届いたあとの形に合わせて並べる。 */
export const 組み合わせ: Story = {
  args: {},
  render: () => (
    <div className="flex gap-3">
      <InkwellSkeleton variant="circle" />
      <div className="flex-1 space-y-2">
        <InkwellSkeleton variant="text" width="60%" />
        <InkwellSkeleton variant="text" />
        <InkwellSkeleton variant="text" width="80%" />
      </div>
    </div>
  ),
};
