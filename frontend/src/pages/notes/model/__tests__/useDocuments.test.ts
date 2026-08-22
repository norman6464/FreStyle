import { renderHook, act, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useDocuments } from '../useDocuments';

const fetchDocuments = vi.fn();
const createDocument = vi.fn();
const deleteDocument = vi.fn();

vi.mock('@/entities/document', () => ({
  DocumentRepository: {
    fetchDocuments: (...a: unknown[]) => fetchDocuments(...a),
    createDocument: (...a: unknown[]) => createDocument(...a),
    deleteDocument: (...a: unknown[]) => deleteDocument(...a),
  },
}));

vi.mock('@/shared/ui/RichTextEditor', () => ({
  emptyRichDoc: () => ({ type: 'doc', content: [{ type: 'paragraph' }] }),
}));

function summary(id: string, over: Record<string, unknown> = {}) {
  return {
    id,
    ownerId: 7,
    kind: 'note',
    title: `T${id}`,
    isPublic: false,
    schemaVersion: 1,
    revision: 1,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-02T00:00:00Z',
    ...over,
  };
}

describe('useDocuments', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchDocuments は kind=note で取得し一覧へ入れる', async () => {
    fetchDocuments.mockResolvedValue([summary('a'), summary('b')]);
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.fetchDocuments();
    });
    expect(fetchDocuments).toHaveBeenCalledWith('note');
    expect(result.current.documents).toHaveLength(2);
  });

  it('createDocument は先頭へ積み選択する（doc 本体はサマリから除く）', async () => {
    createDocument.mockResolvedValue({ ...summary('new'), doc: { type: 'doc', content: [] } });
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.createDocument('無題');
    });
    expect(createDocument).toHaveBeenCalledWith({
      kind: 'note',
      title: '無題',
      doc: { type: 'doc', content: [{ type: 'paragraph' }] },
    });
    expect(result.current.documents[0].id).toBe('new');
    expect(result.current.selectedId).toBe('new');
    // サマリに doc 本体は含めない。
    expect('doc' in result.current.documents[0]).toBe(false);
  });

  it('deleteDocument は一覧から除き、選択中なら解除する', async () => {
    fetchDocuments.mockResolvedValue([summary('a'), summary('b')]);
    deleteDocument.mockResolvedValue(undefined);
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.fetchDocuments();
    });
    act(() => result.current.selectDocument('a'));
    await act(async () => {
      await result.current.deleteDocument('a');
    });
    expect(result.current.documents.map((d) => d.id)).toEqual(['b']);
    expect(result.current.selectedId).toBeNull();
  });

  it('検索はタイトルのみを対象にする', async () => {
    fetchDocuments.mockResolvedValue([summary('a', { title: '買い物' }), summary('b', { title: '仕事' })]);
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.fetchDocuments();
    });
    act(() => result.current.setSearchQuery('買い'));
    expect(result.current.filteredDocuments.map((d) => d.id)).toEqual(['a']);
  });

  it('sort=title でタイトル順に並ぶ', async () => {
    fetchDocuments.mockResolvedValue([summary('a', { title: 'び' }), summary('b', { title: 'あ' })]);
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.fetchDocuments();
    });
    act(() => result.current.setSort('title'));
    expect(result.current.filteredDocuments.map((d) => d.title)).toEqual(['あ', 'び']);
  });

  it('confirmDelete は削除後に次の候補を選択する', async () => {
    fetchDocuments.mockResolvedValue([summary('a'), summary('b'), summary('c')]);
    deleteDocument.mockResolvedValue(undefined);
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.fetchDocuments();
    });
    act(() => result.current.selectDocument('a'));
    act(() => result.current.requestDelete('a'));
    await act(async () => {
      await result.current.confirmDelete();
    });
    await waitFor(() => expect(result.current.selectedId).toBe('b'));
    expect(result.current.deleteTargetId).toBeNull();
  });

  it('syncSummary は該当行の title/updatedAt を更新する', async () => {
    fetchDocuments.mockResolvedValue([summary('a', { title: '旧' })]);
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.fetchDocuments();
    });
    act(() => result.current.syncSummary(summary('a', { title: '新' })));
    expect(result.current.documents[0].title).toBe('新');
  });
});

describe('useDocuments エラー系', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetchDocuments 失敗で error を設定する', async () => {
    fetchDocuments.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.fetchDocuments();
    });
    expect(result.current.error).toBe('ノートの取得に失敗しました');
    expect(result.current.documents).toEqual([]);
  });

  it('createDocument 失敗で null を返し error を設定する', async () => {
    createDocument.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useDocuments());
    let created: unknown;
    await act(async () => {
      created = await result.current.createDocument('無題');
    });
    expect(created).toBeNull();
    expect(result.current.error).toBe('ノートの作成に失敗しました');
  });

  it('deleteDocument 失敗で error を設定し一覧は残す', async () => {
    fetchDocuments.mockResolvedValue([summary('a')]);
    deleteDocument.mockRejectedValue(new Error('boom'));
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.fetchDocuments();
    });
    await act(async () => {
      await result.current.deleteDocument('a');
    });
    expect(result.current.error).toBe('ノートの削除に失敗しました');
    expect(result.current.documents).toHaveLength(1);
  });

  it('confirmDelete は対象が無ければ何もしない', async () => {
    const { result } = renderHook(() => useDocuments());
    await act(async () => {
      await result.current.confirmDelete();
    });
    expect(deleteDocument).not.toHaveBeenCalled();
  });
});
