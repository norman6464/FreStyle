import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import AuthUnavailableNotice from './AuthUnavailableNotice';

/**
 * ログインの仕組みがまだ設定されていないことを伝える帯。
 *
 * ログインのボタンを**消さずに**、押せないまま理由を添える。消してしまうと、外から見て
 * 「壊れているのか、わざと止めているのか」が区別できない。
 *
 * 足りない設定の名前は文には出さず、要素の属性（`data-missing`）に載せてある。
 * 使う人にとって設定の名前は意味が無く、運用する人は要素を見れば分かるため。
 */
const meta = {
  title: 'features/auth/AuthUnavailableNotice',
  component: AuthUnavailableNotice,
  parameters: { layout: 'padded' },
  decorators: [
    (Story) => (
      <div className="max-w-md">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof AuthUnavailableNotice>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 設定が 1 つ足りないとき。 */
export const 一つ足りない: Story = {
  args: { missing: ['VITE_OIDC_CLIENT_ID'] },
  play: async ({ canvasElement }) => {
    const notice = within(canvasElement).getByRole('status');
    await expect(notice).toHaveTextContent('現在ログインを受け付けていません');
    // 足りないものの名前は文には出さず、属性で運用する人に渡す。
    await expect(notice).toHaveAttribute('data-missing', 'VITE_OIDC_CLIENT_ID');
    await expect(notice).not.toHaveTextContent('VITE_OIDC_CLIENT_ID');
  },
};

/** 複数足りないとき。文言は変わらない。 */
export const 複数足りない: Story = {
  args: { missing: ['VITE_OIDC_ISSUER', 'VITE_OIDC_CLIENT_ID'] },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('status')).toHaveAttribute(
      'data-missing',
      'VITE_OIDC_ISSUER,VITE_OIDC_CLIENT_ID',
    );
  },
};
