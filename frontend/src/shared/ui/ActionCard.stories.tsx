import type { Meta, StoryObj } from '@storybook/react-vite';
import { AcademicCapIcon, DocumentTextIcon } from '@heroicons/react/24/outline';
import { expect, fn, userEvent, within } from 'storybook/test';
import { withRouter } from '../../../.storybook/decorators';
import ActionCard from './ActionCard';

/**
 * 「次に何をすればいいか」を示す大きなカード。
 *
 * 行き先がある（`to`）ならリンクに、その場で何かする（`onClick`）ならボタンになる。
 * **どちらか一方**しか渡せない — リンクなのに押しても移らない、という食い違いを型で防いでいる。
 *
 * `emphasis="primary"` は 1 画面に 1 枚だけ。並べると、どれが本命か分からなくなる。
 */
const meta = {
  title: 'shared/ActionCard',
  component: ActionCard,
  parameters: { layout: 'padded' },
  decorators: [withRouter],
} satisfies Meta<typeof ActionCard>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 行き先つき（リンクになる）。 */
export const リンク: Story = {
  args: {
    to: '/code-editor',
    title: '演習をはじめる',
    description: '言語を選んで、手を動かしながら学びます。',
    icon: <AcademicCapIcon className="h-5 w-5" aria-hidden="true" />,
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('link')).toHaveAttribute('href', '/code-editor');
  },
};

/** その場で何かする形（ボタンになる）。 */
export const ボタン: Story = {
  args: {
    onClick: fn(),
    title: 'ページを作る',
    description: 'まっさらなページから書きはじめます。',
    icon: <DocumentTextIcon className="h-5 w-5" aria-hidden="true" />,
  },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button'));
    await expect(args.onClick).toHaveBeenCalledTimes(1);
  },
};

/** いちばん押してほしいもの。地色が付いて浮き上がる。 */
export const 主役: Story = {
  args: {
    to: '/code-editor',
    title: '演習をはじめる',
    description: 'まずはここから。',
    emphasis: 'primary',
    icon: <AcademicCapIcon className="h-5 w-5" aria-hidden="true" />,
  },
};

/** 右上に小さな札を付ける。 */
export const 札つき: Story = {
  args: {
    to: '/code-editor',
    title: 'Go の演習',
    description: '基礎から順に進みます。',
    badge: '初心者向け',
    icon: <AcademicCapIcon className="h-5 w-5" aria-hidden="true" />,
  },
};

/** 説明もアイコンも無い、いちばん短い形。 */
export const 題名だけ: Story = {
  args: { to: '/help', title: '使い方を見る' },
};

/** 実際の並び。主役は 1 枚だけ。 */
export const 並べたところ: Story = {
  args: { to: '/code-editor', title: '演習をはじめる' },
  render: (args) => (
    <div className="flex max-w-xl flex-col gap-3">
      <ActionCard
        {...args}
        emphasis="primary"
        title="演習をはじめる"
        description="まずはここから。"
        icon={<AcademicCapIcon className="h-5 w-5" aria-hidden="true" />}
      />
      <ActionCard to="/kb" title="ナレッジを見る" description="チームで共有する文書です。" />
      <ActionCard to="/help" title="使い方を見る" />
    </div>
  ),
};
