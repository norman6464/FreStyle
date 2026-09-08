import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useKbPageSuggestions } from '../useKbPageSuggestions';

const hoisted = vi.hoisted(() => ({
  listOpenSuggestions: vi.fn(),
  acceptSuggestion: vi.fn(),
  rejectSuggestion: vi.fn(),
}));

vi.mock('@/entities/kb', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@/entities/kb')>();
  return {
    ...actual,
    KbRepository: {
      listOpenSuggestions: hoisted.listOpenSuggestions,
      acceptSuggestion: hoisted.acceptSuggestion,
      rejectSuggestion: hoisted.rejectSuggestion,
    },
  };
});

const suggestion = (id: string) => ({
  id,
  doc: { type: 'doc', content: [] },
  status: 'open' as const,
  author: { userId: 1, name: '田中 太郎' },
  createdAt: '2026-09-01T00:00:00Z',
});

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useKbPageSuggestions', () => {
  it('open=false の間は取得しない', () => {
    renderHook(() => useKbPageSuggestions('w-1', 'p-1', false));
    expect(hoisted.listOpenSuggestions).not.toHaveBeenCalled();
  });

  it('open になると一覧を取得する', async () => {
    hoisted.listOpenSuggestions.mockResolvedValue([suggestion('s-1')]);
    const { result } = renderHook(() => useKbPageSuggestions('w-1', 'p-1', true));

    await waitFor(() => expect(hoisted.listOpenSuggestions).toHaveBeenCalledWith('w-1', 'p-1'));
    await waitFor(() => expect(result.current.suggestions).toHaveLength(1));
    expect(result.current.error).toBeNull();
  });

  it('取得に失敗したらエラーを持つ', async () => {
    hoisted.listOpenSuggestions.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useKbPageSuggestions('w-1', 'p-1', true));

    await waitFor(() => expect(result.current.error).not.toBeNull());
    expect(result.current.suggestions).toEqual([]);
  });

  it('accept が成功すると一覧からその提案が消え、応答（反映後のdoc込み）を返す', async () => {
    hoisted.listOpenSuggestions.mockResolvedValue([suggestion('s-1'), suggestion('s-2')]);
    const accepted = { ...suggestion('s-1'), status: 'accepted' as const, doc: { type: 'doc', content: [{ type: 'paragraph' }] } };
    hoisted.acceptSuggestion.mockResolvedValue(accepted);
    const { result } = renderHook(() => useKbPageSuggestions('w-1', 'p-1', true));
    await waitFor(() => expect(result.current.suggestions).toHaveLength(2));

    let returned: unknown;
    await act(async () => {
      returned = await result.current.accept('s-1');
    });

    expect(hoisted.acceptSuggestion).toHaveBeenCalledWith('w-1', 'p-1', 's-1');
    expect(returned).toEqual(accepted);
    expect(result.current.suggestions.map((s) => s.id)).toEqual(['s-2']);
  });

  it('reject が成功すると一覧からその提案が消える', async () => {
    hoisted.listOpenSuggestions.mockResolvedValue([suggestion('s-1')]);
    hoisted.rejectSuggestion.mockResolvedValue({ ...suggestion('s-1'), status: 'rejected' as const });
    const { result } = renderHook(() => useKbPageSuggestions('w-1', 'p-1', true));
    await waitFor(() => expect(result.current.suggestions).toHaveLength(1));

    await act(async () => {
      await result.current.reject('s-1');
    });

    expect(hoisted.rejectSuggestion).toHaveBeenCalledWith('w-1', 'p-1', 's-1');
    expect(result.current.suggestions).toEqual([]);
  });

  it('accept が失敗したら投げる。一覧は変わらない', async () => {
    hoisted.listOpenSuggestions.mockResolvedValue([suggestion('s-1')]);
    hoisted.acceptSuggestion.mockRejectedValue(new Error('forbidden'));
    const { result } = renderHook(() => useKbPageSuggestions('w-1', 'p-1', true));
    await waitFor(() => expect(result.current.suggestions).toHaveLength(1));

    await expect(result.current.accept('s-1')).rejects.toThrow('forbidden');
    expect(result.current.suggestions).toHaveLength(1);
  });
});
