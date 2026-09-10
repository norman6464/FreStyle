import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import TicketAttachmentSection from './TicketAttachmentSection';
import { withApi, withRawPut, withToast, type ApiStubs } from '../../../../.storybook/decorators';
import type { TicketAttachment } from '@/entities/ticket';

function attachment(over: Partial<TicketAttachment> & { id: string }): TicketAttachment {
  return {
    ticketId: 't-1',
    filename: '設計資料.pdf',
    contentType: 'application/pdf',
    sizeBytes: 234_567,
    uploadedByUserId: 1,
    createdAt: '2026-09-10T00:00:00Z',
    ...over,
  };
}

/**
 * 突き合わせは URL に含まれる文字列の**先勝ち**（.storybook/decorators.tsx 参照）なので、
 * `/attachments/upload-url` のような細かい宛先を、一覧・確定が共有する `/attachments` より
 * 先に書く。
 */
function baseApi(attachments: TicketAttachment[] = [], over: ApiStubs = {}): ApiStubs {
  return {
    '/attachments/upload-url': { url: 'https://storage.example/put?sig', key: 'k-1', expiresIn: 600 },
    '/attachments/at-1/download-url': { url: 'https://storage.example/get?sig', expiresIn: 600 },
    '/attachments/at-1': undefined,
    '/attachments': (config: { method?: string }) =>
      config.method === 'post'
        ? attachment({ id: 'at-new', filename: '設計資料.pdf', contentType: 'application/pdf', sizeBytes: 5 })
        : { attachments },
    ...over,
  };
}

const meta = {
  title: 'pages/backlog/TicketAttachmentSection',
  component: TicketAttachmentSection,
  args: { workspaceSlug: 'acme', ticketId: 't-1', canEdit: true },
  decorators: [withToast, (Story) => <div className="w-80 bg-surface-1 p-3"><Story /></div>],
} satisfies Meta<typeof TicketAttachmentSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 添付なし: Story = {
  decorators: [withApi(baseApi([]))],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('添付はありません')).toBeInTheDocument();
    });
    await expect(canvas.getByRole('button', { name: /ファイルを添付/ })).toBeInTheDocument();
  },
};

export const 添付あり: Story = {
  decorators: [withApi(baseApi([attachment({ id: 'at-1' })]))],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('設計資料.pdf')).toBeInTheDocument();
    });
  },
};

export const 読むだけでは追加ボタンを出さない: Story = {
  args: { canEdit: false },
  decorators: [withApi(baseApi([attachment({ id: 'at-1' })]))],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('設計資料.pdf')).toBeInTheDocument();
    });
    await expect(canvas.queryByRole('button', { name: /ファイルを添付/ })).toBeNull();
    await expect(canvas.queryByLabelText('設計資料.pdf を削除')).toBeNull();
  },
};

export const 取得に失敗: Story = {
  decorators: [
    withApi({
      '/attachments': () => {
        const err = new Error('network') as Error & { isAxiosError: boolean; response: unknown };
        err.isAxiosError = true;
        err.response = { status: 500, data: {} };
        throw err;
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    await waitFor(async () => {
      await expect(within(canvasElement).getByRole('alert')).toHaveTextContent('読み込めませんでした');
    });
  },
};

export const アップロードが成功する: Story = {
  decorators: [withApi(baseApi([])), withRawPut()],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('添付はありません')).toBeInTheDocument();
    });
    const file = new File(['x'], '設計資料.pdf', { type: 'application/pdf' });
    await userEvent.upload(canvas.getByLabelText('添付ファイルを選ぶ'), file);
    await waitFor(async () => {
      await expect(canvas.getAllByText('設計資料.pdf')).toHaveLength(1);
    });
    await expect(canvas.queryByText('添付はありません')).toBeNull();
  },
};

export const アップロードに失敗すると行が残る: Story = {
  decorators: [
    withApi(
      baseApi([], {
        '/attachments/upload-url': () => {
          const err = new Error('network') as Error & { isAxiosError: boolean; response: unknown };
          err.isAxiosError = true;
          err.response = { status: 500, data: {} };
          throw err;
        },
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('添付はありません')).toBeInTheDocument();
    });
    const file = new File(['x'], '設計資料.pdf', { type: 'application/pdf' });
    await userEvent.upload(canvas.getByLabelText('添付ファイルを選ぶ'), file);
    await waitFor(async () => {
      await expect(canvas.getByRole('alert')).toHaveTextContent('アップロードに失敗しました');
    });
    await expect(canvas.getByLabelText('設計資料.pdf のアップロードをやり直す')).toBeInTheDocument();
  },
};

export const 削除する: Story = {
  decorators: [withApi(baseApi([attachment({ id: 'at-1' })]))],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('設計資料.pdf')).toBeInTheDocument();
    });
    await userEvent.click(canvas.getByLabelText('設計資料.pdf を削除'));
    await waitFor(async () => {
      await expect(canvas.getByText('添付はありません')).toBeInTheDocument();
    });
  },
};

export const 削除に失敗するとトーストが出る: Story = {
  decorators: [
    withApi(
      baseApi([attachment({ id: 'at-1' })], {
        '/attachments/at-1': () => {
          const err = new Error('network') as Error & { isAxiosError: boolean; response: unknown };
          err.isAxiosError = true;
          err.response = { status: 500, data: {} };
          throw err;
        },
      }),
    ),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await waitFor(async () => {
      await expect(canvas.getByText('設計資料.pdf')).toBeInTheDocument();
    });
    await userEvent.click(canvas.getByLabelText('設計資料.pdf を削除'));
    // トースト自体は withToast 経由の別コンテナに出るため、ここでは行が残ることだけ確かめる。
    await waitFor(async () => {
      await expect(canvas.getByText('設計資料.pdf')).toBeInTheDocument();
    });
  },
};
