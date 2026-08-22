import type { JSONContent } from '@tiptap/react';

/**
 * RichDocContent は tiptap（ProseMirror）のドキュメント JSON。
 * バックエンド（rich_documents.doc）に保存する正本の形と同じで、ルートは常に type='doc'。
 */
export type RichDocContent = JSONContent;

/**
 * emptyRichDoc は空の tiptap ドキュメント（段落 1 つ）を返す。
 * 新規作成時の初期値に使う。バックエンドの doc CHECK（object かつ type='doc'）を満たす。
 */
export function emptyRichDoc(): RichDocContent {
  return { type: 'doc', content: [{ type: 'paragraph' }] };
}

/**
 * isRichDoc は値が tiptap のドキュメント JSON（object かつ type='doc'）かを判定する。
 * API から受け取った未知の値を描画前にゆるく検証するためのガード。
 */
export function isRichDoc(value: unknown): value is RichDocContent {
  return (
    typeof value === 'object' &&
    value !== null &&
    (value as { type?: unknown }).type === 'doc'
  );
}
