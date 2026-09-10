import type { JSONContent } from '@tiptap/core';
import type { TicketCommentSegment } from '@/entities/ticket';

/**
 * 発言入力欄（tiptap）の中身と `TicketCommentSegment[]` の往復。
 *
 * 入力欄のスキーマは doc → paragraph 1 つ → [text | mention | hardBreak]* の一列
 * （mentionExtension.ts の CommentComposerEnter が Enter を hardBreak に固定するので、
 * 段落は増えない想定。貼り付けで複数段落が紛れ込んでも、段落の境目を '\n' として畳むので
 * 本文は壊れない）。hardBreak はここで '\n' の文字へ変換してしまうので、送信される wire
 * には hardBreak ノードが一切現れない（commentBody.ts の readCommentBody は関与しない）。
 */

/** 編集欄の中身を送信できる区間の列へ畳む。 */
export function editorContentToSegments(doc: JSONContent): TicketCommentSegment[] {
  const segments: TicketCommentSegment[] = [];

  const pushText = (text: string) => {
    if (text === '') return;
    const last = segments[segments.length - 1];
    if (last?.kind === 'text') {
      last.text += text;
    } else {
      segments.push({ kind: 'text', text });
    }
  };

  const walkInline = (node: JSONContent) => {
    if (node.type === 'text' && typeof node.text === 'string') {
      pushText(node.text);
      return;
    }
    if (node.type === 'hardBreak') {
      pushText('\n');
      return;
    }
    if (node.type === 'mention') {
      const userId = node.attrs?.userId;
      if (typeof userId === 'string' && userId !== '') {
        segments.push({ kind: 'mention', userId });
      }
    }
  };

  const blocks = doc.content ?? [];
  blocks.forEach((block, i) => {
    if (i > 0) pushText('\n');
    for (const child of block.content ?? []) walkInline(child);
  });

  return segments;
}

/**
 * 区間の列から編集欄の初期状態を組み立てる（発言の編集を開いたときの下書きの種）。
 * mention は表示名を持たない（wire には userId しかない）ので、呼び出し側が
 * 名前解決を渡す — 引けなければ「不明なユーザー」で埋める（TicketCommentBody と同じ扱い）。
 */
export function segmentsToEditorContent(
  segments: TicketCommentSegment[],
  resolveMentionName: (userId: string) => string | null,
): JSONContent {
  const content: JSONContent[] = [];
  for (const segment of segments) {
    if (segment.kind === 'mention') {
      content.push({
        type: 'mention',
        attrs: { userId: segment.userId, name: resolveMentionName(segment.userId) ?? '不明なユーザー' },
      });
      continue;
    }
    const lines = segment.text.split('\n');
    lines.forEach((line, i) => {
      if (i > 0) content.push({ type: 'hardBreak' });
      if (line !== '') content.push({ type: 'text', text: line });
    });
  }
  return { type: 'doc', content: [{ type: 'paragraph', content }] };
}

/** 編集欄が空かどうか（trim 後の text が全部空 かつ mention も無い）。 */
export function isEditorContentEmpty(doc: JSONContent): boolean {
  return editorContentToSegments(doc).every((s) => s.kind === 'text' && s.text.trim() === '');
}
