import { useState, useCallback, useMemo } from 'react';
import { DocumentRepository, type RichDocumentSummary } from '@/entities/document';
import { emptyRichDoc } from '@/shared/ui/RichTextEditor';
import type { NoteSortOption } from '../config/sortOptions';

/**
 * useDocuments — リッチ文書（kind='note'）の一覧取得・選択・並び替え・作成/削除を担う hook。
 *
 * id は UUID 文字列。一覧はサマリ（doc 本体なし）で、本文は選択時に useDocumentEditor が個別取得する。
 * 旧 note（Markdown・number id・pin）を置き換えるもので、pin と本文プレビューは持たない。
 */
export function useDocuments() {
  const [documents, setDocuments] = useState<RichDocumentSummary[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);
  const [sort, setSort] = useState<NoteSortOption>('default');

  const fetchDocuments = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await DocumentRepository.fetchDocuments('note');
      setDocuments(data);
    } catch {
      setError('ノートの取得に失敗しました');
    } finally {
      setLoading(false);
    }
  }, []);

  const createDocument = useCallback(async (title: string) => {
    try {
      const created = await DocumentRepository.createDocument({
        kind: 'note',
        title,
        doc: emptyRichDoc(),
      });
      // 一覧は doc 本体を持たないサマリなので、作成結果から doc を除いて先頭へ積む。
      const { doc: _doc, ...summary } = created;
      void _doc;
      setDocuments((prev) => [summary, ...prev]);
      setSelectedId(created.id);
      return created;
    } catch {
      setError('ノートの作成に失敗しました');
      return null;
    }
  }, []);

  const deleteDocument = useCallback(async (id: string) => {
    try {
      await DocumentRepository.deleteDocument(id);
      setDocuments((prev) => prev.filter((d) => d.id !== id));
      setSelectedId((prev) => (prev === id ? null : prev));
    } catch {
      setError('ノートの削除に失敗しました');
    }
  }, []);

  const selectDocument = useCallback((id: string | null) => {
    setSelectedId(id);
  }, []);

  // 保存後にサマリ（title / updatedAt / revision）を一覧へ反映する。
  const syncSummary = useCallback((updated: RichDocumentSummary) => {
    setDocuments((prev) => prev.map((d) => (d.id === updated.id ? { ...d, ...updated } : d)));
  }, []);

  const filteredDocuments = useMemo(() => {
    const query = searchQuery.toLowerCase();
    // 一覧は doc 本体を持たないため、検索はタイトルのみを対象にする。
    const filtered = query
      ? documents.filter((d) => d.title.toLowerCase().includes(query))
      : documents;
    const ms = (s: string) => Date.parse(s) || 0;
    return [...filtered].sort((a, b) => {
      if (sort === 'updated-asc') return ms(a.updatedAt) - ms(b.updatedAt);
      if (sort === 'title') return a.title.localeCompare(b.title, 'ja');
      if (sort === 'created-desc') return ms(b.createdAt) - ms(a.createdAt);
      return ms(b.updatedAt) - ms(a.updatedAt);
    });
  }, [documents, searchQuery, sort]);

  const requestDelete = useCallback((id: string) => {
    setDeleteTargetId(id);
  }, []);

  const confirmDelete = useCallback(async () => {
    if (deleteTargetId == null) return;
    // 削除前に次の選択候補を決める（filteredDocuments の順で次→前）。
    const idx = filteredDocuments.findIndex((d) => d.id === deleteTargetId);
    const nextDoc = idx >= 0 ? filteredDocuments[idx + 1] || filteredDocuments[idx - 1] || null : null;

    const deletingSelected = selectedId === deleteTargetId;
    await deleteDocument(deleteTargetId);
    if (deletingSelected) {
      setSelectedId(nextDoc ? nextDoc.id : null);
    }
    setDeleteTargetId(null);
  }, [deleteTargetId, deleteDocument, filteredDocuments, selectedId]);

  const cancelDelete = useCallback(() => {
    setDeleteTargetId(null);
  }, []);

  return {
    documents,
    filteredDocuments,
    selectedId,
    loading,
    error,
    searchQuery,
    setSearchQuery,
    sort,
    setSort,
    fetchDocuments,
    createDocument,
    deleteDocument,
    selectDocument,
    syncSummary,
    deleteTargetId,
    requestDelete,
    confirmDelete,
    cancelDelete,
  };
}
