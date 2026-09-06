import type { Meta, StoryObj } from '@storybook/react-vite';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { expect, userEvent, waitFor, within } from 'storybook/test';
import { withApi, withStore, withToast } from '../../../../.storybook/decorators';
import AppShell from './AppShell';

/**
 * ログイン後の画面ぜんぶを包む外枠（帯・本文・上に戻る・行き先を探す窓）。
 *
 * 帯は本文の**上に重ねてある**。縦に並べると帯の後ろに何も無くなり、半透明とぼかしが
 * 効かない。重ねたぶん本文の先頭に余白を入れて、最初の行が帯の裏に隠れないようにしてある。
 *
 * ⌘K（Windows は Ctrl+K）でどこからでも「行き先を探す窓」が開く。
 *
 * 本文が縦に長い画面では、下へスクロールすると帯が上へ滑って隠れる（本文が全高になる）。
 * 画面を移ると必ず戻す — 前の画面で隠したまま次の画面へ持ち越さない。
 */
const meta = {
  title: 'widgets/app-shell/AppShell',
  component: AppShell,
  parameters: { layout: 'fullscreen' },
  decorators: [
    withStore(),
    withToast,
    withApi({
      '/profile/me': { displayName: '川野 拓馬', avatarUrl: null, email: 'takuma@example.com' },
      '/notifications/unread-count': 2,
      '/kb/workspaces': [
        { slug: 'w-3f2a9c', name: '開発チーム', createdAt: '2026-01-01T00:00:00Z', canManage: true },
      ],
    }),
    // AppShell は「枠」なので、中身は Outlet に入る。router の入れ子まで作らないと描けない。
    (Story) => (
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<Story />}>
            <Route
              index
              element={
                <div className="mx-auto max-w-3xl p-6">
                  <h1 className="mb-4 text-2xl font-bold text-[var(--color-text-primary)]">
                    ここが本文
                  </h1>
                  {Array.from({ length: 30 }, (_, i) => (
                    <p key={i} className="py-2 text-sm text-[var(--color-text-secondary)]">
                      {i + 1} 行目
                    </p>
                  ))}
                </div>
              }
            />
          </Route>
        </Routes>
      </MemoryRouter>
    ),
  ],
} satisfies Meta<typeof AppShell>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ふだんの見え方。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('navigation', { name: 'メインナビゲーション' })).toBeVisible();
    await expect(canvas.getByRole('heading', { name: 'ここが本文' })).toBeVisible();
  },
};

/** ⌘K で「行き先を探す窓」が開く。 */
export const コマンドパレットを開く: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.keyboard('{Meta>}k{/Meta}');
    await waitFor(async () => {
      await expect(canvas.getByPlaceholderText('コマンドを検索...')).toBeVisible();
    });
  },
};

/** 下へスクロールすると「上に戻る」が右下に出る。 */
export const 上に戻るが出る: Story = {
  play: async ({ canvasElement }) => {
    const main = canvasElement.querySelector('#main-content') as HTMLElement;
    main.scrollTop = 500;
    main.dispatchEvent(new Event('scroll'));
    await waitFor(async () => {
      await expect(
        within(canvasElement).getByRole('button', { name: 'ページ上部に戻る' }),
      ).toBeInTheDocument();
    });
  },
};
