import type { Meta, StoryObj } from '@storybook/react-vite';
import { AcademicCapIcon, CodeBracketIcon } from '@heroicons/react/24/outline';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import FeatureCard from './FeatureCard';
import FeatureSection from './FeatureSection';

/**
 * ホームのカードを「学習」「ツール」のような塊にまとめる見出しつきの区画。
 *
 * カードは 2 列で並び、狭い画面では 1 列になる。区画そのものは中身を選ばないので、
 * 何を入れるかは呼び出し側が決める。
 */
const meta = {
  title: 'pages/home/FeatureSection',
  component: FeatureSection,
  parameters: { layout: 'padded' },
  decorators: [
    withRouter,
    (Story) => (
      <div className="max-w-2xl">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof FeatureSection>;

export default meta;
type Story = StoryObj<typeof meta>;

/** カードを 2 枚入れたところ。 */
export const 既定: Story = {
  args: {
    title: '学習',
    children: (
      <>
        <FeatureCard
          to="/code-editor"
          icon={CodeBracketIcon}
          title="演習"
          description="手を動かしながら学びます。"
          color="brand"
        />
        <FeatureCard
          to="/kb"
          icon={AcademicCapIcon}
          title="ナレッジ"
          description="チームで共有する文書です。"
          color="emerald"
        />
      </>
    ),
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { name: '学習' })).toBeVisible();
  },
};

/** 1 枚だけのとき。 */
export const 一枚だけ: Story = {
  args: {
    title: 'ツール',
    children: (
      <FeatureCard
        to="/kb"
        icon={AcademicCapIcon}
        title="ナレッジ"
        description="チームで共有する文書です。"
        color="taupe"
      />
    ),
  },
};
