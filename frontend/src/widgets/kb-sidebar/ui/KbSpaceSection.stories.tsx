import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import type { KbPage, KbPageTreeNode, KbSpace } from '@/entities/kb';
import { withRouter, withToast } from '../../../../.storybook/decorators';
import KbSpaceSection from './KbSpaceSection';

/**
 * スペース 1 つぶんの見出しと、その下のページの木。
 *
 * スペースは**同時に複数見たい**（自分のページと部署のページを行き来する）ので、切り替えでは
 * なく見出しとして並べる。ワークスペースは会社の境目で同時に見る場面が無いから、あちらは
 * 切り替えにしてある。同時に見たいものは並べ、どちらか一方のものは切り替える。
 *
 * 取得に失敗したときは**黙って空にしない**。理由を出して、もう一度試せるようにする。
 * 空の木と、取れなかった木は、見た目が同じでは困る。
 */
const meta = {
  title: 'widgets/kb-sidebar/KbSpaceSection',
  component: KbSpaceSection,
  parameters: { layout: 'padded' },
  decorators: [
    withRouter,
    withToast,
    (Story) => (
      <div className="w-64 bg-surface-1 p-2">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof KbSpaceSection>;

export default meta;
type Story = StoryObj<typeof meta>;

const space: KbSpace = {
  id: 's-1',
  key: 's-1a2b3c',
  name: 'バックエンド定例',
  visibility: 'workspace',
  createdAt: '2026-01-01T00:00:00Z',
};

const page = (id: string, title: string): KbPage => ({
  id,
  spaceId: 's-1',
  title,
  createdByUserId: 1,
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
});

const node = (id: string, title: string, children: KbPageTreeNode[] = []): KbPageTreeNode => ({
  page: page(id, title),
  children,
  hasHiddenChildren: false,
  parentArchived: false,
});

const callbacks = {
  onToggleSpace: fn(),
  onTogglePage: fn(),
  onRetry: fn(),
  onCreatePage: fn(async () => page('p-new', '無題')),
  onRenamePage: fn(async () => page('p-1', '新しい題名')),
  onArchivePage: fn(async () => {}),
  onDeletePage: fn(async () => {}),
  onUnarchivePage: fn(async () => {}),
  onRenameSpace: fn(async () => space),
  onMovePage: fn(async () => {}),
};

const base = {
  ...callbacks,
  space,
  workspaceSlug: 'w-3f2a9c',
  expandedPageIds: new Set<string>(),
  archivedMode: false,
};

/** 閉じているとき。見出しだけ。 */
export const 閉じている: Story = {
  args: { ...base, state: undefined },
  play: async ({ canvasElement }) => {
    await expect(
      within(canvasElement).getByRole('button', { name: /^バックエンド定例$/ }),
    ).toBeVisible();
  },
};

/** 開いてページが並んでいるとき。 */
export const 開いている: Story = {
  args: {
    ...base,
    state: {
      open: true,
      loading: false,
      error: null,
      tree: {
        pages: [node('p-1', 'はじめに'), node('p-2', '議事録', [node('p-3', '9 月 1 日')])],
        hasHiddenChildren: false,
      },
    },
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('link', { name: 'はじめに' })).toBeVisible();
  },
};

/** 読み込んでいる最中。 */
export const 読み込み中: Story = {
  args: { ...base, state: { open: true, loading: true, error: null, tree: null } },
};

/** まだ 1 枚も無いとき。 */
export const ページが無い: Story = {
  args: {
    ...base,
    state: { open: true, loading: false, error: null, tree: { pages: [], hasHiddenChildren: false } },
  },
};

/** 取れなかったとき。理由を出して、もう一度試せるようにする。 */
export const 取得に失敗: Story = {
  args: {
    ...base,
    state: { open: true, loading: false, error: 'ページを取得できませんでした', tree: null },
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/取得できませんでした/)).toBeVisible();
  },
};

/** 見えないページが在るとき。 */
export const 見えないページが在る: Story = {
  args: {
    ...base,
    state: {
      open: true,
      loading: false,
      error: null,
      tree: { pages: [node('p-1', 'はじめに')], hasHiddenChildren: true },
    },
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('表示できないページがあります')).toBeVisible();
  },
};

/** アーカイブ済みを見ているとき。 */
export const アーカイブ済み: Story = {
  args: {
    ...base,
    archivedMode: true,
    state: {
      open: true,
      loading: false,
      error: null,
      tree: { pages: [node('p-9', '古い議事録')], hasHiddenChildren: false },
    },
  },
};

/** 見出しを押すと開閉が親へ伝わる。 */
export const 見出しを押す: Story = {
  args: { ...base, state: undefined },
  play: async ({ args, canvasElement }) => {
    await userEvent.click(
      within(canvasElement).getByRole('button', { name: /^バックエンド定例$/ }),
    );
    await expect(args.onToggleSpace).toHaveBeenCalledWith('s-1');
  },
};
