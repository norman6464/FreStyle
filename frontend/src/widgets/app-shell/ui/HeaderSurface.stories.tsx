import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

/**
 * ヘッダーの地（app-header-surface）の見え方。
 *
 * **ぼかしは「背後に何かがある」ときだけ見える。** ヘッダーと本文を縦に並べると
 * 背後は空なので、半透明にしてもぼけない。本文がヘッダーの下を通る形にして初めて効く。
 * この見本は、その重なりが成立していることを確かめるためのもの。
 */
const meta = {
  title: 'app-shell/HeaderSurface',
  parameters: { layout: 'fullscreen' },
} satisfies Meta;

export default meta;
type Story = StoryObj;

/** 実際のアプリと同じ重ね方（ヘッダーが本文の上・本文は先頭に余白）。 */
export const 本文がヘッダーの下を通る: Story = {
  render: () => (
    <div style={{ height: '360px' }} className="relative">
      <div className="absolute inset-x-0 top-0 z-40" data-testid="header">
        <div className="app-header-surface flex h-16 items-center px-4 font-semibold">
          FreStyle
        </div>
      </div>
      {/* スクロールできる領域はキーボードで到達できる必要がある（本体の main と同じ扱い）。 */}
      <main tabIndex={-1} className="h-full overflow-auto pt-16 outline-none" data-testid="body">
        <div className="mx-auto max-w-2xl px-6 py-6">
          <h1 className="mb-4 text-3xl font-bold">Todoリスト</h1>
          {Array.from({ length: 12 }, (_, i) => (
            <p key={i} className="mb-3 text-[var(--color-text-primary)]">
              スクロールすると、この行がヘッダーの下を通る。地が半透明なので透けて、
              ぼかしがかかって見える（{i + 1} 行目）。
            </p>
          ))}
          {/* 本物の画面と同じく、スクロール領域の中にキーボードで届く要素を置く
              （これが無いと「スクロールできるのにキーボードで到達できない」になる）。 */}
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

    // 本文を少し送って、行がヘッダーの裏へ入る状態にする。
    body.scrollTop = 120;

    const h = header.getBoundingClientRect();
    const b = body.getBoundingClientRect();
    // ヘッダーと本文の領域が**重なっている**こと（縦に並んでいたら重ならない）。
    await expect(h.top >= b.top && h.bottom > b.top).toBe(true);

    // 地にぼかしが効いていること。
    const surface = header.querySelector('.app-header-surface') as HTMLElement;
    const cs = getComputedStyle(surface);
    // Safari 系は -webkit- 付きでしか返さないことがある（型に無いので getPropertyValue で引く）。
    const blur = cs.backdropFilter || cs.getPropertyValue('-webkit-backdrop-filter');
    await expect(blur).toContain('blur');
  },
};
