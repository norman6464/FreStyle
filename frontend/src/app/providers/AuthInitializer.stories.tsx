import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, waitFor, within } from 'storybook/test';
import AuthInitializer from './AuthInitializer';
import { withApi, withStore } from '../../../.storybook/decorators';

/**
 * アプリを開いた直後に「まだサインインが生きているか」を確かめる係。
 *
 * 発行者（GCIP のクライアント SDK / ローカルの Dex）への問い合わせは `subscribe` prop
 * （既定は本物の `subscribeAuthState`）が担う。ここでは発行者そのものを stub できないため
 * （axios のように差し替えのきく通信部分ではない）、`AuthInitializer.test.tsx` と同じく
 * `subscribe` を直接差し替えて見本にする。
 *
 * 答えが返るまでは画面いっぱいの読み込み表示にする。**手元の既定値で先に描かない**のが要点で、
 * 描いてしまうと、ログイン済みの人が一瞬ログイン画面を見ることになる。
 *
 * サインインが確認できたら `authRepository.login()`（backend へのセッション確立。
 * `POST /auth/login`）を呼ぶ。401 / 403（＝サインインが切れたことが確定）のときだけ
 * ログイン済みの目印を消す。通信断やサーバー側の不調で消すと、セッションは生きているのに
 * 次に開いたときの振り分けが効かなくなる。
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
  args: {
    children,
    // callback を呼ばない = 確認が返らない状態のまま止める。
    subscribe: fn(() => () => {}),
  },
  decorators: [withStore({ isAuthenticated: false, loading: true }), withApi({})],
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('status')).toHaveAccessibleName('読み込み中');
  },
};

/** サインインが生きていたとき。セッションを確立して中身が出る。 */
export const ログイン済み: Story = {
  args: {
    children,
    subscribe: fn((callback: (signedIn: boolean) => void) => {
      callback(true);
      return () => {};
    }),
  },
  decorators: [
    withStore({ isAuthenticated: false, loading: true }),
    withApi({ '/auth/login': { message: 'ログインしました。' } }),
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
 * サインインが切れていたとき。
 *
 * この係は行き先を決めない（送るのは門番 `Protected` の仕事）。ここでは手元の状態を
 * 「未サインイン」にして中身へ進むところまでを見る。
 */
export const サインインが切れていたとき: Story = {
  args: {
    children,
    subscribe: fn((callback: (signedIn: boolean) => void) => {
      callback(false);
      return () => {};
    }),
  },
  decorators: [withStore({ isAuthenticated: false, loading: true }), withApi({})],
  play: async ({ canvasElement }) => {
    await waitFor(async () => {
      await expect(
        within(canvasElement).getByRole('heading', { name: 'アプリの中身' }),
      ).toBeVisible();
    });
  },
};
