import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withRouter } from '../../../../.storybook/decorators';
import HelpPage from './HelpPage';

/**
 * 使い方をまとめた画面。
 *
 * 通信をしないので、開けば必ず同じものが出る。どこにも問い合わせない画面は、
 * 落ちる余地も待ち時間も無い。
 */
const meta = {
  title: 'pages/help/HelpPage',
  component: HelpPage,
  parameters: { layout: 'fullscreen' },
  decorators: [withRouter],
} satisfies Meta<typeof HelpPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 既定（唯一の状態）。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toBeVisible();
  },
};
