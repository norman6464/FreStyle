import type { RichDocument, RichDocumentSummary } from '../model/types';

/**
 * toRichDocumentSummary は doc 本体込みの RichDocument から、一覧用の軽量サマリ
 * （doc を除いたフィールド）を取り出す。作成/更新のレスポンスで一覧を同期するときに使う。
 */
export function toRichDocumentSummary(document: RichDocument): RichDocumentSummary {
  const { doc: _doc, ...summary } = document;
  void _doc;
  return summary;
}
