import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withApi, withRouter, withStore } from '../../../../.storybook/decorators';
import SignupPage from './SignupPage';

/**
 * アカウントを作る入口。
 *
 * ログインと同じく、この画面ではパスワードを受け取らない。作るのも発行者の役目で、
 * アプリは入口を出すだけ。
 */
const meta = {
  title: 'pages/signup/SignupPage',
  component: SignupPage,
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
} satisfies Meta<typeof SignupPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ふだんの見え方（見本では発行者の設定が無いので、その案内が出る）。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toBeVisible();
  },
};
