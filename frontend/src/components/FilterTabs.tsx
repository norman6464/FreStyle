import { KeyboardEvent, useCallback, useRef } from 'react';

interface FilterTabsProps<T extends string> {
  tabs: T[];
  selected: T;
  onSelect: (tab: T) => void;
  className?: string;
}

export default function FilterTabs<T extends string>({ tabs, selected, onSelect, className = '' }: FilterTabsProps<T>) {
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const handleKeyDown = useCallback((e: KeyboardEvent, index: number) => {
    let nextIndex: number | null = null;
    switch (e.key) {
      case 'ArrowRight':
        nextIndex = (index + 1) % tabs.length;
        break;
      case 'ArrowLeft':
        nextIndex = (index - 1 + tabs.length) % tabs.length;
        break;
      case 'Home':
        nextIndex = 0;
        break;
      case 'End':
        nextIndex = tabs.length - 1;
        break;
    }
    if (nextIndex !== null) {
      e.preventDefault();
      tabRefs.current[nextIndex]?.focus();
      onSelect(tabs[nextIndex]);
    }
  }, [tabs, onSelect]);

  return (
    <div role="tablist" className={`flex gap-1 border-b border-surface-3 ${className}`}>
      {tabs.map((tab, index) => (
        <button
          key={tab}
          ref={(el) => { tabRefs.current[index] = el; }}
          role="tab"
          aria-selected={selected === tab}
          tabIndex={selected === tab ? 0 : -1}
          onClick={() => onSelect(tab)}
          onKeyDown={(e) => handleKeyDown(e, index)}
          className={`px-3 py-2 text-xs font-medium transition-colors border-b-2 -mb-px ${
            selected === tab
              ? 'border-primary-500 text-primary-400'
              : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
          }`}
        >
          {tab}
        </button>
      ))}
    </div>
  );
}
