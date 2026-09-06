import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';
import GlossaryTerm from './GlossaryTerm';

/**
 * 文章の中の専門用語に、点線の下線と「?」を添える。
 *
 * 用語を知らない人がその場で意味を確かめられるようにするための部品。
 * 説明そのものは HelpTooltip が出す。
 */
const meta = {
  title: 'shared/GlossaryTerm',
  component: GlossaryTerm,
  parameters: { layout: 'centered' },
  decorators: [
    (Story) => (
      <div className="p-24">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof GlossaryTerm>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 単体で置いたところ。 */
export const 既定: Story = {
  args: {
    term: '5軸評価',
    definition: '話し方を 5 つの観点（結論・根拠・具体・簡潔・配慮）で点数にしたものです。',
  },
};

/** 押すと意味が出る。 */
export const 意味を見る: Story = {
  args: {
    term: '5軸評価',
    definition: '話し方を 5 つの観点で点数にしたものです。',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: '5軸評価の意味を表示' }));
    // 説明は 0.15 秒かけて現れる。出きるまで待ってから見る。
    await waitFor(async () => {
      await expect(canvas.getByText('話し方を 5 つの観点で点数にしたものです。')).toBeVisible();
    });
  },
};

/** 文章に混ぜたところ。行の高さが崩れない。 */
export const 文中に置く: Story = {
  args: { term: 'ワークスペース', definition: 'チームごとの入れもの。この中にスペースとページが入ります。' },
  render: (args) => (
    <p className="max-w-md text-sm leading-7 text-[var(--color-text-secondary)]">
      ナレッジは <GlossaryTerm {...args} /> ごとに分かれています。所属していない
      ワークスペースの中身は見えません。
    </p>
  ),
};
