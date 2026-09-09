import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
import TicketDetailPanel from './TicketDetailPanel';
import type { Ticket, TicketChangeGroup, TicketStatus, TicketType } from '@/entities/ticket';
import type { KbGrantablePrincipal } from '@/entities/kb';

const ticket: Ticket = {
  id: 't-1',
  workspaceId: 'w-1',
  spaceId: 's-1',
  number: 457,
  typeId: 'ty-1',
  statusId: 'st-2',
  parentId: null,
  title: '段1: チケットの骨格（9表）',
  doc: { type: 'doc', content: [{ type: 'paragraph', content: [{ type: 'text', text: '本文です。' }] }] },
  priority: 1,
  startDate: null,
  dueDate: '2026-09-12',
  position: 'a0',
  closedAt: null,
  resolution: null,
  createdByUserId: 1,
  archivedAt: null,
  createdAt: '2026-09-08T00:00:00Z',
  updatedAt: '2026-09-09T00:00:00Z',
  assigneePrincipalId: 'p-1',
};

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
    createdAt: '',
    updatedAt: '',
    activeTicketCount: 0,
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
    createdAt: '',
    updatedAt: '',
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
    createdAt: '',
    updatedAt: '',
    activeTicketCount: 1,
  },
];

const principals: KbGrantablePrincipal[] = [{ id: 'p-1', kind: 'user', name: 'norman6464' }];

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
  title: 'pages/backlog/TicketDetailPanel',
  component: TicketDetailPanel,
  parameters: { layout: 'fullscreen' },
  args: {
    ticket,
    spaceKey: 'FRESTYLE',
    statuses,
    types,
    principals,
    parentTicket: undefined,
    history,
    historyLoading: false,
    canEdit: true,
    busy: false,
    onUpdate: fn(async (_id, _input) => ticket),
    onChangeStatus: fn(async () => {}),
    onAssign: fn(async () => {}),
    onUnassign: fn(async () => {}),
    onArchive: fn(async () => {}),
    onRestore: fn(async () => {}),
    onClose: fn(),
  },
  decorators: [
    (Story) => (
      <div className="h-[640px] w-[360px] border-l border-surface-3 bg-surface-1">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof TicketDetailPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 編集できる: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('FRESTYLE-457')).toBeInTheDocument();
    await expect(canvas.getByLabelText('題名')).toHaveValue(ticket.title);
  },
};

export const 読むだけ: Story = {
  args: { canEdit: false },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.queryByLabelText('状態')).toBeNull();
    await expect(canvas.getByText(ticket.title)).toBeInTheDocument();
  },
};

export const 履歴あり: Story = {
  play: async ({ canvasElement }) => {
    // 「開発」は状態セレクトの option にも現れるため、履歴 1 件の行の textContent で確かめる。
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

export const 状態変更が失敗しても表示は元のまま: Story = {
  args: { onChangeStatus: fn(async () => Promise.reject(new Error('409'))) },
  play: async ({ args, canvasElement }) => {
    const canvas = within(canvasElement);
    const select = canvas.getByLabelText('状態') as HTMLSelectElement;
    await userEvent.selectOptions(select, 'st-1');
    await waitFor(() => expect(args.onChangeStatus).toHaveBeenCalledWith('st-1'));
    // ticket prop 自体は変わっていないので、表示は選択中チケットの statusId のまま。
    await expect(select).toHaveValue('st-2');
  },
};
