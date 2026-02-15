import { useEffect, useState, useCallback, useRef } from 'react';
import type { Editor } from '@tiptap/core';

interface SelectionToolbarProps {
  editor: Editor | null;
  containerRef: React.RefObject<HTMLDivElement | null>;
}

const COLORS = [
  { color: '#ef4444', label: '赤' },
  { color: '#f97316', label: 'オレンジ' },
  { color: '#eab308', label: '黄' },
  { color: '#22c55e', label: '緑' },
  { color: '#3b82f6', label: '青' },
  { color: '#a855f7', label: '紫' },
];

export default function SelectionToolbar({ editor, containerRef }: SelectionToolbarProps) {
  const [visible, setVisible] = useState(false);
  const [position, setPosition] = useState({ top: 0, left: 0 });
  const [colorOpen, setColorOpen] = useState(false);
  const [headingOpen, setHeadingOpen] = useState(false);
  const toolbarRef = useRef<HTMLDivElement>(null);

  const updatePosition = useCallback(() => {
    if (!editor || !containerRef.current) return;

    const { from, to } = editor.state.selection;
    if (from === to) {
      setVisible(false);
      return;
    }

    const selection = window.getSelection();
    if (!selection || selection.rangeCount === 0) {
      setVisible(false);
      return;
    }

    const range = selection.getRangeAt(0);
    const rect = range.getBoundingClientRect();
    const containerRect = containerRef.current.getBoundingClientRect();

    if (rect.width === 0) {
      setVisible(false);
      return;
    }

    setPosition({
      top: rect.top - containerRect.top - 48,
      left: rect.left - containerRect.left + rect.width / 2,
    });
    setVisible(true);
    setColorOpen(false);
    setHeadingOpen(false);
  }, [editor, containerRef]);

  useEffect(() => {
    if (!editor) return;

    const onSelectionUpdate = () => {
      requestAnimationFrame(updatePosition);
    };

    const onBlur = () => {
      setTimeout(() => {
        if (!toolbarRef.current?.contains(document.activeElement)) {
          setVisible(false);
        }
      }, 150);
    };

    editor.on('selectionUpdate', onSelectionUpdate);
    editor.on('blur', onBlur);

    return () => {
      editor.off('selectionUpdate', onSelectionUpdate);
      editor.off('blur', onBlur);
    };
  }, [editor, updatePosition]);

  if (!visible || !editor) return null;

  const currentHeading = editor.isActive('heading', { level: 1 }) ? '見出し1'
    : editor.isActive('heading', { level: 2 }) ? '見出し2'
    : editor.isActive('heading', { level: 3 }) ? '見出し3'
    : 'テキスト';

  const setHeading = (level: number) => {
    if (level === 0) {
      editor.chain().focus().setParagraph().run();
    } else {
      editor.chain().focus().toggleHeading({ level: level as 1 | 2 | 3 }).run();
    }
    setHeadingOpen(false);
  };

  const toggleLink = () => {
    if (editor.isActive('link')) {
      editor.chain().focus().unsetLink().run();
    } else {
      const url = window.prompt('URLを入力してください');
      if (url) {
        editor.chain().focus().setLink({ href: url }).run();
      }
    }
  };

  return (
    <div
      ref={toolbarRef}
      className="absolute z-50 pointer-events-auto"
      style={{ top: `${position.top}px`, left: `${position.left}px`, transform: 'translateX(-50%)' }}
    >
      <div className="flex items-center bg-[var(--color-surface-1)] border border-[var(--color-border,var(--color-surface-3))] rounded-lg shadow-2xl px-1 py-0.5 gap-0.5">
        {/* Heading select */}
        <div className="relative">
          <button
            type="button"
            className={`text-xs px-2 py-1 rounded flex items-center gap-1 transition-colors ${
              headingOpen ? 'bg-[var(--color-surface-3)]' : 'hover:bg-[var(--color-surface-2)]'
            } text-[var(--color-text-secondary)]`}
            onMouseDown={(e) => { e.preventDefault(); setHeadingOpen(!headingOpen); setColorOpen(false); }}
          >
            {currentHeading}
            <svg width="10" height="10" viewBox="0 0 10 10" className="opacity-60"><path d="M3 4l2 2 2-2" stroke="currentColor" strokeWidth="1.5" fill="none" /></svg>
          </button>
          {headingOpen && (
            <div className="absolute top-full left-0 mt-1 bg-[var(--color-surface-1)] border border-[var(--color-surface-3)] rounded-lg shadow-xl py-1 min-w-[100px]">
              {[{ label: 'テキスト', level: 0 }, { label: '見出し1', level: 1 }, { label: '見出し2', level: 2 }, { label: '見出し3', level: 3 }].map(({ label, level }) => (
                <button key={level} type="button" className="w-full text-left text-xs px-3 py-1.5 hover:bg-[var(--color-surface-2)] text-[var(--color-text-secondary)]" onMouseDown={(e) => { e.preventDefault(); setHeading(level); }}>
                  {label}
                </button>
              ))}
            </div>
          )}
        </div>

        <Divider />

        {/* Format buttons */}
        <FmtButton label="太字" active={editor.isActive('bold')} onMouseDown={() => editor.chain().focus().toggleBold().run()}>
          <strong>B</strong>
        </FmtButton>
        <FmtButton label="斜体" active={editor.isActive('italic')} onMouseDown={() => editor.chain().focus().toggleItalic().run()}>
          <em>I</em>
        </FmtButton>
        <FmtButton label="下線" active={editor.isActive('underline')} onMouseDown={() => editor.chain().focus().toggleUnderline().run()}>
          <span className="underline">U</span>
        </FmtButton>
        <FmtButton label="取り消し線" active={editor.isActive('strike')} onMouseDown={() => editor.chain().focus().toggleStrike().run()}>
          <span className="line-through">S</span>
        </FmtButton>

        <Divider />

        {/* Code */}
        <FmtButton label="インラインコード" active={editor.isActive('code')} onMouseDown={() => editor.chain().focus().toggleCode().run()}>
          <span className="text-[10px] font-mono">&lt;/&gt;</span>
        </FmtButton>

        {/* Link */}
        <FmtButton label="リンク" active={editor.isActive('link')} onMouseDown={toggleLink}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71" />
            <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71" />
          </svg>
        </FmtButton>

        <Divider />

        {/* Color */}
        <div className="relative">
          <button
            type="button"
            aria-label="文字色"
            className={`w-7 h-7 flex items-center justify-center rounded text-sm font-bold transition-colors ${
              colorOpen ? 'bg-[var(--color-surface-3)]' : 'hover:bg-[var(--color-surface-2)]'
            } text-[var(--color-text-secondary)]`}
            onMouseDown={(e) => { e.preventDefault(); setColorOpen(!colorOpen); setHeadingOpen(false); }}
          >
            A
            <svg width="8" height="8" viewBox="0 0 10 10" className="ml-0.5 opacity-60"><path d="M3 4l2 2 2-2" stroke="currentColor" strokeWidth="1.5" fill="none" /></svg>
          </button>
          {colorOpen && (
            <div className="absolute top-full right-0 mt-1 bg-[var(--color-surface-1)] border border-[var(--color-surface-3)] rounded-lg shadow-xl p-2">
              <div className="flex gap-1">
                {COLORS.map(({ color, label }) => (
                  <button
                    key={color}
                    type="button"
                    aria-label={label}
                    className="w-5 h-5 rounded-full border border-[var(--color-surface-3)] hover:scale-110 transition-transform"
                    style={{ backgroundColor: color }}
                    onMouseDown={(e) => { e.preventDefault(); editor.chain().focus().setColor(color).run(); setColorOpen(false); }}
                  />
                ))}
                <button
                  type="button"
                  aria-label="色をリセット"
                  className="w-5 h-5 rounded-full border border-[var(--color-surface-3)] hover:scale-110 transition-transform flex items-center justify-center bg-[var(--color-surface-1)]"
                  onMouseDown={(e) => { e.preventDefault(); editor.chain().focus().unsetColor().run(); setColorOpen(false); }}
                >
                  <svg width="8" height="8" viewBox="0 0 10 10" className="text-[var(--color-text-secondary)]">
                    <line x1="2" y1="2" x2="8" y2="8" stroke="currentColor" strokeWidth="1.5" />
                    <line x1="8" y1="2" x2="2" y2="8" stroke="currentColor" strokeWidth="1.5" />
                  </svg>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function Divider() {
  return <div className="w-px h-5 bg-[var(--color-surface-3)] mx-0.5" />;
}

interface FmtButtonProps {
  label: string;
  active: boolean;
  onMouseDown: () => void;
  children: React.ReactNode;
}

function FmtButton({ label, active, onMouseDown, children }: FmtButtonProps) {
  return (
    <button
      type="button"
      aria-label={label}
      className={`w-7 h-7 flex items-center justify-center rounded text-sm transition-colors ${
        active
          ? 'bg-[var(--color-surface-3)] text-[var(--color-text-primary)]'
          : 'hover:bg-[var(--color-surface-2)] text-[var(--color-text-secondary)]'
      }`}
      onMouseDown={(e) => { e.preventDefault(); onMouseDown(); }}
    >
      {children}
    </button>
  );
}
