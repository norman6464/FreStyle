import apiClient from '@/shared/api/axios';
import { toArray } from '@/shared/lib/toArray';
import { DOCUMENTS } from '@/shared/config/apiRoutes';
import type {
  CreateDocumentInput,
  DocumentKind,
  RichDocument,
  RichDocumentSummary,
  UpdateDocumentInput,
} from '../model/types';

/**
 * リッチ文書 API（/api/v2/documents）の薄いラッパ。
 * フロント型は backend domain.RichDocument と 1:1 なのでマッパーは不要（DOP 整合性方針）。
 * 認可（owner スコープ）・doc 検証・楽観ロックは backend が担う。
 */
const DocumentRepository = {
  /** 自分の文書一覧（doc 本体なしの軽量サマリ・更新日降順）。kind で絞り込み可。 */
  async fetchDocuments(kind?: DocumentKind): Promise<RichDocumentSummary[]> {
    const res = await apiClient.get<RichDocumentSummary[]>(DOCUMENTS.list, {
      params: kind ? { kind } : undefined,
    });
    return toArray<RichDocumentSummary>(res.data);
  },

  /** 1 件を doc 本体込みで取得する（所有者 or 公開のみ・それ以外は 404）。 */
  async fetchDocument(id: string): Promise<RichDocument> {
    const res = await apiClient.get<RichDocument>(DOCUMENTS.byId(id));
    return res.data;
  },

  /** 新規作成（current user 名義・201）。 */
  async createDocument(input: CreateDocumentInput): Promise<RichDocument> {
    const res = await apiClient.post<RichDocument>(DOCUMENTS.list, input);
    return res.data;
  },

  /** 更新（所有者のみ・楽観ロック）。revision 不一致は 409。 */
  async updateDocument(id: string, input: UpdateDocumentInput): Promise<RichDocument> {
    const res = await apiClient.put<RichDocument>(DOCUMENTS.byId(id), input);
    return res.data;
  },

  /** 論理削除（所有者のみ・204）。 */
  async deleteDocument(id: string): Promise<void> {
    await apiClient.delete(DOCUMENTS.byId(id));
  },
};

export default DocumentRepository;
