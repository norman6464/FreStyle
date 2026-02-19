import { useMemo, useCallback, useRef } from 'react';
import { ArrowDownTrayIcon, ClipboardDocumentIcon } from '@heroicons/react/24/outline';
import { getNoteStats } from '../utils/noteStats';
import { useTableOfContents } from '../hooks/useTableOfContents';
import { useMarkdownExport } from '../hooks/useMarkdownExport';
import { useToast } from '../hooks/useToast';
import type { SaveStatus } from '../hooks/useNoteEditor';
import BlockEditor from './BlockEditor';
import type { BlockEditorHandle } from './BlockEditor';
import TableOfContents from './TableOfContents';
import WordCount from './WordCount';
import ReadingTime from './ReadingTime';
import LineCount from './LineCount';

interface NoteEditorProps {
  title: string;
  content: string;
  noteId: string | null;
  saveStatus: SaveStatus;
  onTitleChange: (title: string) => void;
  onContentChange: (content: string) => void;
}

const SAVE_STATUS_CONFIG: Record<Exclude<SaveStatus, 'idle'>, { label: string; color: string }> = {
  unsaved: { label: '未保存', color: 'text-amber-500' },
  saving: { label: '保存中...', color: 'text-[var(--color-text-muted)]' },
  saved: { label: '保存済み', color: 'text-emerald-500' },
};

export default function NoteEditor({
  title,
  content,
  noteId,
  saveStatus,
  onTitleChange,
  onContentChange,
}: NoteEditorProps) {
  const titleRef = useRef<HTMLInputElement>(null);
  const editorRef = useRef<BlockEditorHandle>(null);

  const stats = useMemo(() => getNoteStats(content), [content]);
  const { headings, isOpen, toggle } = useTableOfContents(content);
  const { exportAsMarkdown, copyAsMarkdown } = useMarkdownExport();
  const { showToast } = useToast();

  const handleExport = useCallback(() => {
    exportAsMarkdown(title, content);
  }, [exportAsMarkdown, title, content]);

  const handleCopyMarkdown = useCallback(async () => {
    const md = copyAsMarkdown(title, content);
    try {
      await navigator.clipboard.writeText(md);
      showToast('success', 'Markdownをコピーしました');
    } catch {
      showToast('error', 'コピーに失敗しました');
    }
  }, [copyAsMarkdown, title, content, showToast]);

  const handleTitleFocus = useCallback(() => {
    const proseMirror = document.querySelector('.ProseMirror') as HTMLElement;
    proseMirror?.blur();
  }, []);

  const handleTitleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      editorRef.current?.focusAtStart();
    }
  }, []);

  const handleBackspaceAtStart = useCallback(() => {
    titleRef.current?.focus();
    if (titleRef.current) {
      const len = titleRef.current.value.length;
      titleRef.current.setSelectionRange(len, len);
    }
  }, []);

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
        ref={titleRef}
        type="text"
        value={title}
        onChange={(e) => onTitleChange(e.target.value)}
        onFocus={handleTitleFocus}
        onKeyDown={handleTitleKeyDown}
        placeholder="無題"
        aria-label="ノートのタイトル"
        className="text-3xl font-bold text-[var(--color-text-primary)] bg-transparent border-none outline-none w-full mb-4 pl-10 placeholder:text-[var(--color-text-faint)]"
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

      <BlockEditor ref={editorRef} content={content} onChange={onContentChange} noteId={noteId} onBackspaceAtStart={handleBackspaceAtStart} />

      <div className="flex items-center gap-3 pt-3 border-t border-surface-3" aria-label="ノート統計">
        <WordCount charCount={stats.charCount} />
        <LineCount lineCount={stats.lineCount} />
        <ReadingTime charCount={stats.charCount} />
        {saveStatus !== 'idle' && (
          <span className={`text-xs ${SAVE_STATUS_CONFIG[saveStatus].color}`} aria-label="保存状態">
            {SAVE_STATUS_CONFIG[saveStatus].label}
          </span>
        )}
        <div className="ml-auto flex items-center gap-1">
          <button
            type="button"
            onClick={handleCopyMarkdown}
            aria-label="Markdownをコピー"
            className="p-1.5 rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
            title="Markdownをコピー"
          >
            <ClipboardDocumentIcon className="w-4 h-4" />
          </button>
          <button
            type="button"
            onClick={handleExport}
            aria-label="Markdownでダウンロード"
            className="p-1.5 rounded hover:bg-[var(--color-surface-3)] text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
            title="Markdownでダウンロード"
          >
            <ArrowDownTrayIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
