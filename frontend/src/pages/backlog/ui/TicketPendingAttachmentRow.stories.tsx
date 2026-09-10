import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';
import TicketPendingAttachmentRow from './TicketPendingAttachmentRow';
import type { PendingAttachment } from '../model/useTicketAttachments';

function pending(over: Partial<PendingAttachment> = {}): PendingAttachment {
  return {
    clientId: 'p-1',
    file: new File(['x'], '設計資料.pdf', { type: 'application/pdf' }),
    status: 'uploading',
    error: null,
    ...over,
  };
}

const meta = {
  title: 'pages/backlog/TicketPendingAttachmentRow',
  component: TicketPendingAttachmentRow,
  args: { pending: pending(), onRetry: fn(), onDismiss: fn() },
  decorators: [(Story) => <ul className="w-80 bg-surface-1 p-3"><Story /></ul>],
} satisfies Meta<typeof TicketPendingAttachmentRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const アップロード中: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('設計資料.pdf')).toBeInTheDocument();
    await expect(canvas.getByRole('status')).toHaveTextContent('アップロード中');
    await expect(canvas.queryByRole('button')).toBeNull();
  },
};

export const 失敗: Story = {
  args: { pending: pending({ status: 'failed', error: 'アップロードに失敗しました。' }) },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole('alert')).toHaveTextContent('失敗しました');
    await expect(canvas.getByLabelText('設計資料.pdf のアップロードをやり直す')).toBeInTheDocument();
    await expect(canvas.getByLabelText('設計資料.pdf を取り消す')).toBeInTheDocument();
  },
};

export const やり直す: Story = {
  args: { pending: pending({ status: 'failed', error: 'アップロードに失敗しました。' }) },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByLabelText('設計資料.pdf のアップロードをやり直す'));
    await expect(args.onRetry).toHaveBeenCalled();
  },
};

export const 取り消す: Story = {
  args: { pending: pending({ status: 'failed', error: 'アップロードに失敗しました。' }) },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    await userEvent.click(canvas.getByLabelText('設計資料.pdf を取り消す'));
    await expect(args.onDismiss).toHaveBeenCalled();
  },
};
