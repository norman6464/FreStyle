import { describe, it, expect, vi, beforeEach } from 'vitest';
import TicketRepository from '../ticketRepository';
import apiClient from '@/shared/api/axios';

vi.mock('@/shared/api/axios');

const mockGet = vi.mocked(apiClient.get);
const mockPost = vi.mocked(apiClient.post);
const mockPut = vi.mocked(apiClient.put);
const mockDelete = vi.mocked(apiClient.delete);

beforeEach(() => {
  vi.clearAllMocks();
});

const wireTicket = (over: Record<string, unknown> = {}) => ({
  id: 't-1',
  workspaceId: 'w-1',
  spaceId: 's-1',
  number: 457,
  typeId: 'ty-1',
  statusId: 'st-1',
  title: '段1: チケットの骨格',
  doc: { type: 'doc', content: [] },
  priority: 2,
  position: 'a0',
  createdByUserId: 1,
  createdAt: '2026-09-08T00:00:00Z',
  updatedAt: '2026-09-09T00:00:00Z',
  ...over,
});

describe('TicketRepository.fetchTickets', () => {
  it('GET /kb/workspaces/:slug/spaces/:spaceId/tickets を叩き、omitempty のキーを null に正規化する', async () => {
    mockGet.mockResolvedValue({ data: { tickets: [wireTicket()] } });
    const list = await TicketRepository.fetchTickets('acme', 's-1');
    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s-1/tickets', { params: {} });
    expect(list).toHaveLength(1);
    expect(list[0]).toMatchObject({
      id: 't-1',
      parentId: null,
      startDate: null,
      dueDate: null,
      closedAt: null,
      resolution: null,
      archivedAt: null,
      assigneePrincipalId: null,
    });
  });

  it('tickets が null で返っても空配列にする', async () => {
    mockGet.mockResolvedValue({ data: { tickets: null } });
    await expect(TicketRepository.fetchTickets('acme', 's-1')).resolves.toEqual([]);
  });

  it('絞り込みをクエリパラメータへ渡す', async () => {
    mockGet.mockResolvedValue({ data: { tickets: [] } });
    await TicketRepository.fetchTickets('acme', 's-1', {
      statusId: 'st-1',
      typeId: 'ty-1',
      assigneePrincipalId: 'p-1',
      archived: true,
    });
    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s-1/tickets', {
      params: { statusId: 'st-1', typeId: 'ty-1', assigneePrincipalId: 'p-1', archived: 'true' },
    });
  });

  it('値がある omitempty フィールドはそのまま通す', async () => {
    mockGet.mockResolvedValue({
      data: { tickets: [wireTicket({ parentId: 't-0', assigneePrincipalId: 'p-1' })] },
    });
    const [ticket] = await TicketRepository.fetchTickets('acme', 's-1');
    expect(ticket.parentId).toBe('t-0');
    expect(ticket.assigneePrincipalId).toBe('p-1');
  });
});

describe('TicketRepository.createTicket', () => {
  it('POST で作成し、省略項目は空文字/0 で送る', async () => {
    mockPost.mockResolvedValue({ data: wireTicket() });
    await TicketRepository.createTicket('acme', 's-1', { title: '新しいチケット' });
    expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s-1/tickets', {
      parentId: '',
      typeId: '',
      statusId: '',
      title: '新しいチケット',
      doc: undefined,
      priority: 0,
      startDate: undefined,
      dueDate: undefined,
    });
  });
});

describe('TicketRepository.moveTicket', () => {
  it('POST /move へ anchor を送る（204・戻り値なし）', async () => {
    mockPost.mockResolvedValue({ data: undefined });
    await TicketRepository.moveTicket('acme', 't-1', { anchorTicketId: 't-2', anchorAfter: true });
    expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/move', {
      anchorTicketId: 't-2',
      anchorAfter: true,
    });
  });

  it('anchor 省略時は末尾へ（空文字を送る）', async () => {
    mockPost.mockResolvedValue({ data: undefined });
    await TicketRepository.moveTicket('acme', 't-1', {});
    expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/move', {
      anchorTicketId: '',
      anchorAfter: false,
    });
  });
});

describe('TicketRepository.unassignTicket / archiveTicketStatus', () => {
  it('DELETE /assignee を叩く', async () => {
    mockDelete.mockResolvedValue({ data: undefined });
    await TicketRepository.unassignTicket('acme', 't-1');
    expect(mockDelete).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/assignee');
  });

  it('POST /ticket-statuses/:id/archive を叩く', async () => {
    mockPost.mockResolvedValue({ data: undefined });
    await TicketRepository.archiveTicketStatus('acme', 's-1', 'st-1');
    expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s-1/ticket-statuses/st-1/archive');
  });
});

describe('TicketRepository.fetchTicketStatuses', () => {
  it('activeTicketCount を含めて正規化する', async () => {
    mockGet.mockResolvedValue({
      data: {
        statuses: [
          {
            id: 'st-1',
            workspaceId: 'w-1',
            spaceId: 's-1',
            name: 'To Do',
            category: 'todo',
            color: '#5b6b7a',
            position: 'a0',
            isInitial: true,
            createdAt: '2026-09-08T00:00:00Z',
            updatedAt: '2026-09-08T00:00:00Z',
            activeTicketCount: 3,
          },
        ],
      },
    });
    const [status] = await TicketRepository.fetchTicketStatuses('acme', 's-1');
    expect(status.activeTicketCount).toBe(3);
    expect(status.archivedAt).toBeNull();
  });
});

describe('TicketRepository.updateTicketStatus', () => {
  it('応答に activeTicketCount が無くても 0 に正規化する（backend は domain 構造体を素で返す）', async () => {
    mockPut.mockResolvedValue({
      data: {
        id: 'st-1',
        workspaceId: 'w-1',
        spaceId: 's-1',
        name: 'To Do',
        category: 'todo',
        color: '#5b6b7a',
        position: 'a0',
        isInitial: true,
        createdAt: '2026-09-08T00:00:00Z',
        updatedAt: '2026-09-09T00:00:00Z',
      },
    });
    const status = await TicketRepository.updateTicketStatus('acme', 's-1', 'st-1', {
      name: 'To Do',
      category: 'todo',
      color: '#5b6b7a',
    });
    expect(status.activeTicketCount).toBe(0);
  });
});
