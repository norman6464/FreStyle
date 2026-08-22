import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { AxiosError } from 'axios';
import { useDocumentEditor } from '../useDocumentEditor';
import type { RichDocument, RichDocumentSummary } from '@/entities/document';

const fetchDocument = vi.fn();
const updateDocument = vi.fn();

vi.mock('@/entities/document', () => ({
  DocumentRepository: {
    fetchDocument: (...a: unknown[]) => fetchDocument(...a),
    updateDocument: (...a: unknown[]) => updateDocument(...a),
  },
  toRichDocumentSummary: (document: RichDocument): RichDocumentSummary => {
    const { doc: _doc, ...summary } = document;
    void _doc;
    return summary;
  },
}));

vi.mock('@/shared/ui/RichTextEditor', () => ({
  emptyRichDoc: () => ({ type: 'doc', content: [{ type: 'paragraph' }] }),
}));

const docBody = { type: 'doc', content: [{ type: 'text', text: 'x' }] };

function fullDoc(over: Partial<RichDocument> = {}): RichDocument {
  return {
    id: 'a',
    ownerId: 7,
    kind: 'note',
    title: 'タイトル',
    isPublic: false,
    schemaVersion: 1,
    revision: 3,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-02T00:00:00Z',
    doc: docBody,
    ...over,
  };
}

function conflict409(): AxiosError {
  return new AxiosError('conflict', 'ERR_CONFLICT', undefined, undefined, {
    status: 409,
    statusText: 'Conflict',
    data: {},
    headers: {},
    config: {} as never,
  });
}

describe('useDocumentEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('選択が変わると doc 本体を取得してエディタへ入れる', async () => {
    fetchDocument.mockResolvedValue(fullDoc({ title: '読込済み', revision: 5 }));
    const { result } = renderHook(() => useDocumentEditor('a'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(fetchDocument).toHaveBeenCalledWith('a');
    expect(result.current.editTitle).toBe('読込済み');
    expect(result.current.loadingDoc).toBe(false);
    expect(result.current.loadError).toBe(false);
  });

  it('doc 取得失敗で loadError=true、loadingDoc=false になる', async () => {
    fetchDocument.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useDocumentEditor('a'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(result.current.loadError).toBe(true);
    expect(result.current.loadingDoc).toBe(false);
  });

  it('reload で取得失敗から復帰できる', async () => {
    fetchDocument.mockRejectedValueOnce(new Error('boom'));
    fetchDocument.mockResolvedValueOnce(fullDoc({ title: '復帰' }));
    const { result } = renderHook(() => useDocumentEditor('a'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(result.current.loadError).toBe(true);
    await act(async () => {
      result.current.reload();
      await vi.runAllTimersAsync();
    });
    expect(result.current.loadError).toBe(false);
    expect(result.current.editTitle).toBe('復帰');
  });

  it('編集は debounce 後に revision 付きで保存し、saveStatus と onSynced を更新する', async () => {
    fetchDocument.mockResolvedValue(fullDoc({ revision: 3 }));
    updateDocument.mockResolvedValue(fullDoc({ title: '新', revision: 4 }));
    const onSynced = vi.fn();
    const { result } = renderHook(() => useDocumentEditor('a', { onSynced }));
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    act(() => result.current.handleTitleChange('新'));
    expect(result.current.saveStatus).toBe('unsaved');

    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(updateDocument).toHaveBeenCalledWith('a', { title: '新', doc: docBody, revision: 3 });
    expect(result.current.saveStatus).toBe('saved');
    // onSynced はサマリ（doc 本体なし）で呼ばれる。
    expect(onSynced).toHaveBeenCalled();
    expect(onSynced.mock.calls.at(-1)?.[0]).not.toHaveProperty('doc');
  });

  it('版不一致(409・AxiosError)は最新版を取り直し onConflict を呼ぶ', async () => {
    fetchDocument.mockResolvedValueOnce(fullDoc({ title: '初期', revision: 3 }));
    updateDocument.mockRejectedValue(conflict409());
    fetchDocument.mockResolvedValueOnce(fullDoc({ title: 'サーバ最新', revision: 9 }));
    const onConflict = vi.fn();
    const onSynced = vi.fn();
    const { result } = renderHook(() => useDocumentEditor('a', { onConflict, onSynced }));
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    act(() => result.current.handleTitleChange('ローカル編集'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    expect(onConflict).toHaveBeenCalled();
    expect(result.current.editTitle).toBe('サーバ最新');
    expect(result.current.saveStatus).toBe('saved');
    expect(onSynced).toHaveBeenCalled();
  });

  it('保存が 409 以外のエラーなら saveStatus を unsaved に戻す（無表示にしない）', async () => {
    fetchDocument.mockResolvedValue(fullDoc({ revision: 3 }));
    updateDocument.mockRejectedValue(new Error('500'));
    const { result } = renderHook(() => useDocumentEditor('a'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    act(() => result.current.handleTitleChange('x'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(result.current.saveStatus).toBe('unsaved');
  });

  it('forceSave は 800ms を待たず即保存し、その後タイマーが進んでも二重保存しない', async () => {
    fetchDocument.mockResolvedValue(fullDoc({ revision: 3 }));
    updateDocument.mockResolvedValue(fullDoc({ revision: 4 }));
    const { result } = renderHook(() => useDocumentEditor('a'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    act(() => result.current.handleTitleChange('即'));
    // debounce(800ms) を進めずに forceSave → 0ms 進行（＝マイクロタスクのみ）で保存される。
    await act(async () => {
      result.current.forceSave();
      await vi.advanceTimersByTimeAsync(0);
    });
    expect(updateDocument).toHaveBeenCalledTimes(1);

    // forceSave が debounce タイマーを解除しているので、800ms 進めても再保存しない。
    await act(async () => {
      await vi.advanceTimersByTimeAsync(800);
    });
    expect(updateDocument).toHaveBeenCalledTimes(1);
  });

  it('debounce 中に別文書へ切り替えると、前の文書の保存が実行される（編集を失わない）', async () => {
    fetchDocument.mockResolvedValue(fullDoc({ id: 'a', revision: 3 }));
    updateDocument.mockResolvedValue(fullDoc({ id: 'a', revision: 4 }));
    const { result, rerender } = renderHook(({ id }) => useDocumentEditor(id), {
      initialProps: { id: 'a' as string | null },
    });
    await act(async () => {
      await vi.runAllTimersAsync();
    });

    // a を編集（保存タイマー保留）→ 800ms 以内に b へ切替。
    act(() => result.current.handleTitleChange('a の編集'));
    fetchDocument.mockResolvedValue(fullDoc({ id: 'b', title: 'B', revision: 1 }));
    await act(async () => {
      rerender({ id: 'b' });
      await vi.runAllTimersAsync();
    });

    // 前の文書 a に対して保存が走っている（編集が捨てられない）。
    expect(updateDocument).toHaveBeenCalledWith('a', expect.objectContaining({ title: 'a の編集' }));
  });

  it('未選択のときは何も取得しない', async () => {
    const { result } = renderHook(() => useDocumentEditor(null));
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(fetchDocument).not.toHaveBeenCalled();
    expect(result.current.editTitle).toBe('');
  });
});
