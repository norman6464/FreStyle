import { useEffect, useMemo, useRef } from 'react';
import type { SlashCommand } from '../constants/slashCommands';

interface SlashCommandMenuProps {
  items: SlashCommand[];
  selectedIndex: number;
  onSelect: (index: number) => void;
}

export default function SlashCommandMenu({ items, selectedIndex, onSelect }: SlashCommandMenuProps) {
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const selected = listRef.current?.querySelector<HTMLElement>('[aria-current="true"]');
    if (selected && typeof selected.scrollIntoView === 'function') {
      selected.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex]);

  const categories = useMemo(() => {
    const map = new Map<string, { items: SlashCommand[]; startIndex: number }>();
    let idx = 0;
    for (const item of items) {
      const cat = item.category || 'その他';
      if (!map.has(cat)) {
        map.set(cat, { items: [], startIndex: idx });
      }
      map.get(cat)!.items.push(item);
      idx++;
    }
    return map;
  }, [items]);

  const activeId = items[selectedIndex] ? `slash-cmd-${selectedIndex}` : undefined;

  return (
    <div
      ref={listRef}
      role="menu"
      aria-activedescendant={activeId}
      aria-label="スラッシュコマンド"
      className="slash-command-menu bg-[var(--color-surface-1)] border border-[var(--color-surface-3)] rounded-xl shadow-2xl overflow-y-auto max-h-80 min-w-[280px]"
    >
      {[...categories.entries()].map(([category, { items: catItems, startIndex }]) => (
        <div key={category}>
          <div className="px-3 pt-3 pb-1">
            <span className="text-[11px] font-semibold text-[var(--color-text-muted)] uppercase tracking-wider">
              {category}
            </span>
          </div>
          {catItems.map((item, i) => {
            const globalIndex = startIndex + i;
            return (
              <button
                key={`${item.action}-${globalIndex}`}
                id={`slash-cmd-${globalIndex}`}
                role="menuitem"
                aria-current={globalIndex === selectedIndex}
                className={`w-full flex items-center gap-3 px-3 py-1.5 text-left transition-colors ${
                  globalIndex === selectedIndex
                    ? 'bg-[var(--color-surface-3)] text-[var(--color-text-primary)]'
                    : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-2)]'
                }`}
                onClick={() => onSelect(globalIndex)}
                type="button"
              >
                <span className="w-7 h-7 flex items-center justify-center rounded bg-[var(--color-surface-2)] border border-[var(--color-surface-3)] text-xs font-semibold shrink-0">
                  {item.icon}
                </span>
                <span className="flex-1 text-sm font-medium truncate">{item.label}</span>
                {item.shortcut && (
                  <span className="text-xs text-[var(--color-text-faint)] ml-auto shrink-0">
                    {item.shortcut}
                  </span>
                )}
              </button>
            );
          })}
        </div>
      ))}
      <div className="flex items-center justify-between px-3 py-2 border-t border-[var(--color-surface-3)] mt-1">
        <span className="text-[11px] text-[var(--color-text-faint)]">メニューを閉じる</span>
        <kbd className="text-[11px] text-[var(--color-text-faint)] bg-[var(--color-surface-2)] px-1.5 py-0.5 rounded border border-[var(--color-surface-3)]">
          esc
        </kbd>
      </div>
    </div>
  );
}
