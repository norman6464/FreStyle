import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useTicketAttachments } from '../useTicketAttachments';
import type { TicketAttachment } from '@/entities/ticket';

const hoisted = vi.hoisted(() => ({
  fetchTicketAttachments: vi.fn(),
  issueTicketAttachmentUploadUrl: vi.fn(),
  putTicketAttachmentFile: vi.fn(),
  createTicketAttachment: vi.fn(),
  deleteTicketAttachment: vi.fn(),
}));

vi.mock('@/entities/ticket', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/ticket')>();
  return {
    ...actual,
    TicketRepository: {
      fetchTicketAttachments: hoisted.fetchTicketAttachments,
      issueTicketAttachmentUploadUrl: hoisted.issueTicketAttachmentUploadUrl,
      putTicketAttachmentFile: hoisted.putTicketAttachmentFile,
      createTicketAttachment: hoisted.createTicketAttachment,
      deleteTicketAttachment: hoisted.deleteTicketAttachment,
    },
  };
});

function fixtureAttachment(over: Partial<TicketAttachment> & { id: string }): TicketAttachment {
  return {
    ticketId: 't-1',
    filename: 'file.png',
    contentType: 'image/png',
    sizeBytes: 100,
    uploadedByUserId: 1,
    createdAt: '',
    ...over,
  };
}

const SLUG = 'acme';
const TICKET = 't-1';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useTicketAttachments', () => {
  it('宛先が揃ったら取得する', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([fixtureAttachment({ id: 'a-1' })]);
    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.attachments).toHaveLength(1);
    expect(hoisted.fetchTicketAttachments).toHaveBeenCalledWith(SLUG, TICKET);
  });

  it('取得失敗は文言を出す', async () => {
    hoisted.fetchTicketAttachments.mockRejectedValue(new Error('network'));
    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.error).not.toBeNull());
  });

  it('宛先が揃っていなければ何もしない', () => {
    const { result } = renderHook(() => useTicketAttachments(undefined, undefined));
    expect(result.current.attachments).toEqual([]);
    expect(hoisted.fetchTicketAttachments).not.toHaveBeenCalled();
  });

  it('アップロードは presign → PUT → 確定の順で進み、成功すると確定側へ移る', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([]);
    hoisted.issueTicketAttachmentUploadUrl.mockResolvedValue({ url: 'https://gcs/put', key: 'k-1', expiresIn: 600 });
    hoisted.putTicketAttachmentFile.mockResolvedValue(undefined);
    const created = fixtureAttachment({ id: 'a-1', filename: 'diagram.png' });
    hoisted.createTicketAttachment.mockResolvedValue(created);

    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));

    const file = new File(['x'], 'diagram.png', { type: 'image/png' });
    act(() => {
      result.current.upload(file);
    });
    expect(result.current.pending).toHaveLength(1);
    expect(result.current.pending[0].status).toBe('uploading');

    await waitFor(() => expect(result.current.pending).toHaveLength(0));
    expect(result.current.attachments).toEqual([created]);

    expect(hoisted.issueTicketAttachmentUploadUrl).toHaveBeenCalledWith(SLUG, TICKET, 'image/png', file.size);
    expect(hoisted.putTicketAttachmentFile).toHaveBeenCalledWith('https://gcs/put', file);
    expect(hoisted.createTicketAttachment).toHaveBeenCalledWith(SLUG, TICKET, {
      key: 'k-1',
      filename: 'diagram.png',
      contentType: 'image/png',
      sizeBytes: file.size,
    });
  });

  it('対応していない形式は API を呼ばずに失敗行になる', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([]);
    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));

    const file = new File(['x'], 'script.exe', { type: 'application/x-msdownload' });
    act(() => {
      result.current.upload(file);
    });

    expect(result.current.pending).toHaveLength(1);
    expect(result.current.pending[0].status).toBe('failed');
    expect(hoisted.issueTicketAttachmentUploadUrl).not.toHaveBeenCalled();
  });

  it('上限を超えるサイズは API を呼ばずに失敗行になる', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([]);
    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));

    const big = new File([new Uint8Array(10)], 'big.pdf', { type: 'application/pdf' });
    Object.defineProperty(big, 'size', { value: 26 * 1024 * 1024 });
    act(() => {
      result.current.upload(big);
    });

    expect(result.current.pending[0].status).toBe('failed');
    expect(hoisted.issueTicketAttachmentUploadUrl).not.toHaveBeenCalled();
  });

  it('アップロード失敗は行を failed のまま残す', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([]);
    hoisted.issueTicketAttachmentUploadUrl.mockRejectedValue(new Error('network'));

    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));

    const file = new File(['x'], 'diagram.png', { type: 'image/png' });
    act(() => {
      result.current.upload(file);
    });

    await waitFor(() => expect(result.current.pending[0]?.status).toBe('failed'));
    expect(result.current.attachments).toEqual([]);
  });

  it('retry は同じファイルを持ち回して再送する', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([]);
    hoisted.issueTicketAttachmentUploadUrl.mockRejectedValueOnce(new Error('network'));

    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));

    const file = new File(['x'], 'diagram.png', { type: 'image/png' });
    act(() => {
      result.current.upload(file);
    });
    await waitFor(() => expect(result.current.pending[0]?.status).toBe('failed'));

    const created = fixtureAttachment({ id: 'a-1', filename: 'diagram.png' });
    hoisted.issueTicketAttachmentUploadUrl.mockResolvedValue({ url: 'https://gcs/put', key: 'k-1', expiresIn: 600 });
    hoisted.putTicketAttachmentFile.mockResolvedValue(undefined);
    hoisted.createTicketAttachment.mockResolvedValue(created);

    const clientId = result.current.pending[0].clientId;
    act(() => {
      result.current.retry(clientId);
    });

    await waitFor(() => expect(result.current.pending).toHaveLength(0));
    expect(result.current.attachments).toEqual([created]);
  });

  it('dismiss は失敗行を消す（API は呼ばない）', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([]);
    hoisted.issueTicketAttachmentUploadUrl.mockRejectedValue(new Error('network'));

    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => {
      result.current.upload(new File(['x'], 'diagram.png', { type: 'image/png' }));
    });
    await waitFor(() => expect(result.current.pending[0]?.status).toBe('failed'));

    act(() => {
      result.current.dismiss(result.current.pending[0].clientId);
    });
    expect(result.current.pending).toEqual([]);
  });

  it('削除すると一覧から外す', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([fixtureAttachment({ id: 'a-1' })]);
    hoisted.deleteTicketAttachment.mockResolvedValue(undefined);

    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.remove('a-1');
    });

    expect(result.current.attachments).toEqual([]);
    expect(result.current.busyId).toBeNull();
  });

  it('削除失敗は投げて busyId を戻す', async () => {
    hoisted.fetchTicketAttachments.mockResolvedValue([fixtureAttachment({ id: 'a-1' })]);
    hoisted.deleteTicketAttachment.mockRejectedValue(new Error('403'));

    const { result } = renderHook(() => useTicketAttachments(SLUG, TICKET));
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await expect(result.current.remove('a-1')).rejects.toThrow();
    });

    expect(result.current.attachments).toHaveLength(1);
    expect(result.current.busyId).toBeNull();
  });
});
