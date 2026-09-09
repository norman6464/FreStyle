import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, within } from 'storybook/test';
import BacklogRow from './BacklogRow';
import type { Ticket, TicketStatus, TicketType } from '@/entities/ticket';

const baseTicket: Ticket = {
  id: 't-1',
  workspaceId: 'w-1',
  spaceId: 's-1',
  number: 457,
  typeId: 'ty-1',
  statusId: 'st-2',
  parentId: null,
  title: '段1: チケットの骨格（9表）',
  doc: { type: 'doc', content: [] },
  priority: 1,
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
};

const devType: TicketType = {
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
  activeTicketCount: 3,
};

const devStatus: TicketStatus = {
  id: 'st-2',
  workspaceId: 'w-1',
  spaceId: 's-1',
  name: '開発',
  category: 'in_progress',
  color: '#a0661a',
  position: 'a0',
  isInitial: false,
  archivedAt: null,
  createdAt: '2026-09-08T00:00:00Z',
  updatedAt: '2026-09-08T00:00:00Z',
  activeTicketCount: 2,
};

const doneStatus: TicketStatus = { ...devStatus, id: 'st-5', name: 'リリース', category: 'done', color: '#2f6b47' };

const meta = {
  title: 'pages/backlog/BacklogRow',
  component: BacklogRow,
  parameters: { layout: 'padded' },
  args: {
    ticket: baseTicket,
    spaceKey: 'FRESTYLE',
    type: devType,
    status: devStatus,
    assigneeName: '',
    assigneeInitials: '',
    selected: false,
    busy: false,
    indented: false,
    canEdit: true,
    onOpen: fn(),
  },
  decorators: [
    (Story) => (
      <div className="w-[560px] border border-surface-3">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof BacklogRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const 担当あり: Story = {
  args: {
    ticket: { ...baseTicket, assigneePrincipalId: 'p-nor' },
    assigneeName: 'norman6464',
    assigneeInitials: 'NO',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByText('FRESTYLE-457')).toBeInTheDocument();
    await expect(canvas.getByText('NO')).toBeInTheDocument();
  },
};

export const 未割り当て: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByLabelText('未割り当て')).toBeInTheDocument();
  },
};

export const 子チケット_字下げ: Story = {
  args: { ticket: { ...baseTicket, parentId: 't-0' }, indented: true },
};

export const 期限切れ: Story = {
  args: { ticket: { ...baseTicket, dueDate: '2020-01-01' } },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText('01/01')).toHaveClass('text-red-600');
  },
};

export const 完了枠_打ち消し線: Story = {
  args: { status: doneStatus },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(baseTicket.title)).toHaveClass('line-through');
  },
};

export const 操作中: Story = {
  args: { busy: true },
};

export const 読むだけ_chevronを出さない: Story = {
  args: { canEdit: false },
};
