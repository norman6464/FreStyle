import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, screen, userEvent, waitFor, within } from 'storybook/test';
import TicketCommentSection from './TicketCommentSection';
import { withApi, withToast, type ApiStubs } from '../../../../.storybook/decorators';

const PROFILE = { userId: 1, displayName: 'norman6464', email: '', bio: '', avatarUrl: '', status: '', updatedAt: '' };

function comment(over: Record<string, unknown> & { id: string }) {
  return {
    parentCommentId: undefined,
    author: { userId: 1, name: '田中 太郎' },
    body: [{ type: 'text', text: 'コメント本文' }],
    edited: false,
    reactions: [],
    createdAt: '2026-09-10T00:00:00Z',
    updatedAt: '2026-09-10T00:00:00Z',
    ...over,
  };
}

const MEMBERS = [
  { principalId: 'p-1', userId: 1, name: 'norman6464' },
  { principalId: 'p-2', userId: 2, name: '佐藤 花子' },
];

function baseApi(over: ApiStubs = {}): ApiStubs {
  return {
    '/profile/me': PROFILE,
    '/kb/workspaces/acme/tickets/t-1/comments': { comments: [] },
    '/kb/workspaces/acme/members': MEMBERS,
    ...over,
  };
}

const meta = {
  title: 'pages/backlog/TicketCommentSection',
  component: TicketCommentSection,
  args: { workspaceSlug: 'acme', ticketId: 't-1' },
  decorators: [withToast, (Story) => <div className="w-[420px] bg-surface-1 p-3"><Story /></div>],
} satisfies Meta<typeof TicketCommentSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 空: Story = {
  decorators: [withApi(baseApi())],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('まだコメントはありません')).toBeInTheDocument();
    });
    await expect(canvas.getByRole('textbox', { name: 'コメントを書く' })).toBeInTheDocument();
  },
};

export const 失敗: Story = {
  decorators: [
    withApi({
      '/profile/me': PROFILE,
      '/kb/workspaces/acme/tickets/t-1/comments': () => {
        const err = new Error('network') as Error & { isAxiosError: boolean; response: unknown };
        err.isAxiosError = true;
        err.response = { status: 500, data: {} };
        throw err;
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByRole('alert')).toHaveTextContent('読み込めませんでした');
    });
    await expect(canvas.getByRole('button', { name: '再読み込み' })).toBeInTheDocument();
  },
};

export const 返信つき: Story = {
  decorators: [
    withApi(
      baseApi({
        '/kb/workspaces/acme/tickets/t-1/comments': {
          comments: [
            comment({ id: 'c-1' }),
            comment({ id: 'c-2', parentCommentId: 'c-1', author: { userId: 2, name: '佐藤 花子' }, body: [{ type: 'text', text: '返信です' }] }),
          ],
        },
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('返信です')).toBeInTheDocument();
    });
    await expect(canvas.getAllByRole('article')).toHaveLength(2);
  },
};

export const 反応つき自分の発言には操作が出る: Story = {
  decorators: [
    withApi(
      baseApi({
        '/kb/workspaces/acme/tickets/t-1/comments': {
          comments: [comment({ id: 'c-1', reactions: [{ userId: 1, emoji: '👍' }, { userId: 2, emoji: '👍' }] })],
        },
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByLabelText('👍 の反応 2 件（自分も押しています）')).toBeInTheDocument();
    });
    // author.userId(1) === profile.userId(1) なので自分の発言。操作メニューが出る。
    await expect(canvas.getByRole('button', { name: '田中 太郎 の発言の操作' })).toBeInTheDocument();
  },
};

export const 他人の発言には操作が出ない: Story = {
  decorators: [
    withApi(
      baseApi({
        '/kb/workspaces/acme/tickets/t-1/comments': {
          comments: [comment({ id: 'c-1', author: { userId: 99, name: '鈴木 一郎' } })],
        },
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('鈴木 一郎')).toBeInTheDocument();
    });
    await expect(canvas.queryByRole('button', { name: '鈴木 一郎 の発言の操作' })).toBeNull();
  },
};

export const 反応ピッカーを開く: Story = {
  decorators: [withApi(baseApi({ '/kb/workspaces/acme/tickets/t-1/comments': { comments: [comment({ id: 'c-1' })] } }))],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByLabelText('反応を付ける')).toBeInTheDocument();
    });
    await userEvent.click(canvas.getByLabelText('反応を付ける'));
    await expect(canvas.getByRole('group', { name: '反応を選ぶ' })).toBeInTheDocument();
    await expect(canvas.getByLabelText('🚀 の反応を付ける')).toBeInTheDocument();
  },
};

export const 編集済みの履歴を開く: Story = {
  decorators: [
    // withApi は部分一致なので、より具体的な /edits を /comments より先に置く
    // （baseApi のマージだと /comments が先に定義されていて先に当たってしまう）。
    withApi({
      '/kb/workspaces/acme/tickets/t-1/comments/c-1/edits': {
        edits: [{ id: 'e-1', editor: { userId: 1, name: '田中 太郎' }, previousBody: [{ type: 'text', text: '直す前' }], editedAt: '2026-09-09T00:00:00Z' }],
      },
      ...baseApi({ '/kb/workspaces/acme/tickets/t-1/comments': { comments: [comment({ id: 'c-1', edited: true })] } }),
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByRole('button', { name: '田中 太郎 の編集履歴を開く' })).toBeInTheDocument();
    });
    await userEvent.click(canvas.getByRole('button', { name: '田中 太郎 の編集履歴を開く' }));
    await waitFor(async () => {
      await expect(canvas.getByText('直す前')).toBeInTheDocument();
    });
  },
};

export const 空白だけの返信は送信できない: Story = {
  decorators: [withApi(baseApi({ '/kb/workspaces/acme/tickets/t-1/comments': { comments: [comment({ id: 'c-1' })] } }))],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByRole('button', { name: '返信' })).toBeInTheDocument();
    });
    await userEvent.click(canvas.getByRole('button', { name: '返信' }));
    const replyBox = await canvas.findByRole('textbox', { name: '返信を書く' });
    await userEvent.type(replyBox, '   ');
    // 「返信」という名前のボタンが開閉トグルと送信の 2 つあるので、送信ボタン
    // （入力欄と同じコンポーザの中）に絞って確かめる。
    const composer = within(replyBox.closest('.rounded-xl') as HTMLElement);
    await expect(composer.getByRole('button', { name: '返信' })).toBeDisabled();
  },
};

export const 名指しの候補から選んで送信する: Story = {
  decorators: [
    withApi(
      baseApi({
        '/kb/workspaces/acme/tickets/t-1/comments': (config: { method?: string; data?: unknown }) => {
          if (config.method !== 'post') return { comments: [] };
          const sent = typeof config.data === 'string' ? JSON.parse(config.data) : config.data;
          return comment({ id: 'c-new', author: { userId: 1, name: 'norman6464' }, body: (sent as { body: unknown }).body });
        },
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const box = await canvas.findByRole('textbox', { name: 'コメントを書く' });
    await userEvent.type(box, '@佐藤');
    // 候補の一覧は floating-ui でキャンバスの外（document 直下）へ描かれるため、
    // canvasElement には閉じない screen で探す（slashCommand.integration.test.tsx と同じ理由）。
    await waitFor(async () => {
      await expect(screen.getByRole('option', { name: /佐藤 花子/ })).toBeInTheDocument();
    });
    await userEvent.click(screen.getByRole('option', { name: /佐藤 花子/ }));
    await userEvent.click(canvas.getByRole('button', { name: '送信' }));
    await waitFor(async () => {
      await expect(canvas.getByText('@佐藤 花子')).toBeInTheDocument();
    });
  },
};

/**
 * Enter は「選んでいる候補を確定する」（改行にならない）。CommentComposerEnter（改行を
 * hardBreak に固定する拡張）が候補一覧より先に Enter を食ってしまう回帰を防ぐための story
 * （実際に一度この壊れ方をした — mentionExtension.ts の doc 参照）。
 */
export const 候補が出ているときのEnterは改行にならず確定する: Story = {
  decorators: [withApi(baseApi())],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const box = await canvas.findByRole('textbox', { name: 'コメントを書く' });
    await userEvent.type(box, '@');
    await waitFor(async () => {
      await expect(screen.getByRole('option', { name: /norman6464/ })).toBeInTheDocument();
    });
    await userEvent.keyboard('{Enter}');
    await waitFor(async () => {
      await expect(canvas.getByText('@norman6464')).toBeInTheDocument();
    });
    // 改行（hardBreak）を挟んで二重に確定していない — 名指しは 1 個だけ。
    await expect(canvas.getAllByText('@norman6464')).toHaveLength(1);
  },
};

export const 反応12種の格子が全部見える: Story = {
  decorators: [withApi(baseApi({ '/kb/workspaces/acme/tickets/t-1/comments': { comments: [comment({ id: 'c-1' })] } }))],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByLabelText('反応を付ける')).toBeInTheDocument();
    });
    await userEvent.click(canvas.getByLabelText('反応を付ける'));
    for (const emoji of ['👍', '🙏', '🎉', '👀', '✅', '❤️', '😄', '🤔', '🚀', '🔥', '⚠️', '😢']) {
      await expect(canvas.getByLabelText(`${emoji} の反応を付ける`)).toBeInTheDocument();
    }
  },
};
