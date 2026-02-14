import { useState, useEffect } from 'react';
import { useSessionNote } from '../hooks/useSessionNote';

interface SessionNoteEditorProps {
  sessionId: number;
}

export default function SessionNoteEditor({ sessionId }: SessionNoteEditorProps) {
  const { note, saveNote } = useSessionNote(sessionId);
  const [text, setText] = useState(note);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    setText(note);
  }, [note]);

  const handleSave = () => {
    saveNote(text);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-2">振り返りメモ</p>
      <textarea
        value={text}
        onChange={(e) => setText(e.target.value)}
        placeholder="振り返りメモを入力..."
        className="w-full text-sm text-[var(--color-text-secondary)] border border-surface-3 rounded-lg p-2 resize-none h-20 focus:outline-none focus:ring-1 focus:ring-primary-400"
      />
      <div className="flex items-center justify-end gap-2 mt-2">
        {saved && <span className="text-xs text-emerald-400">保存しました</span>}
        <button
          onClick={handleSave}
          className="text-xs font-medium text-white bg-primary-500 px-3 py-1.5 rounded-lg hover:bg-primary-600 transition-colors"
        >
          保存
        </button>
      </div>
    </div>
  );
}
