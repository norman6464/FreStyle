import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { useDocumentEditor } from '../useDocumentEditor';

const fetchDocument = vi.fn();
const updateDocument = vi.fn();

vi.mock('@/entities/document', () => ({
  DocumentRepository: {
    fetchDocument: (...a: unknown[]) => fetchDocument(...a),
    updateDocument: (...a: unknown[]) => updateDocument(...a),
  },
}));

vi.mock('@/shared/ui/RichTextEditor', () => ({
  emptyRichDoc: () => ({ type: 'doc', content: [{ type: 'paragraph' }] }),
}));

const docBody = { type: 'doc', content: [{ type: 'text', text: 'x' }] };

function fullDoc(over: Record<string, unknown> = {}) {
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
  });

  it('編集は debounce 後に revision 付きで保存し、saveStatus と revision を更新する', async () => {
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
    expect(onSynced).toHaveBeenCalled();
  });

  it('版不一致(409)は最新版を取り直し onConflict を呼ぶ', async () => {
    fetchDocument.mockResolvedValueOnce(fullDoc({ title: '初期', revision: 3 }));
    const conflictErr = Object.assign(new Error('conflict'), {
      isAxiosError: true,
      response: { status: 409 },
    });
    updateDocument.mockRejectedValue(conflictErr);
    // 409 後の取り直しでサーバ最新版を返す。
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
    expect(onSynced).toHaveBeenCalled(); // 取り直したサマリで一覧も同期
  });

  it('forceSave は debounce を待たず即保存する', async () => {
    fetchDocument.mockResolvedValue(fullDoc({ revision: 3 }));
    updateDocument.mockResolvedValue(fullDoc({ revision: 4 }));
    const { result } = renderHook(() => useDocumentEditor('a'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    act(() => result.current.handleTitleChange('即'));
    await act(async () => {
      result.current.forceSave();
      await vi.runAllTimersAsync();
    });
    expect(updateDocument).toHaveBeenCalled();
    expect(result.current.saveStatus).toBe('saved');
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

describe('useDocumentEditor エラー系', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });
  afterEach(() => vi.useRealTimers());

  it('doc 取得失敗でも落ちず loadingDoc は false に戻る', async () => {
    fetchDocument.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useDocumentEditor('a'));
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(result.current.loadingDoc).toBe(false);
  });

  it('保存が 409 以外のエラーなら saveStatus を idle に戻す', async () => {
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
    expect(result.current.saveStatus).toBe('idle');
  });
});
