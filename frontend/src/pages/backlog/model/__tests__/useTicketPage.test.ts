import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useTicketPage } from '../useTicketPage';
import type { Label, Ticket, TicketPermission } from '@/entities/ticket';

const hoisted = vi.hoisted(() => ({
  resolveTicket: vi.fn(),
  updateTicket: vi.fn(),
  changeTicketStatus: vi.fn(),
  assignTicket: vi.fn(),
  unassignTicket: vi.fn(),
  archiveTicket: vi.fn(),
  restoreTicket: vi.fn(),
  fetchSpaces: vi.fn(),
  addTicketLabel: vi.fn(),
  removeTicketLabel: vi.fn(),
}));

vi.mock('@/entities/ticket', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/ticket')>();
  return {
    ...actual,
    TicketRepository: {
      resolveTicket: hoisted.resolveTicket,
      updateTicket: hoisted.updateTicket,
      changeTicketStatus: hoisted.changeTicketStatus,
      assignTicket: hoisted.assignTicket,
      unassignTicket: hoisted.unassignTicket,
      archiveTicket: hoisted.archiveTicket,
      restoreTicket: hoisted.restoreTicket,
      addTicketLabel: hoisted.addTicketLabel,
      removeTicketLabel: hoisted.removeTicketLabel,
    },
  };
});

vi.mock('@/entities/kb', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/kb')>();
  return {
    ...actual,
    KbRepository: { fetchSpaces: hoisted.fetchSpaces },
  };
});

const permission: TicketPermission = { canView: true, canComment: true, canEdit: true, canManage: false };

function ticket(over: Partial<Ticket> = {}): Ticket {
  return {
    id: 't-1',
    workspaceId: 'w-1',
    spaceId: 's-1',
    number: 12,
    typeId: 'ty-1',
    statusId: 'st-1',
    parentId: null,
    title: '本文',
    doc: { type: 'doc', content: [] },
    priority: 2,
    startDate: null,
    dueDate: null,
    position: 'a0',
    closedAt: null,
    resolution: null,
    createdByUserId: 1,
    archivedAt: null,
    createdAt: '',
    updatedAt: '',
    assigneePrincipalId: null,
    labels: [],
    ...over,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.fetchSpaces.mockResolvedValue([{ id: 's-1', key: 'FRESTYLE', name: 'FreStyle' }]);
});

describe('useTicketPage', () => {
  it('解決してワークスペース・チケット・祖先・権限・スペースを持つ', async () => {
    hoisted.resolveTicket.mockResolvedValue({
      workspaceSlug: 'acme',
      workspaceName: 'Acme',
      ticket: ticket(),
      canEdit: true,
      ancestors: [],
      permission,
    });

    const { result } = renderHook(() => useTicketPage('t-1'));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.workspaceSlug).toBe('acme');
    expect(result.current.ticket?.id).toBe('t-1');
    expect(result.current.permission).toEqual(permission);
    expect(result.current.space?.key).toBe('FRESTYLE');
  });

  it('404はチケットが見つからない文言、それ以外は読み込み失敗の文言', async () => {
    hoisted.resolveTicket.mockRejectedValueOnce({ response: { status: 404 } });
    const notFound = renderHook(() => useTicketPage('t-404'));
    await waitFor(() => expect(notFound.result.current.error).toBe('チケットが見つかりませんでした。'));

    hoisted.resolveTicket.mockRejectedValueOnce({ response: { status: 500 } });
    const failed = renderHook(() => useTicketPage('t-500'));
    await waitFor(() =>
      expect(failed.result.current.error).toBe(
        'チケットを開けませんでした。時間をおいて開き直すと最新の状態が出ます。',
      ),
    );
  });

  it('スペースが引けなくてもチケットは表示する（キーだけ出ない）', async () => {
    hoisted.resolveTicket.mockResolvedValue({
      workspaceSlug: 'acme',
      workspaceName: 'Acme',
      ticket: ticket(),
      canEdit: true,
      ancestors: [],
      permission,
    });
    hoisted.fetchSpaces.mockRejectedValue(new Error('network'));

    const { result } = renderHook(() => useTicketPage('t-1'));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBeNull();
    expect(result.current.ticket?.id).toBe('t-1');
    expect(result.current.space).toBeNull();
  });

  it('宛先を切り替えたら古い応答を無視する', async () => {
    let resolveFirst!: (v: unknown) => void;
    hoisted.resolveTicket.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveFirst = resolve;
        }),
    );
    hoisted.resolveTicket.mockResolvedValueOnce({
      workspaceSlug: 'acme',
      workspaceName: 'Acme',
      ticket: ticket({ id: 't-2', title: '2件目' }),
      canEdit: true,
      ancestors: [],
      permission,
    });

    const { result, rerender } = renderHook(({ id }) => useTicketPage(id), { initialProps: { id: 't-1' } });
    rerender({ id: 't-2' });

    await waitFor(() => expect(result.current.ticket?.id).toBe('t-2'));

    resolveFirst({
      workspaceSlug: 'acme',
      workspaceName: 'Acme',
      ticket: ticket({ id: 't-1', title: '1件目（古い）' }),
      canEdit: true,
      ancestors: [],
      permission,
    });

    await new Promise((r) => setTimeout(r, 0));
    expect(result.current.ticket?.id).toBe('t-2');
  });

  it('204で本体が返らない担当解除は手元で外す', async () => {
    hoisted.resolveTicket.mockResolvedValue({
      workspaceSlug: 'acme',
      workspaceName: 'Acme',
      ticket: ticket({ assigneePrincipalId: 'p-1' }),
      canEdit: true,
      ancestors: [],
      permission,
    });
    hoisted.unassignTicket.mockResolvedValue(undefined);

    const { result } = renderHook(() => useTicketPage('t-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.unassign();
    });

    expect(result.current.ticket?.assigneePrincipalId).toBeNull();
  });

  it('書き込みの失敗は投げ直し、busyを戻す', async () => {
    hoisted.resolveTicket.mockResolvedValue({
      workspaceSlug: 'acme',
      workspaceName: 'Acme',
      ticket: ticket(),
      canEdit: true,
      ancestors: [],
      permission,
    });
    hoisted.archiveTicket.mockRejectedValue(new Error('403'));

    const { result } = renderHook(() => useTicketPage('t-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await expect(
      act(async () => {
        await result.current.archive();
      }),
    ).rejects.toThrow('403');
    expect(result.current.busy).toBe(false);
  });
});

describe('addLabel / removeLabel', () => {
  const permission: TicketPermission = { canView: true, canComment: true, canEdit: true, canManage: false };
  const label = (over: Partial<Label> & { id: string }): Label => ({
    spaceId: 's-1',
    name: 'ラベル',
    color: '#1d4ed8',
    createdAt: '',
    updatedAt: '',
    ...over,
  });

  it('付けると手元の labels に足す（重複しては足さない）', async () => {
    hoisted.resolveTicket.mockResolvedValue({
      workspaceSlug: 'acme',
      workspaceName: 'Acme',
      ticket: ticket(),
      canEdit: true,
      ancestors: [],
      permission,
    });
    hoisted.addTicketLabel.mockResolvedValue(undefined);
    const l1 = label({ id: 'l-1' });

    const { result } = renderHook(() => useTicketPage('t-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.addLabel(l1);
    });
    expect(result.current.ticket?.labels).toEqual([l1]);
    expect(hoisted.addTicketLabel).toHaveBeenCalledWith('acme', 't-1', 'l-1');

    await act(async () => {
      await result.current.addLabel(l1);
    });
    expect(result.current.ticket?.labels).toEqual([l1]);
  });

  it('外すと手元の labels から取り除く', async () => {
    hoisted.resolveTicket.mockResolvedValue({
      workspaceSlug: 'acme',
      workspaceName: 'Acme',
      ticket: ticket({ labels: [label({ id: 'l-1' })] }),
      canEdit: true,
      ancestors: [],
      permission,
    });
    hoisted.removeTicketLabel.mockResolvedValue(undefined);

    const { result } = renderHook(() => useTicketPage('t-1'));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.removeLabel('l-1');
    });
    expect(result.current.ticket?.labels).toEqual([]);
  });
});
