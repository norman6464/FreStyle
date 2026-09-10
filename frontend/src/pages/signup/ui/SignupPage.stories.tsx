import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withApi, withRouter, withStore, withToast } from '../../../../.storybook/decorators';
import SignupPage from './SignupPage';

/**
 * アカウントを作る入口。`pages/login/ui/LoginPage.stories.tsx` と対になる。
 *
 * 発行者が 2 通りある——本番（GCIP）はメールとパスワードをこの場で受け取り、
 * その場でアカウントを作る。ローカル開発（Dex）は自己登録の手段を持たないため、
 * 発行者のログイン画面へ送るだけの入口になる。どちらを見せるかは手元の `.env`
 * （`VITE_FIREBASE_*` / `VITE_OIDC_*`）に従う。
 */
const meta = {
  title: 'pages/signup/SignupPage',
  component: SignupPage,
  parameters: { layout: 'fullscreen' },
  decorators: [
    withRouter,
    withStore({ isAuthenticated: false, loading: false }),
    withApi({}),
    withToast,
    (Story) => (
      <div className="h-[640px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SignupPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ふだんの見え方。表示される姿はこの env（`.env`）の設定に従う。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toBeVisible();
  },
};
