import type { ReactNode } from 'react';
import { MemoryRouter } from 'react-router-dom';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, waitFor, within } from 'storybook/test';
import type { NotePage, NotePageTreeNode } from '@/entities/note';
import NotePageRow from './NotePageRow';
import NoteHiddenChildrenRow from './NoteHiddenChildrenRow';

/** 行ごとに変わらない値。story では ID と題名だけを差し替える。 */
const PAGE_DEFAULTS: Omit<NotePage, 'id' | 'title'> = {
  spaceId: 'space-1',
  createdByUserId: 1,
  createdAt: '2026-08-01T00:00:00Z',
  updatedAt: '2026-08-01T00:00:00Z',
};

function makeNode(
  id: string,
  title: string,
  options: {
    children?: NotePageTreeNode[];
    hasHiddenChildren?: boolean;
    archived?: boolean;
    parentArchived?: boolean;
  } = {},
): NotePageTreeNode {
  return {
    page: {
      ...PAGE_DEFAULTS,
      id,
      title,
      // アーカイブ済みかどうかは日時の有無で表す（真偽値の列は無い）。
      ...(options.archived ? { archivedAt: '2026-08-20T09:00:00Z' } : {}),
    },
    children: options.children ?? [],
    hasHiddenChildren: options.hasHiddenChildren ?? false,
    parentArchived: options.parentArchived ?? false,
  };
}

/**
 * 木の 1 段と同じ入れ物（ul > li）。行そのものは div なので、
 * 段の入れ子は外側の ul が表す（NoteTreeList と同じ形）。
 */
function RowList({ children }: { children: ReactNode }) {
  return <ul className="space-y-px">{children}</ul>;
}

const minutesNode = makeNode('page-minutes', '今日の議事録');
const treeDesignNode = makeNode('page-tree', '木の描き方');
const permissionDesignNode = makeNode('page-permission', '権限の解決');
const designNode = makeNode('page-design', '設計メモ', {
  children: [treeDesignNode, permissionDesignNode],
});

/**
 * サイドバーの木の 1 行。題名へのリンク・開閉の三角・葉の「・」・
 * 紙とフォルダの印・触れている間だけ出る ＋ と ⋯ が同じ行に載る。
 *
 * 入れ物の幅を実物のサイドバーと同じ 256px に決め打ちしてあるのは、
 * 「長い題名が折り返さず省略されるか」まで見本で確かめられるようにするため。
 */
const meta = {
  title: 'note-sidebar/NotePageRow',
  component: NotePageRow,
  parameters: { layout: 'centered' },
  decorators: [
    (Story) => (
      // 題名は Link。木は必ずルーターの下に置かれるので、見本でも同じにする。
      <MemoryRouter>
        <div className="w-64 rounded-lg border border-surface-3 bg-surface-1 p-2">
          <Story />
        </div>
      </MemoryRouter>
    ),
  ],
  args: {
    node: minutesNode,
    depth: 0,
    siblings: [minutesNode],
    index: 0,
    parentId: null,
    expanded: false,
    workspaceSlug: 'frestyle',
    active: false,
    renaming: false,
    archivedMode: false,
    dragging: false,
    dropZone: null,
    onToggle: fn(),
    onStartRename: fn(),
    onCancelRename: fn(),
    onCommitRename: fn(),
    onCreateChild: fn(),
    onArchive: fn(),
    onUnarchive: fn(),
    onDelete: fn(),
    onMove: fn(),
    onDragStart: fn(),
    onDragEnd: fn(),
    onDragOverRow: fn(),
    onDropOnRow: fn(),
  },
} satisfies Meta<typeof NotePageRow>;

export default meta;
type Story = StoryObj<typeof meta>;

/** 何も起きていない行。子が無いので三角の位置は「・」、印は紙。 */
export const 通常の行: Story = {
  render: (args) => (
    <RowList>
      <li>
        <NotePageRow {...args} />
      </li>
    </RowList>
  ),
};

/**
 * いま開いているページの行。**背景と字の太さ**だけで示す。
 *
 * 文字色まで変えると行の中の ＋ や ⋯ まで色を継ぎ、木全体が青く見える。
 * ここでは文字が本文と同じ黒のままであることを実際に測って確かめる。
 */
export const いま開いている行: Story = {
  args: { active: true },
  render: (args) => (
    <RowList>
      <li>
        <NotePageRow {...args} />
      </li>
    </RowList>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const link = canvas.getByRole('link', { name: '今日の議事録' });
    // 開いているページであることは role ではなく aria-current が表す。
    await expect(link).toHaveAttribute('aria-current', 'page');
    await expect(getComputedStyle(link).color).toBe(getComputedStyle(document.body).color);
  },
};

/**
 * 葉・閉じた親・開いた親の 3 つを並べる。
 *
 * 三角は段の折り畳み、印は行の性質を表す。子を持つ行は開いている間だけ
 * 開いたフォルダになり、三角と印が同じ状態を指す。
 */
export const 葉と親の見分け: Story = {
  render: (args) => (
    <RowList>
      <li>
        <NotePageRow
          {...args}
          node={minutesNode}
          siblings={[minutesNode, designNode]}
          index={0}
        />
      </li>
      <li>
        <NotePageRow
          {...args}
          node={designNode}
          siblings={[minutesNode, designNode]}
          index={1}
          expanded={false}
        />
      </li>
      <li>
        <NotePageRow {...args} node={designNode} siblings={[designNode]} index={0} expanded />
        <RowList>
          <li>
            <NotePageRow
              {...args}
              node={treeDesignNode}
              depth={1}
              siblings={designNode.children}
              index={0}
              parentId={designNode.page.id}
            />
          </li>
          <li>
            <NotePageRow
              {...args}
              node={permissionDesignNode}
              depth={1}
              siblings={designNode.children}
              index={1}
              parentId={designNode.page.id}
            />
          </li>
        </RowList>
      </li>
    </RowList>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // 印は aria-hidden なので role では引けない。data-icon で見分ける。
    await expect(canvasElement.querySelectorAll('[data-icon="page"]')).toHaveLength(3);
    await expect(canvasElement.querySelectorAll('[data-icon="page-group"]')).toHaveLength(1);
    await expect(canvasElement.querySelectorAll('[data-icon="page-group-open"]')).toHaveLength(1);
    // 同じページでも、閉じている行と開いている行で三角の名前が変わる。
    await expect(canvas.getByRole('button', { name: '設計メモ を開く' })).toBeInTheDocument();
    await expect(canvas.getByRole('button', { name: '設計メモ を閉じる' })).toBeInTheDocument();
  },
};

/** 3 段の入れ子。1 段につき 14px 下がるので、段が目で追える。 */
export const 深い入れ子: Story = {
  render: (args) => {
    const grandChild = makeNode('page-grandchild', '分数インデックス');
    const child = makeNode('page-child', '並び順の決め方', { children: [grandChild] });
    const parent = makeNode('page-parent', 'ナレッジ基盤', { children: [child] });
    return (
      <RowList>
        <li>
          <NotePageRow {...args} node={parent} siblings={[parent]} index={0} expanded />
          <RowList>
            <li>
              <NotePageRow
                {...args}
                node={child}
                depth={1}
                siblings={[child]}
                index={0}
                parentId={parent.page.id}
                expanded
              />
              <RowList>
                <li>
                  <NotePageRow
                    {...args}
                    node={grandChild}
                    depth={2}
                    siblings={[grandChild]}
                    index={0}
                    parentId={child.page.id}
                  />
                </li>
              </RowList>
            </li>
          </RowList>
        </li>
      </RowList>
    );
  },
};

/**
 * 幅に収まらない題名。折り返して行の高さが変わると木の並びが崩れるので、
 * 1 行のまま末尾を省略する。切り詰められていることを幅を測って確かめる。
 */
export const 長い題名: Story = {
  render: (args) => {
    const longTitleNode = makeNode(
      'page-long-title',
      'ページの移動でどの兄弟の隣に置くかを指定できるようにするための設計メモ',
    );
    return (
      <RowList>
        <li>
          <NotePageRow {...args} node={longTitleNode} siblings={[longTitleNode]} index={0} />
        </li>
      </RowList>
    );
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const title = canvas.getByText(/ページの移動でどの兄弟の隣に/);
    // 見えている幅より中身が広い = 折り返さずに省略されている。
    await expect(title.scrollWidth).toBeGreaterThan(title.clientWidth);
  },
};

/**
 * アーカイブ一覧での見え方。作る・名前を変える・動かすは出さず、「復帰」だけ。
 *
 * 復帰できるのはアーカイブの根だけなので、親がまだアーカイブ中の子には出さない
 * （押せるのに必ず断られるボタンになる）。触れている間だけ出るボタンなので、
 * 見本では根の「復帰」にフォーカスを当てて見えるようにしてある。
 */
export const アーカイブ済みの行: Story = {
  args: { archivedMode: true },
  parameters: {
    // 「復帰」の文字色は部品が持っている --color-text-muted（#A4A4A0）で、白地に対して
    // 2.5:1 しかない。見えている状態を見本に出す以上ここで必ず引っかかるが、直すのは
    // トークン側（部品の実装）であって story ではないので、この 1 つだけ外す。
    a11y: { config: { rules: [{ id: 'color-contrast', enabled: false }] } },
  },
  render: (args) => {
    const archivedChild = makeNode('page-archived-child', '古い議事録', {
      archived: true,
      parentArchived: true,
    });
    const archivedRoot = makeNode('page-archived-root', '2025 年度の記録', {
      archived: true,
      children: [archivedChild],
    });
    return (
      <RowList>
        <li>
          <NotePageRow {...args} node={archivedRoot} siblings={[archivedRoot]} index={0} expanded />
          <RowList>
            <li>
              <NotePageRow
                {...args}
                node={archivedChild}
                depth={1}
                siblings={[archivedChild]}
                index={0}
                parentId={archivedRoot.page.id}
              />
            </li>
          </RowList>
        </li>
      </RowList>
    );
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    // 復帰はアーカイブの根の 1 行だけ。子（親がアーカイブ中）には出ない。
    const restore = canvas.getAllByRole('button', { name: '復帰' });
    await expect(restore).toHaveLength(1);
    // ＋ と ⋯ はアーカイブ一覧では出さない（現役に戻してから）。
    await expect(canvas.queryByRole('button', { name: /の操作$/ })).toBeNull();
    restore[0].focus();
    // 出方は transition なので、濃くなり切るまで待つ（見本の絵もこの状態で撮られる）。
    await waitFor(async () => {
      await expect(restore[0]).toBeVisible();
    });
  },
};

/**
 * 自分には見えない子を持つ行。
 *
 * 印だけを出し、枚数も題名も出さない（そもそも API が返してこない）。
 * 開ける子が 1 枚も無いので三角は出ず、行そのものは葉と同じ「・」＋紙のまま。
 */
export const 見えない子がある行: Story = {
  parameters: {
    // 印の文字色も --color-text-muted（白地に 2.5:1）。上と同じ理由でここだけ外す。
    a11y: { config: { rules: [{ id: 'color-contrast', enabled: false }] } },
  },
  render: (args) => {
    const restrictedNode = makeNode('page-restricted', '人事の共有メモ', {
      hasHiddenChildren: true,
    });
    return (
      <RowList>
        <li>
          <NotePageRow {...args} node={restrictedNode} siblings={[restrictedNode]} index={0} />
          {/* 印は行の 1 段内側に出す。置き場所も NoteTreeList と同じにしてある。 */}
          <NoteHiddenChildrenRow depth={1} />
        </li>
      </RowList>
    );
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('表示できないページがあります')).toBeInTheDocument();
    // 伏せた子しか居ない行はフォルダにしない（開閉の三角が無いのにフォルダ、を避ける）。
    await expect(canvasElement.querySelectorAll('[data-icon="page"]')).toHaveLength(1);
    await expect(canvas.queryByRole('button', { name: /を開く$/ })).toBeNull();
  },
};

/**
 * ＋ と ⋯ は普段は消えていて、触れているか、キーボードでたどり着いたときだけ出る。
 * 見本では ⋯ にフォーカスを当てて、出たところを見せる。
 */
export const 操作が出ている行: Story = {
  render: (args) => (
    <RowList>
      <li>
        <NotePageRow {...args} />
      </li>
    </RowList>
  ),
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    canvas.getByRole('button', { name: '今日の議事録 の操作' }).focus();
    // opacity で消しているだけで DOM からは外していない（Tab の順序を変えないため）。
    // 濃くなるのは transition なので、待ってから見る。
    await waitFor(async () => {
      await expect(canvas.getByRole('button', { name: '今日の議事録 の下にページを追加' })).toBeVisible();
    });
  },
};
