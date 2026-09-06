import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import SkipLink from './SkipLink';

/**
 * キーボードで操作する人が、ヘッダーのリンクを全部たどらずに本文へ飛ぶためのリンク。
 *
 * **ふだんは見えない。** Tab キーで最初に当たったときだけ左上に現れる。
 * マウスを使う人には要らないが、キーボードだけの人にとっては、毎ページで
 * ナビの項目を延々と Tab で越える手間が消える。
 */
const meta = {
  title: 'widgets/app-shell/SkipLink',
  component: SkipLink,
  parameters: { layout: 'fullscreen' },
  decorators: [
    (Story) => (
      <div className="min-h-[240px] bg-surface p-6">
        <Story />
        <main id="main" tabIndex={-1} className="mt-8 rounded border border-surface-3 p-4">
          <p className="text-sm text-[var(--color-text-secondary)]">ここが本文です。</p>
        </main>
      </div>
    ),
  ],
} satisfies Meta<typeof SkipLink>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ふだんの状態。目には見えないが、要素としては在る。 */
export const 隠れている: Story = {
  args: { targetId: 'main' },
  play: async ({ canvasElement }) => {
    // 読み上げソフトからは見つかる。目に見えないだけ。
    await expect(
      within(canvasElement).getByRole('link', { name: 'メインコンテンツへスキップ' }),
    ).toBeInTheDocument();
  },
};

/** Tab で当たったとき。左上に現れる。 */
export const フォーカスすると現れる: Story = {
  args: { targetId: 'main' },
  play: async ({ canvasElement }) => {
    const link = within(canvasElement).getByRole('link', { name: 'メインコンテンツへスキップ' });
    await userEvent.tab();
    await expect(link).toHaveFocus();
    await expect(link).toBeVisible();
  },
};

/** 押すと本文へフォーカスが移る。 */
export const 押すと本文へ移る: Story = {
  args: { targetId: 'main' },
  play: async ({ canvasElement }) => {
    const link = within(canvasElement).getByRole('link', { name: 'メインコンテンツへスキップ' });
    await userEvent.click(link);
    await expect(document.getElementById('main')).toHaveFocus();
  },
};

/** 文言は変えられる（行き先が本文以外のとき）。 */
export const 文言を変える: Story = {
  args: { targetId: 'main', label: '記事本文へ移動' },
};
