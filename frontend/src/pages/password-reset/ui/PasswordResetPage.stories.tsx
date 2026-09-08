import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import PasswordResetPage from './PasswordResetPage';

/**
 * パスワード再設定画面。GCIP（Firebase）だけの機能で、`pages/login/ui/LoginPage.tsx` の
 * 「パスワードをお忘れですか？」から辿り着く。
 *
 * ローカル開発（Dex）はこの機能を持たないため、その旨の案内に置き換わる。
 * どちらを見せるかは手元の `.env`（`VITE_FIREBASE_*` / `VITE_OIDC_*`）に従う。
 */
const meta = {
  title: 'pages/password-reset/PasswordResetPage',
  component: PasswordResetPage,
  parameters: { layout: 'fullscreen' },
  decorators: [
    withRouter,
    (Story) => (
      <div className="h-[640px]">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof PasswordResetPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ふだんの見え方。表示される姿はこの env（`.env`）の設定に従う。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toBeVisible();
  },
};
