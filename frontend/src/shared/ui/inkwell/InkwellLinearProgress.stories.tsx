import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import InkwellLinearProgress from './InkwellLinearProgress';

/**
 * inkwell の横長の進み具合。
 *
 * 数（0〜100）を渡すと**どこまで進んだか**を表し、渡さないと帯が流れ続ける
 * （終わりが読めない待ち時間に使う）。
 *
 * 動きを減らす設定にしている人には、流れを止めて静かな帯に切り替わる。
 */
const meta = {
  title: 'shared/inkwell/InkwellLinearProgress',
  component: InkwellLinearProgress,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="w-96">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof InkwellLinearProgress>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 終わりが読めない待ち。帯が流れ続ける。 */
export const 終わりが読めない待ち: Story = {
  args: {},
  play: async ({ canvasElement }) => {
    const bar = within(canvasElement).getByRole('progressbar');
    await expect(bar).toHaveAccessibleName('読み込み中');
    // 終わりが読めないので「いま何 %」は持たせない（読み上げが嘘をつかないように）。
    await expect(bar).not.toHaveAttribute('aria-valuenow');
  },
};

/** どこまで進んだかが分かる待ち。 */
export const 進み具合: Story = {
  args: { value: 60, 'aria-label': 'アップロードの進み具合' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('progressbar')).toHaveAttribute(
      'aria-valuenow',
      '60',
    );
  },
};

/** 0 % から 100 % まで並べて見る。 */
export const 段階: Story = {
  args: {},
  render: () => (
    <div className="flex flex-col gap-4">
      {[0, 25, 50, 75, 100].map((value) => (
        <InkwellLinearProgress key={value} value={value} aria-label={`進み具合 ${value}%`} />
      ))}
    </div>
  ),
};

/** 範囲の外の数を渡しても 0〜100 に収まる。 */
export const 範囲外は丸める: Story = {
  args: { value: 140, 'aria-label': '進み具合' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('progressbar')).toHaveAttribute(
      'aria-valuenow',
      '100',
    );
  },
};
