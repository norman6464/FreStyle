import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

/**
 * ヘッダーの地（app-header-surface）の見え方。
 *
 * ヘッダーは常時表示で、本文とは縦に並ぶ（重ねない）。不透明な地 + 1px の下罫だけで
 * 境界を表す（2026-09-12。以前は本文に重ねて半透明 + ぼかしにしていたが、常時表示化に
 * 伴い重ねる理由が無くなったのでやめた）。
 */
const meta = {
  title: 'widgets/app-shell/HeaderSurface',
  parameters: { layout: 'fullscreen' },
} satisfies Meta;

export default meta;
type Story = StoryObj;

/** ヘッダーと本文は縦に並ぶ（重ねない）。地は不透明。 */
export const ヘッダーと本文が縦に並ぶ: Story = {
  render: () => (
    <div style={{ height: '360px' }} className="flex flex-col">
      <div className="app-header-surface flex h-14 flex-shrink-0 items-center px-4 font-semibold" data-testid="header">
        FreStyle
      </div>
      {/* スクロールできる領域はキーボードで到達できる必要がある（本体の main と同じ扱い）。 */}
      <main tabIndex={-1} className="flex-1 min-h-0 overflow-auto outline-none" data-testid="body">
        <div className="mx-auto max-w-2xl px-6 py-6">
          <h1 className="mb-4 text-3xl font-bold">Todoリスト</h1>
          {Array.from({ length: 12 }, (_, i) => (
            <p key={i} className="mb-3 text-[var(--color-text-primary)]">
              ヘッダーは重ならないので、本文の先頭行は隠れない（{i + 1} 行目）。
            </p>
          ))}
          <a href="#" className="text-[var(--color-text-primary)] underline">
            本文中のリンク
          </a>
        </div>
      </main>
    </div>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const header = canvas.getByTestId('header');
    const body = canvas.getByTestId('body');

    const h = header.getBoundingClientRect();
    const b = body.getBoundingClientRect();
    // 縦に並んでいる（重ならない）こと。
    await expect(h.bottom <= b.top + 1).toBe(true);

    // 地は不透明（ぼかしは使わない）。
    const surface = header;
    const cs = getComputedStyle(surface);
    const blur = cs.backdropFilter || cs.getPropertyValue('-webkit-backdrop-filter');
    await expect(blur === 'none' || blur === '').toBe(true);
  },
};
