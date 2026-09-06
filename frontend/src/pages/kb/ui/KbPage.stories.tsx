import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, waitFor, within } from 'storybook/test';
import { withApi, withToast, routerWithParam } from '../../../../.storybook/decorators';
import KbPage from './KbPage';

/**
 * ナレッジの画面（左に木、右に本文）。
 *
 * URL は `/kb/:pageId` だけで、ワークスペースやスペースは出さない。どのページかが決まれば
 * 場所は引けるので、URL に階層を並べても長くなるだけで、階層を組み替えるたびにリンクが
 * 切れることになる。
 *
 * ページを指さずに `/kb` を開いたときは、前に見ていたページか、見えるうちの最初のページへ
 * その場で移る（行き止まりの空の画面を出さないため）。
 */
const meta = {
  title: 'pages/kb/KbPage',
  component: KbPage,
  parameters: { layout: 'fullscreen' },
  decorators: [withToast],
} satisfies Meta<typeof KbPage>;

export default meta;
type Story = StoryObj<typeof meta>;

const workspaces = [
  { slug: 'w-3f2a9c', name: '開発チーム', createdAt: '2026-01-01T00:00:00Z', canManage: true },
];

const spaces = [
  { id: 's-1', key: 's-1a2b3c', name: 'バックエンド定例', createdAt: '2026-01-01T00:00:00Z' },
];

const page = {
  id: 'p-1',
  spaceId: 's-1',
  title: '設計メモ',
  createdByUserId: 1,
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
};

const tree = {
  pages: [
    { page, children: [], hasHiddenChildren: false, parentArchived: false },
    {
      page: { ...page, id: 'p-2', title: '議事録' },
      children: [],
      hasHiddenChildren: false,
      parentArchived: false,
    },
  ],
  hasHiddenChildren: false,
};

const doc = {
  type: 'doc',
  content: [
    {
      type: 'paragraph',
      content: [{ type: 'text', text: 'ページとブロックを分けて持つ、という決めごとの記録。' }],
    },
  ],
};

const resolved = (over: Record<string, unknown> = {}) => ({
  workspaceSlug: 'w-3f2a9c',
  workspaceName: '開発チーム',
  page,
  doc,
  canEdit: true,
  canManage: true,
  ...over,
});

// 突き合わせは前から順。細かい宛先を先に書く。
const api = (over: Record<string, unknown> = {}) => ({
  '/kb/pages/p-1': resolved(),
  '/spaces/s-1/pages': tree,
  '/spaces': spaces,
  '/kb/workspaces': workspaces,
  ...over,
});

/** ページを開いているとき。 */
export const ページを開く: Story = {
  decorators: [routerWithParam('/kb/:pageId', '/kb/p-1'), withApi(api())],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(
      async () => {
        await expect(canvas.getByDisplayValue('設計メモ')).toBeInTheDocument();
      },
      { timeout: 5000 },
    );
  },
};

/** 読むだけの人が開いたとき。題名も本文も打ち替えられない。 */
export const 読むだけ: Story = {
  decorators: [
    routerWithParam('/kb/:pageId', '/kb/p-1'),
    withApi(api({ '/kb/pages/p-1': resolved({ canEdit: false, canManage: false }) })),
  ],
};

/** 見られないページ・存在しないページ。どちらも同じ見え方にする（実在を読ませない）。 */
export const 見られないページ: Story = {
  decorators: [routerWithParam('/kb/:pageId', '/kb/p-404'), withApi(api())],
};
