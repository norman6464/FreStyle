import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withApi, withRouter, withStore } from '../../../../.storybook/decorators';
import LoginCallback from './LoginCallback';

/**
 * ログインの発行者から戻ってきた直後に、一瞬だけ通る画面。
 *
 * 出るのは「ログイン中…」だけ。ここで受け取った合図をサーバーに渡し、済んだら次の画面へ移る。
 * 人が留まる場所ではないので、案内も操作も置かない。
 *
 * 一瞬しか見えない画面ほど、見本が要る（実際に踏むのが難しく、壊れても気づきにくい）。
 */
const meta = {
  title: 'pages/login-callback/LoginCallback',
  component: LoginCallback,
  parameters: { layout: 'fullscreen' },
  decorators: [withRouter, withStore({ isAuthenticated: false, loading: true }), withApi({})],
} satisfies Meta<typeof LoginCallback>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 受け渡している最中。 */
export const 処理中: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('ログイン中...')).toBeVisible();
  },
};
