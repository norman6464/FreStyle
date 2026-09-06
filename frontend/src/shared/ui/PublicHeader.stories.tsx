import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { routerAt } from '../../../.storybook/decorators';
import PublicHeader from './PublicHeader';

/**
 * ログイン前の画面（ログイン / アカウント作成）で共通の帯。
 *
 * **いま居るページへのリンクは出さない。** アカウント作成の画面で「アカウントを作成」を
 * 出しても、押した人は同じ場所に留まるだけで、何も起きなかったように見えるため。
 * 出すのは常に「反対側」だけ。
 */
const meta = {
  title: 'shared/PublicHeader',
  component: PublicHeader,
  parameters: { layout: 'fullscreen' },
} satisfies Meta<typeof PublicHeader>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ログイン画面にいるとき。案内は「アカウントを作成」。 */
export const ログイン画面: Story = {
  decorators: [routerAt('/login')],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('link', { name: /アカウントを作成/ })).toBeVisible();
    // 自分自身への案内は出さない。
    await expect(canvas.queryByRole('link', { name: 'ログイン' })).toBeNull();
  },
};

/** アカウント作成の画面にいるとき。案内は「ログイン」に入れ替わる。 */
export const アカウント作成画面: Story = {
  decorators: [routerAt('/signup')],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('link', { name: 'ログイン' })).toBeVisible();
    await expect(canvas.queryByRole('link', { name: /アカウントを作成/ })).toBeNull();
  },
};
