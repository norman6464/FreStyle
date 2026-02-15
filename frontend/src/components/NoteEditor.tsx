import { useMemo, useCallback } from 'react';
import { getNoteStats } from '../utils/noteStats';
import { useTableOfContents } from '../hooks/useTableOfContents';
import BlockEditor from './BlockEditor';
import TableOfContents from './TableOfContents';
import WordCount from './WordCount';
import ReadingTime from './ReadingTime';
import LineCount from './LineCount';

interface NoteEditorProps {
  title: string;
  content: string;
  noteId: string | null;
  onTitleChange: (title: string) => void;
  onContentChange: (content: string) => void;
}

export default function NoteEditor({
  title,
  content,
  noteId,
  onTitleChange,
  onContentChange,
}: NoteEditorProps) {
  const stats = useMemo(() => getNoteStats(content), [content]);
  const { headings, isOpen, toggle } = useTableOfContents(content);

  const handleHeadingClick = useCallback((id: string) => {
    const index = parseInt(id.replace('heading-', ''), 10);
    const editorEl = document.querySelector('.ProseMirror');
    if (!editorEl) return;

    const headingEls = editorEl.querySelectorAll('h1, h2, h3');
    const target = headingEls[index];
    if (target) {
      target.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, []);

  return (
    <div className="flex flex-col h-full p-6 max-w-3xl mx-auto w-full">
      <input
        type="text"
        value={title}
        onChange={(e) => onTitleChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault();
            const editor = document.querySelector('.ProseMirror') as HTMLElement;
            editor?.focus();
          }
        }}
        placeholder="無題"
        aria-label="ノートのタイトル"
        className="text-3xl font-bold text-[var(--color-text-primary)] bg-transparent border-none outline-none w-full mb-4 placeholder:text-[var(--color-text-faint)]"
      />

      {headings.length > 0 && (
        <div className="mb-2">
          <button
            onClick={toggle}
            className="text-xs text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)] transition-colors"
            aria-label="目次の表示切替"
          >
            {isOpen ? '▼ 目次を閉じる' : '▶ 目次を表示'}
          </button>
          {isOpen && (
            <TableOfContents headings={headings} onHeadingClick={handleHeadingClick} />
          )}
        </div>
      )}

      <BlockEditor content={content} onChange={onContentChange} noteId={noteId} />

      <div className="flex items-center gap-3 pt-3 border-t border-surface-3" aria-label="ノート統計">
        <WordCount charCount={stats.charCount} />
        <LineCount lineCount={stats.lineCount} />
        <ReadingTime charCount={stats.charCount} />
      </div>
    </div>
  );
}
