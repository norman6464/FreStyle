import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import KbSpaceOverviewPage from './KbSpaceOverviewPage';
import { routerWithParam, withApi, withToast } from '../../../../.storybook/decorators';

const workspaces = [{ slug: 'acme', name: 'Acme 社', createdAt: '2026-01-01T00:00:00Z', canManage: true }];
const mySpaces = [{ id: 'space-1', name: '開発部', role: 'editor' }];
const spaces = [{ id: 'space-1', key: 'space-1', name: '開発部', visibility: 'workspace', createdAt: '2026-01-01T00:00:00Z' }];

/** スペース単位の「概要」画面（段14）。凝った内容は無く、自分の役割程度を示す最小限。 */
const meta = {
  title: 'pages/kb-space-overview/KbSpaceOverviewPage',
  component: KbSpaceOverviewPage,
  parameters: { layout: 'fullscreen' },
  decorators: [withToast],
} satisfies Meta<typeof KbSpaceOverviewPage>;

export default meta;
type Story = StoryObj<typeof meta>;

// 突き合わせは前から順なので、細かい宛先を先に書く（/spaces が先だと木の要求まで拾う）。
export const ふつう: Story = {
  decorators: [
    routerWithParam('/kb/spaces/:spaceId', '/kb/spaces/space-1'),
    withApi({
      '/spaces/space-1/pages': { pages: [], hasHiddenChildren: false },
      '/me/spaces': mySpaces,
      '/spaces': spaces,
      '/kb/workspaces': workspaces,
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // スペース名は見出しとして 1 回だけ出る（サイドバー内の表示はサイドバー自身の
    // KbSidebar.stories.tsx が確かめる。ここは本文だけを見る）。
    await waitFor(async () => {
      await expect(canvas.getByRole('heading', { name: '開発部' })).toBeInTheDocument();
    });
    await expect(canvas.getByText(/編集者/)).toBeInTheDocument();
  },
};

export const アクセスできるスペースが無い: Story = {
  // spaceId 無しの入口（/kb/spaces）でだけ再現する。特定の spaceId を指しての「見つからない」
  // とは別（そちらは別の文言になる。resolveKbSpace の doc 参照）。
  decorators: [routerWithParam('/kb/spaces', '/kb/spaces'), withApi({ '/me/spaces': [], '/kb/workspaces': workspaces })],
  play: async ({ canvasElement }) => {
    await expect(
      await within(canvasElement).findByText('アクセスできるスペースがありません'),
    ).toBeInTheDocument();
  },
};
