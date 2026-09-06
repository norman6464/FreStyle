import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import SNSSignInButton from './SNSSignInButton';

/**
 * 外部サービスのアカウントでログインするボタン。
 *
 * ロゴは各社が配っているものを指しているため、**インターネットに繋がっていないと絵が出ない**。
 * 絵が出なくても文字（「Googleでログイン」）だけで用は足りる作りにしてある。
 */
const meta = {
  title: 'shared/SNSSignInButton',
  component: SNSSignInButton,
  parameters: { layout: 'centered' },
  args: { onClick: fn() },
  decorators: [
    (Story) => (
      <div className="w-80">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SNSSignInButton>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Google。 */
export const Google: Story = {
  args: { provider: 'google' },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(within(canvasElement).getByRole('button', { name: /Googleでログイン/ }));
    await expect(args.onClick).toHaveBeenCalledTimes(1);
  },
};

/** Facebook。 */
export const Facebook: Story = {
  args: { provider: 'facebook' },
};

/** X（旧 Twitter）。 */
export const X: Story = {
  args: { provider: 'x' },
};

/** 3 つ並べたところ。ログイン画面での見え方。 */
export const 並べたところ: Story = {
  args: { provider: 'google' },
  render: (args) => (
    <div>
      <SNSSignInButton {...args} provider="google" />
      <SNSSignInButton {...args} provider="facebook" />
      <SNSSignInButton {...args} provider="x" />
    </div>
  ),
};

/** 押せないとき（送信中など）。 */
export const 押せない: Story = {
  args: { provider: 'google', disabled: true },
};
