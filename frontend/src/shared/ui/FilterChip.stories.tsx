import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import FilterChip from './FilterChip';

/**
 * 一覧を絞り込む、丸い切り替えボタン。
 *
 * 押されているかどうかは色だけでなく `aria-pressed` でも伝わる。色が見えない人にも
 * 「いまこれで絞り込んでいる」が分かるようにするため。
 */
const meta = {
  title: 'shared/FilterChip',
  component: FilterChip,
  parameters: { layout: 'centered' },
  args: { onClick: fn() },
} satisfies Meta<typeof FilterChip>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 選んでいないとき。 */
export const 未選択: Story = {
  args: { label: 'すべて', active: false },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('button')).toHaveAttribute(
      'aria-pressed',
      'false',
    );
  },
};

/** 選んでいるとき。既定では brand 色。 */
export const 選択中: Story = {
  args: { label: 'すべて', active: true },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('button')).toHaveAttribute('aria-pressed', 'true');
  },
};

/** 色を指定したとき（言語ごとの色などを渡す）。 */
export const 色を指定: Story = {
  args: {
    label: 'Go',
    active: true,
    activeClass: 'bg-sky-500/15 text-sky-800 border-sky-500/30',
  },
};

/** 実際の並び。1 つだけが選ばれている状態。 */
export const 並べたところ: Story = {
  args: { label: 'すべて', active: true },
  render: (args) => (
    <div className="flex flex-wrap gap-2">
      <FilterChip {...args} label="すべて" active />
      <FilterChip {...args} label="Go" active={false} />
      <FilterChip {...args} label="TypeScript" active={false} />
      <FilterChip {...args} label="Docker" active={false} />
    </div>
  ),
};

/** 押すと onClick が呼ばれる（選択状態そのものは親が持つ）。 */
export const 押したとき: Story = {
  args: { label: 'Go', active: false },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: 'Go' }));
    await expect(args.onClick).toHaveBeenCalledTimes(1);
  },
};
