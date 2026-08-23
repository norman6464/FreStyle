import type { RichDocContent } from '@/shared/ui/RichTextEditor';

/**
 * stripLeadingDocTitle は doc（tiptap JSON）の先頭が h1 のとき、それを除いた doc を返す。
 *
 * 章タイトルはカードヘッダーで material.title として大きく表示するため、本文先頭の h1 を
 * 残すとタイトルが二重に見える（Markdown 版の stripLeadingTitle と同じ目的。FRESTYLE-131）。
 * 先頭ノードが h1 でなければそのまま返す。
 */
export function stripLeadingDocTitle(doc: RichDocContent): RichDocContent {
  const [first, ...rest] = doc.content ?? [];
  if (first?.type === 'heading' && first.attrs?.level === 1) {
    // 本文が h1 だけの章でも tiptap が受理できるよう、空になったら段落 1 つを置く。
    return { ...doc, content: rest.length > 0 ? rest : [{ type: 'paragraph' }] };
  }
  return doc;
}
