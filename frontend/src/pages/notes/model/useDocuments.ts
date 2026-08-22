import { useState, useCallback, useMemo } from 'react';
import { DocumentRepository, toRichDocumentSummary, type RichDocumentSummary } from '@/entities/document';
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
    setError(null);
    try {
      const created = await DocumentRepository.createDocument({
        kind: 'note',
        title,
        doc: emptyRichDoc(),
      });
      // 一覧は doc 本体を持たないサマリなので、作成結果から doc を除いて先頭へ積む。
      setDocuments((prev) => [toRichDocumentSummary(created), ...prev]);
      setSelectedId(created.id);
      return created;
    } catch {
      setError('ノートの作成に失敗しました');
      return null;
    }
  }, []);

  // deleteDocument は成功時 true / 失敗時 false を返す。失敗時は一覧も選択も変えない。
  const deleteDocument = useCallback(async (id: string) => {
    setError(null);
    try {
      await DocumentRepository.deleteDocument(id);
      setDocuments((prev) => prev.filter((d) => d.id !== id));
      setSelectedId((prev) => (prev === id ? null : prev));
      return true;
    } catch {
      setError('ノートの削除に失敗しました');
      return false;
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
      ? documents.filter((document) => document.title.toLowerCase().includes(query))
      : documents;
    const toMillis = (rfc3339: string) => Date.parse(rfc3339) || 0;
    return [...filtered].sort((a, b) => {
      if (sort === 'updated-asc') return toMillis(a.updatedAt) - toMillis(b.updatedAt);
      if (sort === 'title') return a.title.localeCompare(b.title, 'ja');
      if (sort === 'created-desc') return toMillis(b.createdAt) - toMillis(a.createdAt);
      return toMillis(b.updatedAt) - toMillis(a.updatedAt);
    });
  }, [documents, searchQuery, sort]);

  const requestDelete = useCallback((id: string) => {
    setDeleteTargetId(id);
  }, []);

  // confirmDelete は成功時 true / 失敗時 false を返す。成功したときだけ選択を次候補へ移す。
  const confirmDelete = useCallback(async () => {
    if (deleteTargetId == null) return false;
    // 削除前に次の選択候補を決める（filteredDocuments の順で次→前）。
    const idx = filteredDocuments.findIndex((d) => d.id === deleteTargetId);
    const nextDoc = idx >= 0 ? filteredDocuments[idx + 1] || filteredDocuments[idx - 1] || null : null;

    const deletingSelected = selectedId === deleteTargetId;
    const ok = await deleteDocument(deleteTargetId);
    if (ok && deletingSelected) {
      setSelectedId(nextDoc ? nextDoc.id : null);
    }
    setDeleteTargetId(null);
    return ok;
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
