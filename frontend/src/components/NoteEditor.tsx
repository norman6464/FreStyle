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
      <textarea
        value={content}
        onChange={(e) => onContentChange(e.target.value)}
        placeholder="ここに入力..."
        aria-label="ノートの内容"
        className="flex-1 text-sm text-[var(--color-text-primary)] bg-transparent border-none outline-none w-full resize-none leading-relaxed placeholder:text-[var(--color-text-faint)]"
      />
    </div>
  );
}
