import type { Meta, StoryObj } from '@storybook/react-vite';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { expect, within } from 'storybook/test';
import Protected from './Protected';
import { withStore } from '../../../.storybook/decorators';

/**
 * ログインが要る画面の門番。
 *
 * ログインしていれば中身をそのまま出し、していなければログイン画面へ送る。
 * 自分では何も描かない — 「権限がありません」の画面を出さないのは、未ログインの人にとって
 * その画面は行き止まりで、いちばんしたいこと（ログイン）から遠ざかるため。
 *
 * ここで見ているのは手元の状態だけ。本当の権限はサーバーが判定する（この門番を素通りしても、
 * 中身が使うデータは返ってこない）。
 */
const meta = {
  title: 'app/Protected',
  component: Protected,
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof Protected>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 見本用の道。送られた先が分かるよう、ログイン画面の代わりを置く。 */
function Stage({ children }: { children: React.ReactNode }) {
  return (
    <MemoryRouter initialEntries={['/secret']}>
      <Routes>
        <Route path="/secret" element={<Protected>{children}</Protected>} />
        <Route
          path="/login"
          element={<p className="p-8 text-sm text-[var(--color-text-secondary)]">ログイン画面</p>}
        />
      </Routes>
    </MemoryRouter>
  );
}

const secret = (
  <div className="p-8">
    <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">中身</h1>
  </div>
);

/** ログイン済み。中身がそのまま出る。 */
export const ログイン済み: Story = {
  args: { children: secret },
  decorators: [withStore({ isAuthenticated: true, loading: false })],
  render: () => <Stage>{secret}</Stage>,
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { name: '中身' })).toBeVisible();
  },
};

/** 未ログイン。中身は出さず、ログイン画面へ送る。 */
export const 未ログイン: Story = {
  args: { children: secret },
  decorators: [withStore({ isAuthenticated: false, loading: false })],
  render: () => <Stage>{secret}</Stage>,
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('ログイン画面')).toBeVisible();
    await expect(canvas.queryByRole('heading', { name: '中身' })).toBeNull();
  },
};
