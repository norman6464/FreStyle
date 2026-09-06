import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';
import { withApi, withRouter, withStore, withToast } from '../../../../.storybook/decorators';
import Header from './Header';

/**
 * 画面のいちばん上に固定される帯。左に印、中央に行き先、右に知らせと自分のメニュー。
 *
 * 下の境目は線ではなく**ぼかし**にしてある。線だと区切りが強すぎて、本文が続いていく画面で
 * 視線がそこで止まる。半透明の地色越しに下が透けることで、境目だけが分かる。
 *
 * 狭い画面ではナビが畳まれ、三本線のボタンから縦に開く。
 *
 * 行き先の一覧は 1 か所（`model/navigation.ts`）だけが持っていて、この帯・畳んだメニュー・
 * サイドバーが同じものを読む。増やすときも 1 行足せば全部に出る。
 */
const meta = {
  title: 'widgets/app-shell/Header',
  component: Header,
  parameters: { layout: 'fullscreen' },
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

/** ふだんの見え方。 */
export const 既定: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('navigation', { name: 'メインナビゲーション' })).toBeVisible();
    await expect(canvas.getByRole('link', { name: 'ナレッジ' })).toBeVisible();
    await expect(await canvas.findByText('川野 拓馬')).toBeVisible();
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
    await expect(await within(canvasElement).findByText('99+')).toBeVisible();
  },
};

/** 自分の情報が取れなかったとき。壊れずに「ユーザー」と出す。 */
export const 情報が取れないとき: Story = {
  decorators: [withApi({ '/kb/workspaces': [] })],
  play: async ({ canvasElement }) => {
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
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: 'メニュー' }));
    await expect(canvas.getByRole('navigation', { name: 'モバイルナビゲーション' })).toBeVisible();
  },
};
