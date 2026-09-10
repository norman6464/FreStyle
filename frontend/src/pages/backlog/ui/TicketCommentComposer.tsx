import { useEffect, useMemo, useState } from 'react';
import { EditorContent, useEditor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import { Placeholder } from '@tiptap/extensions';
import type { KbWorkspaceMember } from '@/entities/kb';
import type { TicketCommentSegment } from '@/entities/ticket';
import { CommentComposerEnter, Mention, setMentionMembers } from './mentionExtension';
import { editorContentToSegments, isEditorContentEmpty, segmentsToEditorContent } from '../lib/mentionComposerContent';

export interface TicketCommentComposerProps {
  /** 失敗は投げてくる前提（投げられたら入力を保つ）。 */
  onSubmit: (body: TicketCommentSegment[]) => Promise<void>;
  /** '@' の候補。ワークスペースに属する人（useWorkspaceMembers）。 */
  members: KbWorkspaceMember[];
  /** 発言の編集を開いたときの下書きの種。省略時は空欄から始める。 */
  initialSegments?: TicketCommentSegment[];
  /** initialSegments の mention に表示名を当てる（引けなければ「不明なユーザー」）。 */
  resolveMentionName?: (userId: string) => string | null;
  placeholder?: string;
  submitLabel?: string;
  autoFocus?: boolean;
}

/**
 * 発言の入力欄。'@' に続けて日本語で打つと、ワークスペースに属する人の候補が出る
 * （tiptap の Suggestion。shared/ui/RichTextEditor の '/' コマンドと同じ仕組み）。
 * 選ぶと名指しは 1 個の不可分な単位になり、Backspace で丸ごと消える。
 *
 * エディタのスキーマは本文エディタ（RichTextEditor）とは別（段落を持たない一列・marks 無し）
 * — 発言の本文は marks を一切保持しない方針（entities/ticket/lib/commentBody.ts）に、
 * 入力側のスキーマも最初から合わせてある。
 *
 * 空判定は「文字も名指しも無い」。backend は「配列が空」または「text ノードだけで trim 後が
 * 全部空」を本文全体ごと 400 で拒む境界を持つので、ここで先に止める。
 */
export default function TicketCommentComposer({
  onSubmit,
  members,
  initialSegments = [],
  resolveMentionName,
  placeholder = 'コメントを書く',
  submitLabel = '送信',
  autoFocus = false,
}: TicketCommentComposerProps) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // マウント時の下書きの種だけを見る（以降 initialSegments が変わっても打ち直さない —
  // 発言の編集はコンポーザごと開閉されるたびに新しく積むので、この eslint-disable で十分）。
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const initialContent = useMemo(() => segmentsToEditorContent(initialSegments, resolveMentionName ?? (() => null)), []);
  const [empty, setEmpty] = useState(() => isEditorContentEmpty(initialContent));

  const editor = useEditor({
    editable: !submitting,
    extensions: [
      StarterKit.configure({
        heading: false,
        code: false,
        codeBlock: false,
        link: false,
        blockquote: false,
        bulletList: false,
        orderedList: false,
        listItem: false,
        horizontalRule: false,
        bold: false,
        italic: false,
        strike: false,
        dropcursor: false,
        gapcursor: false,
      }),
      // members は初回描画時点の値で足りる。読み込みが遅れて後から届いた分は下の
      // useEffect が editor.storage.mention へ書き足す（拡張一覧は生成時に固定されるため）。
      Mention.configure({ members }),
      CommentComposerEnter,
      Placeholder.configure({ placeholder }),
    ],
    content: initialContent,
    autofocus: autoFocus,
    editorProps: {
      attributes: {
        class: 'text-sm text-[var(--color-text-primary)] focus:outline-none',
        role: 'textbox',
        'aria-multiline': 'true',
        'aria-label': placeholder,
      },
    },
    onUpdate: ({ editor: currentEditor }) => {
      setEmpty(isEditorContentEmpty(currentEditor.getJSON()));
    },
  });

  useEffect(() => {
    if (editor && !editor.isDestroyed) setMentionMembers(editor, members);
  }, [editor, members]);

  const canSubmit = !empty && !submitting;

  const handleSubmit = async () => {
    if (!editor || !canSubmit) return;
    const segments = editorContentToSegments(editor.getJSON());
    setSubmitting(true);
    setError(null);
    try {
      await onSubmit(segments);
      editor.commands.clearContent();
      setEmpty(true);
    } catch {
      setError('送信できませんでした。もう一度お試しください。');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="rounded-xl border border-surface-3 p-2">
      <EditorContent editor={editor} />
      {error && (
        <p role="alert" className="mt-1 text-xs leading-relaxed text-red-700">
          {error}
        </p>
      )}
      <div className="mt-1.5 flex items-center">
        <span className="text-[11px] text-[var(--color-text-muted)]">@ で名前を挙げると通知が届きます</span>
        <button
          type="button"
          onClick={() => void handleSubmit()}
          disabled={!canSubmit}
          className="ml-auto rounded border-0 bg-brand-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-brand-700 disabled:opacity-50"
        >
          {submitting ? '送信中…' : submitLabel}
        </button>
      </div>
    </div>
  );
}
