import { describe, it, expect, vi, beforeEach } from 'vitest';
import DocumentRepository from '../documentRepository';
import apiClient from '@/shared/api/axios';
import type { RichDocument, RichDocumentSummary } from '../../model/types';

vi.mock('@/shared/api/axios');

const summary: RichDocumentSummary = {
  id: '31400a07-297e-8057-884b-c05dbdf9fa53',
  ownerId: 7,
  kind: 'note',
  title: 'メモ',
  isPublic: false,
  schemaVersion: 1,
  revision: 2,
  createdAt: '2026-08-01T00:00:00Z',
  updatedAt: '2026-08-02T00:00:00Z',
};

const full: RichDocument = {
  ...summary,
  doc: { type: 'doc', content: [{ type: 'paragraph' }] },
};

describe('DocumentRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetchDocuments は GET /documents（params なし）で配列を返す', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: [summary] });
    const result = await DocumentRepository.fetchDocuments();
    expect(apiClient.get).toHaveBeenCalledWith('/api/v2/documents', { params: undefined });
    expect(result).toEqual([summary]);
  });

  it('fetchDocuments は kind を params で渡す', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: [summary] });
    await DocumentRepository.fetchDocuments('note');
    expect(apiClient.get).toHaveBeenCalledWith('/api/v2/documents', { params: { kind: 'note' } });
  });

  it('fetchDocuments は配列以外を空配列にフォールバックする', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: null });
    expect(await DocumentRepository.fetchDocuments()).toEqual([]);
  });

  it('fetchDocument は GET /documents/:id で doc 込みの 1 件を返す', async () => {
    vi.mocked(apiClient.get).mockResolvedValue({ data: full });
    const result = await DocumentRepository.fetchDocument(summary.id);
    expect(apiClient.get).toHaveBeenCalledWith(`/api/v2/documents/${summary.id}`);
    expect(result).toEqual(full);
  });

  it('createDocument は POST /documents に入力を送り 1 件を返す', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({ data: full });
    const input = { kind: 'note' as const, title: 'メモ', doc: full.doc };
    const result = await DocumentRepository.createDocument(input);
    expect(apiClient.post).toHaveBeenCalledWith('/api/v2/documents', input);
    expect(result).toEqual(full);
  });

  it('updateDocument は PUT /documents/:id に revision 込みで送る', async () => {
    vi.mocked(apiClient.put).mockResolvedValue({ data: { ...full, revision: 3 } });
    const input = { title: '更新', doc: full.doc, revision: 2 };
    const result = await DocumentRepository.updateDocument(summary.id, input);
    expect(apiClient.put).toHaveBeenCalledWith(`/api/v2/documents/${summary.id}`, input);
    expect(result.revision).toBe(3);
  });

  it('deleteDocument は DELETE /documents/:id を叩く', async () => {
    vi.mocked(apiClient.delete).mockResolvedValue({ data: undefined });
    await DocumentRepository.deleteDocument(summary.id);
    expect(apiClient.delete).toHaveBeenCalledWith(`/api/v2/documents/${summary.id}`);
  });
});
