import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import KbCommentThreadCard from './KbCommentThreadCard';
import type { KbCommentThread } from '@/entities/kb';

/**
 * コメントスレッド 1 件（作成者・コメントの列挙・解決状態）。
 *
 * 読むことは誰でもできるが、書く（返信）・解決する・再開するのは canComment が true の
 * 人だけ。false の人には返信欄も解決/再開ボタンも出さない — 押しても 403 が返るだけの
 * ボタンは、権限が無いことすら伝えない（SharePanel の共有ボタンと同じ理由）。
 */
const meta = {
  title: 'pages/kb/KbCommentThreadCard',
  component: KbCommentThreadCard,
  parameters: { layout: 'padded' },
  args: {
    canComment: true,
    onReply: fn(async () => {}),
    onResolve: fn(async () => {}),
    onReopen: fn(async () => {}),
  },
  decorators: [
    (Story) => (
      <div className="w-96">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof KbCommentThreadCard>;

export default meta;
type Story = StoryObj<typeof meta>;

const unresolvedThread: KbCommentThread = {
  id: 't-1',
  createdBy: { userId: 1, name: '田中 太郎' },
  resolvedAt: null,
  resolvedBy: null,
  createdAt: '2026-09-01T10:00:00Z',
  comments: [
    {
      id: 'c-1',
      author: { userId: 1, name: '田中 太郎' },
      body: [{ type: 'text', text: 'この段落、もう少し具体例が欲しいです。' }],
      createdAt: '2026-09-01T10:00:00Z',
      updatedAt: '2026-09-01T10:00:00Z',
    },
    {
      id: 'c-2',
      author: { userId: 2, name: '鈴木 花子' },
      body: [{ type: 'text', text: '同感です。追記しますね。' }],
      createdAt: '2026-09-01T11:00:00Z',
      updatedAt: '2026-09-01T11:00:00Z',
    },
  ],
};

const resolvedThread: KbCommentThread = {
  ...unresolvedThread,
  id: 't-2',
  resolvedAt: '2026-09-02T09:00:00Z',
  resolvedBy: { userId: 2, name: '鈴木 花子' },
  comments: [unresolvedThread.comments[0]],
};

/** 未解決・コメントできる。返信欄と「解決」ボタンが出る。 */
export const 未解決_コメントできる: Story = {
  args: { thread: unresolvedThread },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    // 「田中 太郎」はスレッドの作成者名とコメントの投稿者名の両方に出るので複数件ヒットする。
    await expect(canvas.getAllByText('田中 太郎').length).toBeGreaterThan(0);
    await expect(canvas.getByText('この段落、もう少し具体例が欲しいです。')).toBeVisible();
    await expect(canvas.getByText('同感です。追記しますね。')).toBeVisible();

    const resolveButton = canvas.getByRole('button', { name: '解決' });
    await userEvent.click(resolveButton);
    await expect(args.onResolve).toHaveBeenCalledWith('t-1');

    // 返信も送れる。
    const reply = canvas.getByPlaceholderText('返信を書く…');
    await userEvent.type(reply, '直しました');
    await userEvent.click(canvas.getByRole('button', { name: '送信' }));
    await expect(args.onReply).toHaveBeenCalledWith('t-1', [{ type: 'text', text: '直しました' }]);
  },
};

/** 未解決・コメントできない。読めるが、返信欄も「解決」ボタンも出ない。 */
export const 未解決_コメントできない: Story = {
  args: { thread: unresolvedThread, canComment: false },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('この段落、もう少し具体例が欲しいです。')).toBeVisible();
    await expect(canvas.queryByRole('button', { name: '解決' })).not.toBeInTheDocument();
    await expect(canvas.queryByPlaceholderText('返信を書く…')).not.toBeInTheDocument();
  },
};

/** 解決済み・コメントできる。解決者名が出て、ボタンは「再開」に変わる。 */
export const 解決済み_コメントできる: Story = {
  args: { thread: resolvedThread },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText(/鈴木 花子 が解決済みにしました/)).toBeVisible();

    const reopenButton = canvas.getByRole('button', { name: '再開' });
    await expect(canvas.queryByRole('button', { name: '解決' })).not.toBeInTheDocument();
    await userEvent.click(reopenButton);
    await expect(args.onReopen).toHaveBeenCalledWith('t-2');
  },
};

/** 解決済み・コメントできない。解決者名は見えるが、再開ボタンは出ない。 */
export const 解決済み_コメントできない: Story = {
  args: { thread: resolvedThread, canComment: false },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText(/鈴木 花子 が解決済みにしました/)).toBeVisible();
    await expect(canvas.queryByRole('button', { name: '再開' })).not.toBeInTheDocument();
  },
};
