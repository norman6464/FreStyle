import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';
import ScrollToTop from './ScrollToTop';

/**
 * 下までスクロールしたときに右下へ出る「上に戻る」ボタン。
 *
 * 上のほうに居るあいだは**出さない**。まだ戻る必要が無いのに常に出ていると、
 * 右下の内容をずっと隠し続けることになる。
 *
 * 見張る相手はウィンドウではなく、id で渡された箱。この画面では本文だけが
 * スクロールする作りだから（ウィンドウ側は動かない）。
 */
const meta = {
  title: 'widgets/app-shell/ScrollToTop',
  component: ScrollToTop,
  parameters: { layout: 'fullscreen' },
  decorators: [
    (Story) => (
      <div
        id="sb-scroll-area"
        // 縦に流れる枠はキーボードでも動かせる必要がある（実物の本文も同じ）。
        tabIndex={0}
        className="h-[360px] overflow-y-auto bg-surface p-6"
      >
        <Story />
        {Array.from({ length: 40 }, (_, i) => (
          <p key={i} className="py-2 text-sm text-[var(--color-text-secondary)]">
            {i + 1} 行目
          </p>
        ))}
      </div>
    ),
  ],
} satisfies Meta<typeof ScrollToTop>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 上のほうに居るとき。ボタンは出ない。 */
export const 出ていない: Story = {
  args: { targetId: 'sb-scroll-area' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).queryByRole('button', { name: 'ページ上部に戻る' })).toBeNull();
  },
};

/** 下までスクロールしたとき。右下に出る。 */
export const 下へ行くと出る: Story = {
  args: { targetId: 'sb-scroll-area', threshold: 200 },
  play: async ({ canvasElement }) => {
    const area = document.getElementById('sb-scroll-area') as HTMLElement;
    area.scrollTop = 400;
    area.dispatchEvent(new Event('scroll'));
    await waitFor(async () => {
      await expect(
        within(canvasElement).getByRole('button', { name: 'ページ上部に戻る' }),
      ).toBeInTheDocument();
    });
  },
};

/** 押すと先頭へ戻る。 */
export const 押すと先頭へ戻る: Story = {
  args: { targetId: 'sb-scroll-area', threshold: 200 },
  play: async ({ canvasElement }) => {
    const area = document.getElementById('sb-scroll-area') as HTMLElement;
    area.scrollTop = 400;
    area.dispatchEvent(new Event('scroll'));
    const button = await within(canvasElement).findByRole('button', { name: 'ページ上部に戻る' });
    await userEvent.click(button);
    await waitFor(async () => {
      await expect(area.scrollTop).toBe(0);
    });
  },
};

/** 出はじめる位置を変える（既定は 200px）。 */
export const 出はじめる位置を変える: Story = {
  args: { targetId: 'sb-scroll-area', threshold: 50 },
};
