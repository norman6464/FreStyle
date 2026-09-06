import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../.storybook/decorators';
import LinkText from './LinkText';

/**
 * 文章の中に置く、アプリ内リンク。
 *
 * ブラウザの再読み込みを起こさずに画面を移る（react-router の `<Link>`）。
 * 外部サイトへ飛ぶときは、この部品ではなく素の `<a>` を使う。
 */
const meta = {
  title: 'shared/LinkText',
  component: LinkText,
  parameters: { layout: 'centered' },
  decorators: [withRouter],
} satisfies Meta<typeof LinkText>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 単体で置いたところ。 */
export const 既定: Story = {
  args: { to: '/signup', children: 'アカウントを作成する' },
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByRole('link', { name: 'アカウントを作成する' }),
    ).toHaveAttribute('href', '/signup');
  },
};

/** 文章の中に混ぜたところ。前後の文と高さが揃う。 */
export const 文中に置く: Story = {
  args: { to: '/login', children: 'ログイン' },
  render: (args) => (
    <p className="text-sm text-[var(--color-text-secondary)]">
      すでにアカウントをお持ちですか？ <LinkText {...args} />
    </p>
  ),
};
