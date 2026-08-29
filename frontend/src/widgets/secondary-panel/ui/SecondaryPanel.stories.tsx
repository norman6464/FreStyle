import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';
import SecondaryPanel from './SecondaryPanel';

/**
 * 本文の脇に出す面（コースの章一覧・ノートのサイドバー）。
 *
 * **同じ部品を両方の画面で使う。** 画面ごとに別の開閉を作ると、覚えることが増える。
 * `peekable` では « で隠すと本文が全幅になり、左端に触れると一時的に浮いて出る。
 * ⌘\ でも切り替わる。表示の状態は `storageKey` ごとに憶えておく。
 */
const meta = {
  title: 'widgets/SecondaryPanel',
  component: SecondaryPanel,
  parameters: { layout: 'fullscreen' },
  decorators: [
    (Story) => (
      <div className="flex h-[420px]">
        <Story />
        <div className="min-w-0 flex-1 overflow-auto bg-surface p-6">
          <h1 className="mb-3 text-2xl font-bold text-[var(--color-text-primary)]">本文</h1>
          <p className="text-[var(--color-text-primary)]">
            隠すと、この本文が全幅に広がる。左端に触れると面が浮いて出る。
          </p>
        </div>
      </div>
    ),
  ],
} satisfies Meta<typeof SecondaryPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 中身の見本（章の一覧・ノートの木のかわり）。 */
function PanelBody() {
  return (
    <ul className="p-2">
      {['はじめに', '環境を整える', '最初のプログラム', 'まとめ'].map((t) => (
        <li key={t}>
          <button
            type="button"
            className="w-full rounded-md px-2 py-1.5 text-left text-sm text-[var(--color-text-primary)] hover:bg-surface-2"
          >
            {t}
          </button>
        </li>
      ))}
    </ul>
  );
}

/** 出したまま（既定）。 */
export const 出したまま: Story = {
  args: {
    title: 'Go 入門',
    peekable: true,
    storageKey: 'sb.panel.demo.open',
    children: <PanelBody />,
  },
};

/** 見出しに数を添える（章数・ページ数）。 */
export const 数を添える: Story = {
  args: {
    title: 'Go 入門',
    badge: '4 章',
    peekable: true,
    storageKey: 'sb.panel.demo.badge',
    children: <PanelBody />,
  },
};

/** « を押して隠したところ。本文が全幅になる。 */
export const 隠したところ: Story = {
  args: {
    title: 'Go 入門',
    peekable: true,
    storageKey: 'sb.panel.demo.hidden',
    children: <PanelBody />,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const hide = canvas.getByRole('button', { name: 'サイドバーを閉じる' });
    await userEvent.click(hide);
    // 隠れたら「閉じる」は消え、戻す入口が残る（☰ と、左端に触れたとき出る »）。
    await waitFor(async () => {
      await expect(canvas.queryByRole('button', { name: 'サイドバーを閉じる' })).toBeNull();
    });
    await expect(
      canvas.getAllByRole('button', { name: 'サイドバーを固定表示する' }).length,
    ).toBeGreaterThan(0);
  },
};

/** 折りたたみ式（peekable ではない従来の形）。 */
export const 折りたたみ式: Story = {
  args: {
    title: 'Go 入門',
    collapsible: true,
    collapsed: false,
    children: <PanelBody />,
  },
};
