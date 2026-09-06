import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import type { KbSpace } from '@/entities/kb';
import { withApi, withRouter } from '../../../../.storybook/decorators';
import KbSearchDialog from './KbSearchDialog';

/**
 * ワークスペース全体から題名でページを探す窓。
 *
 * サイドバー本体は「場所（木）」を示すことに徹し、探すのはこの窓に分けてある。
 * 常設の入力欄を置くと、狭いサイドバーの中で木と結果が同じ面を取り合うことになる。
 *
 * 探すのはサーバー。返るのは木と同じ規則で見てよい現役のページだけで、
 * **検索だけ別の判定にしない**（別にすると、見えないはずのページが検索からだけ見える）。
 *
 * 打つたびに投げず 250 ミリ秒待つ。速く打ったときに古い応答が新しい結果を上書きしないよう、
 * 世代番号で捨てている。
 */
const meta = {
  title: 'widgets/kb-sidebar/KbSearchDialog',
  component: KbSearchDialog,
  parameters: { layout: 'fullscreen' },
  args: { workspaceSlug: 'w-3f2a9c', onClose: fn() },
  decorators: [
    withRouter,
    (Story) => (
      <div className="min-h-[520px] bg-surface p-6">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof KbSearchDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

const spaces: KbSpace[] = [
  {
    id: 's-1',
    key: 's-1a2b3c',
    name: 'バックエンド定例',
    visibility: 'workspace',
    createdAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 's-2',
    key: 's-9d8c7b',
    name: '営業定例',
    visibility: 'private',
    createdAt: '2026-02-01T00:00:00Z',
  },
];

const page = (id: string, spaceId: string, title: string) => ({
  id,
  spaceId,
  title,
  createdByUserId: 1,
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
});

/** 開いた直後。まだ何も打っていない。 */
export const 開いた直後: Story = {
  decorators: [withApi({ '/search': [] })],
  args: { spaces },
  play: async ({ canvasElement }) => {
    const input = within(canvasElement).getByRole('combobox');
    await expect(input).toHaveFocus();
  },
};

/** 見つかったとき。スペースごとに見出しを付けて並べる。 */
export const 見つかった: Story = {
  decorators: [
    withApi({
      '/search': [page('p-1', 's-1', '設計メモ'), page('p-2', 's-2', '設計レビューの進め方')],
    }),
  ],
  args: { spaces },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(canvas.getByRole('combobox'), '設計');
    await waitFor(
      async () => {
        await expect(canvas.getByText('設計メモ')).toBeVisible();
      },
      { timeout: 5000 },
    );
    await expect(canvas.getByText('設計レビューの進め方')).toBeVisible();
  },
};

/** 見つからなかったとき。空欄にせず、その旨を出す。 */
export const 見つからない: Story = {
  decorators: [withApi({ '/search': [] })],
  args: { spaces },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(canvas.getByRole('combobox'), 'みつからない語');
    await waitFor(
      async () => {
        await expect(canvas.getByText('一致するページがありません')).toBeVisible();
      },
      { timeout: 5000 },
    );
  },
};

/** 探せなかったとき（通信の失敗）。黙って空にせず、もう一度試せるようにする。 */
export const 失敗したとき: Story = {
  // 見本に無い宛先は 404 を返すので、失敗の道筋がそのまま通る。
  decorators: [withApi({})],
  args: { spaces },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.type(canvas.getByRole('combobox'), '設計');
    await waitFor(
      async () => {
        await expect(canvas.getByText('検索に失敗しました')).toBeVisible();
        await expect(canvas.getByRole('button', { name: '再試行' })).toBeVisible();
      },
      { timeout: 5000 },
    );
  },
};

/** Escape で閉じる。 */
export const Escapeで閉じる: Story = {
  decorators: [withApi({ '/search': [] })],
  args: { spaces },
  play: async ({ args, canvasElement }) => {
    await userEvent.type(within(canvasElement).getByRole('combobox'), '{Escape}');
    await expect(args.onClose).toHaveBeenCalled();
  },
};
