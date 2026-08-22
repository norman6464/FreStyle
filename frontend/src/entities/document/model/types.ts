import type { RichDocContent } from '@/shared/ui/RichTextEditor';

/**
 * リッチ文書（rich_documents）entity のドメイン型。
 * backend の `/api/v2/documents`（domain.RichDocument）と 1:1 で対応する。
 * 本文（doc）は tiptap の JSON（RichDocContent）で、正本として保存する。
 */

/** 文書の用途区分。backend の domain.DocumentKind と一致。 */
export type DocumentKind = 'note' | 'course-chapter';

/**
 * RichDocumentSummary は一覧の 1 件（doc 本体を含まない軽量サマリ）。
 * `GET /api/v2/documents` が返す形と 1:1。
 */
export interface RichDocumentSummary {
  id: string;
  ownerId: number;
  companyId?: number;
  kind: DocumentKind;
  title: string;
  isPublic: boolean;
  schemaVersion: number;
  revision: number;
  createdAt: string;
  updatedAt: string;
}

/** RichDocument はサマリに doc 本体（tiptap JSON）を加えた 1 件分。個別取得・保存で使う。 */
export interface RichDocument extends RichDocumentSummary {
  doc: RichDocContent;
}

/** 作成入力。ownerId は current user 固定なので送らない（backend が付与）。 */
export interface CreateDocumentInput {
  kind: DocumentKind;
  title: string;
  doc: RichDocContent;
  isPublic?: boolean;
  schemaVersion?: number;
}

/** 更新入力。revision は楽観ロックのため必須（不一致は 409）。 */
export interface UpdateDocumentInput {
  title: string;
  doc: RichDocContent;
  revision: number;
  isPublic?: boolean;
  schemaVersion?: number;
}
