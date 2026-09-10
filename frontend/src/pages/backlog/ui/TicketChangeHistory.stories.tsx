import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';
import TicketChangeHistory from './TicketChangeHistory';
import type { TicketChangeGroup } from '@/entities/ticket';

const history: TicketChangeGroup[] = [
  {
    id: 'g-1',
    workspaceId: 'w-1',
    ticketId: 't-1',
    actorUserId: 1,
    createdAt: '2026-09-09T15:20:00Z',
    items: [{ id: 'i-1', groupId: 'g-1', field: 'status', oldValue: 'st-1', newValue: 'st-2', oldLabel: 'To Do', newLabel: '開発' }],
  },
];

const meta = {
  title: 'pages/backlog/TicketChangeHistory',
  component: TicketChangeHistory,
  args: { history, loading: false, error: null },
  decorators: [(Story) => <div className="w-72 bg-surface-1 p-3"><Story /></div>],
} satisfies Meta<typeof TicketChangeHistory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 履歴あり: Story = {
  play: async ({ canvasElement }) => {
    // 「開発」は他の場面で状態セレクトの option にも現れうるため、行の textContent で確かめる。
    const item = within(canvasElement).getByRole('listitem');
    await expect(item).toHaveTextContent('状態を');
    await expect(item).toHaveTextContent('開発');
  },
};

export const 履歴なし: Story = {
  args: { history: [] },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('まだ変更はありません')).toBeInTheDocument();
  },
};

export const 読込中: Story = {
  args: { loading: true },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('status')).toBeInTheDocument();
  },
};

export const 失敗: Story = {
  args: { history: [], loading: false, error: '変更履歴を読み込めませんでした。' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('alert')).toHaveTextContent('読み込めませんでした');
  },
};
