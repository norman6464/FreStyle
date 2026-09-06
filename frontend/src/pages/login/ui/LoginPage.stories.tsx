import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withApi, withRouter, withStore } from '../../../../.storybook/decorators';
import LoginPage from './LoginPage';

/**
 * ログインの入口。
 *
 * メールとパスワードはこの画面では受け取らない。受け取るのは発行者（ログインを預かる側）の
 * 役目で、アプリが受け取ると、二要素・連続失敗の締め出し・パスワードの強さといった
 * 発行者側の守りを素通りする経路を自分で開くことになる。
 *
 * 設定が揃っていないときは、ボタンを消さずに押せないままにして理由を添える。
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
 * ふだんの見え方。
 *
 * 見本では発行者の設定を入れていないので、**設定が揃っていないときの案内**が出る。
 * これは壊れているのではなく、設定が無いときに実際に出る姿。
 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toBeVisible();
  },
};
