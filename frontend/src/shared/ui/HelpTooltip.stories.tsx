import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';
import HelpTooltip from './HelpTooltip';

/**
 * 「?」を押すと説明が出る小さな部品。
 *
 * マウスを乗せただけでは開かない。**押して開き、押して閉じる**（開閉パネル）形にしてある。
 * 勝手に出る吹き出しは、触れるつもりのない人の邪魔になり、指で操作する端末では
 * そもそも「乗せる」ができないため。
 *
 * Escape と、外側を押すことでも閉じる。
 */
const meta = {
  title: 'shared/HelpTooltip',
  component: HelpTooltip,
  parameters: { layout: 'centered' },
  decorators: [
    (Story) => (
      // 上に出す向きが既定なので、上下に余白を取らないと画面外に出て見えない。
      <div className="p-24">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof HelpTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 閉じているとき。「?」だけが見える。 */
export const 閉じている: Story = {
  args: { label: '5軸評価について', children: '話し方を 5 つの観点で点数にしたものです。' },
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByRole('button', { name: '5軸評価について' }),
    ).toHaveAttribute('aria-expanded', 'false');
  },
};

/** 押して開いたところ。 */
export const 開いたところ: Story = {
  args: { label: '5軸評価について', children: '話し方を 5 つの観点で点数にしたものです。' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const trigger = canvas.getByRole('button', { name: '5軸評価について' });
    await userEvent.click(trigger);
    await expect(trigger).toHaveAttribute('aria-expanded', 'true');
    // パネルは 0.15 秒かけて現れる。出きる前に見ると opacity が 0 のままなので待つ。
    await waitFor(async () => {
      await expect(canvas.getByText('話し方を 5 つの観点で点数にしたものです。')).toBeVisible();
    });
  },
};

/** 出す向きは 4 方向から選べる。ここでは下に出す。 */
export const 下に出す: Story = {
  args: { placement: 'bottom', label: '用語の説明', children: '画面の下側にある要素で使います。' },
  play: async ({ canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: '用語の説明' }));
  },
};

/** 説明は文字だけでなく、強調や改行も置ける。 */
export const 中身に飾りを置く: Story = {
  args: {
    label: '保存のしくみについて',
    children: (
      <>
        <strong className="text-[var(--color-text-primary)]">自動で保存されます。</strong>
        <br />
        手が止まってから少し待つと、そのときの内容が残ります。
      </>
    ),
  },
  play: async ({ canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: '保存のしくみについて' }));
  },
};

/** Escape で閉じる。 */
export const Escapeで閉じる: Story = {
  args: { label: '用語の説明', children: '押して開き、Escape で閉じます。' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const trigger = canvas.getByRole('button', { name: '用語の説明' });
    await userEvent.click(trigger);
    await expect(trigger).toHaveAttribute('aria-expanded', 'true');
    await userEvent.keyboard('{Escape}');
    await expect(trigger).toHaveAttribute('aria-expanded', 'false');
  },
};
