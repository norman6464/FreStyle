interface FilterTabsProps<T extends string> {
  tabs: T[];
  selected: T;
  onSelect: (tab: T) => void;
  className?: string;
}

export default function FilterTabs<T extends string>({ tabs, selected, onSelect, className = '' }: FilterTabsProps<T>) {
  return (
    <div role="tablist" className={`flex gap-1 border-b border-surface-3 ${className}`}>
      {tabs.map((tab) => (
        <button
          key={tab}
          role="tab"
          aria-selected={selected === tab}
          onClick={() => onSelect(tab)}
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
