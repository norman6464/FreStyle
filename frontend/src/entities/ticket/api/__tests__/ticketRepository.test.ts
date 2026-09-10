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

describe('TicketRepository.fetchTicketComments', () => {
  it('GET /comments を叩き、本文をインラインノードの配列から区間の列へ畳む', async () => {
    mockGet.mockResolvedValue({
      data: {
        comments: [
          {
            id: 'c-1',
            author: { userId: 1, name: '田中 太郎' },
            body: [{ type: 'text', text: 'こんにちは' }],
            edited: false,
            reactions: [{ userId: 2, emoji: '👍' }],
            createdAt: '2026-09-10T00:00:00Z',
            updatedAt: '2026-09-10T00:00:00Z',
          },
        ],
      },
    });

    const comments = await TicketRepository.fetchTicketComments('acme', 't-1');

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/comments');
    expect(comments).toEqual([
      {
        id: 'c-1',
        parentCommentId: null,
        author: { userId: 1, name: '田中 太郎' },
        body: [{ kind: 'text', text: 'こんにちは' }],
        edited: false,
        reactions: [{ userId: 2, emoji: '👍' }],
        createdAt: '2026-09-10T00:00:00Z',
        updatedAt: '2026-09-10T00:00:00Z',
      },
    ]);
  });

  it('reactions が無ければ空配列にする', async () => {
    mockGet.mockResolvedValue({
      data: {
        comments: [
          {
            id: 'c-1',
            author: { userId: 1, name: '田中 太郎' },
            body: [],
            edited: false,
            createdAt: '',
            updatedAt: '',
          },
        ],
      },
    });
    const [comment] = await TicketRepository.fetchTicketComments('acme', 't-1');
    expect(comment.reactions).toEqual([]);
  });

  it('comments が null でも空配列にする', async () => {
    mockGet.mockResolvedValue({ data: { comments: null } });
    await expect(TicketRepository.fetchTicketComments('acme', 't-1')).resolves.toEqual([]);
  });
});

describe('TicketRepository.createTicketComment', () => {
  it('POST で区間の列を送信できる本文へ組み立てて送る', async () => {
    mockPost.mockResolvedValue({
      data: {
        id: 'c-1',
        author: { userId: 1, name: '田中 太郎' },
        body: [{ type: 'text', text: 'お願いします' }],
        edited: false,
        reactions: [],
        createdAt: '',
        updatedAt: '',
      },
    });

    await TicketRepository.createTicketComment('acme', 't-1', [{ kind: 'text', text: 'お願いします' }], 'c-parent');

    expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/comments', {
      parentCommentId: 'c-parent',
      body: [{ type: 'text', text: 'お願いします' }],
    });
  });
});

describe('TicketRepository.updateTicketComment', () => {
  it('PUT で本文を置き換える（応答の reactions は常に空配列で返る）', async () => {
    mockPut.mockResolvedValue({
      data: {
        id: 'c-1',
        author: { userId: 1, name: '田中 太郎' },
        body: [{ type: 'text', text: '直しました' }],
        edited: true,
        reactions: [],
        createdAt: '',
        updatedAt: '',
      },
    });

    const updated = await TicketRepository.updateTicketComment('acme', 't-1', 'c-1', [
      { kind: 'text', text: '直しました' },
    ]);

    expect(mockPut).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/comments/c-1', {
      body: [{ type: 'text', text: '直しました' }],
    });
    expect(updated.reactions).toEqual([]);
  });
});

describe('TicketRepository.deleteTicketComment', () => {
  it('DELETE を叩く（204）', async () => {
    mockDelete.mockResolvedValue({ data: undefined });
    await TicketRepository.deleteTicketComment('acme', 't-1', 'c-1');
    expect(mockDelete).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/comments/c-1');
  });
});

describe('TicketRepository.fetchTicketCommentEdits', () => {
  it('GET /edits を叩き、編集前の本文も区間の列へ畳む', async () => {
    mockGet.mockResolvedValue({
      data: {
        edits: [
          {
            id: 'e-1',
            editor: { userId: 1, name: '田中 太郎' },
            previousBody: [{ type: 'text', text: '直す前' }],
            editedAt: '2026-09-10T00:00:00Z',
          },
        ],
      },
    });

    const edits = await TicketRepository.fetchTicketCommentEdits('acme', 't-1', 'c-1');

    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/comments/c-1/edits');
    expect(edits).toEqual([
      {
        id: 'e-1',
        editor: { userId: 1, name: '田中 太郎' },
        previousBody: [{ kind: 'text', text: '直す前' }],
        editedAt: '2026-09-10T00:00:00Z',
      },
    ]);
  });
});

describe('TicketRepository.addTicketCommentReaction / removeTicketCommentReaction', () => {
  it('絵文字を URL エンコードして PUT/DELETE する（204・冪等）', async () => {
    mockPut.mockResolvedValue({ data: undefined });
    mockDelete.mockResolvedValue({ data: undefined });

    await TicketRepository.addTicketCommentReaction('acme', 't-1', 'c-1', '👍');
    await TicketRepository.removeTicketCommentReaction('acme', 't-1', 'c-1', '👍');

    const expectedUrl = `/api/v2/kb/workspaces/acme/tickets/t-1/comments/c-1/reactions/${encodeURIComponent('👍')}`;
    expect(mockPut).toHaveBeenCalledWith(expectedUrl);
    expect(mockDelete).toHaveBeenCalledWith(expectedUrl);
  });
});

describe('TicketRepository.fetchLabels', () => {
  it('GET /labels を叩く', async () => {
    mockGet.mockResolvedValue({
      data: { labels: [{ id: 'l-1', spaceId: 's-1', name: '不具合', color: '#1d4ed8', createdAt: '', updatedAt: '' }] },
    });
    const labels = await TicketRepository.fetchLabels('acme', 's-1');
    expect(mockGet).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s-1/labels');
    expect(labels).toHaveLength(1);
  });

  it('labels が null でも空配列にする', async () => {
    mockGet.mockResolvedValue({ data: { labels: null } });
    await expect(TicketRepository.fetchLabels('acme', 's-1')).resolves.toEqual([]);
  });
});

describe('TicketRepository.createLabel / updateLabel', () => {
  it('POST で作成する', async () => {
    mockPost.mockResolvedValue({ data: { id: 'l-1', spaceId: 's-1', name: '検索', color: '#dbeafe', createdAt: '', updatedAt: '' } });
    await TicketRepository.createLabel('acme', 's-1', { name: '検索', color: '#dbeafe' });
    expect(mockPost).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s-1/labels', { name: '検索', color: '#dbeafe' });
  });

  it('PUT で更新する', async () => {
    mockPut.mockResolvedValue({ data: { id: 'l-1', spaceId: 's-1', name: '検索2', color: '#dbeafe', createdAt: '', updatedAt: '' } });
    await TicketRepository.updateLabel('acme', 's-1', 'l-1', { name: '検索2', color: '#dbeafe' });
    expect(mockPut).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s-1/labels/l-1', { name: '検索2', color: '#dbeafe' });
  });
});

describe('TicketRepository.deleteLabel', () => {
  it('DELETE を叩く（204）', async () => {
    mockDelete.mockResolvedValue({ data: undefined });
    await TicketRepository.deleteLabel('acme', 's-1', 'l-1');
    expect(mockDelete).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/spaces/s-1/labels/l-1');
  });
});

describe('TicketRepository.addTicketLabel / removeTicketLabel', () => {
  it('PUT/DELETE でチケットへの付け外しをする（どちらも204・冪等）', async () => {
    mockPut.mockResolvedValue({ data: undefined });
    mockDelete.mockResolvedValue({ data: undefined });
    await TicketRepository.addTicketLabel('acme', 't-1', 'l-1');
    await TicketRepository.removeTicketLabel('acme', 't-1', 'l-1');
    expect(mockPut).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/labels/l-1');
    expect(mockDelete).toHaveBeenCalledWith('/api/v2/kb/workspaces/acme/tickets/t-1/labels/l-1');
  });
});
