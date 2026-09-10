import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import TicketAttachmentRow from './TicketAttachmentRow';
import { withApi } from '../../../../.storybook/decorators';
import type { TicketAttachment } from '@/entities/ticket';

const attachment: TicketAttachment = {
  id: 'at-1',
  ticketId: 't-1',
  filename: '設計資料.pdf',
  contentType: 'application/pdf',
  sizeBytes: 234_567,
  uploadedByUserId: 1,
  createdAt: '2026-09-10T00:00:00Z',
};

const meta = {
  title: 'pages/backlog/TicketAttachmentRow',
  component: TicketAttachmentRow,
  args: {
    workspaceSlug: 'acme',
    ticketId: 't-1',
    attachment,
    canEdit: true,
    busy: false,
    onRemove: fn(),
  },
  decorators: [(Story) => <ul className="w-80 bg-surface-1 p-3"><Story /></ul>],
} satisfies Meta<typeof TicketAttachmentRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 編集できる: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('設計資料.pdf')).toBeInTheDocument();
    await expect(canvas.getByText('229.1 KB')).toBeInTheDocument();
    await expect(canvas.getByLabelText('設計資料.pdf を削除')).toBeInTheDocument();
  },
};

export const 読むだけでは削除ボタンを出さない: Story = {
  args: { canEdit: false },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.queryByLabelText('設計資料.pdf を削除')).toBeNull();
  },
};

export const 削除中: Story = {
  args: { busy: true },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByLabelText('設計資料.pdf を削除')).toBeDisabled();
  },
};

export const 画像はアイコンが変わる: Story = {
  args: { attachment: { ...attachment, filename: '完成図.png', contentType: 'image/png' } },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('完成図.png')).toBeInTheDocument();
  },
};

export const ダウンロードURLを発行して開く: Story = {
  decorators: [withApi({ '/attachments/at-1/download-url': { url: 'https://storage.example/signed', expiresIn: 600 } })],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const originalOpen = window.open;
    const opened: string[] = [];
    window.open = ((url?: string | URL) => {
      opened.push(String(url));
      return null;
    }) as typeof window.open;
    try {
      await userEvent.click(canvas.getByText('設計資料.pdf'));
      await waitFor(() => expect(opened).toEqual(['https://storage.example/signed']));
    } finally {
      window.open = originalOpen;
    }
  },
};

export const ダウンロードURLの発行に失敗: Story = {
  decorators: [
    withApi({
      '/attachments/at-1/download-url': () => {
        const err = new Error('network') as Error & { isAxiosError: boolean; response: unknown };
        err.isAxiosError = true;
        err.response = { status: 500, data: {} };
        throw err;
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByText('設計資料.pdf'));
    await waitFor(async () => {
      await expect(canvas.getByRole('alert')).toHaveTextContent('取得できませんでした');
    });
  },
};
