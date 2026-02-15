import { KeyboardEvent, useCallback, useMemo, useRef, useState } from 'react';
import { PencilIcon, EyeIcon } from '@heroicons/react/24/outline';
import { getNoteStats } from '../utils/noteStats';
import MarkdownRenderer from './MarkdownRenderer';

interface NoteEditorProps {
  title: string;
  content: string;
  onTitleChange: (title: string) => void;
  onContentChange: (content: string) => void;
}

const TAB_EDIT_ID = 'note-tab-edit';
const TAB_PREVIEW_ID = 'note-tab-preview';
const PANEL_ID = 'note-tabpanel';

export default function NoteEditor({
  title,
  content,
  onTitleChange,
  onContentChange,
}: NoteEditorProps) {
  const stats = useMemo(() => getNoteStats(content), [content]);
  const [isPreview, setIsPreview] = useState(false);
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const handleKeyDown = useCallback((e: KeyboardEvent, index: number) => {
    let nextIndex: number | null = null;
    switch (e.key) {
      case 'ArrowRight':
        nextIndex = (index + 1) % 2;
        break;
      case 'ArrowLeft':
        nextIndex = (index - 1 + 2) % 2;
        break;
      case 'Home':
        nextIndex = 0;
        break;
      case 'End':
        nextIndex = 1;
        break;
    }
    if (nextIndex !== null) {
      e.preventDefault();
      tabRefs.current[nextIndex]?.focus();
      setIsPreview(nextIndex === 1);
    }
  }, []);

  const activeTabId = isPreview ? TAB_PREVIEW_ID : TAB_EDIT_ID;

  return (
    <div className="flex flex-col h-full p-6 max-w-3xl mx-auto w-full">
      <input
        type="text"
        value={title}
        onChange={(e) => onTitleChange(e.target.value)}
        placeholder="無題"
        aria-label="ノートのタイトル"
        className="text-xl font-bold text-[var(--color-text-primary)] bg-transparent border-none outline-none w-full mb-4 placeholder:text-[var(--color-text-faint)]"
      />

      <div className="flex border-b border-surface-3 mb-4" role="tablist">
        <button
          type="button"
          role="tab"
          id={TAB_EDIT_ID}
          ref={(el) => { tabRefs.current[0] = el; }}
          aria-selected={!isPreview}
          aria-controls={PANEL_ID}
          tabIndex={!isPreview ? 0 : -1}
          onClick={() => setIsPreview(false)}
          onKeyDown={(e) => handleKeyDown(e, 0)}
          className={`flex items-center gap-1.5 px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px ${
            !isPreview
              ? 'border-primary-400 text-primary-400'
              : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
          }`}
        >
          <PencilIcon className="w-4 h-4" />
          編集
        </button>
        <button
          type="button"
          role="tab"
          id={TAB_PREVIEW_ID}
          ref={(el) => { tabRefs.current[1] = el; }}
          aria-selected={isPreview}
          aria-controls={PANEL_ID}
          tabIndex={isPreview ? 0 : -1}
          onClick={() => setIsPreview(true)}
          onKeyDown={(e) => handleKeyDown(e, 1)}
          className={`flex items-center gap-1.5 px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px ${
            isPreview
              ? 'border-primary-400 text-primary-400'
              : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
          }`}
        >
          <EyeIcon className="w-4 h-4" />
          プレビュー
        </button>
      </div>

      <div role="tabpanel" id={PANEL_ID} aria-labelledby={activeTabId} className="flex-1 flex flex-col overflow-hidden">
        {isPreview ? (
          <div className="flex-1 overflow-y-auto">
            <MarkdownRenderer content={content} />
          </div>
        ) : (
          <textarea
            value={content}
            onChange={(e) => onContentChange(e.target.value)}
            placeholder="ここに入力..."
            aria-label="ノートの内容"
            className="flex-1 text-sm text-[var(--color-text-primary)] bg-transparent border-none outline-none w-full resize-none leading-relaxed placeholder:text-[var(--color-text-faint)]"
          />
        )}
      </div>

      <div className="flex items-center gap-3 pt-3 border-t border-surface-3 text-[11px] text-[var(--color-text-faint)]" aria-label="ノート統計">
        <span>{stats.charCount}文字</span>
        {stats.readingTimeMin > 0 && <span>約{stats.readingTimeMin}分</span>}
      </div>
    </div>
  );
}
