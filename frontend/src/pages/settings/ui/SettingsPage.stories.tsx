import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withApi, withRouter, withToast } from '../../../../.storybook/decorators';
import SettingsPage from './SettingsPage';

/**
 * 設定の画面。左（狭い画面では上）で区分を選び、右に中身が出る。
 *
 * 区分を増やすときはこの画面の一覧に 1 つ足す。画面ごとに URL を分けていないのは、
 * 設定は行き来しながら見るもので、戻る操作で 1 つずつ遡らせたくないため。
 */
const meta = {
  title: 'pages/settings/SettingsPage',
  component: SettingsPage,
  parameters: { layout: 'fullscreen' },
  decorators: [
    withRouter,
    withToast,
    withApi({
      '/profile/me': {
        displayName: '川野 拓馬',
        email: 'takuma@example.com',
        avatarUrl: null,
        bio: 'Go と React を勉強しています。',
      },
    }),
    (Story) => (
      <div className="bg-surface">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SettingsPage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 既定。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('heading', { level: 1 })).toBeVisible();
  },
};

/** 狭い画面。 */
export const 狭い画面: Story = {
  globals: { viewport: { value: 'mobile1', isRotated: false } },
};
