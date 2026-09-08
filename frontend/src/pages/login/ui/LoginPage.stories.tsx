import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withApi, withRouter, withStore } from '../../../../.storybook/decorators';
import LoginPage from './LoginPage';

/**
 * ログインの入口。
 *
 * 発行者が 2 通りある——本番（GCIP）はメールとパスワードをこの画面がその場で受け取る
 * （発行者の SDK が直接検証するため）。ローカル開発（Dex）は発行者のログイン画面へ
 * 丸ごと送るだけで、パスワードは受け取らない（Dex はその場で受け取る手段を持たない）。
 *
 * どちらを見せるかはビルド時の env（`VITE_FIREBASE_*` / `VITE_OIDC_*`）で決まり、
 * Storybook もこのリポジトリの `.env` を読んで起動するため、実際に表示される姿は
 * 手元の `.env` の中身に従う（env はストーリーの play では差し替えられない）。
 *
 * 設定が両方とも欠けているときは、ボタンを消さずに押せないままにして理由を添える。
 * 消すと、外から見て「壊れているのか、わざと止めているのか」が分からない。
 */
const meta = {
  title: 'pages/login/LoginPage',
  component: LoginPage,
  parameters: { layout: 'fullscreen' },
  decorators: [
    withRouter,
    withStore({ isAuthenticated: false, loading: false }),
    withApi({}),
    (Story) => (
      <div className="h-[640px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof LoginPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * ふだんの見え方。表示される姿はこの env（`.env`）の設定に従う。
 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toBeVisible();
  },
};
