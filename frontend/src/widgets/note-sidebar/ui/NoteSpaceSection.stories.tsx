import type { ReactNode } from 'react';
import { MemoryRouter } from 'react-router-dom';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import { ToastContext, type ToastContextValue } from '@/shared/lib/hooks/useToastContext';
import type { NotePage, NotePageTree, NotePageTreeNode, NoteSpace } from '@/entities/note';
import type { NoteSpaceState } from '../model/useNoteTree';
import NoteSpaceSection from './NoteSpaceSection';

/**
 * 失敗の知らせは Context 越しに出るので、見本でも入れ物だけ用意する。
 * 本物の Provider は app 層にあり、widgets からは参照できない（層は下向きの一方通行）ので、
 * ここでは何もしない値を差す。**この見本では失敗を起こす操作をしない**ので中身は要らない。
 */
const toastStub: ToastContextValue = { toasts: [], showToast: fn(), removeToast: fn() };

function ToastStub({ children }: { children: ReactNode }) {
  return <ToastContext.Provider value={toastStub}>{children}</ToastContext.Provider>;
}

/** 行ごとに変わらない値。見本では ID と題名だけを差し替える。 */
const PAGE_DEFAULTS: Omit<NotePage, 'id' | 'title'> = {
  spaceId: 'space-design',
  createdByUserId: 1,
  createdAt: '2026-08-01T00:00:00Z',
  updatedAt: '2026-08-01T00:00:00Z',
};

function makeNode(id: string, title: string, children: NotePageTreeNode[] = []): NotePageTreeNode {
  return {
    page: { ...PAGE_DEFAULTS, id, title },
    children,
    hasHiddenChildren: false,
    parentArchived: false,
  };
}

const space: NoteSpace = {
  id: 'space-design',
  key: 'design',
  name: '設計チーム',
  createdAt: '2026-08-01T00:00:00Z',
};

const designNode = makeNode('page-design', '設計メモ', [
  makeNode('page-tree', '木の描き方'),
  makeNode('page-permission', '権限の解決'),
]);
const tree: NotePageTree = {
  pages: [designNode, makeNode('page-minutes', '今日の議事録')],
  hasHiddenChildren: false,
};

/**
 * 読み込み状態は「開いているか」「読み込み中か」「失敗したか」「木があるか」の
 * 組み合わせでしか変わらない。既定を 1 つ置いて、story では差分だけ書く。
 */
function spaceState(overrides: Partial<NoteSpaceState> = {}): NoteSpaceState {
  return { open: true, loading: false, error: null, tree, ...overrides };
}

/**
 * スペース 1 つ分の見出しと、その配下のページの木。
 *
 * スペースは同時に複数見たいので、切り替えではなく見出しとして縦に並ぶ。
 * 見出しの ＋ と ⋯ は行と違って**常時**出る（面の入口なので、見えていないと
 * どこから作るのか分からない）。
 *
 * 入れ物の幅を実物のサイドバーと同じ 256px に決め打ちしてあるのは、見出しの名前と
 * 右端の操作の詰まり具合、木の字下げまで実物と同じ絵で確かめられるようにするため。
 */
const meta = {
  title: 'note-sidebar/NoteSpaceSection',
  component: NoteSpaceSection,
  parameters: {
    layout: 'centered',
    // 見出し・読み込み中・空の知らせはどれも部品が持っている --color-text-muted（#A4A4A0）で、
    // 白地に対して 2.5:1 しかない。この部品は必ず見出しを描くので全 story で引っかかるが、
    // 直すのはトークン側（部品の実装）であって story ではないため、ここでは外す。
    a11y: { config: { rules: [{ id: 'color-contrast', enabled: false }] } },
  },
  decorators: [
    (Story) => (
      // 題名は Link、失敗の知らせは Context。実物と同じ入れ子の下に置く。
      <MemoryRouter>
        <ToastStub>
          <div className="w-64 rounded-lg border border-surface-3 bg-surface-1 p-2">
            <Story />
          </div>
        </ToastStub>
      </MemoryRouter>
    ),
  ],
  args: {
    space,
    state: spaceState(),
    workspaceSlug: 'frestyle',
    expandedPageIds: new Set(['page-design']),
    archivedMode: false,
    onToggleSpace: fn(),
    onTogglePage: fn(),
    onRetry: fn(),
    onCreatePage: fn(async () => designNode.page),
    onRenamePage: fn(async () => designNode.page),
    onArchivePage: fn(async () => {}),
    onDeletePage: fn(async () => {}),
    onUnarchivePage: fn(async () => {}),
    onRenameSpace: fn(async () => space),
    onMovePage: fn(async () => {}),
  },
} satisfies Meta<typeof NoteSpaceSection>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 開いた見出しと木。三角は下を向き、木のうち「設計メモ」だけ子を開いてある。 */
export const 開いている: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // 開閉は role ではなく aria-expanded が表す（見出しは menu を名乗らない）。
    await expect(canvas.getByRole('button', { name: '設計チーム' })).toHaveAttribute(
      'aria-expanded',
      'true',
    );
    await expect(canvas.getByRole('list', { name: '設計チーム のページ' })).toBeInTheDocument();
  },
};

/**
 * 閉じた見出し。木はそもそも描かない（開くまで取りに行かないので、
 * 閉じている間は state の中身に関わらず見出しだけになる）。
 */
export const 閉じている: Story = {
  args: { state: spaceState({ open: false, tree: null }) },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('button', { name: '設計チーム' })).toHaveAttribute(
      'aria-expanded',
      'false',
    );
    await expect(canvas.queryByRole('list', { name: '設計チーム のページ' })).toBeNull();
  },
};

/** 開いた直後、木がまだ届いていないところ。 */
export const 読み込み中: Story = {
  args: { state: spaceState({ loading: true, tree: null }) },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('読み込み中…')).toBeInTheDocument();
  },
};

/**
 * 取得に失敗したところ。理由と再試行を必ず出す。
 *
 * 黙って空にすると「ページが 1 枚も無いスペース」と見分けが付かず、
 * 利用者はページが消えたと読む。
 */
export const 読み込みに失敗: Story = {
  args: {
    state: spaceState({ error: 'ページを読み込めませんでした', tree: null }),
  },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('ページを読み込めませんでした')).toBeInTheDocument();
    await userEvent.click(canvas.getByRole('button', { name: '再試行' }));
    // やり直しはスペース単位。どのスペースを取り直すかまで渡っていること。
    await expect(args.onRetry).toHaveBeenCalledWith(space.id);
  },
};

/** ページが 1 枚も無いスペース。失敗と区別が付く言葉を置く。 */
export const ページが1枚も無い: Story = {
  args: { state: spaceState({ tree: { pages: [], hasHiddenChildren: false } }) },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('ページがありません')).toBeInTheDocument();
  },
};

/**
 * 見出しの ⋯ を開いたところ。いまは「スペースの名前を変更」だけが入る。
 *
 * 素のボタンの一覧として出している（menu を名乗ると矢印キーでの移動を約束することになる）。
 */
export const 見出しのメニューを開いたところ: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const trigger = canvas.getByRole('button', { name: '設計チーム の操作' });
    await userEvent.click(trigger);
    await expect(trigger).toHaveAttribute('aria-expanded', 'true');
    await expect(
      await canvas.findByRole('button', { name: 'スペースの名前を変更' }),
    ).toBeVisible();
  },
};

/**
 * 見出しの名前を書き換えているところ。見出しが入力欄に置き換わり、
 * 三角だけがその場に残る（畳んだり開いたりの状態は書き換え中も見えていてほしい）。
 *
 * 入口はメニューにしかないので、見本でもそこから開く。
 */
export const 見出しの改名中: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: '設計チーム の操作' }));
    await userEvent.click(await canvas.findByRole('button', { name: 'スペースの名前を変更' }));
    const input = await canvas.findByRole('textbox', { name: 'スペースの名前' });
    // いまの名前が入った状態で始まる（選択済みなので、そのまま打てば置き換わる）。
    await expect(input).toHaveValue('設計チーム');
    // 書き換え中は作る・操作の入口を出さない（名前が確定するまで押させない）。
    await expect(canvas.queryByRole('button', { name: '設計チーム にページを追加' })).toBeNull();
  },
};

/**
 * アーカイブ済みを見ているとき。見出しの ＋ と ⋯ は出さない
 * （そこには作れず、名前も現役へ戻してから変える）。空のときの言葉も変わる。
 */
export const アーカイブ済みを見ている: Story = {
  args: {
    archivedMode: true,
    state: spaceState({ tree: { pages: [], hasHiddenChildren: false } }),
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('アーカイブしたページはありません')).toBeInTheDocument();
    await expect(canvas.queryByRole('button', { name: '設計チーム の操作' })).toBeNull();
    await expect(canvas.queryByRole('button', { name: '設計チーム にページを追加' })).toBeNull();
  },
};
