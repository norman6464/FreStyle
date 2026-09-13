import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import { withApi, withRouter } from '../../../../.storybook/decorators';
import HeaderRecentPagesNav from './HeaderRecentPagesNav';

const NAV_CLASS = 'px-3 py-1.5 rounded-md text-sm font-medium text-[var(--color-text-tertiary)]';

/**
 * ヘッダーの「最近見たページ ▾」（段2・段3）。ワークスペース横断の閲覧履歴で、
 * 「今いるスペース」に関わらず常にドロップダウンとして使える。
 */
const meta = {
  title: 'widgets/app-shell/HeaderRecentPagesNav',
  component: HeaderRecentPagesNav,
  args: { className: NAV_CLASS },
  decorators: [withRouter],
} satisfies Meta<typeof HeaderRecentPagesNav>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 開くと最近見たページの一覧が出る。 */
export const 一覧が出る: Story = {
  decorators: [
    withApi({
      '/kb/me/recent-pages': [
        {
          pageId: 'p-1',
          workspaceSlug: 'w-3f2a9c',
          title: '設計メモ',
          spaceId: 's-1',
          spaceName: '開発チーム',
          viewedAt: '2026-09-13T00:00:00Z',
        },
        {
          pageId: 'p-2',
          workspaceSlug: 'w-3f2a9c',
          title: '議事録',
          icon: { type: 'emoji', value: '📝' },
          spaceId: 's-1',
          spaceName: '開発チーム',
          viewedAt: '2026-09-12T00:00:00Z',
        },
      ],
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: /最近見たページ/ }));
    await expect(await canvas.findByRole('link', { name: /設計メモ/ })).toBeVisible();
    await expect(canvas.getByRole('link', { name: /議事録/ })).toBeVisible();
  },
};

/** 0 件のとき。 */
export const まだ無いとき: Story = {
  decorators: [withApi({ '/kb/me/recent-pages': [] })],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: /最近見たページ/ }));
    await expect(await canvas.findByText('まだ最近見たページはありません')).toBeVisible();
  },
};
