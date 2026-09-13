import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';
import { expect, userEvent, within } from 'storybook/test';
import { withApi, withRouter, withStore, withToast } from '../../../../.storybook/decorators';
import { setCurrentKbSpace } from '@/entities/kb';
import Header from './Header';

/**
 * 画面のいちばん上に固定される帯。左に印、中央に行き先と検索、右に知らせと自分のメニュー。
 * 常時表示（本文の上には重ねない・自動的には隠れない）で、地は不透明。
 *
 * 狭い画面ではナビが畳まれ、三本線のボタンから縦に開く。検索は虫眼鏡アイコンに畳む。
 *
 * 素のリンクで表せる項目（ホーム）は `model/navigation.ts` の MAIN_NAV_ITEMS が持つ。
 * 「スペース ▾」「最近見たページ ▾」はドロップダウンを持つため専用コンポーネント
 * （HeaderSpacesNav・HeaderRecentPagesNav）で描画する。
 *
 * ワークスペース切替はここには無い（`KbSidebar` 先頭にある）。
 */
const meta = {
  title: 'widgets/app-shell/Header',
  component: Header,
  parameters: { layout: 'fullscreen' },
  args: {
    onOpenSearch: fn(),
  },
  decorators: [
    withRouter,
    withStore(),
    withToast,
    withApi({
      '/profile/me': { displayName: '川野 拓馬', avatarUrl: null, email: 'takuma@example.com' },
      // 件数の宛先は一覧の宛先を含むので、細かいほうを先に書く。
      '/notifications/unread-count': 0,
      '/kb/workspaces': [
        { slug: 'w-3f2a9c', name: '開発チーム', createdAt: '2026-01-01T00:00:00Z', canManage: true },
      ],
    }),
    (Story) => (
      <div className="min-h-[320px] bg-surface">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof Header>;

export default meta;
type Story = StoryObj<typeof meta>;

/** ふだんの見え方。今いるスペースが分からない（KB のページを開いていない）ときは
 *  「スペース」が素のリンクになり、「作成」は出ない。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    setCurrentKbSpace(null);
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('navigation', { name: 'メインナビゲーション' })).toBeVisible();
    await expect(canvas.getByRole('link', { name: 'スペース' })).toBeVisible();
    await expect(canvas.getByRole('button', { name: /最近見たページ/ })).toBeVisible();
    await expect(canvas.queryByRole('button', { name: /作成/ })).toBeNull();
    await expect(await canvas.findByText('川野 拓馬')).toBeVisible();
  },
};

/** ナレッジのページを開いていて「今いるスペース」が分かっているとき。
 *  「スペース」がドロップダウンになり、「作成」ボタンも出る。 */
export const スペースが分かっているとき: Story = {
  decorators: [
    withApi({
      '/profile/me': { displayName: '川野 拓馬', avatarUrl: null, email: 'takuma@example.com' },
      '/notifications/unread-count': 0,
      '/kb/workspaces/w-3f2a9c/me/spaces': [
        { id: 's-1', name: '開発チーム', role: 'editor' },
        { id: 's-2', name: '営業定例', role: 'viewer' },
      ],
      '/kb/me/recent-pages': [
        {
          pageId: 'p-1',
          workspaceSlug: 'w-3f2a9c',
          title: '設計メモ',
          spaceId: 's-1',
          spaceName: '開発チーム',
          viewedAt: '2026-09-13T00:00:00Z',
        },
      ],
      '/kb/workspaces': [
        { slug: 'w-3f2a9c', name: '開発チーム', createdAt: '2026-01-01T00:00:00Z', canManage: true },
      ],
    }),
  ],
  play: async ({ canvasElement }) => {
    setCurrentKbSpace({ workspaceSlug: 'w-3f2a9c', spaceId: 's-1' });
    const canvas = within(canvasElement);
    await expect(await canvas.findByRole('button', { name: /作成/ })).toBeVisible();

    await userEvent.click(canvas.getByRole('button', { name: /^スペース/ }));
    await expect(await canvas.findByRole('link', { name: '営業定例' })).toBeVisible();

    await userEvent.click(canvas.getByRole('button', { name: /最近見たページ/ }));
    await expect(await canvas.findByRole('link', { name: /設計メモ/ })).toBeVisible();
  },
};

/** 中央の検索ボタンを押すと onOpenSearch が呼ばれる。 */
export const 検索ボタンを押す: Story = {
  play: async ({ canvasElement, args }) => {
    setCurrentKbSpace(null);
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: '検索' }));
    await expect(args.onOpenSearch).toHaveBeenCalledTimes(1);
  },
};

/** 未読の知らせがあるとき。ベルに件数が付く。 */
export const 未読あり: Story = {
  decorators: [
    withApi({
      '/profile/me': { displayName: '川野 拓馬', avatarUrl: null, email: 'takuma@example.com' },
      '/notifications/unread-count': 5,
      '/kb/workspaces': [],
    }),
  ],
  play: async ({ canvasElement }) => {
    setCurrentKbSpace(null);
    await expect(
      await within(canvasElement).findByRole('link', { name: '通知 (未読 5 件)' }),
    ).toBeVisible();
  },
};

/** 未読が多すぎるとき。3 桁は「99+」に丸めて、ベルを押し広げない。 */
export const 未読が多い: Story = {
  decorators: [
    withApi({
      '/profile/me': { displayName: '川野 拓馬', avatarUrl: null, email: 'takuma@example.com' },
      '/notifications/unread-count': 128,
      '/kb/workspaces': [],
    }),
  ],
  play: async ({ canvasElement }) => {
    setCurrentKbSpace(null);
    await expect(await within(canvasElement).findByText('99+')).toBeVisible();
  },
};

/** 自分の情報が取れなかったとき。壊れずに「ユーザー」と出す。 */
export const 情報が取れないとき: Story = {
  decorators: [withApi({ '/kb/workspaces': [] })],
  play: async ({ canvasElement }) => {
    setCurrentKbSpace(null);
    const canvas = within(canvasElement);
    // ナビは出る。名前だけが既定の文言になる。
    await expect(canvas.getByRole('navigation', { name: 'メインナビゲーション' })).toBeVisible();
    await expect(await canvas.findByText('ユーザー')).toBeVisible();
  },
};

/** 狭い画面。ナビが畳まれ、三本線から開く。 */
export const 狭い画面: Story = {
  globals: { viewport: { value: 'mobile1', isRotated: false } },
  play: async ({ canvasElement }) => {
    setCurrentKbSpace(null);
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'メニュー' }));
    await expect(canvas.getByRole('navigation', { name: 'モバイルナビゲーション' })).toBeVisible();
  },
};
