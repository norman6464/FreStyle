import { useEffect, useRef } from 'react';
import type { SlashCommand } from '../constants/slashCommands';

interface SlashCommandMenuProps {
  items: SlashCommand[];
  selectedIndex: number;
  onSelect: (index: number) => void;
}

export default function SlashCommandMenu({ items, selectedIndex, onSelect }: SlashCommandMenuProps) {
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const selected = listRef.current?.querySelector<HTMLElement>('[aria-selected="true"]');
    if (selected && typeof selected.scrollIntoView === 'function') {
      selected.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex]);

  return (
    <div
      ref={listRef}
      role="listbox"
      className="slash-command-menu bg-[var(--color-surface-1)] border border-[var(--color-border)] rounded-lg shadow-lg overflow-y-auto max-h-64 py-1 min-w-[220px]"
    >
      {items.map((item, index) => (
        <button
          key={item.action}
          role="option"
          aria-selected={index === selectedIndex}
          className={`w-full flex items-center gap-3 px-3 py-2 text-left transition-colors ${
            index === selectedIndex
              ? 'bg-[var(--color-surface-3)] text-[var(--color-text-primary)]'
              : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-2)]'
          }`}
          onClick={() => onSelect(index)}
          type="button"
        >
          <span className="w-8 h-8 flex items-center justify-center rounded bg-[var(--color-surface-2)] text-sm font-semibold shrink-0">
            {item.icon}
          </span>
          <div className="flex flex-col min-w-0">
            <span className="text-sm font-medium">{item.label}</span>
            <span className="text-xs text-[var(--color-text-faint)] truncate">{item.description}</span>
          </div>
        </button>
      ))}
    </div>
  );
}
