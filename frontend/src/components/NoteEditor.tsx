import { useMemo } from 'react';
import { getNoteStats } from '../utils/noteStats';
import BlockEditor from './BlockEditor';

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

      <BlockEditor content={content} onChange={onContentChange} />

      <div className="flex items-center gap-3 pt-3 border-t border-surface-3 text-[11px] text-[var(--color-text-faint)]" aria-label="ノート統計">
        <span>{stats.charCount}文字</span>
        {stats.readingTimeMin > 0 && <span>約{stats.readingTimeMin}分</span>}
      </div>
    </div>
  );
}
