import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, waitFor, within } from 'storybook/test';
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

/**
 * PUT(設定) / DELETE(解除) を method で撃ち分ける。
 *
 * withApi は URL の**部分一致**で見本を選び、method は見ない。icon の宛先
 * （`/kb/workspaces/…/pages/p-1/icon`）は `/kb/workspaces` を部分文字列として含むので、
 * この宛先を `/kb/workspaces` より**前**に置かないと、そちらの見本（一覧）に取られる。
 */
const iconStub = (config: { method?: string }) =>
  config.method === 'delete'
    ? { ...page, icon: null }
    : { ...page, icon: { type: 'emoji', value: '📘' } };

// 突き合わせは前から順。細かい宛先を先に書く。
const api = (over: Record<string, unknown> = {}) => ({
  '/kb/pages/p-1': resolved(),
  '/pages/p-1/icon': iconStub,
  // 逆リンクの宛先（.../pages/p-1/backlinks）は `/kb/workspaces` を部分文字列として
  // 含むため、iconStub と同じ理由でそちらより前に置く（後ろだと workspaces の一覧が
  // 誤って backlinks の応答として使われ、意図しない逆リンクセクションが出てしまう）。
  '/pages/p-1/backlinks': [],
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

/** 読むだけの人が開いたとき。題名も本文も打ち替えられない。アイコンも押せない。 */
export const 読むだけ: Story = {
  decorators: [
    routerWithParam('/kb/:pageId', '/kb/p-1'),
    withApi(
      api({
        '/kb/pages/p-1': resolved({
          canEdit: false,
          canManage: false,
          page: { ...page, icon: { type: 'emoji', value: '📘' } },
        }),
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(await canvas.findByRole('heading', { name: '設計メモ' })).toBeInTheDocument();
    // アイコンは img として出るだけで、押せるボタンにはならない。
    await expect(canvas.getByRole('img', { name: 'ページのアイコン' })).toHaveTextContent('📘');
    await expect(canvas.queryByRole('button', { name: 'アイコンを追加' })).toBeNull();
    await expect(canvas.queryByRole('button', { name: 'ページのアイコンを変更' })).toBeNull();
  },
};

/** 見られないページ・存在しないページ。どちらも同じ見え方にする（実在を読ませない）。 */
export const 見られないページ: Story = {
  decorators: [routerWithParam('/kb/:pageId', '/kb/p-404'), withApi(api())],
};

/** アイコンを付ける。一覧から選ぶと保存され、頭部の絵文字に変わる。 */
export const アイコンを付ける: Story = {
  decorators: [routerWithParam('/kb/:pageId', '/kb/p-1'), withApi(api())],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const addButton = await canvas.findByRole('button', { name: 'アイコンを追加' });
    await userEvent.click(addButton);

    const dialog = await canvas.findByRole('dialog', { name: 'ページのアイコンを選ぶ' });
    await userEvent.click(within(dialog).getByRole('button', { name: 'アイコンを 📘 にする' }));

    await waitFor(async () => {
      await expect(canvas.getByRole('button', { name: 'ページのアイコンを変更' })).toBeInTheDocument();
    });
    await expect(canvasElement.querySelector('[data-icon="emoji"]')).toHaveTextContent('📘');
  },
};

/** アイコンを外す。 */
export const アイコンを外す: Story = {
  decorators: [
    routerWithParam('/kb/:pageId', '/kb/p-1'),
    withApi(
      api({
        '/kb/pages/p-1': resolved({ page: { ...page, icon: { type: 'emoji', value: '📘' } } }),
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const iconButton = await canvas.findByRole('button', { name: 'ページのアイコンを変更' });
    await userEvent.click(iconButton);

    const dialog = await canvas.findByRole('dialog', { name: 'ページのアイコンを選ぶ' });
    await userEvent.click(within(dialog).getByRole('button', { name: 'アイコンを外す' }));

    await waitFor(async () => {
      await expect(canvas.getByRole('button', { name: 'アイコンを追加' })).toBeInTheDocument();
    });
  },
};

/** アイコンの変更に失敗。トーストで知らせ、ピッカーは開いたまま。 */
export const アイコンの変更に失敗: Story = {
  decorators: [
    routerWithParam('/kb/:pageId', '/kb/p-1'),
    withApi(
      api({
        '/pages/p-1/icon': () => {
          throw new Error('invalid_icon');
        },
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(await canvas.findByRole('button', { name: 'アイコンを追加' }));
    const dialog = await canvas.findByRole('dialog', { name: 'ページのアイコンを選ぶ' });
    await userEvent.click(within(dialog).getByRole('button', { name: 'アイコンを 📘 にする' }));

    await waitFor(async () => {
      await expect(canvas.getByRole('alert')).toHaveTextContent('アイコンを変更できませんでした');
    });
    // 失敗したのでピッカーは開いたまま。
    await expect(canvas.getByRole('dialog', { name: 'ページのアイコンを選ぶ' })).toBeInTheDocument();
  },
};

/** 最終編集が出る。 */
export const 最終編集が出る: Story = {
  decorators: [
    routerWithParam('/kb/:pageId', '/kb/p-1'),
    withApi(
      api({
        '/kb/pages/p-1': resolved({
          lastEditedBy: { userId: 1, name: '田中 太郎' },
          lastEditedAt: '2026-09-01T10:00:00',
        }),
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    await expect(
      await within(canvasElement).findByText(/最終編集 田中 太郎/),
    ).toBeVisible();
  },
};
