import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import NoteWorkspaceSwitcher from './NoteWorkspaceSwitcher';
import type { NoteWorkspace } from '@/entities/note';

/**
 * サイドバー最上段のワークスペース切替。押すまで**閉じている**のが既定で、
 * 開くと一覧といまの 1 つの印、末尾に追加の入口が出る。
 *
 * ワークスペースが会社の境界なので「同時に 1 つ」の切替になっている
 * （並べて同時に見せる形は採らない）。ここの story はその見え方を並べて確かめるためのもの。
 */
const meta = {
  title: 'note-sidebar/NoteWorkspaceSwitcher',
  component: NoteWorkspaceSwitcher,
  parameters: { layout: 'centered' },
  decorators: [
    (Story) => (
      // 実際の置き場所（幅の決まった細いサイドバーの最上段）と同じ幅で見る。
      // ポップアップは入れ物の幅いっぱいに開くので、幅が違うと長い名前の省略のされ方が変わる。
      <div className="w-64 rounded-lg border border-surface-3 bg-surface-1 p-2">
        <Story />
      </div>
    ),
  ],
  args: {
    onSelect: fn(),
    onCreate: fn(async () => {}),
  },
} satisfies Meta<typeof NoteWorkspaceSwitcher>;

export default meta;
type Story = StoryObj<typeof meta>;

const workspace = (slug: string, name: string): NoteWorkspace => ({
  slug,
  name,
  createdAt: '2026-08-01T00:00:00Z',
});

// 3 社ぶん。長い名前を 1 つ混ぜてあるのは、幅を超えたときに省略されることまで見たいため。
const workspaces = [
  workspace('shinjuku-tech', '新宿テクノロジー'),
  workspace('marine-works', 'マリンワークス株式会社 開発本部'),
  workspace('kb-lab', 'ナレッジ基盤ラボ'),
];

/** 既定。押すまでは一覧を出さず、いまのワークスペース名だけを見せる。 */
export const 閉じている: Story = {
  args: { workspaces, activeSlug: 'shinjuku-tech' },
};

/** 開いたところ。いま選んでいる行にだけ印（チェック）が付く。 */
export const 開いている: Story = {
  args: { workspaces, activeSlug: 'marine-works' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: /マリンワークス/ }));
    const active = await canvas.findByRole('button', { name: 'マリンワークス株式会社 開発本部' });
    // 印は「いまの 1 つ」だけ。読み上げにも aria-current で同じことが伝わる。
    await expect(active).toHaveAttribute('aria-current', 'true');
    await expect(canvas.getByRole('button', { name: '新宿テクノロジー' })).toHaveAttribute(
      'aria-current',
      'false',
    );
  },
};

/**
 * 所属が 1 つだけのとき。切替の意味は無いが、**追加の入口はここにしか無い**ので
 * ポップアップ自体は出す（この入口が消えると 2 つめを作る手段が UI から無くなる）。
 */
export const ワークスペースが1件だけ: Story = {
  args: { workspaces: [workspaces[0]], activeSlug: 'shinjuku-tech' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: '新宿テクノロジー' }));
    await canvas.findByRole('button', { name: 'ワークスペースを追加' });
  },
};

/** どれも選ばれていないとき（切替前）。名前の代わりに促しが出る。 */
export const 未選択: Story = {
  args: { workspaces, activeSlug: null },
};

/** ポップアップの中で追加フォームを開いたところ。一覧は畳まずそのまま下に足す。 */
export const 追加フォームを開いたところ: Story = {
  args: { workspaces, activeSlug: 'kb-lab' },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByRole('button', { name: /ナレッジ基盤ラボ/ }));
    await userEvent.click(await canvas.findByRole('button', { name: 'ワークスペースを追加' }));
    await canvas.findByRole('textbox', { name: 'ワークスペースの名前' });
  },
};
