import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import { withApi, withRouter, withToast } from '../../../../.storybook/decorators';
import NotificationPage from './NotificationPage';

/**
 * 届いた知らせの一覧。
 *
 * 未読は地色が濃く、丸い印が付く。読んだかどうかを、色だけでなく要素の有無でも
 * 区別できるようにしてある。
 */
const meta = {
  title: 'pages/notifications/NotificationPage',
  component: NotificationPage,
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
} satisfies Meta<typeof NotificationPage>;

export default meta;
type Story = StoryObj<typeof meta>;

// 種別は実在の値ではなく、素性の分かる仮の値。backend にはまだ通知を作る処理が無く、
// 日本語の名前を当てる対応表も空なので、バッジには種別の文字がそのまま出る。
const notification = (id: number, title: string, body: string, isRead: boolean) => ({
  id,
  type: 'sample_type',
  title,
  body,
  isRead,
  createdAt: '2026-09-06T09:41:00Z',
});

/** 未読と既読が混ざっているとき。 */
export const 既定: Story = {
  decorators: [
    withApi({
      // 件数の宛先は一覧の宛先を含むので、細かいほうを先に書く。
      '/notifications/unread-count': 2,
      '/notifications': [
        notification(
          1,
          'コメントに返信がありました',
          '「設計メモ」のコメントに返信が付きました。',
          false,
        ),
        notification(2, 'ページが共有されました', '「設計メモ」が閲覧できるようになりました。', false),
        notification(3, 'コメントに返信がありました', '「議事録」のコメントに返信が付きました。', true),
      ],
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('ページが共有されました')).toBeVisible();
    });
  },
};

/** 1 件も無いとき。 */
export const 空: Story = {
  decorators: [withApi({ '/notifications/unread-count': 0, '/notifications': [] })],
};

/** 取れなかったとき。 */
export const 取得に失敗: Story = {
  decorators: [withApi({})],
};
