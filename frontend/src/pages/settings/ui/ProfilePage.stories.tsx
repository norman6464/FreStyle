import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import { withApi, withRouter, withToast } from '../../../../.storybook/decorators';
import ProfilePage from './ProfilePage';

/**
 * 自分の情報を見て直す画面。
 *
 * 保存の結果は知らせ（トースト）で伝える。入力欄はそのまま残すので、失敗しても
 * 打ち直しにはならない。
 */
const meta = {
  title: 'pages/settings/ProfilePage',
  component: ProfilePage,
  parameters: { layout: 'fullscreen' },
  decorators: [
    withRouter,
    withToast,
    (Story) => (
      <div className="bg-surface p-6">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ProfilePage>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 情報が取れたとき。 */
export const 既定: Story = {
  decorators: [
    withApi({
      '/profile/me': {
        displayName: '川野 拓馬',
        email: 'takuma@example.com',
        avatarUrl: null,
        bio: 'Go と React を勉強しています。',
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    await expect(await within(canvasElement).findByDisplayValue('川野 拓馬')).toBeVisible();
  },
};

/** 写真を登録しているとき。 */
export const 写真つき: Story = {
  decorators: [
    withApi({
      '/profile/me': {
        displayName: '川野 拓馬',
        email: 'takuma@example.com',
        avatarUrl:
          'data:image/svg+xml;utf8,' +
          encodeURIComponent(
            '<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96">' +
              '<rect width="96" height="96" fill="#7c6f64"/>' +
              '<circle cx="48" cy="38" r="18" fill="#f2e5bc"/>' +
              '<circle cx="48" cy="86" r="30" fill="#f2e5bc"/>' +
              '</svg>',
          ),
        bio: '',
      },
    }),
  ],
};

/** 情報が取れなかったとき。 */
export const 取得に失敗: Story = {
  decorators: [withApi({})],
};
