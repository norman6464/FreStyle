import { useMemo, useState } from 'react';
import { PencilIcon, EyeIcon } from '@heroicons/react/24/outline';
import { getNoteStats } from '../utils/noteStats';
import MarkdownRenderer from './MarkdownRenderer';

interface NoteEditorProps {
  title: string;
  content: string;
  onTitleChange: (title: string) => void;
  onContentChange: (content: string) => void;
}

export default function NoteEditor({
  title,
  content,
  onTitleChange,
  onContentChange,
}: NoteEditorProps) {
  const stats = useMemo(() => getNoteStats(content), [content]);
  const [isPreview, setIsPreview] = useState(false);

  return (
    <div className="flex flex-col h-full p-6 max-w-3xl mx-auto w-full">
      <div className="flex items-center justify-between mb-4">
        <input
          type="text"
          value={title}
          onChange={(e) => onTitleChange(e.target.value)}
          placeholder="無題"
          aria-label="ノートのタイトル"
          className="text-xl font-bold text-[var(--color-text-primary)] bg-transparent border-none outline-none flex-1 placeholder:text-[var(--color-text-faint)]"
        />
        <button
          onClick={() => setIsPreview(!isPreview)}
          className="ml-3 p-1.5 rounded-lg hover:bg-surface-2 transition-colors text-[var(--color-text-muted)]"
          aria-label={isPreview ? '編集' : 'プレビュー'}
        >
          {isPreview ? <PencilIcon className="w-4 h-4" /> : <EyeIcon className="w-4 h-4" />}
        </button>
      </div>

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

      <div className="flex items-center gap-3 pt-3 border-t border-surface-3 text-[11px] text-[var(--color-text-faint)]" aria-label="ノート統計">
        <span>{stats.charCount}文字</span>
        {stats.readingTimeMin > 0 && <span>約{stats.readingTimeMin}分</span>}
      </div>
    </div>
  );
}
