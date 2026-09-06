import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import AuthInitializer from './AuthInitializer';
import { withApi, withStore } from '../../../.storybook/decorators';

/**
 * アプリを開いた直後に「まだログインが生きているか」をサーバーへ聞きに行く係。
 *
 * 答えが返るまでは画面いっぱいの読み込み表示にする。**手元の既定値で先に描かない**のが要点で、
 * 描いてしまうと、ログイン済みの人が一瞬ログイン画面を見ることになる。
 *
 * 401 / 403（＝ログインが切れたことが確定）のときだけ、ログイン済みの目印を消す。通信断や
 * サーバー側の不調で消すと、セッションは生きているのに次に開いたときの振り分けが効かなくなる。
 */
const meta = {
  title: 'app/AuthInitializer',
  component: AuthInitializer,
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof AuthInitializer>;

export default meta;
type Story = StoryObj<typeof meta>;

const children = (
  <div className="p-8">
    <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">アプリの中身</h1>
  </div>
);

/** 聞いている最中。画面いっぱいの読み込み表示。 */
export const 確認中: Story = {
  args: { children },
  decorators: [
    withStore({ isAuthenticated: false, loading: true }),
    // 答えを返さない宛先にして、聞いている最中の姿で止める。
    withApi({ '/auth/me': () => new Promise(() => {}) }),
  ],
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('status')).toHaveAccessibleName('読み込み中');
  },
};

/** ログインが生きていたとき。中身が出る。 */
export const ログイン済み: Story = {
  args: { children },
  decorators: [
    withStore({ isAuthenticated: false, loading: true }),
    withApi({ '/auth/me': { id: 1, email: 'takuma@example.com', name: '川野 拓馬' } }),
  ],
  play: async ({ canvasElement }) => {
    await waitFor(async () => {
      await expect(
        within(canvasElement).getByRole('heading', { name: 'アプリの中身' }),
      ).toBeVisible();
    });
  },
};

/**
 * ログインが切れていたとき。
 *
 * この係は行き先を決めない（送るのは門番の仕事）。ここでは手元の状態を「未ログイン」にして
 * 中身へ進むところまでを見る。
 */
export const ログインが切れていたとき: Story = {
  args: { children },
  decorators: [
    withStore({ isAuthenticated: false, loading: true }),
    // 見本に無い宛先は 404 で投げるので、聞きに行って断られた道筋がそのまま通る。
    withApi({}),
  ],
  play: async ({ canvasElement }) => {
    await waitFor(async () => {
      await expect(
        within(canvasElement).getByRole('heading', { name: 'アプリの中身' }),
      ).toBeVisible();
    });
  },
};
