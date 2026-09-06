import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import InkwellCircularProgress from './InkwellCircularProgress';

/**
 * inkwell の丸い進み具合。
 *
 * 横長のもの（InkwellLinearProgress）との使い分けは置き場所で決める。ボタンの中や
 * 一覧の行など**幅が取れないところ**は丸、画面の上端に渡すなら横長。
 */
const meta = {
  title: 'shared/inkwell/InkwellCircularProgress',
  component: InkwellCircularProgress,
  parameters: { layout: 'centered' },
} satisfies Meta<typeof InkwellCircularProgress>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 終わりが読めない待ち。回り続ける。 */
export const 終わりが読めない待ち: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('progressbar')).toHaveAccessibleName('読み込み中');
  },
};

/** どこまで進んだかが分かる待ち。 */
export const 進み具合: Story = {
  args: { value: 70, 'aria-label': '読み込みの進み具合' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('progressbar')).toHaveAttribute(
      'aria-valuenow',
      '70',
    );
  },
};

/** 大きさと線の太さを変える。 */
export const 大きさと太さ: Story = {
  args: {},
  render: () => (
    <div className="flex items-center gap-6">
      <InkwellCircularProgress size={24} thickness={3} value={40} aria-label="小" />
      <InkwellCircularProgress size={40} thickness={4} value={40} aria-label="中" />
      <InkwellCircularProgress size={64} thickness={6} value={40} aria-label="大" />
    </div>
  ),
};

/** 0 % から 100 % まで並べて見る。 */
export const 段階: Story = {
  args: {},
  render: () => (
    <div className="flex items-center gap-4">
      {[0, 25, 50, 75, 100].map((value) => (
        <InkwellCircularProgress key={value} value={value} aria-label={`進み具合 ${value}%`} />
      ))}
    </div>
  ),
};
