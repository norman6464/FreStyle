import { useEffect, useRef, useState } from 'react';
import { ArrowsUpDownIcon, CheckIcon, ChevronDownIcon } from '@heroicons/react/24/outline';
import { NOTE_SORT_OPTIONS, type NoteSortOption } from '../constants/sortOptions';

interface NoteSortMenuProps {
  selected: NoteSortOption;
  onChange: (value: NoteSortOption) => void;
}

/**
 * NoteSortMenu — ノート一覧の「並べ替え」をドロップダウンメニューで選択する。
 *
 * 旧: タブ風の SortSelector（4 つのボタンを横並び）
 * 新: 1 行の「並べ替え: <現在の選択> ▾」ボタンを押すとメニューが開き、
 *     チェックマーク付きの一覧から選ぶ。 Notion の並べ替えメニューに近い見た目。
 *
 * メニュー外クリックで閉じる。
 */
export default function NoteSortMenu({ selected, onChange }: NoteSortMenuProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!open) return;
    const onClick = (e: MouseEvent) => {
      if (!containerRef.current) return;
      if (!containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', onClick);
    return () => document.removeEventListener('mousedown', onClick);
  }, [open]);

  const currentLabel =
    NOTE_SORT_OPTIONS.find((o) => o.value === selected)?.label ?? NOTE_SORT_OPTIONS[0].label;

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={open}
        className="w-full flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-md text-xs text-[var(--color-text-secondary)] hover:bg-surface-2 transition-colors"
      >
        <span className="flex items-center gap-1.5">
          <ArrowsUpDownIcon className="w-3.5 h-3.5 text-[var(--color-text-muted)]" />
          並べ替え
        </span>
        <span className="flex items-center gap-1 text-[var(--color-text-muted)]">
          {currentLabel}
          <ChevronDownIcon className={`w-3 h-3 transition-transform ${open ? 'rotate-180' : ''}`} />
        </span>
      </button>
      {open && (
        <div
          role="menu"
          aria-label="並べ替えオプション"
          className="absolute z-20 left-0 right-0 mt-1 rounded-md border border-surface-3 bg-[var(--color-surface)] shadow-lg overflow-hidden"
        >
          {NOTE_SORT_OPTIONS.map(({ value, label }) => {
            const isSelected = selected === value;
            return (
              <button
                key={value}
                role="menuitemradio"
                aria-checked={isSelected}
                onClick={() => {
                  onChange(value);
                  setOpen(false);
                }}
                className={`w-full flex items-center justify-between gap-2 px-3 py-1.5 text-xs transition-colors ${
                  isSelected
                    ? 'text-[var(--color-text-primary)] bg-surface-2'
                    : 'text-[var(--color-text-secondary)] hover:bg-surface-2'
                }`}
              >
                <span>{label}</span>
                {isSelected && <CheckIcon className="w-3.5 h-3.5 text-taupe-400" />}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
