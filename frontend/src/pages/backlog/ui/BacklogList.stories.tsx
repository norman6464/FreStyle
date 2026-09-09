import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, within } from 'storybook/test';
import BacklogList from './BacklogList';
import type { Ticket, TicketStatus, TicketType } from '@/entities/ticket';

const statuses: TicketStatus[] = [
  {
    id: 'st-1',
    workspaceId: 'w-1',
    spaceId: 's-1',
    name: 'To Do',
    category: 'todo',
    color: '#5b6b7a',
    position: 'a0',
    isInitial: true,
    archivedAt: null,
    createdAt: '2026-09-08T00:00:00Z',
    updatedAt: '2026-09-08T00:00:00Z',
    activeTicketCount: 1,
  },
  {
    id: 'st-2',
    workspaceId: 'w-1',
    spaceId: 's-1',
    name: '開発',
    category: 'in_progress',
    color: '#a0661a',
    position: 'a1',
    isInitial: false,
    archivedAt: null,
    createdAt: '2026-09-08T00:00:00Z',
    updatedAt: '2026-09-08T00:00:00Z',
    activeTicketCount: 1,
  },
];

const types: TicketType[] = [
  {
    id: 'ty-1',
    workspaceId: 'w-1',
    spaceId: 's-1',
    name: '開発タスク',
    hierarchyLevel: 0,
    color: '#2563eb',
    position: 'a0',
    isDefault: true,
    templateTitle: null,
    templateDoc: null,
    archivedAt: null,
    createdAt: '2026-09-08T00:00:00Z',
    updatedAt: '2026-09-08T00:00:00Z',
    activeTicketCount: 2,
  },
];

function ticket(over: Partial<Ticket>): Ticket {
  return {
    id: 't-1',
    workspaceId: 'w-1',
    spaceId: 's-1',
    number: 457,
    typeId: 'ty-1',
    statusId: 'st-2',
    parentId: null,
    title: '段1: チケットの骨格',
    doc: { type: 'doc', content: [] },
    priority: 2,
    startDate: null,
    dueDate: null,
    position: 'a0',
    closedAt: null,
    resolution: null,
    createdByUserId: 1,
    archivedAt: null,
    createdAt: '2026-09-08T00:00:00Z',
    updatedAt: '2026-09-09T00:00:00Z',
    assigneePrincipalId: null,
    ...over,
  };
}

const tickets: Ticket[] = [
  ticket({ id: 't-1', number: 457 }),
  ticket({ id: 't-2', number: 458, statusId: 'st-1', parentId: 't-1' }),
];

const meta = {
  title: 'pages/backlog/BacklogList',
  component: BacklogList,
  parameters: { layout: 'fullscreen' },
  args: {
    tickets,
    statuses,
    types,
    spaceKey: 'FRESTYLE',
    loading: false,
    error: null,
    archived: false,
    canEdit: true,
    selectedId: null,
    busyId: null,
    nameOf: () => '',
    initialsOf: () => '',
    onSelect: fn(),
    onCreate: fn(async () => {}),
    onMove: fn(async () => {}),
    onRetry: fn(),
  },
  decorators: [
    (Story) => (
      <div className="h-[520px] w-[640px] bg-surface-1">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof BacklogList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ふつう: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('FRESTYLE-457')).toBeInTheDocument();
    await expect(canvas.getByText('FRESTYLE-458')).toBeInTheDocument();
  },
};

export const 空: Story = {
  args: { tickets: [] },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('まだチケットがありません')).toBeInTheDocument();
  },
};

export const 読み取りだけ: Story = {
  args: { canEdit: false },
  play: async ({ canvasElement }) => {
    // 作成行・並び替えの帯が出ない。
    await expect(within(canvasElement).queryByLabelText('新しいチケットの題名')).toBeNull();
  },
};

export const アーカイブ: Story = {
  args: { archived: true, tickets: [ticket({ id: 't-3', number: 402, archivedAt: '2026-09-01T00:00:00Z' })] },
};

export const 絞り込みで0件: Story = {
  args: { tickets: [] },
};

export const 読み込み失敗: Story = {
  args: { error: 'チケットを読み込めませんでした。' },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByRole('button', { name: '再読み込み' })).toBeInTheDocument();
  },
};
