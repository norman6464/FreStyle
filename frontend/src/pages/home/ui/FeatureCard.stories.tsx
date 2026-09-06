import type { Meta, StoryObj } from '@storybook/react-vite';
import { AcademicCapIcon, ChatBubbleLeftRightIcon, CodeBracketIcon } from '@heroicons/react/24/outline';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import FeatureCard from './FeatureCard';

/**
 * ホームに並ぶ、機能への入口カード 1 枚。
 *
 * 色は「どの系統の機能か」を見分けるためのもので、優先度ではない。同じ系統には同じ色を使う。
 *
 * 学べる技術のロゴを添えられる。ここに渡してよいのは**手元に絵がある技術だけ**で、
 * 無いものを混ぜると 1 つだけ汎用の記号になり、列が不揃いに見える。
 */
const meta = {
  title: 'pages/home/FeatureCard',
  component: FeatureCard,
  parameters: { layout: 'padded' },
  decorators: [
    withRouter,
    (Story) => (
      <div className="max-w-sm">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof FeatureCard>;

export default meta;
type Story = StoryObj<typeof meta>;

/** いちばん短い形。 */
export const 既定: Story = {
  args: {
    to: '/code-editor',
    icon: CodeBracketIcon,
    title: '演習',
    description: '手を動かしながら学びます。',
    color: 'brand',
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('link', { name: /演習/ })).toHaveAttribute(
      'href',
      '/code-editor',
    );
  },
};

/** 右上に小さな札を付ける。 */
export const 札つき: Story = {
  args: {
    to: '/code-editor',
    icon: AcademicCapIcon,
    title: '演習',
    description: '手を動かしながら学びます。',
    color: 'emerald',
    badge: 'NEW',
  },
};

/** 学べる技術のロゴを添えたところ。 */
export const ロゴつき: Story = {
  args: {
    to: '/code-editor',
    icon: CodeBracketIcon,
    title: '演習',
    description: '言語を選んで、手を動かしながら学びます。',
    color: 'brand',
    techLogos: ['go', 'typescript', 'python', 'docker'],
  },
};

/** 4 つの色を並べて見分けを確かめる。 */
export const 色ぜんぶ: Story = {
  args: {
    to: '/code-editor',
    icon: CodeBracketIcon,
    title: '演習',
    description: '手を動かしながら学びます。',
    color: 'brand',
  },
  render: (args) => (
    <div className="grid max-w-2xl grid-cols-2 gap-3">
      <FeatureCard {...args} color="brand" title="ブランド" />
      <FeatureCard {...args} color="emerald" title="エメラルド" icon={AcademicCapIcon} />
      <FeatureCard {...args} color="taupe" title="トープ" icon={ChatBubbleLeftRightIcon} />
      <FeatureCard {...args} color="blue" title="ブルー" />
    </div>
  ),
};

/** 説明が長いとき。カードの高さは揃ったままにする。 */
export const 説明が長い: Story = {
  args: {
    to: '/code-editor',
    icon: CodeBracketIcon,
    title: '演習',
    description:
      '言語を選び、問題を読み、その場でコードを書いて動かします。通らなかったテストケースは、どの行がどう違うかまで出ます。',
    color: 'brand',
  },
};
